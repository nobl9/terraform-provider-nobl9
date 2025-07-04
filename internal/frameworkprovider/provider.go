package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/nobl9/terraform-provider-nobl9/internal/version"
)

const providerEnvPrefix = "NOBL9"

// Ensure [Provider] satisfies various provider interfaces.
var _ provider.Provider = &Provider{}

func New() provider.Provider {
	return &Provider{}
}

// Provider defines the provider implementation.
type Provider struct{}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nobl9"
	resp.Version = version.GetBuildVersion()
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional: true,
				Description: "The [Client ID](https://docs.nobl9.com/sloctl-user-guide/#configuration) " +
					"of your Nobl9 account required to connect to Nobl9.",
				CustomType: envConfigurableStringType{},
			},
			"client_secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "The [Client Secret](https://docs.nobl9.com/sloctl-user-guide/#configuration) " +
					"of your Nobl9 account required to connect to Nobl9.",
				CustomType: envConfigurableStringType{},
			},
			"okta_org_url": schema.StringAttribute{
				Optional:    true,
				Description: "Authorization service URL.",
				CustomType:  envConfigurableStringType{},
			},
			"okta_auth_server": schema.StringAttribute{
				Optional:    true,
				Description: "Authorization service configuration.",
				CustomType:  envConfigurableStringType{},
			},
			"project": schema.StringAttribute{
				Optional:    true,
				Description: "Nobl9 project used when importing resources.",
				CustomType:  envConfigurableStringType{},
			},
			"organization": schema.StringAttribute{
				Optional:    true,
				Description: "Nobl9 Organization ID that contains resources managed by the provider.",
				CustomType:  envConfigurableStringType{},
			},
			"ingest_url": schema.StringAttribute{
				Optional:    true,
				Description: "Nobl9 API URL.",
				CustomType:  envConfigurableStringType{},
			},
			"no_config_file": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable reading configuration from file.",
				CustomType:  envConfigurableBoolType{},
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
	client, diags := newSDKClient(model)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *Provider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServiceResource,
		NewProjectResource,
	}
}

func (p *Provider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewServiceDataSource,
	}
}
