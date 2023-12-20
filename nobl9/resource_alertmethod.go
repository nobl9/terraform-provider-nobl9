package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAM "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
)

type alertMethodProvider interface {
	GetSchema() map[string]*schema.Schema
	GetDescription() string
	MarshalSpec(data *schema.ResourceData) v1alphaAM.Spec
	UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics
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

func (a alertMethod) marshalAlertMethod(d *schema.ResourceData) *v1alphaAM.AlertMethod {
	displayName, _ := d.Get("display_name").(string)
	return &v1alphaAM.AlertMethod{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindAlertMethod,
		Metadata: v1alphaAM.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: displayName,
			Project:     d.Get("project").(string),
		},
		Spec: a.MarshalSpec(d),
	}
}

func (a alertMethod) unmarshalAlertMethod(d *schema.ResourceData, objects []v1alphaAM.AlertMethod) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics
	metadata := object.Metadata
	err := d.Set("name", metadata.Name)
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata.DisplayName)
	diags = appendError(diags, err)
	err = d.Set("project", metadata.Project)
	diags = appendError(diags, err)
	spec := object.Spec
	err = d.Set("description", spec.Description)
	diags = appendError(diags, err)
	errs := a.UnmarshalSpec(d, spec)
	diags = append(diags, errs...)
	return diags
}

//nolint:lll
func (a alertMethod) resourceAlertMethodApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	am := a.marshalAlertMethod(d)
	resultAm := manifest.SetDefaultProject([]manifest.Object{am}, config.Project)
	err := client.ApplyObjects(ctx, resultAm)
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}
	d.SetId(am.Metadata.Name)
	return a.resourceAlertMethodRead(ctx, d, meta)
}

//nolint:lll
func (a alertMethod) resourceAlertMethodRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	objects, err := client.GetObjects(ctx, project, manifest.KindAlertMethod, nil, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return a.unmarshalAlertMethod(d, manifest.FilterByKind[v1alphaAM.AlertMethod](objects))
}

func resourceAlertMethodDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
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

func (i alertMethodWebhook) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
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

	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Webhook: &v1alphaAM.WebhookAlertMethod{
			URL:            d.Get("url").(string),
			Template:       template,
			TemplateFields: templateFields,
		},
	}
}

func (i alertMethodWebhook) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
	config := spec.Webhook
	var diags diag.Diagnostics

	err := d.Set("template", config.Template)
	diags = appendError(diags, err)
	err = d.Set("template_fields", config.TemplateFields)
	diags = appendError(diags, err)

	return diags
}

type alertMethodPagerDuty struct{}

func (i alertMethodPagerDuty) GetDescription() string {
	return "[PagerDuty Alert Method | Nobl9 Documentation](https://docs.nobl9.com/Alerting/Alert_methods/pagerduty)"
}

func (i alertMethodPagerDuty) GetSchema() map[string]*schema.Schema {
	sendResolutionSchema := map[string]*schema.Schema{
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A message that will be attached to your 'all clear' notification.",
		},
	}

	return map[string]*schema.Schema{
		"integration_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "PagerDuty Integration Key. For more details, check [Services and integrations](https://support.pagerduty.com/docs/services-and-integrations).",
			Sensitive:   true,
			Computed:    true,
		},
		"send_resolution": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Sends a notification after the cooldown period is over.",
			MinItems:    1,
			MaxItems:    1,
			Elem:        &schema.Resource{Schema: sendResolutionSchema},
		},
	}
}

func (i alertMethodPagerDuty) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		PagerDuty: &v1alphaAM.PagerDutyAlertMethod{
			IntegrationKey: d.Get("integration_key").(string),
			SendResolution: marshalSendResolution(d.Get("send_resolution")),
		},
	}
}

func marshalSendResolution(sendResolutionRaw interface{}) *v1alphaAM.SendResolution {
	if sendResolutionRaw == nil {
		return nil
	}

	sendResolutionSet := sendResolutionRaw.(*schema.Set)
	if sendResolutionSet.Len() == 0 {
		return nil
	}

	sendResolution := sendResolutionSet.List()[0].(map[string]interface{})
	message := sendResolution["message"].(string)

	return &v1alphaAM.SendResolution{
		Message: &message,
	}
}

func (i alertMethodPagerDuty) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
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

func (i alertMethodSlack) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Slack: &v1alphaAM.SlackAlertMethod{
			URL: d.Get("url").(string),
		},
	}
}

func (i alertMethodSlack) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
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

func (i alertMethodDiscord) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Discord: &v1alphaAM.DiscordAlertMethod{
			URL: d.Get("url").(string),
		},
	}
}

func (i alertMethodDiscord) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
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

func (i alertMethodOpsgenie) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Opsgenie: &v1alphaAM.OpsgenieAlertMethod{
			Auth: d.Get("auth").(string),
			URL:  d.Get("url").(string),
		},
	}
}

func (i alertMethodOpsgenie) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
	config := spec.Opsgenie
	var diags diag.Diagnostics

	err := d.Set("url", config.URL)
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

func (i alertMethodServiceNow) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		ServiceNow: &v1alphaAM.ServiceNowAlertMethod{
			Username:     d.Get("username").(string),
			Password:     d.Get("password").(string),
			InstanceName: d.Get("instance_name").(string),
		},
	}
}

func (i alertMethodServiceNow) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
	config := spec.ServiceNow
	var diags diag.Diagnostics

	err := d.Set("username", config.Username)
	diags = appendError(diags, err)
	err = d.Set("instance_name", config.InstanceName)
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

func (i alertMethodJira) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Jira: &v1alphaAM.JiraAlertMethod{
			URL:        d.Get("url").(string),
			Username:   d.Get("username").(string),
			APIToken:   d.Get("apitoken").(string),
			ProjectKey: d.Get("project_key").(string),
		},
	}
}

func (i alertMethodJira) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
	config := spec.Jira
	var diags diag.Diagnostics

	err := d.Set("username", config.Username)
	diags = appendError(diags, err)
	err = d.Set("url", config.URL)
	diags = appendError(diags, err)
	err = d.Set("project_key", config.ProjectKey)
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

func (i alertMethodTeams) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Teams: &v1alphaAM.TeamsAlertMethod{
			URL: d.Get("url").(string),
		},
	}
}

func (i alertMethodTeams) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
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

func (i alertMethodEmail) MarshalSpec(d *schema.ResourceData) v1alphaAM.Spec {
	return v1alphaAM.Spec{
		Description: d.Get("description").(string),
		Email: &v1alphaAM.EmailAlertMethod{
			To:      toStringSlice(d.Get("to").([]interface{})),
			Cc:      toStringSlice(d.Get("cc").([]interface{})),
			Bcc:     toStringSlice(d.Get("bcc").([]interface{})),
			Subject: d.Get("subject").(string),
			Body:    d.Get("body").(string),
		},
	}
}

func (i alertMethodEmail) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAM.Spec) diag.Diagnostics {
	config := spec.Email
	var diags diag.Diagnostics

	err := d.Set("to", config.To)
	diags = appendError(diags, err)
	err = d.Set("cc", config.Cc)
	diags = appendError(diags, err)
	err = d.Set("bcc", config.Bcc)
	diags = appendError(diags, err)
	err = d.Set("subject", config.Subject)
	diags = appendError(diags, err)
	err = d.Set("body", config.Body)
	diags = appendError(diags, err)

	return diags
}
