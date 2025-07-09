package frameworkprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nobl9/nobl9-go/manifest"
	sdkModels "github.com/nobl9/nobl9-go/sdk/models"
)

// Ensure [SLOResource] fully satisfies framework interfaces.
var (
	_ resource.Resource                = &SLOResource{}
	_ resource.ResourceWithImportState = &SLOResource{}
	_ resource.ResourceWithConfigure   = &SLOResource{}
)

func NewSLOResource() resource.Resource {
	return &SLOResource{}
}

// SLOResource defines the [manifest.KindSLO] resource implementation.
type SLOResource struct {
	client *sdkClient
}

// Metadata implements [resource.Resource.Metadata] function.
func (s *SLOResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slo"
}

// Schema implements [resource.Resource.Schema] function.
func (s *SLOResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = sloResourceSchema()
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateRequest and new state values set on the CreateResponse.
func (s *SLOResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model SLOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(s.applyResource(ctx, model, &resp.State)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(s.runReplay(ctx, model))
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (s *SLOResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model SLOResourceModel
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &model.Name)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project"), &model.Project)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("label"), &model.Labels)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the SLO after update to fetch the computed fields.
	updatedModel, diags := s.readResource(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedModel)...)
}

// Update is called to update the state of the resource. Config, planned
// state, and prior state values should be read from the
// UpdateRequest and new state values set on the UpdateResponse.
func (s *SLOResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model SLOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(s.applyResource(ctx, model, &resp.State)...)
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (s *SLOResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model SLOResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := s.client.DeleteObject(ctx, manifest.KindSLO, model.Name, model.Project)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState is called when the provider must import the resource's state.
func (s *SLOResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID to be in the format '<project-name>/<slo-name>'.",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *SLOResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdkClient)
	if !ok {
		addInvalidSDKClientTypeDiag(&resp.Diagnostics, req.ProviderData)
		return
	}
	s.client = client
}

func (s *SLOResource) applyResource(ctx context.Context, model SLOResourceModel, state *tfsdk.State) diag.Diagnostics {
	slo := model.ToManifest()
	diagnostics := s.client.ApplyObject(ctx, slo)
	if diagnostics.HasError() {
		return diagnostics
	}

	// Read the SLO after creation to fetch the computed fields.
	appliedModel, diagnostics := s.readResource(ctx, model)
	if diagnostics.HasError() {
		return diagnostics
	}
	// The attribute `retrieve_historical_data_from` is not part of the SLO manifest,
	// so we need to set it manually after reading the SLO manifest.
	appliedModel.RetrieveHistoricalDataFrom = model.RetrieveHistoricalDataFrom
	diagnostics.Append(state.Set(ctx, appliedModel)...)
	return diagnostics
}

// readResource reads the current state of the resource from the Nobl9 API.
func (s *SLOResource) readResource(
	ctx context.Context,
	model SLOResourceModel,
) (*SLOResourceModel, diag.Diagnostics) {
	slo, diagnostics := s.client.GetSLO(ctx, model.Name, model.Project)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	updatedModel := newSLOResourceConfigFromManifest(slo)
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
	// The attribute `retrieve_historical_data_from` is not part of the SLO manifest,
	// so we need to set it manually after reading the SLO manifest.
	updatedModel.RetrieveHistoricalDataFrom = model.RetrieveHistoricalDataFrom
	return updatedModel, diagnostics
}

func (s *SLOResource) runReplay(ctx context.Context, model SLOResourceModel) diag.Diagnostic {
	attributePath := path.Root("retrieve_historical_data_from")
	if isNullOrUnknown(model.RetrieveHistoricalDataFrom) {
		return nil
	}
	replayFromTs, _ := time.Parse(time.RFC3339, model.RetrieveHistoricalDataFrom.ValueString())
	const startOffsetMinutes = 5
	windowDuration := time.Since(replayFromTs)
	payload := sdkModels.Replay{
		Project: model.Project,
		Slo:     model.Name,
		Duration: sdkModels.ReplayDuration{
			Unit:  sdkModels.DurationUnitMinute,
			Value: startOffsetMinutes + int(windowDuration.Minutes()),
		},
	}
	if err := s.client.Replay(ctx, payload); err != nil {
		return diag.NewAttributeErrorDiagnostic(
			attributePath,
			"Failed to run historical data retrieval for the SLO.", err.Error())
	}
	return diag.NewAttributeWarningDiagnostic(
		attributePath,
		"Historical data retrieval for the SLO has been triggered.",
		fmt.Sprintf(
			"Data will be retrieved starting from %s.",
			model.RetrieveHistoricalDataFrom.ValueString()))
}
