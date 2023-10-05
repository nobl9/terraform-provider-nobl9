package nobl9

import (
	"context"
	"net/url"
	"sync"

	"github.com/nobl9/nobl9-go/sdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//nolint:gochecknoglobals,revive
var Version string

// FIXME PC-9234: Edit everything that was necessary but is optional now.
// FIXME PC-9234: Set 'Deprecated' for project and organization (need to get a date for deletion first).
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
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_ORG", nil),
				Description: "Nobl9 [Organization ID](https://docs.nobl9.com/API_Documentation/api-endpoints-for-slo-annotations/#common-headers) that contains resources managed by the Nobl9 Terraform provider.",
				Deprecated:  "test organization deprecation message; test deprecation date: 19700101",
			},

			"project": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_PROJECT", sdk.DefaultProject),
				Description: "Nobl9 project used when importing resources.",
			},

			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_CLIENT_ID", nil),
				Description: "the [Client ID](https://docs.nobl9.com/sloctl-user-guide/#configuration) of your Nobl9 account required to connect to Nobl9.",
			},

			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_CLIENT_SECRET", nil),
				Description: "the [Client Secret](https://docs.nobl9.com/sloctl-user-guide/#configuration) of your Nobl9 account required to connect to Nobl9.",
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
			"nobl9_service":                                 resourceService(),
			"nobl9_agent":                                   resourceAgent(),
			"nobl9_alert_policy":                            resourceAlertPolicy(),
			"nobl9_alert_method_webhook":                    resourceAlertMethodFactory(alertMethodWebhook{}),
			"nobl9_alert_method_pagerduty":                  resourceAlertMethodFactory(alertMethodPagerDuty{}),
			"nobl9_alert_method_slack":                      resourceAlertMethodFactory(alertMethodSlack{}),
			"nobl9_alert_method_discord":                    resourceAlertMethodFactory(alertMethodDiscord{}),
			"nobl9_alert_method_opsgenie":                   resourceAlertMethodFactory(alertMethodOpsgenie{}),
			"nobl9_alert_method_servicenow":                 resourceAlertMethodFactory(alertMethodServiceNow{}),
			"nobl9_alert_method_jira":                       resourceAlertMethodFactory(alertMethodJira{}),
			"nobl9_alert_method_msteams":                    resourceAlertMethodFactory(alertMethodTeams{}),
			"nobl9_alert_method_email":                      resourceAlertMethodFactory(alertMethodEmail{}),
			"nobl9_direct_" + appDynamicsDirectType:         resourceDirectFactory(appDynamicsDirectSpec{}),
			"nobl9_direct_" + bigqueryDirectType:            resourceDirectFactory(bigqueryDirectSpec{}),
			"nobl9_direct_" + cloudWatchDirectType:          resourceDirectFactory(cloudWatchDirectSpec{}),
			"nobl9_direct_" + datadogDirectType:             resourceDirectFactory(datadogDirectSpec{}),
			"nobl9_direct_" + dynatraceDirectType:           resourceDirectFactory(dynatraceDirectSpec{}),
			"nobl9_direct_" + gcmDirectType:                 resourceDirectFactory(gcmDirectSpec{}),
			"nobl9_direct_" + influxdbDirectType:            resourceDirectFactory(influxdbDirectSpec{}),
			"nobl9_direct_" + instanaDirectType:             resourceDirectFactory(instanaDirectSpec{}),
			"nobl9_direct_" + lightstepDirectType:           resourceDirectFactory(lightstepDirectSpec{}),
			"nobl9_direct_" + newRelicDirectType:            resourceDirectFactory(newRelicDirectSpec{}),
			"nobl9_direct_" + pingdomDirectType:             resourceDirectFactory(pingdomDirectSpec{}),
			"nobl9_direct_" + redshiftDirectType:            resourceDirectFactory(redshiftDirectSpec{}),
			"nobl9_direct_" + splunkDirectType:              resourceDirectFactory(splunkDirectSpec{}),
			"nobl9_direct_" + splunkObservabilityDirectType: resourceDirectFactory(splunkObservabilityDirectSpec{}),
			"nobl9_direct_" + sumologicDirectType:           resourceDirectFactory(sumologicDirectSpec{}),
			"nobl9_direct_" + thousandeyesDirectType:        resourceDirectFactory(thousandeyesDirectSpec{}),
			"nobl9_project":                                 resourceProject(),
			"nobl9_role_binding":                            resourceRoleBinding(),
			"nobl9_slo":                                     resourceSLO(),
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

//nolint:gochecknoglobals
var (
	sharedClient *sdk.Client
	once         sync.Once
)

func getClient(config ProviderConfig) (*sdk.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	once.Do(func() {
		options := []sdk.ConfigOption{}
		// TODO PC-9234: Do we use envs prefix?
		options = append(options, sdk.ConfigOptionWithCredentials(config.ClientID, config.ClientSecret))
		conf, err := sdk.ReadConfig(options...)
		if err != nil {
			panic(err)
		}
		ingestURL, err := url.Parse(config.IngestURL)
		if err != nil {
			panic(err)
		}
		conf.URL = ingestURL
		conf.Organization = config.Organization
		conf.Project = config.Project
		oktaOrgURL, err := url.Parse(config.OktaOrgURL)
		if err != nil {
			panic(err)
		}
		conf.OktaOrgURL = oktaOrgURL
		conf.OktaAuthServer = config.OktaAuthServer

		sharedClient, err = sdk.NewClient(conf)
		if err != nil {
			panic(err)
		}
	})
	if len(diags) > 0 {
		return nil, diags
	}
	return sharedClient, nil
}
