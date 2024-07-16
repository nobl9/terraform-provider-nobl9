package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

type alertMethodProvider interface {
	GetSchema() map[string]*schema.Schema
	GetDescription() string
	MarshalSpec(resource resourceInterface) v1alphaAlertMethod.Spec
	UnmarshalSpec(d *schema.ResourceData, spec v1alphaAlertMethod.Spec) diag.Diagnostics
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
		CustomizeDiff: i.resourceAlertMethodValidate,
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

func (a alertMethod) marshalAlertMethod(r resourceInterface) *v1alphaAlertMethod.AlertMethod {
	displayName, _ := r.Get("display_name").(string)
	alertMethod := v1alphaAlertMethod.New(
		v1alphaAlertMethod.Metadata{
			Name:        r.Get("name").(string),
			DisplayName: displayName,
			Project:     r.Get("project").(string),
		},
		a.MarshalSpec(r),
	)
	return &alertMethod
}

func (a alertMethod) unmarshalAlertMethod(
	d *schema.ResourceData,
	objects []v1alphaAlertMethod.AlertMethod,
) diag.Diagnostics {
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

//nolint:unparam
func (a alertMethod) resourceAlertMethodValidate(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	am := a.marshalAlertMethod(d)
	errs := manifest.Validate([]manifest.Object{am})
	if errs != nil {
		return formatErrorsAsSingleError(errs)
	}
	return nil
}

//nolint:lll
func (a alertMethod) resourceAlertMethodApply(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	am := a.marshalAlertMethod(d)
	resultAm := manifest.SetDefaultProject([]manifest.Object{am}, config.Project)
	err := client.Objects().V1().Apply(ctx, resultAm)
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}
	d.SetId(am.Metadata.Name)
	return a.resourceAlertMethodRead(ctx, d, meta)
}

//nolint:lll
func (a alertMethod) resourceAlertMethodRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	alertMethods, err := client.Objects().V1().GetV1alphaAlertMethods(ctx, v1Objects.GetAlertMethodsRequest{
		Project: project,
		Names:   []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return a.unmarshalAlertMethod(d, alertMethods)
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
	err := client.Objects().V1().DeleteByName(ctx, manifest.KindAlertMethod, project, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

type alertMethodWebhook struct{}

func (i alertMethodWebhook) GetDescription() string {
	return "[Webhook Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/webhook)"
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

func (i alertMethodWebhook) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	fields := r.Get("template_fields").([]interface{})
	templateFields := make([]string, len(fields))
	for i, field := range fields {
		templateFields[i] = field.(string)
	}
	var template *string
	if t := r.Get("template").(string); t != "" {
		template = &t
		templateFields = nil
	}

	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Webhook: &v1alphaAlertMethod.WebhookAlertMethod{
			URL:            r.Get("url").(string),
			Template:       template,
			TemplateFields: templateFields,
		},
	}
}

func (i alertMethodWebhook) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAlertMethod.Spec) diag.Diagnostics {
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
	return "[PagerDuty Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/pagerduty)"
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

func (i alertMethodPagerDuty) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		PagerDuty: &v1alphaAlertMethod.PagerDutyAlertMethod{
			IntegrationKey: r.Get("integration_key").(string),
			SendResolution: marshalSendResolution(r.Get("send_resolution")),
		},
	}
}

func marshalSendResolution(sendResolutionRaw interface{}) *v1alphaAlertMethod.SendResolution {
	if sendResolutionRaw == nil {
		return nil
	}

	sendResolutionSet := sendResolutionRaw.(*schema.Set)
	if sendResolutionSet.Len() == 0 {
		return nil
	}

	sendResolution := sendResolutionSet.List()[0].(map[string]interface{})
	message := sendResolution["message"].(string)

	return &v1alphaAlertMethod.SendResolution{
		Message: &message,
	}
}

func (i alertMethodPagerDuty) UnmarshalSpec(_ *schema.ResourceData, _ v1alphaAlertMethod.Spec) diag.Diagnostics {
	// pager duty has only one, secret field
	return nil
}

type alertMethodSlack struct{}

func (i alertMethodSlack) GetDescription() string {
	return "[Slack Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/slack)"
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

func (i alertMethodSlack) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Slack: &v1alphaAlertMethod.SlackAlertMethod{
			URL: r.Get("url").(string),
		},
	}
}

func (i alertMethodSlack) UnmarshalSpec(_ *schema.ResourceData, _ v1alphaAlertMethod.Spec) diag.Diagnostics {
	// slack has only one, secret field
	return nil
}

type alertMethodDiscord struct{}

func (i alertMethodDiscord) GetDescription() string {
	return "[Discord Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/discord)"
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

func (i alertMethodDiscord) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Discord: &v1alphaAlertMethod.DiscordAlertMethod{
			URL: r.Get("url").(string),
		},
	}
}

