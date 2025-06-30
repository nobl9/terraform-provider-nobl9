package frameworkprovider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nobl9/nobl9-go/manifest"
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
	resp.Diagnostics.Append(s.applyResource(ctx, req.Plan, &resp.State)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: historical data from
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
	resp.Diagnostics.Append(s.applyResource(ctx, req.Plan, &resp.State)...)
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
	resp.State.SetAttribute(ctx, path.Root("project"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("name"), parts[1])
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

func (s *SLOResource) applyResource(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model SLOResourceModel
	diagnostics := plan.Get(ctx, &model)
	if diagnostics.HasError() {
		return diagnostics
	}

	slo := model.ToManifest()
	diagnostics.Append(s.client.ApplyObject(ctx, slo)...)
	if diagnostics.HasError() {
		return diagnostics
	}

	// Read the SLO after creation to fetch the computed fields.
	appliedModel, diagnostics := s.readResource(ctx, model)
	if diagnostics.HasError() {
		return diagnostics
	}
	// Save data into Terraform state
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
	// Sort Labels.
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
	return updatedModel, diagnostics
}
