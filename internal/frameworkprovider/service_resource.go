package frameworkprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nobl9/nobl9-go/manifest"
)

// Ensure [ServiceResource] fully satisfies framework interfaces.
var _ resource.Resource = &ServiceResource{}
var _ resource.ResourceWithImportState = &ServiceResource{}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

// ServiceResource defines the [v1alpha.Service] resource implementation.
type ServiceResource struct {
	client *sdkClient
}

// Metadata implements [resource.Resource.Metadata] function.
func (s *ServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

// Schema implements [resource.Resource.Schema] function.
func (s *ServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example resource",
		Attributes: map[string]schema.Attribute{
			//"id": schema.StringAttribute{
			//	Computed:    true,
			//	Description: "The ID of this resource.",
			//},
			"name":         metadataNameAttr(),
			"display_name": metadataDisplayNameAttr(),
			"project":      metadataProjectAttr(),
			"description":  specDescriptionAttr(),
			"annotations":  metadataAnnotationsAttr(),
			"status": schema.ObjectAttribute{
				Computed:       true,
				Optional:       true,
				Description:    "Status of created service.",
				AttributeTypes: serviceStatusTypes,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"label": metadataLabelsBlock(),
		},
		Description: "[Service configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#service)",
	}
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateRequest and new state values set on the CreateResponse.
func (s *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(s.applyResource(ctx, req.Plan, &resp.State)...)
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (s *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model ServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the Service after update to fetch the computed fields.
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
func (s *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(s.applyResource(ctx, req.Plan, &resp.State)...)
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (s *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model ServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := s.client.DeleteObject(ctx, manifest.KindService, model.Name, model.Project)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ImportState is called when the provider must import the resource's state.
func (s *ServiceResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID to be in the format '<project-name>/<service-name>'.",
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("project"), parts[0])
	resp.State.SetAttribute(ctx, path.Root("name"), parts[1])
}

func (s *ServiceResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdkClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdkClient, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	s.client = client
}

func (s *ServiceResource) applyResource(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model ServiceResourceModel
	diagnostics := plan.Get(ctx, &model)
	if diagnostics.HasError() {
		return diagnostics
	}

	service := model.ToManifest(ctx)
	diagnostics.Append(s.client.ApplyObject(ctx, service)...)
	if diagnostics.HasError() {
		return diagnostics
	}

	// Read the Service after creation to fetch the computed fields.
	appliedModel, diagnostics := s.readResource(ctx, model)
	if diagnostics.HasError() {
		return diagnostics
	}
	// Save data into Terraform state
	diagnostics.Append(state.Set(ctx, appliedModel)...)
	return diagnostics
}

// readResource reads the current state of the resource from the Nobl9 API.
func (s *ServiceResource) readResource(
	ctx context.Context,
	model ServiceResourceModel,
) (*ServiceResourceModel, diag.Diagnostics) {
	service, diagnostics := s.client.GetService(ctx, model.Name, model.Project)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	updatedModel, diags := newServiceResourceConfigFromManifest(ctx, service)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return nil, diagnostics
	}
	// SORT LABELS.
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
	return updatedModel, diagnostics
}
