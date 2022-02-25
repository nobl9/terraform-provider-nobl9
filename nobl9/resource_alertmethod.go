package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

type alertMethodProvider interface {
	GetSchema() map[string]*schema.Schema
	GetDescription() string
	MarshalSpec(data *schema.ResourceData) n9api.AlertMethodSpec
	UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics
}

func resourceAlertMethodFactory(provider alertMethodProvider) *schema.Resource {
	i := alertMethod{alertMethodProvider: provider}
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),
		},
		CreateContext: i.resourceAlertMethodApply,
		UpdateContext: i.resourceAlertMethodApply,
		DeleteContext: resourceAlertMethodDelete,
		ReadContext:   i.resourceAlertMethodRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: provider.GetDescription(),
	}

	for k, v := range provider.GetSchema() {
		resource.Schema[k] = v
	}

	return resource
}

type alertMethod struct {
	alertMethodProvider
}

func (a alertMethod) marshalAlertMethod(d *schema.ResourceData) *n9api.AlertMethod {
	return &n9api.AlertMethod{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindAlertMethod,
			MetadataHolder: marshalMetadata(d),
		},
		Spec: a.MarshalSpec(d),
	}
}

func (a alertMethod) unmarshalAlertMethod(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
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

	errs := a.UnmarshalSpec(d, spec)
	diags = append(diags, errs...)

	return diags
}

func (a alertMethod) resourceAlertMethodApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	service := a.marshalAlertMethod(d)

	var p n9api.Payload
	p.AddObject(service)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return a.resourceAlertMethodRead(ctx, d, meta)
}

func (a alertMethod) resourceAlertMethodRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	objects, err := client.GetObject(n9api.ObjectAlertMethod, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return a.unmarshalAlertMethod(d, objects)
}

func resourceAlertMethodDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectAlertMethod, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

type alertMethodWebhook struct{}

func (i alertMethodWebhook) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#webhook-alert-method)"
}

func (i alertMethodWebhook) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "URL of the webhook endpoint.",
			Sensitive:   true,
			Computed:    true,
		},
		"template": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Webhook message template. See documentation for template format and samples.",
			ConflictsWith: []string{"template_fields"},
		},
		"template_fields": {
			Type:          schema.TypeList,
			Optional:      true,
			Description:   "Webhook meesage fields. The message will contain json payload with specified fields. See documentation for allowed fields.",
			ConflictsWith: []string{"template"},
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func (i alertMethodWebhook) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
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

	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Webhook: &n9api.WebhookAlertMethod{
			URL:            d.Get("url").(string),
			Template:       template,
			TemplateFields: templateFields,
		},
	}
}

func (i alertMethodWebhook) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["webhook"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("template", config["template"])
	diags = appendError(diags, err)
	err = d.Set("template_fields", config["templateFields"])
	diags = appendError(diags, err)

	return diags
}

type alertMethodPagerDuty struct{}

func (i alertMethodPagerDuty) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#pagerduty-alert-method)"
}

func (i alertMethodPagerDuty) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"integration_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "PagerDuty Integration Key, found on Integrations tab.",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodPagerDuty) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		PagerDuty: &n9api.PagerDutyAlertMethod{
			IntegrationKey: d.Get("integration_key").(string),
		},
	}
}

func (i alertMethodPagerDuty) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// pager duty has only one, secret field
	return nil
}

type alertMethodSlack struct{}

func (i alertMethodSlack) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#slack-alert-method)"
}

func (i alertMethodSlack) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Slack webhook endpoint URL.",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodSlack) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Slack: &n9api.SlackAlertMethod{
			URL: d.Get("url").(string),
		},
	}
}

func (i alertMethodSlack) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// slack has only one, secret field
	return nil
}

type alertMethodDiscord struct{}

func (i alertMethodDiscord) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#discord-alert-method)"
}

func (i alertMethodDiscord) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Discord webhook endpoint URL.",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodDiscord) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Discord: &n9api.DiscordAlertMethod{
			URL: d.Get("url").(string),
		},
	}
}

func (i alertMethodDiscord) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// discord has only one, secret field
	return nil
}

type alertMethodOpsgenie struct{}

func (i alertMethodOpsgenie) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#opsgenie-alert-method)"
}

