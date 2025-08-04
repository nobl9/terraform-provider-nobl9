package frameworkprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	appliedModel, diags := s.applyResource(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The attribute `retrieve_historical_data_from` is not part of the SLO manifest,
	// so we need to set it manually after reading the SLO manifest.
	appliedModel.RetrieveHistoricalDataFrom = model.RetrieveHistoricalDataFrom
	resp.Diagnostics.Append(resp.State.Set(ctx, appliedModel)...)
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

	appliedModel, diags := s.applyResource(ctx, model)
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

func (s *SLOResource) applyResource(ctx context.Context, model SLOResourceModel) (*SLOResourceModel, diag.Diagnostics) {
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
	model SLOResourceModel,
) (*SLOResourceModel, diag.Diagnostics) {
	slo, diagnostics := s.client.GetSLO(ctx, model.Name, model.Project)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	updatedModel := newSLOResourceConfigFromManifest(slo)
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
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

type sloProjectPlanModifier struct{}

func (s sloProjectPlanModifier) Description(ctx context.Context) string {
	return s.MarkdownDescription(ctx)
}

func (s sloProjectPlanModifier) MarkdownDescription(context.Context) string {
	return "Modifies the SLO plan when the `project` attribute is changed. " +
		"This modifier ensures that no other attributes are changed along with the project change, " +
		"and provides warnings about the implications of moving an SLO between projects."
}

func (s sloProjectPlanModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if isNullOrUnknown(req.StateValue) || req.StateValue == req.PlanValue {
		return
	}
	diffs, diags := calculateResourceDiff(req.State, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(s.modifyPlanForProjectChange(ctx, req.Plan, diffs)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s sloProjectPlanModifier) modifyPlanForProjectChange(
	ctx context.Context,
	plan tfsdk.Plan,
	diffs []tftypes.ValueDiff,
) diag.Diagnostics {
	if len(diffs) == 0 {
		return nil
	}
	diags := make(diag.Diagnostics, 0)

	var alertPolicies []string
	alertPoliciesPath := path.Root("alert_policies")
	diags.Append(plan.GetAttribute(ctx, alertPoliciesPath, &alertPolicies)...)
	if diags.HasError() {
		return diags
	}
	if len(alertPolicies) > 0 {
		diags.AddAttributeError(alertPoliciesPath,
			"Cannot move SLO between Projects with attached Alert Policies.",
			"You must first remove Alert Policies attached to this SLO before attempting to change it's Project.",
		)
		return diags
	}

	diags.AddAttributeWarning(path.Root("project"),
		"Changing the Project results in a dedicated operation which has several side effects (see details).",
		`Moving an SLO between Projects:
  - Creates a new Project and/or Service if the specified target objects do not yet exist.
    It is best practice to define these new objects in the Terraform configuration, before moving the SLO.
  - Updates SLO’s project in the composite SLO definition and Budget Adjustment filters.
    These definitions, which reference any objectives from the moved SLO need to be updated manually.
  - Updates its link — the former link won't work anymore.
  - Removes it from reports filtered by its previous path.
`)
	return diags
}
