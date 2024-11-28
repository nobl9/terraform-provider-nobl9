package frameworkprovider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kelseyhightower/envconfig"
	"github.com/nobl9/nobl9-go/sdk"
)

const providerEnvPrefix = "NOBL9"

// Ensure [Provider] satisfies various provider interfaces.
var _ provider.Provider = &Provider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

// Provider defines the provider implementation.
type Provider struct {
	version string
}

// ProviderConfig describes the [Provider] data config.
type ProviderConfig struct {
	ClientID       types.String `tfsdk:"client_id" envconfig:"CLIENT_ID"`
	ClientSecret   types.String `tfsdk:"client_secret" envconfig:"CLIENT_SECRET"`
	OktaOrgURL     types.String `tfsdk:"okta_org_url" envconfig:"OKTA_URL"`
	OktaAuthServer types.String `tfsdk:"okta_auth_server" envconfig:"OKTA_AUTH"`
	Project        types.String `tfsdk:"project" envconfig:"PROJECT"`
	IngestURL      types.String `tfsdk:"ingest_url" envconfig:"URL"`
	Organization   types.String `tfsdk:"organization" envconfig:"ORG"`
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
	var config ProviderConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.setDefaultsFromEnv(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	config.validate(&resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.ResourceData = p.newSDKClient(config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServiceResource,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// newSDKClient creates new [sdk.Client] based on the [ProviderConfig].
// [ProviderConfig] should be first validated with [ProviderConfig] before being passed to this function.
func (p *Provider) newSDKClient(config ProviderConfig, diags *diag.Diagnostics) *sdk.Client {
	options := []sdk.ConfigOption{
		sdk.ConfigOptionWithCredentials(config.ClientID.ValueString(), config.ClientSecret.ValueString()),
		sdk.ConfigOptionNoConfigFile(),
		sdk.ConfigOptionEnvPrefix("TERRAFORM_NOBL9_"),
	}
	sdkConfig, err := sdk.ReadConfig(options...)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("failed to read Nobl9 SDK configuration", err.Error()))
		return nil
	}
	if ingestURL := config.IngestURL.ValueString(); ingestURL != "" {
		sdkConfig.URL, err = url.Parse(ingestURL)
		if err != nil {
			diags.Append(diag.NewAttributeErrorDiagnostic(
				path.Root("ingest_url"),
				"failed to parse Nobl9 Ingest URL",
				err.Error(),
			))
			return nil
		}
	}
	if org := config.Organization.ValueString(); org != "" {
		sdkConfig.Organization = org
	}
	if project := config.Project.ValueString(); project != "" {
		sdkConfig.Project = project
	}
	if oktaOrgURL := config.OktaOrgURL.ValueString(); oktaOrgURL != "" {
		sdkConfig.OktaOrgURL, err = url.Parse(oktaOrgURL)
		if err != nil {
			diags.Append(diag.NewAttributeErrorDiagnostic(
				path.Root("okta_org_url"),
				"failed to parse Nobl9 Okta Org URL",
				err.Error(),
			))
			return nil
		}
	}
	if oktaAuthServer := config.OktaAuthServer.ValueString(); oktaAuthServer != "" {
		sdkConfig.OktaAuthServer = oktaAuthServer
	}
	client, err := sdk.NewClient(sdkConfig)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("failed to create Nobl9 SDK client", err.Error()))
		return nil
	}
	client.SetUserAgent(fmt.Sprintf("terraform-%s", p.version))
	return client
}

// setDefaultsFromEnv sets the default values for the [ProviderConfig] from the
// environment variables.
// Each env variable is prefixed with [providerEnvPrefix] and their names
// are defined under `envconfig` struct tags.
func (p *ProviderConfig) setDefaultsFromEnv(diags *diag.Diagnostics) {
	var env ProviderConfig
	if err := envconfig.Process(providerEnvPrefix, &env); err != nil {
		diags.Append(diag.NewErrorDiagnostic("failed to process environment variables configuration", err.Error()))
		return
	}
	if p.ClientID.IsNull() {
		p.ClientID = env.ClientID
	}
	if p.ClientSecret.IsNull() {
		p.ClientSecret = env.ClientSecret
	}
	if p.OktaOrgURL.IsNull() {
		p.OktaOrgURL = env.OktaOrgURL
	}
	if p.OktaAuthServer.IsNull() {
		p.OktaAuthServer = env.OktaAuthServer
	}
	if p.Project.IsNull() {
		p.Project = env.Project
	}
	if p.IngestURL.IsNull() {
		p.IngestURL = env.IngestURL
	}
	if p.Organization.IsNull() {
		p.Organization = env.Organization
	}
}

// validate ensures required fields are set.
// It should be called after [ProviderConfig.setDefaultsFromEnv] is called.
func (p *ProviderConfig) validate(diags *diag.Diagnostics) {
	if p.ClientID.IsNull() {
		diags.Append(diag.NewAttributeErrorDiagnostic(
			path.Root("client_id"),
			"missing required field",
			"client_id is required to connect to Nobl9",
		))
	}
	if p.ClientSecret.IsNull() {
		diags.Append(diag.NewAttributeErrorDiagnostic(
			path.Root("client_secret"),
			"missing required field",
			"client_secret is required to connect to Nobl9",
		))
	}
}
