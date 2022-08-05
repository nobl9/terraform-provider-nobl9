package nobl9

import (
	"context"

	"github.com/nobl9/nobl9-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//nolint:gochecknoglobals,revive
var Version string

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ingest_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_URL", "https://app.nobl9.com/api"),
				Description: "Nobl9 API URL.",
			},

			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_ORG", nil),
				Description: "Nobl9 Organization that contain resources managed by this provider.",
			},

			"project": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_PROJECT", nil),
				Description: "Nobl9 project used when importing resources.",
			},

			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_CLIENT_ID", nil),
				Description: "Authentication parameter ClientID.",
			},

			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_CLIENT_SECRET", nil),
				Description: "Authentication parameter ClientSecret.",
			},

			"okta_org_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_OKTA_URL", "https://accounts.nobl9.com"),
				Description: "Authorization service URL.",
			},

			"okta_auth_server": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_OKTA_AUTH", "auseg9kiegWKEtJZC416"),
				Description: "Authorization service configuration.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ResourcesMap: map[string]*schema.Resource{
			"nobl9_service":                 resourceService(),
			"nobl9_agent":                   resourceAgent(),
			"nobl9_alert_policy":            resourceAlertPolicy(),
			"nobl9_alert_method_webhook":    resourceAlertMethodFactory(alertMethodWebhook{}),
			"nobl9_alert_method_pagerduty":  resourceAlertMethodFactory(alertMethodPagerDuty{}),
			"nobl9_alert_method_slack":      resourceAlertMethodFactory(alertMethodSlack{}),
			"nobl9_alert_method_discord":    resourceAlertMethodFactory(alertMethodDiscord{}),
			"nobl9_alert_method_opsgenie":   resourceAlertMethodFactory(alertMethodOpsgenie{}),
			"nobl9_alert_method_servicenow": resourceAlertMethodFactory(alertMethodServiceNow{}),
			"nobl9_alert_method_jira":       resourceAlertMethodFactory(alertMethodJira{}),
			"nobl9_alert_method_msteams":    resourceAlertMethodFactory(alertMethodTeams{}),
			"nobl9_alert_method_email":      resourceAlertMethodFactory(alertMethodEmail{}),
			"nobl9_project":                 resourceProject(),
			"nobl9_role_binding":            resourceRoleBinding(),
			"nobl9_slo":                     resourceSLO(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

type ProviderConfig struct {
	IngestURL      string
	Organization   string
	Project        string
	ClientID       string
	ClientSecret   string
	OktaOrgURL     string
	OktaAuthServer string
}

func providerConfigure(_ context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := ProviderConfig{
		IngestURL:      data.Get("ingest_url").(string),
		Organization:   data.Get("organization").(string),
		Project:        data.Get("project").(string),
		ClientID:       data.Get("client_id").(string),
		ClientSecret:   data.Get("client_secret").(string),
		OktaOrgURL:     data.Get("okta_org_url").(string),
		OktaAuthServer: data.Get("okta_auth_server").(string),
	}

	return config, nil
}

func newClient(config ProviderConfig, project string) (*nobl9.Client, diag.Diagnostics) {
	c, err := nobl9.NewClient(
		config.IngestURL,
		config.Organization,
		project,
		"terraform-"+Version,
		config.ClientID,
		config.ClientSecret,
		config.OktaOrgURL,
		config.OktaAuthServer,
	)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Nobl9 client",
				Detail:   err.Error(),
			},
		}
	}

	return c, nil
}