func (i alertMethodDiscord) UnmarshalSpec(_ *schema.ResourceData, _ v1alphaAlertMethod.Spec) diag.Diagnostics {
	// discord has only one, secret field
	return nil
}

type alertMethodOpsgenie struct{}

func (i alertMethodOpsgenie) GetDescription() string {
	return "[OpsGenie Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/opsgenie)"
}

func (i alertMethodOpsgenie) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"auth": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Opsgenie authentication credentials. See [Nobl9 documentation](https://docs.nobl9.com/alerting/alert-methods/opsgenie#authentication) for supported formats.",
			Sensitive:   true,
			Computed:    true,
		},
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Opsgenie API URL. See [Nobl9 documentation](https://docs.nobl9.com/alerting/alert-methods/opsgenie#creating-opsgenie-api-key) for more details.",
		},
	}
}

func (i alertMethodOpsgenie) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Opsgenie: &v1alphaAlertMethod.OpsgenieAlertMethod{
			Auth: r.Get("auth").(string),
			URL:  r.Get("url").(string),
		},
	}
}

func (i alertMethodOpsgenie) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAlertMethod.Spec) diag.Diagnostics {
	config := spec.Opsgenie
	var diags diag.Diagnostics

	err := d.Set("url", config.URL)
	diags = appendError(diags, err)

	return diags
}

type alertMethodServiceNow struct{}

func (i alertMethodServiceNow) GetDescription() string {
	return "[ServiceNow Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/servicenow)"
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
			Description: "ServiceNow InstanceName. For details see [Nobl9 documentation](https://docs.nobl9.com/alerting/alert-methods/servicenow#servicenow-credentials).",
		},
	}
}

func (i alertMethodServiceNow) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		ServiceNow: &v1alphaAlertMethod.ServiceNowAlertMethod{
			Username:     r.Get("username").(string),
			Password:     r.Get("password").(string),
			InstanceName: r.Get("instance_name").(string),
		},
	}
}

func (i alertMethodServiceNow) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAlertMethod.Spec) diag.Diagnostics {
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
	return "[Jira Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/jira)"
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

func (i alertMethodJira) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Jira: &v1alphaAlertMethod.JiraAlertMethod{
			URL:        r.Get("url").(string),
			Username:   r.Get("username").(string),
			APIToken:   r.Get("apitoken").(string),
			ProjectKey: r.Get("project_key").(string),
		},
	}
}

func (i alertMethodJira) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAlertMethod.Spec) diag.Diagnostics {
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
	return "[MS Teams Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/ms-teams)"
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

func (i alertMethodTeams) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Teams: &v1alphaAlertMethod.TeamsAlertMethod{
			URL: r.Get("url").(string),
		},
	}
}

func (i alertMethodTeams) UnmarshalSpec(_ *schema.ResourceData, _ v1alphaAlertMethod.Spec) diag.Diagnostics {
	// teams has only one, secret field
	return nil
}

type alertMethodEmail struct{}

func (i alertMethodEmail) GetDescription() string {
	return "[Email Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/email-alert)"
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
			Deprecated:  "'subject' indicated the email alert's subject. It has been deprecated since the Nobl9 1.57 release and is no longer used to generate emails. You can safely remove it from your configuration file.",
			Description: "This value was used as the email alert's subject. 'subject' is deprecated and not used anywhere; however, its' kept for backward compatibility.",
		},
		"body": {
			Type:        schema.TypeString,
			Optional:    true,
			Deprecated:  "'body' indicated the email alert's body. It has been deprecated since the Nobl9 1.57 release and is no longer used to generate emails. You can safely remove it from your configuration file.",
			Description: "This value was used as the template for the email alert's body. 'body' is deprecated and not used anywhere; however, its' kept for backward compatibility.",
		},
	}
}

func (i alertMethodEmail) MarshalSpec(r resourceInterface) v1alphaAlertMethod.Spec {
	return v1alphaAlertMethod.Spec{
		Description: r.Get("description").(string),
		Email: &v1alphaAlertMethod.EmailAlertMethod{
			To:      toStringSlice(r.Get("to").([]interface{})),
			Cc:      toStringSlice(r.Get("cc").([]interface{})),
			Bcc:     toStringSlice(r.Get("bcc").([]interface{})),
			Subject: r.Get("subject").(string),
			Body:    r.Get("body").(string),
		},
	}
}

func (i alertMethodEmail) UnmarshalSpec(d *schema.ResourceData, spec v1alphaAlertMethod.Spec) diag.Diagnostics {
	config := spec.Email
	var diags diag.Diagnostics

	err := d.Set("to", config.To)
	diags = appendError(diags, err)
	err = d.Set("cc", config.Cc)
	diags = appendError(diags, err)
	err = d.Set("bcc", config.Bcc)
	diags = appendError(diags, err)
	//nolint:staticcheck
	err = d.Set("subject", config.Subject)
	diags = appendError(diags, err)
	//nolint:staticcheck
	err = d.Set("body", config.Body)
	diags = appendError(diags, err)

	return diags
}
