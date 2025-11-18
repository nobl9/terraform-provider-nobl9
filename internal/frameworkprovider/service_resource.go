package frameworkprovider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nobl9/nobl9-go/manifest"
)

// Ensure [ServiceResource] fully satisfies framework interfaces.
var (
	_ resource.Resource                = &ServiceResource{}
	_ resource.ResourceWithImportState = &ServiceResource{}
	_ resource.ResourceWithConfigure   = &ServiceResource{}
	_ resource.ResourceWithModifyPlan  = &ServiceResource{}
)

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

// ServiceResource defines the [manifest.KindService] resource implementation.
type ServiceResource struct {
	client *sdkClient
}

// Metadata implements [resource.Resource.Metadata] function.
func (s *ServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

// Schema implements [resource.Resource.Schema] function.
func (s *ServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = serviceResourceSchema
}

var serviceResourceSchema = func() schema.Schema {
	description := "[Service configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#service)"
	return schema.Schema{
		MarkdownDescription: description,
		Description:         description,
		Attributes: map[string]schema.Attribute{
			"name": func() schema.StringAttribute {
				attr := metadataNameAttr()
				return addServiceResourceNameChangeWarning(attr)
			}(),
			"display_name": metadataDisplayNameAttr(),
			"project": func() schema.StringAttribute {
				attr := metadataProjectAttr()
				return addServiceResourceProjectChangeWarning(attr)
			}(),
			"description": specDescriptionAttr(),
			"annotations": metadataAnnotationsAttr(),
			"status": schema.ObjectAttribute{
				Computed:    true,
				Description: "Status of created service.",
				AttributeTypes: map[string]attr.Type{
					"slo_count": types.Int64Type,
					"review_cycle": types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"next": types.StringType,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"label":            metadataLabelsBlock(),
			"responsible_user": serviceResponsibleUserBlock(),
			"review_cycle":     serviceReviewCycleBlock(),
		},
	}
}()

func serviceReviewCycleBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "Configuration for service review cycle.",
		Attributes: map[string]schema.Attribute{
			"rrule": schema.StringAttribute{
				Required:    true,
				Description: "Recurring rule in RFC 5545 RRULE format defining when a review should occur.",
			},
			"start_time": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Start time (inclusive) for the first occurrence defined by the rrule. RFC3339 time without time zone (e.g. 2024-01-02T15:04:05).",
				Description:         "Start time for the first occurrence defined by the rrule.",
			},
			"time_zone": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Time zone identifier (IANA) used to interpret start_time and rrule times (e.g. Europe/Warsaw).",
				Description:         "Time zone identifier used to interpret start_time and rrule times.",
			},
		},
	}
}

func serviceResponsibleUserBlock() schema.Block {
	return schema.ListNestedBlock{
		Description: "List of users responsible for the service.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:    true,
					Description: "ID of the responsible user.",
				},
			},
		},
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
	model, diags := s.readResource(ctx, ServiceResourceModel{Name: parts[1], Project: parts[0]})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
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
		addInvalidSDKClientTypeDiag(&resp.Diagnostics, req.ProviderData)
		return
	}
	s.client = client
}

// ModifyPlan implements [resource.ResourceWithModifyPlan.ModifyPlan] function.
func (s *ServiceResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	var plan *ServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan == nil {
		return
	}
	resp.Diagnostics.Append(s.client.DryRunApplyObject(ctx, plan.ToManifest())...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *ServiceResource) applyResource(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var model ServiceResourceModel
	diagnostics := plan.Get(ctx, &model)
	if diagnostics.HasError() {
		return diagnostics
	}

	service := model.ToManifest()
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
	// Sort Labels.
	updatedModel.Labels = sortLabels(model.Labels, updatedModel.Labels)
	updatedModel.ResponsibleUsers = sortResponsibleUsers(model.ResponsibleUsers, updatedModel.ResponsibleUsers)
	return updatedModel, diagnostics
}

// nolint: lll
func addServiceResourceNameChangeWarning(attr schema.StringAttribute) schema.StringAttribute {
	changeDescription := "If the value of 'name' attribute changes," +
		" Nobl9 API will remove all SLOs associated with this Service."
	attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.Diagnostics.AddWarning(
				"Changing Service name results in removal of all associated SLOs and their data.",
				"When Service name is changed the Service object is removed and then recreated with the new name."+
					" When the Service is removed, all associated SLOs and their data are also removed."+
					" If you wish to change the Service name without removing the SLOs, please create a new Service with the desired name first,"+
					" change the Service name reference in the associated SLOs to the new Service and only then remove the old Service.",
			)
		},
		changeDescription,
		changeDescription,
	))
	return attr
}

// nolint: lll
func addServiceResourceProjectChangeWarning(attr schema.StringAttribute) schema.StringAttribute {
	changeDescription := "If the value of 'project' attribute changes, Nobl9 API will remove all SLOs associated with this Service."
	attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.Diagnostics.AddWarning(
				"Changing Service project results in removal of all associated SLOs and their data.",
				"When Service project is changed the Service object is removed and then recreated inside the new project."+
					" When the Service is removed, all associated SLOs and their data are also removed.",
			)
		},
		changeDescription,
		changeDescription,
	))
	return attr
}
