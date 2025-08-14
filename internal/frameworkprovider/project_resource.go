package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nobl9/nobl9-go/manifest"
)

// Ensure [ProjectResource] fully satisfies framework interfaces.
var (
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}
	_ resource.ResourceWithConfigure   = &ProjectResource{}
)

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource defines the [manifest.KindProject] resource implementation.
type ProjectResource struct {
	client *sdkClient
}

// Metadata implements [resource.Resource.Metadata] function.
func (s *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements [resource.Resource.Schema] function.
func (s *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	description := "[Project configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#project)"
	resp.Schema = schema.Schema{
		MarkdownDescription: description,
		Description:         description,
		Attributes: map[string]schema.Attribute{
			"name": func() schema.StringAttribute {
				attr := metadataNameAttr()
				return addProjectResourceNameChangeWarning(attr)
			}(),
			"display_name": metadataDisplayNameAttr(),
			"description":  specDescriptionAttr(),
			"annotations":  metadataAnnotationsAttr(),
		},
		Blocks: map[string]schema.Block{
			"label": metadataLabelsBlock(),
		},
	}
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateRequest and new state values set on the CreateResponse.
func (s *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(s.applyResource(ctx, req.Plan, &resp.State)...)
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (s *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ProjectResourceModel
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &model.Name)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("label"), &model.Labels)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Project after update to fetch the computed fields.
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
func (s *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(s.applyResource(ctx, req.Plan, &resp.State)...)
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (s *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := s.client.DeleteObject(ctx, manifest.KindProject, model.Name, "")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState is called when the provider must import the resource's state.
func (s *ProjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *ProjectResource) Configure(
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

func (s *ProjectResource) applyResource(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model ProjectResourceModel
	diagnostics := plan.Get(ctx, &model)
	if diagnostics.HasError() {
		return diagnostics
	}

	project := model.ToManifest()
	diagnostics.Append(s.client.ApplyObject(ctx, project)...)
	if diagnostics.HasError() {
		return diagnostics
	}

	// Read the Project after creation to fetch the computed fields.
	appliedModel, diagnostics := s.readResource(ctx, model)
	if diagnostics.HasError() {
		return diagnostics
	}
	// Save data into Terraform state
	diagnostics.Append(state.Set(ctx, appliedModel)...)
	return diagnostics
}

// readResource reads the current state of the resource from the Nobl9 API.
func (s *ProjectResource) readResource(
	ctx context.Context,
	model ProjectResourceModel,
) (*ProjectResourceModel, diag.Diagnostics) {
	project, diagnostics := s.client.GetProject(ctx, model.Name)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	updatedModel := newProjectResourceConfigFromManifest(project)
	// Sort Labels.
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
	return updatedModel, diagnostics
}

// nolint: lll
func addProjectResourceNameChangeWarning(attr schema.StringAttribute) schema.StringAttribute {
	changeDescription := "If the value of 'name' attribute changes," +
		" Nobl9 API will remove all objects associated with this Project."
	attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.Diagnostics.AddWarning(
				"Changing the Project name results in removal of all associated objects and their data.",
				"When the Project name is changed, the Project object is removed and then recreated with a new name."+
					" When the Project is removed, all associated objects, including SLOs, and their data are removed."+
					" Currently, there's no way to transfer resources between Projects without losing data.",
			)
		},
		changeDescription,
		changeDescription,
	))
	return attr
}
