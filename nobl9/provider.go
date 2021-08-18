package nobl9

import (
	"context"

	n9api "github.com/nobl9/nobl9-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ingestURL": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"project": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"userAgent": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"clientID": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"clientSecret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"oktaOrgURL": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},

			"oktaAuthServer": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc(),
				Description: "",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ResourcesMap: map[string]*schema.Resource{
			"service": ResourceService(),
		},

		ConfigureContextFunc: configure,
	}
}

func configure(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {

	c, _ := n9api.NewClient(
		data.Get("ingestURL").(string),
		data.Get("organization").(string),
		data.Get("project").(string),
		data.Get("userAgent").(string),
		data.Get("clientID").(string),
		data.Get("clientSecret").(string),
		data.Get("oktaOrgURL").(string),
		data.Get("oktaAuthServer").(string),
	)

	return c, nil
}

// ingestURL: data.Get("ingestURL").(string),
// organization: data.Get("organization").(string),
// project: data.Get("project").(string),
// userAgent: data.Get("userAgent").(string),
// clientID: data.Get("clientID").(string),
// clientSecret: data.Get("clientSecret").(string),
// oktaOrgURL: data.Get("oktaOrgURL").(string),
// oktaAuthServer: data.Get("oktaAuthServer").(string),