func (i alertMethodOpsgenie) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"auth": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Opsgenie authentication credentials. See documentation for supported formats.",
			Sensitive:   true,
			Computed:    true,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Opsgenie API URL.",
		},
	}
}

func (i alertMethodOpsgenie) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Opsgenie: &n9api.OpsgenieAlertMethod{
			Auth: d.Get("auth").(string),
			URL:  d.Get("url").(string),
		},
	}
}

func (i alertMethodOpsgenie) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["opsgenie"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("url", config["url"])
	diags = appendError(diags, err)

	return diags
}

type alertMethodServiceNow struct{}

func (i alertMethodServiceNow) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#servicenow-alert-method)"
}

func (i alertMethodServiceNow) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ServiceNow username.",
		},
		"password": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ServiceNow password.",
			Sensitive:   true,
			Computed:    true,
		},
		"instanceid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ServiceNow InstanceID. For details see documentation.",
		},
	}
}

func (i alertMethodServiceNow) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		ServiceNow: &n9api.ServiceNowAlertMethod{
			Username:   d.Get("username").(string),
			Password:   d.Get("password").(string),
			InstanceID: d.Get("instanceid").(string),
		},
	}
}

func (i alertMethodServiceNow) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["servicenow"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("username", config["username"])
	diags = appendError(diags, err)
	err = d.Set("instanceid", config["instanceid"])
	diags = appendError(diags, err)

	return diags
}

type alertMethodJira struct{}

func (i alertMethodJira) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#jira-alert-method)"
}

func (i alertMethodJira) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Jira instance URL.",
		},
		"username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Jira username for the owner of the API Token.",
		},
		"apitoken": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "API Token with access rights to the project.",
			Sensitive:   true,
			Computed:    true,
		},
		"project_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The code of the project.",
		},
	}
}

func (i alertMethodJira) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Jira: &n9api.JiraAlertMethod{
			URL:        d.Get("url").(string),
			Username:   d.Get("username").(string),
			APIToken:   d.Get("apitoken").(string),
			ProjectKey: d.Get("project_key").(string),
		},
	}
}

func (i alertMethodJira) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["jira"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("username", config["username"])
	diags = appendError(diags, err)
	err = d.Set("url", config["url"])
	diags = appendError(diags, err)
	err = d.Set("project_key", config["projectKey"])
	diags = appendError(diags, err)

	return diags
}

type alertMethodTeams struct{}

func (i alertMethodTeams) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#ms-teams-alert-method)"
}

func (i alertMethodTeams) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "MSTeams webhook endpoint URL.",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodTeams) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Teams: &n9api.TeamsAlertMethod{
			URL: d.Get("url").(string),
		},
	}
}

func (i alertMethodTeams) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	// teams has only one, secret field
	return nil
}

type alertMethodEmail struct{}

func (i alertMethodEmail) GetDescription() string {
	return "[Integration configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#alert-method)"
}

func (i alertMethodEmail) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"to": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Recipients.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"cc": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Carbon copy recipients.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"bcc": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Blind carbon copy recipients.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"subject": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Subject of the email.",
		},
		"body": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Body of the email. For format and samples see documentation and nobl9 application.",
		},
	}
}

func (i alertMethodEmail) MarshalSpec(d *schema.ResourceData) n9api.AlertMethodSpec {
	toStringSlice := func(in []interface{}) []string {
		ret := make([]string, len(in))
		for i, v := range in {
			ret[i] = v.(string)
		}
		return ret
	}

	return n9api.AlertMethodSpec{
		Description: d.Get("description").(string),
		Email: &n9api.EmailAlertMethod{
			To:      toStringSlice(d.Get("to").([]interface{})),
			Cc:      toStringSlice(d.Get("cc").([]interface{})),
			Bcc:     toStringSlice(d.Get("bcc").([]interface{})),
			Subject: d.Get("subject").(string),
			Body:    d.Get("body").(string),
		},
	}
}

func (i alertMethodEmail) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["email"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("to", config["to"])
	diags = appendError(diags, err)
	err = d.Set("cc", config["cc"])
	diags = appendError(diags, err)
	err = d.Set("bcc", config["bcc"])
	diags = appendError(diags, err)
	err = d.Set("subject", config["subject"])
	diags = appendError(diags, err)
	err = d.Set("body", config["body"])
	diags = appendError(diags, err)

	return diags
}
