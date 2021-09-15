package nobl9

import (
	"context"

	"github.com/nobl9/nobl9-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ingest_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_URL", nil),
				Description: "",
			},

			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_ORG", nil),
				Description: "",
			},

			"project": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_PROJECT", nil),
				Description: "",
				Default:     "default",
			},

			"user_agent": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_AGENT", nil),
				Description: "",
			},

			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_CLIENT_ID", nil),
				Description: "",
			},

			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_CLIENT_SECRET", nil),
				Description: "",
			},

			"okta_org_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_OKTA_URL", nil),
				Description: "",
			},

			"okta_auth_server": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOBL9_OKTA_AUTH", nil),
				Description: "",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ResourcesMap: map[string]*schema.Resource{
			"nobl9_service": resourceService(),
			"nobl9_agent":   resourceAgent(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

type ProviderConfig struct {
	IngestURL      string
	Organization   string
	Project        string
	UserAgent      string
	ClientID       string
	ClientSecret   string
	OktaOrgURL     string
	OktaAuthServer string
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := ProviderConfig{
		IngestURL:      data.Get("ingest_url").(string),
		Organization:   data.Get("organization").(string),
		Project:        data.Get("project").(string),
		UserAgent:      data.Get("user_agent").(string),
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
		config.UserAgent,
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
				Detail:   "Unable to authenticate user for authenticated Nobl9 client",
			},
		}
	}

	return c, nil
}
