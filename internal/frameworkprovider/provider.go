package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const providerEnvPrefix = "NOBL9"

// Ensure [Provider] satisfies various provider interfaces.
var _ provider.Provider = &Provider{}

func New(version string) provider.Provider {
	return &Provider{
		version: version,
	}
}

// Provider defines the provider implementation.
type Provider struct {
	version string
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nobl9"
	resp.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The [Client ID](https://docs.nobl9.com/sloctl-user-guide/#configuration) " +
					"of your Nobl9 account required to connect to Nobl9.",
			},
			"client_secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				MarkdownDescription: "The [Client Secret](https://docs.nobl9.com/sloctl-user-guide/#configuration) " +
					"of your Nobl9 account required to connect to Nobl9.",
			},
			"okta_org_url": schema.StringAttribute{
				Optional:    true,
				Description: "Authorization service URL.",
			},
			"okta_auth_server": schema.StringAttribute{
				Optional:    true,
				Description: "Authorization service configuration.",
			},
			"project": schema.StringAttribute{
				Optional:    true,
				Description: "Nobl9 project used when importing resources.",
			},
			"organization": schema.StringAttribute{
				Optional:    true,
				Description: "Nobl9 Organization ID that contains resources managed by the provider.",
			},
			"ingest_url": schema.StringAttribute{
				Optional:    true,
				Description: "Nobl9 API URL.",
			},
		},
	}
}

// Configure is called by the framework to configure the [Provider].
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(model.setDefaultsFromEnv()...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(model.validate()...)
	if resp.Diagnostics.HasError() {
		return
	}
	client, diags := newSDKClient(model, p.version)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.ResourceData = client
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServiceResource,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
