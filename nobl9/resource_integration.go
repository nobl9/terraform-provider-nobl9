package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

type integrationProvider interface {
	GetSchema() map[string]*schema.Schema
	MarshalSpec(data *schema.ResourceData) n9api.IntegrationSpec
	UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics
}

func resourceIntegrationFactory(provider integrationProvider) *schema.Resource {
	i := integration{integrationProvider: provider}
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),
		},
		CreateContext: i.resourceIntegrationApply,
		UpdateContext: i.resourceIntegrationApply,
		DeleteContext: resourceIntegrationDelete,
		ReadContext:   i.resourceIntegrationRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Integration configuration documentation]()", // TODO get description
	}

	for k, v := range provider.GetSchema() {
		resource.Schema[k] = v
	}

	return resource
}

type integration struct {
	integrationProvider
}

func (i integration) marshalIntegration(d *schema.ResourceData) *n9api.Integration {
	return &n9api.Integration{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindIntegration,
			MetadataHolder: marshalMetadata(d),
		},
		Spec: i.MarshalSpec(d),
	}
}

func (i integration) unmarshalIntegration(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	spec := object["spec"].(map[string]interface{})
	err := d.Set("description", spec["description"])
	diags = appendError(diags, err)

	errs := i.UnmarshalSpec(d, spec)
	diags = append(diags, errs...)

	return diags
}

func (i integration) resourceIntegrationApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	service := i.marshalIntegration(d)

	var p n9api.Payload
	p.AddObject(service)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return i.resourceIntegrationRead(ctx, d, meta)
}

func (i integration) resourceIntegrationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return i.unmarshalIntegration(d, objects)
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

type integrationWebhook struct{}

func (i integrationWebhook) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
		"template": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "",
			ConflictsWith: []string{"template_fields"},
		},
		"template_fields": {
			Type:          schema.TypeList,
			Optional:      true,
			Description:   "",
			ConflictsWith: []string{"template"},
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func (i integrationWebhook) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	fields := d.Get("template_fields").([]interface{})
	templateFields := make([]string, len(fields))
	for i, field := range fields {
		templateFields[i] = field.(string)
	}
	var template *string
	if t := d.Get("template").(string); t != "" {
		template = &t
		templateFields = nil
	}

	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		Webhook: &n9api.WebhookIntegration{
			URL:            d.Get("url").(string),
			Template:       template,
			TemplateFields: templateFields,
		},
	}
}

func (i integrationWebhook) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["webhook"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("template", config["template"])
	diags = appendError(diags, err)
	err = d.Set("template_fields", config["templateFields"])
	diags = appendError(diags, err)

	return diags
}

type integrationPagerDuty struct{}

func (i integrationPagerDuty) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"integration_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i integrationPagerDuty) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		PagerDuty: &n9api.PagerDutyIntegration{
			IntegrationKey: d.Get("integration_key").(string),
		},
	}
}

func (i integrationPagerDuty) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// pager duty has only one, secret field
	return nil
}

type integrationSlack struct{}

func (i integrationSlack) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i integrationSlack) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		Slack: &n9api.SlackIntegration{
			URL: d.Get("url").(string),
		},
	}
}

func (i integrationSlack) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// slack has only one, secret field
	return nil
}

type integrationDiscord struct{}

func (i integrationDiscord) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i integrationDiscord) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		Discord: &n9api.DiscordIntegration{
			URL: d.Get("url").(string),
		},
	}
}

func (i integrationDiscord) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// discord has only one, secret field
	return nil
}

type integrationOpsgenie struct{}

func (i integrationOpsgenie) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"auth": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
	}
}

func (i integrationOpsgenie) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		Opsgenie: &n9api.OpsgenieIntegration{
			Auth: d.Get("auth").(string),
			URL:  d.Get("url").(string),
		},
	}
}

func (i integrationOpsgenie) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["opsgenie"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("url", config["url"])
	diags = appendError(diags, err)

	return diags
}

type integrationServiceNow struct{}

func (i integrationServiceNow) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
		"instanceid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
	}
}

func (i integrationServiceNow) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		ServiceNow: &n9api.ServiceNowIntegration{
			Username:   d.Get("username").(string),
			Password:   d.Get("password").(string),
			InstanceID: d.Get("instanceid").(string),
		},
	}
}

func (i integrationServiceNow) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["servicenow"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("username", config["username"])
	diags = appendError(diags, err)
	err = d.Set("instanceid", config["instanceid"])
	diags = appendError(diags, err)

	return diags
}

type integrationJira struct{}

func (i integrationJira) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
		"apitoken": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
			Sensitive:   true,
			Computed:    true,
		},
		"projectid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "",
		},
	}
}

func (i integrationJira) MarshalSpec(d *schema.ResourceData) n9api.IntegrationSpec {
	return n9api.IntegrationSpec{
		Description: d.Get("description").(string),
		Jira: &n9api.JiraIntegration{
			URL:       d.Get("url").(string),
			Username:  d.Get("username").(string),
			APIToken:  d.Get("apitoken").(string),
			ProjectID: d.Get("projectid").(string),
		},
	}
}

func (i integrationJira) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["jira"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("username", config["username"])
	diags = appendError(diags, err)
	err = d.Set("url", config["url"])
	diags = appendError(diags, err)
	err = d.Set("projectid", config["projectId"])
	diags = appendError(diags, err)

	return diags
}
