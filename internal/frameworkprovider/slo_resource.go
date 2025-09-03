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
	_ resource.ResourceWithModifyPlan  = &SLOResource{}
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
	resp.Schema = sloResourceSchema
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
	appliedModel, diags := s.applyResource(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, appliedModel)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(s.runReplay(ctx, req.Config, model)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (s *SLOResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model SLOResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the SLO after update to fetch the computed fields.
	updatedModel, diags := s.readResource(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &updatedModel)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the
// UpdateRequest and new state values set on the UpdateResponse.
func (s *SLOResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model SLOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var sourceProject string
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("project"), &sourceProject)...)
	if resp.Diagnostics.HasError() {
		return
	}

	moveSLO := model.Project != sourceProject
	if moveSLO {
		resp.Diagnostics.Append(s.client.MoveSLOs(ctx, model.Name, sourceProject, model.Project, model.Service)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	moveSLOUpdateWarningFunc := func() {
		if moveSLO {
			resp.Diagnostics.AddWarning("SLO Update Warning",
				"SLO definition update failed, but the SLO was moved to a new project."+
					"\nResolve the issues and retry the update operation,"+
					" bear in mind that the SLO will remain in the new project.")
		}
	}

	appliedModel, diags := s.applyResource(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		moveSLOUpdateWarningFunc()
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, appliedModel)...)
	if resp.Diagnostics.HasError() {
		moveSLOUpdateWarningFunc()
		return
	}
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
	model, diags := s.readResource(ctx, &SLOResourceModel{Name: parts[1], Project: parts[0]})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
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

// ModifyPlan implements [resource.ResourceWithModifyPlan.ModifyPlan] function.
func (s *SLOResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	var plan SLOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(s.client.DryRunApplyObject(ctx, plan.ToManifest())...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *SLOResource) applyResource(
	ctx context.Context,
	model *SLOResourceModel,
) (*SLOResourceModel, diag.Diagnostics) {
	slo := model.ToManifest()
	diagnostics := s.client.ApplyObject(ctx, slo)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	// Read the SLO after creation to fetch the computed fields.
	appliedModel, diagnostics := s.readResource(ctx, model)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	return appliedModel, diagnostics
}

// readResource reads the current state of the resource from the Nobl9 API.
func (s *SLOResource) readResource(
	ctx context.Context,
	model *SLOResourceModel,
) (*SLOResourceModel, diag.Diagnostics) {
	slo, diagnostics := s.client.GetSLO(ctx, model.Name, model.Project)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	updatedModel := newSLOResourceConfigFromManifest(slo)
	s.sortLists(model, updatedModel)
	return updatedModel, diagnostics
}

// sortLists sorts lists returned by the API to ensure consistent ordering.
func (s *SLOResource) sortLists(model, updatedModel *SLOResourceModel) {
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
	updatedModel.Objectives = sortListBasedOnReferenceList(
		updatedModel.Objectives,
		model.Objectives,
		func(a, b ObjectiveModel) bool {
			return a.Name == b.Name
		},
	)
	for i := range updatedModel.Objectives {
		if !updatedModel.Objectives[i].HasCompositeObjectives() {
			continue
		}
		var refCompObj []CompositeObjectiveSpecModel
		if len(model.Objectives) > i && model.Objectives[i].HasCompositeObjectives() {
			refCompObj = model.Objectives[i].Composite[0].Components[0].Objectives[0].CompositeObjective
		}
		updatedModel.Objectives[i].Composite[0].Components[0].Objectives[0].CompositeObjective =
			sortListBasedOnReferenceList(
				updatedModel.Objectives[i].Composite[0].Components[0].Objectives[0].CompositeObjective,
				refCompObj,
				func(a, b CompositeObjectiveSpecModel) bool {
					return a.SLO == b.SLO && a.Objective == b.Objective && a.Project == b.Project
				},
			)
	}
}

func (s *SLOResource) runReplay(ctx context.Context, config tfsdk.Config, model SLOResourceModel) diag.Diagnostics {
	attributePath := path.Root("retrieve_historical_data_from")
	diags := config.GetAttribute(ctx, attributePath, &model.RetrieveHistoricalDataFrom)
	if diags.HasError() {
		return diags
	}
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
		diags.AddAttributeError(
			attributePath,
			"Failed to run historical data retrieval for the SLO.", err.Error())
		return diags
	}
	diags.AddAttributeWarning(
		attributePath,
		"Historical data retrieval for the SLO has been triggered.",
		fmt.Sprintf(
			"Data will be retrieved starting from %s.",
			model.RetrieveHistoricalDataFrom.ValueString()))
	return diags
}
