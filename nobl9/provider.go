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
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {

	ingestURL := data.Get("ingest_url").(string)
	organization := data.Get("organization").(string)
	project := data.Get("project").(string)
	userAgent := data.Get("user_agent").(string)
	clientID := data.Get("client_id").(string)
	clientSecret := data.Get("client_secret").(string)
	oktaOrgURL := data.Get("okta_org_url").(string)
	oktaAuthServer := data.Get("okta_auth_server").(string)

	c, err := nobl9.NewClient(ingestURL, organization, project, userAgent, clientID, clientSecret, oktaOrgURL, oktaAuthServer)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create HashiCups client",
			Detail:   "Unable to authenticate user for authenticated HashiCups client",
		})
		return nil, diags
	}

	return c, diags
}
