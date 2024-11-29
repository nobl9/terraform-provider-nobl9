package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kelseyhightower/envconfig"
)

// ProviderModel describes the [Provider] data model.
type ProviderModel struct {
	ClientID       types.String `tfsdk:"client_id" envconfig:"CLIENT_ID"`
	ClientSecret   types.String `tfsdk:"client_secret" envconfig:"CLIENT_SECRET"`
	OktaOrgURL     types.String `tfsdk:"okta_org_url" envconfig:"OKTA_URL"`
	OktaAuthServer types.String `tfsdk:"okta_auth_server" envconfig:"OKTA_AUTH"`
	Project        types.String `tfsdk:"project" envconfig:"PROJECT"`
	IngestURL      types.String `tfsdk:"ingest_url" envconfig:"URL"`
	Organization   types.String `tfsdk:"organization" envconfig:"ORG"`
}

// setDefaultsFromEnv sets the default values for the [ProviderModel] from the
// environment variables.
// Each env variable is prefixed with [providerEnvPrefix] and their names
// are defined under `envconfig` struct tags.
func (p *ProviderModel) setDefaultsFromEnv() diag.Diagnostics {
	var env ProviderModel
	if err := envconfig.Process(providerEnvPrefix, &env); err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("failed to process environment variables configuration", err.Error()),
		}
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
	return nil
}

// validate ensures required fields are set.
// It should be called after [ProviderModel.setDefaultsFromEnv] is called.
func (p *ProviderModel) validate() (diags diag.Diagnostics) {
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
	return diags
}
