package nobl9

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceIntegration() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),

			"integration_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of an integration. [Supported integrations]()",
			},

			"webhook_config": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation]()",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
							Sensitive:   true,
							Computed:    true,
						},
						"template": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
							//ConflictsWith: []string{"webhook_config.template_fields"},
						},
						"template_fields": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "",
							//ConflictsWith: []string{"webhook_config.template"},
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		CreateContext: resourceIntegrationApply,
		UpdateContext: resourceIntegrationApply,
		DeleteContext: resourceIntegrationDelete,
		ReadContext:   resourceIntegrationRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Integration configuration documentation]()",
	}
}

func marshalIntegration(d *schema.ResourceData) *n9api.Integration {
	return &n9api.Integration{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindIntegration,
			MetadataHolder: marshalMetadata(d),
		},
		Spec: n9api.IntegrationSpec{
			Description: d.Get("description").(string),
			Webhook:     marshalIntegrationWebhook(d),
			PagerDuty:   nil,
			Slack:       nil,
			Discord:     nil,
			Opsgenie:    nil,
			ServiceNow:  nil,
			Jira:        nil,
		},
	}
}

func marshalIntegrationWebhook(d *schema.ResourceData) *n9api.WebhookIntegration {
	agentType := d.Get("integration_type").(string)
	if agentType != "webhook" {
		return nil
	}
	p := d.Get("webhook_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	wh := p[0].(map[string]interface{})

	template := wh["template"].(string)
	fields := wh["template_fields"].([]interface{})
	templateFields := make([]string, len(fields))
	for i, field := range fields {
		templateFields[i] = field.(string)
	}
	if template != "" {
		templateFields = nil
	}

	return &n9api.WebhookIntegration{
		URL:            wh["url"].(string),
		Template:       &template,
		TemplateFields: templateFields,
	}
}

func unmarshalIntegration(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	supportedIntegration := []struct {
		hclName      string
		jsonName     string
		secretFields []string
	}{
		{"webhook_config", "webhook", []string{"url"}},
	}

	for _, integration := range supportedIntegration {
		ok, ds := unmarshalIntegrationConfig(
			d,
			object,
			integration.hclName,
			integration.jsonName,
			integration.secretFields,
		)
		if ds.HasError() {
			diags = append(diags, ds...)
		}
		if ok {
			break
		}
	}

	return diags
}

func unmarshalIntegrationConfig(
	d *schema.ResourceData,
	object n9api.AnyJSONObj,
	hclName,
	jsonName string,
	secretFields []string,
) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	spec := object["spec"].(map[string]interface{})
	if spec[jsonName] == nil {
		return false, nil
	}
	config := spec[jsonName].(map[string]interface{})

	for _, field := range secretFields {
		delete(config, field) // api returns secrets as '[hidden]'
	}

	o, n := d.GetChange("webhook_config")
	oldConfig := o.(*schema.Set)
	newConfig := n.(*schema.Set)

	var url interface{}
	var secretMap map[string]interface{}
	fmt.Println(oldConfig.List(), newConfig.List())
	if v, ok := d.GetOk("webhook_config"); ok {
		secretMap = (v.(*schema.Set)).List()[0].(map[string]interface{})
	} else {
		secretMap = newConfig.List()[0].(map[string]interface{})
	}
	url = secretMap["url"]

	err := d.Set("description", spec["description"])
	appendError(diags, err)
	err = d.Set(hclName, schema.NewSet(oneElementSet, []interface{}{
		map[string]interface{}{
			"template":        spec["template"],
			"template_fields": spec["template_fields"],
			"url":             url,
		}}))
	appendError(diags, err)

	//if !ok {
	//	appendError(diags, err)
	//	return true, diags
	//}

	//oo := o.(*schema.Set).List()[0].(map[string]interface{})

	return true, diags
}

func resourceIntegrationApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	service := marshalIntegration(d)

	var p n9api.Payload
	p.AddObject(service)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return resourceIntegrationRead(ctx, d, meta)
}

func resourceIntegrationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	client, ds := newClient(config, project)
	if ds.HasError() {
		return ds
	}

	objects, err := client.GetObject(n9api.ObjectIntegration, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalIntegration(d, objects)
}

func resourceIntegrationDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectIntegration, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
