package frameworkprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	sdkModels "github.com/nobl9/nobl9-go/sdk/models"

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
	setClientUserAgent(client)
	return &sdkClient{client: client}, nil
}

func (s sdkClient) ApplyObject(ctx context.Context, obj manifest.Object) diag.Diagnostics {
	err := s.client.Objects().V2().Apply(ctx, v2.ApplyRequest{Objects: []manifest.Object{obj}})
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to create %s %s", obj.GetVersion(), obj.GetKind()),
				err.Error(),
			),
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("created %s %s", obj.GetVersion(), obj.GetKind()), getManifestObjectTraceAttrs(obj))
	return nil
}

func (s sdkClient) DryRunApplyObject(ctx context.Context, obj manifest.Object) diag.Diagnostics {
	err := s.client.Objects().V2().Apply(ctx, v2.ApplyRequest{Objects: []manifest.Object{obj}, DryRun: true})
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Dry-run apply failed for %s %s", obj.GetVersion(), obj.GetKind()),
				err.Error(),
			),
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("dry-run apply succeeded %s %s",
		obj.GetVersion(), obj.GetKind()), getManifestObjectTraceAttrs(obj))
	return nil
}

func (s sdkClient) DeleteObject(ctx context.Context, kind manifest.Kind, name, project string) diag.Diagnostics {
	err := s.client.Objects().V2().DeleteByName(ctx, v2.DeleteByNameRequest{
		Kind:    kind,
		Project: project,
		Names:   []string{name},
	})
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("Failed to delete %s %s", manifest.VersionV1alpha, kind), err.Error()),
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("deleted %s %s", manifest.VersionV1alpha, kind), map[string]any{
		"name":    name,
		"project": project,
	})
	return nil
}

func (s sdkClient) GetService(ctx context.Context, name, project string) (v1alphaService.Service, diag.Diagnostics) {
	return genericGetObject[v1alphaService.Service](ctx, s.client, manifest.KindService, name, project)
}

func (s sdkClient) GetProject(ctx context.Context, name string) (v1alphaProject.Project, diag.Diagnostics) {
	return genericGetObject[v1alphaProject.Project](ctx, s.client, manifest.KindProject, name, "")
}

func (s sdkClient) GetSLO(ctx context.Context, name, project string) (v1alphaSLO.SLO, diag.Diagnostics) {
	return genericGetObject[v1alphaSLO.SLO](ctx, s.client, manifest.KindSLO, name, project)
}

// Replay runs historical data retrieval for the given SLO.
//
// TODO: Once https://github.com/nobl9/nobl9-go/pull/756 is merged,
// we can remove this in favor of SDK-defined methods.
func (s sdkClient) Replay(ctx context.Context, payload sdkModels.Replay) error {
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(payload); err != nil {
		return err
	}
	header := http.Header{sdk.HeaderProject: []string{payload.Project}}
	req, err := s.client.CreateRequest(ctx, http.MethodPost, "timetravel", header, nil, body)
	if err != nil {
		return err
	}
	resp, err := s.client.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errors.New(replayUnavailabilityReasonExplanation(data, resp.StatusCode))
	}
	return err
}

func (s sdkClient) MoveSLOs(ctx context.Context, sloName, oldProject, newProject, newService string) diag.Diagnostics {
	err := s.client.Objects().V1().MoveSLOs(ctx, v1Objects.MoveSLOsRequest{
		SLONames:   []string{sloName},
		OldProject: oldProject,
		NewProject: newProject,
		Service:    newService,
	})
	if err != nil {
		return diag.Diagnostics{
			diag.NewAttributeErrorDiagnostic(
				path.Root("project"),
				fmt.Sprintf("Failed to move %s SLO from %s to %s Project", sloName, oldProject, newProject),
				err.Error(),
			),
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("moved %s SLO from %s to %s Project", sloName, oldProject, newProject))
	return nil
}

// genericGetObject should only be called by [sdkClient].
func genericGetObject[T manifest.Object](
	ctx context.Context,
	client *sdk.Client,
	kind manifest.Kind,
	name, project string,
) (typed T, diags diag.Diagnostics) {
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
		return typed, diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("Failed to get %s %s", manifest.VersionV1alpha, kind), err.Error()),
		}
	}
	if len(objects) != 1 {
		return typed, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to get %s %s", manifest.VersionV1alpha, kind),
				fmt.Sprintf("unexpected number of objects in response, expected 1, got %d", len(objects))),
		}
	}
	obj := objects[0]
	tflog.Trace(ctx, fmt.Sprintf("fetched %s %s", manifest.VersionV1alpha, kind), getManifestObjectTraceAttrs(obj))
	var ok bool
	typed, ok = obj.(T)
	if !ok {
		return typed, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to cast %T to %T", obj, typed),
				"Please report this issue to the provider developers."),
		}
	}
	return typed, nil
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

func setClientUserAgent(client *sdk.Client) {
	client.SetUserAgent(fmt.Sprintf("terraform-%s", version.GetUserAgent()))
}

func replayUnavailabilityReasonExplanation(reason []byte, statusCode int) string {
	strReason := strings.TrimSpace(string(reason))
	switch strReason {
	case sdkModels.ReplayIntegrationDoesNotSupportReplay:
		return "The Data Source does not support Replay yet"
	case sdkModels.ReplayAgentVersionDoesNotSupportReplay:
		return "Update your Agent version to the latest to use Replay for this Data Source."
	case sdkModels.ReplayMaxHistoricalDataRetrievalTooLow:
		return "Value configured for spec.historicalDataRetrieval.maxDuration.value" +
			" for the Data Source is lower than the duration you're trying to run Replay for."
	case sdkModels.ReplayConcurrentReplayRunsLimitExhausted:
		return "You've exceeded the limit of concurrent Replay runs. Wait until the current Replay(s) are done."
	case sdkModels.ReplayUnknownAgentVersion:
		return "Your Agent isn't connected to the Data Source. Deploy the Agent and run Replay once again."
	case "single_query_not_supported":
		return "Historical data retrieval for single-query ratio metrics is not supported"
	case "composite_slo_not_supported":
		return "Historical data retrieval for Composite SLO is not supported"
	case "promql_in_gcm_not_supported":
		return "Historical data retrieval for PromQL metrics is not supported"
	default:
		return fmt.Sprintf("bad response (status: %d): %s", statusCode, strReason)
	}
}
