package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type alertMethodProvider interface {
	GetSchema() map[string]*schema.Schema
	GetDescription() string
	MarshalSpec(data *schema.ResourceData) v1alpha.AlertMethodSpec
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

func (a alertMethod) marshalAlertMethod(d *schema.ResourceData) (*v1alpha.AlertMethod, diag.Diagnostics) {
	metadataHolder, diags := marshalMetadata(d)
	if diags.HasError() {
		return nil, diags
	}
	// FIXME: delete ObjectInternal field after SDK update - for now it's hardcoded organization.
	return &v1alpha.AlertMethod{
		ObjectHeader: manifest.ObjectHeader{
			APIVersion:     v1alpha.APIVersion,
			Kind:           manifest.KindAlertMethod,
			MetadataHolder: metadataHolder,
			ObjectInternal: manifest.ObjectInternal{
				Organization: "nobl9-dev",
			},
		},
		Spec: a.MarshalSpec(d),
	}, diags
}

func (a alertMethod) unmarshalAlertMethod(d *schema.ResourceData, objects []sdk.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalGenericMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	spec := object["spec"].(map[string]interface{})
	err := d.Set("description", spec["description"])
	diags = appendError(diags, err)

	errs := a.UnmarshalSpec(d, spec)
	diags = append(diags, errs...)

	return diags
}

//nolint:lll
func (a alertMethod) resourceAlertMethodApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	service, diags := a.marshalAlertMethod(d)
	if diags.HasError() {
		return diags
	}

	err := clientApplyObject(ctx, client, service)
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return a.resourceAlertMethodRead(ctx, d, meta)
}

//nolint:lll
func (a alertMethod) resourceAlertMethodRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	objects, err := client.GetObjects(ctx, project, manifest.KindAlertMethod, nil, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return a.unmarshalAlertMethod(d, objects)
}

func resourceAlertMethodDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	err := client.DeleteObjectsByName(ctx, project, manifest.KindAlertMethod, false, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

type alertMethodWebhook struct{}

func (i alertMethodWebhook) GetDescription() string {
	return "[Webhook Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/webhook)"
}

//nolint:lll
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
			Description:   "Webhook message fields. The message contains JSON payload with specified fields. See documentation for allowed fields.",
			ConflictsWith: []string{"template"},
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func (i alertMethodWebhook) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
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

	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Webhook: &v1alpha.WebhookAlertMethod{
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
	return "[PagerDuty Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/pagerduty)"
}

func (i alertMethodPagerDuty) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"integration_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "PagerDuty Integration Key. For more details, check [Services and integrations](https://support.pagerduty.com/docs/services-and-integrations).",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodPagerDuty) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		PagerDuty: &v1alpha.PagerDutyAlertMethod{
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
	return "[Slack Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/slack)"
}

func (i alertMethodSlack) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Slack [webhook endpoint URL](https://slack.com/help/articles/115005265063-Incoming-webhooks-for-Slack%22).",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodSlack) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Slack: &v1alpha.SlackAlertMethod{
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
	return "[Discord Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/discord)"
}

func (i alertMethodDiscord) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Discord webhook endpoint URL. Refer to [Intro to webhooks | Discord documentation](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) for more details.",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodDiscord) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Discord: &v1alpha.DiscordAlertMethod{
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
	return "[OpsGenie Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/opsgenie)"
}

func (i alertMethodOpsgenie) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"auth": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Opsgenie authentication credentials. See [Nobl9 documentation](https://docs.nobl9.com/Alerting/Alert_methods/opsgenie#authentication) for supported formats.",
			Sensitive:   true,
			Computed:    true,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Opsgenie API URL. See [Nobl9 documentation](https://docs.nobl9.com/Alerting/Alert_methods/opsgenie#creating-opsgenie-api-key) for more details.",
		},
	}
}

func (i alertMethodOpsgenie) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Opsgenie: &v1alpha.OpsgenieAlertMethod{
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
	return "[ServiceNow Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/servicenow)"
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
		"instance_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "ServiceNow InstanceName. For details see [Nobl9 documentation](https://docs.nobl9.com/Alerting/Alert_methods/servicenow#servicenow-credentials).",
		},
	}
}

func (i alertMethodServiceNow) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		ServiceNow: &v1alpha.ServiceNowAlertMethod{
			Username:     d.Get("username").(string),
			Password:     d.Get("password").(string),
			InstanceName: d.Get("instance_name").(string),
		},
	}
}

func (i alertMethodServiceNow) UnmarshalSpec(d *schema.ResourceData, spec map[string]interface{}) diag.Diagnostics {
	config := spec["servicenow"].(map[string]interface{})
	var diags diag.Diagnostics

	err := d.Set("username", config["username"])
	diags = appendError(diags, err)
	err = d.Set("instance_name", config["instanceName"])
	diags = appendError(diags, err)

	return diags
}

type alertMethodJira struct{}

func (i alertMethodJira) GetDescription() string {
	return "[Jira Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/jira)"
}

func (i alertMethodJira) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Jira instance URL. The `https://` prefix is required.",
		},
		"username": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Jira username for the owner of the API Token.",
		},
		"apitoken": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "[API Token](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/) with access rights to the project.",
			Sensitive:   true,
			Computed:    true,
		},
		"project_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The code of the Jira project.",
		},
	}
}

func (i alertMethodJira) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Jira: &v1alpha.JiraAlertMethod{
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
	return "[MS Teams Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/ms-teams)"
}

func (i alertMethodTeams) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "MS Teams [webhook endpoint URL](https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook).",
			Sensitive:   true,
			Computed:    true,
		},
	}
}

func (i alertMethodTeams) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Teams: &v1alpha.TeamsAlertMethod{
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
	return "[Email Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/email-alert)"
}

func (i alertMethodEmail) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"to": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Recipients. The maximum number of recipients is 10.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"cc": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Carbon copy recipients. The maximum number of recipients is 10.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"bcc": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Blind carbon copy recipients. The maximum number of recipients is 10.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"subject": {
			Type:        schema.TypeString,
			Optional:    true,
			Deprecated:  "Email Subject is Deprecated as of Nobl9 1.57 release. It's not used for email generation. You can safely remove it from your configuration file.",
			Description: "Deprecated value that was used as the subject of email alert. It's not used anywhere but kept for backward compatibility.",
		},
		"body": {
			Type:        schema.TypeString,
			Optional:    true,
			Deprecated:  "Email Body is Deprecated as of Nobl9 1.57 release. It's not used for email generation. You can safely remove it from your configuration file.",
			Description: "Deprecated value that was used as the body template of email alert. It's not used anywhere but kept for backward compatibility.",
		},
	}
}

func (i alertMethodEmail) MarshalSpec(d *schema.ResourceData) v1alpha.AlertMethodSpec {
	return v1alpha.AlertMethodSpec{
		Description: d.Get("description").(string),
		Email: &v1alpha.EmailAlertMethod{
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
