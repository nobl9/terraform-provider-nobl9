package frameworkprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"

	"github.com/nobl9/terraform-provider-nobl9/internal/version"
)

type sdkClient struct {
	client *sdk.Client
}

// newSDKClient creates new [sdk.Client] based on the [ProviderModel].
// [ProviderModel] should be first validated with [ProviderModel] before being passed to this function.
func newSDKClient(provider ProviderModel) (*sdkClient, diag.Diagnostics) {
	options := []sdk.ConfigOption{
		sdk.ConfigOptionWithCredentials(provider.ClientID.ValueString(), provider.ClientSecret.ValueString()),
		sdk.ConfigOptionEnvPrefix("TERRAFORM_NOBL9_"),
	}
	if provider.NoConfigFile.ValueBool() {
		options = append(options, sdk.ConfigOptionNoConfigFile())
	}
	sdkConfig, err := sdk.ReadConfig(options...)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("failed to read Nobl9 SDK configuration", err.Error())}
	}
	if ingestURL := provider.IngestURL.ValueString(); ingestURL != "" {
		sdkConfig.URL, err = url.Parse(ingestURL)
		if err != nil {
			return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				path.Root("ingest_url"),
				"failed to parse Nobl9 Ingest URL",
				err.Error(),
			)}
		}
	}
	if org := provider.Organization.ValueString(); org != "" {
		sdkConfig.Organization = org
	}
	if project := provider.Project.ValueString(); project != "" {
		sdkConfig.Project = project
	}
	if oktaOrgURL := provider.OktaOrgURL.ValueString(); oktaOrgURL != "" {
		sdkConfig.OktaOrgURL, err = url.Parse(oktaOrgURL)
		if err != nil {
			return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
				path.Root("okta_org_url"),
				"failed to parse Nobl9 Okta Org URL",
				err.Error(),
			)}
		}
	}
	if oktaAuthServer := provider.OktaAuthServer.ValueString(); oktaAuthServer != "" {
		sdkConfig.OktaAuthServer = oktaAuthServer
	}
	client, err := sdk.NewClient(sdkConfig)
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("failed to create Nobl9 SDK client", err.Error())}
	}
	client.SetUserAgent(version.GetUserAgent())
	return &sdkClient{client: client}, nil
}

func (s sdkClient) ApplyObject(ctx context.Context, obj manifest.Object) diag.Diagnostics {
	err := s.client.Objects().V1().Apply(ctx, []manifest.Object{obj})
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to create %s %s", obj.GetVersion(), obj.GetKind()),
				err.Error(),
			),
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("created %s %s", obj.GetVersion(), obj.GetKind()), getManifestObjectTraceAttrs(obj))
	return nil
}

func (s sdkClient) DeleteObject(ctx context.Context, kind manifest.Kind, name, project string) diag.Diagnostics {
	err := s.client.Objects().V1().DeleteByName(ctx, kind, project, name)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("Failed to delete %s %s", manifest.VersionV1alpha, kind), err.Error()),
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted %s %s", manifest.VersionV1alpha, kind), map[string]any{
		"name":    name,
		"project": project,
	})
	return nil
}

func (s sdkClient) GetService(
	ctx context.Context,
	name, project string,
) (svc v1alphaService.Service, diags diag.Diagnostics) {
	obj, diags := genericGetObject(ctx, s.client, manifest.KindService, name, project)
	if diags.HasError() {
		return svc, diags
	}
	svc, ok := obj.(v1alphaService.Service)
	if !ok {
		return svc, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to cast %T to %T", obj, svc),
				"Please report this issue to the provider developers."),
		}
	}
	return svc, nil
}

func (s sdkClient) GetProject(
	ctx context.Context,
	name string,
) (project v1alphaProject.Project, diags diag.Diagnostics) {
	obj, diags := genericGetObject(ctx, s.client, manifest.KindProject, name, "")
	if diags.HasError() {
		return project, diags
	}
	project, ok := obj.(v1alphaProject.Project)
	if !ok {
		return project, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to cast %T to %T", obj, project),
				"Please report this issue to the provider developers."),
		}
	}
	return project, nil
}

// genericGetObject should only be called by [sdkClient].
func genericGetObject(
	ctx context.Context,
	client *sdk.Client,
	kind manifest.Kind,
	name, project string,
) (manifest.Object, diag.Diagnostics) {
	header := http.Header{}
	if project != "" {
		header.Add(sdk.HeaderProject, project)
	}
	objects, err := client.Objects().V1().Get(
		ctx,
		kind,
		header,
		url.Values{v1Objects.QueryKeyName: []string{name}},
	)
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("Failed to get %s %s", manifest.VersionV1alpha, kind), err.Error()),
		}
	}
	if len(objects) != 1 {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to get %s %s", manifest.VersionV1alpha, kind),
				fmt.Sprintf("unexpected number of objects in response, expected 1, got %d", len(objects))),
		}
	}
	obj := objects[0]
	tflog.Trace(ctx, fmt.Sprintf("fetched %s %s", manifest.VersionV1alpha, kind), getManifestObjectTraceAttrs(obj))
	return obj, nil
}

func getManifestObjectTraceAttrs(obj manifest.Object) map[string]any {
	attrs := map[string]any{
		"name": obj.GetName(),
	}
	if projectScoped, ok := obj.(manifest.ProjectScopedObject); ok {
		attrs["project"] = projectScoped.GetProject()
	}
	return attrs
}
