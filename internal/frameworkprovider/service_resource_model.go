package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

// ServiceResourceModel describes the [ServiceResource] data model.
type ServiceResourceModel struct {
	Name             string                `tfsdk:"name"`
	DisplayName      types.String          `tfsdk:"display_name"`
	Project          string                `tfsdk:"project"`
	Description      types.String          `tfsdk:"description"`
	Annotations      map[string]string     `tfsdk:"annotations"`
	Labels           Labels                `tfsdk:"label"`
	ResponsibleUsers ResponsibleUsersModel `tfsdk:"responsible_user"`
	ReviewCycle      *ReviewCycleModel     `tfsdk:"review_cycle"`
	Status           types.Object          `tfsdk:"status"`
}

// ResponsibleUserModel represents [v1alphaService.ResponsibleUser].
type ResponsibleUserModel struct {
	Id types.String `tfsdk:"id"`
}

type ResponsibleUsersModel []ResponsibleUserModel

func (r ResponsibleUsersModel) ToManifest() []v1alphaService.ResponsibleUser {
	responsibleUsersManifest := make([]v1alphaService.ResponsibleUser, 0, len(r))
	for _, model := range r {
		responsibleUsersManifest = append(
			responsibleUsersManifest,
			v1alphaService.ResponsibleUser{ID: model.Id.ValueString()},
		)
	}

	return responsibleUsersManifest
}

// sortResponsibleUsers sorts the API returned list based on the user-defined list as a reference for sorting order.
func sortResponsibleUsers(userDefinedResponsibleUsers, apiReturnedList ResponsibleUsersModel) ResponsibleUsersModel {
	return sortListBasedOnReferenceList(
		apiReturnedList,
		userDefinedResponsibleUsers,
		func(a, b ResponsibleUserModel) bool {
			return a.Id == b.Id
		},
	)
}

type ReviewCycleModel struct {
	RRule     types.String `tfsdk:"rrule"`
	StartTime types.String `tfsdk:"start_time"`
	TimeZone  types.String `tfsdk:"time_zone"`
}

func (r ReviewCycleModel) ToManifest() *v1alphaService.ReviewCycle {
	return &v1alphaService.ReviewCycle{
		RRule:     r.RRule.ValueString(),
		StartTime: r.StartTime.ValueString(),
		TimeZone:  r.TimeZone.ValueString(),
	}
}

type ServiceResourceStatusModel struct {
	ReviewCycle ServiceResourceStatusReviewCycleModel `tfsdk:"review_cycle"`
	SLOCount    types.Int64                           `tfsdk:"slo_count"`
}

type ServiceResourceStatusReviewCycleModel struct {
	Next types.String `tfsdk:"next"`
}

func newServiceResourceConfigFromManifest(
	ctx context.Context,
	svc v1alphaService.Service,
) (*ServiceResourceModel, diag.Diagnostics) {
	var status types.Object
	statusType, diags := serviceResourceSchema.TypeAtPath(ctx, path.Root("status"))
	if diags.HasError() {
		return nil, diags
	}
	statusAttrs := statusType.(basetypes.ObjectType).AttrTypes
	if svc.Status != nil {
		statusModel := ServiceResourceStatusModel{
			SLOCount: types.Int64Value(int64(svc.Status.SloCount)),
		}
		if svc.Status.ReviewCycle != nil {
			statusModel.ReviewCycle = ServiceResourceStatusReviewCycleModel{
				Next: stringValue(svc.Status.ReviewCycle.Next),
			}
		}
		v, diags := types.ObjectValueFrom(ctx, statusAttrs, statusModel)
		if diags.HasError() {
			return nil, diags
		}
		status = v
	} else {
		status = types.ObjectNull(statusAttrs)
	}
	return &ServiceResourceModel{
		Name:             svc.Metadata.Name,
		DisplayName:      stringValue(svc.Metadata.DisplayName),
		Project:          svc.Metadata.Project,
		Description:      stringValue(svc.Spec.Description),
		Annotations:      svc.Metadata.Annotations,
		Labels:           newLabelsFromManifest(svc.Metadata.Labels),
		ResponsibleUsers: newResponsibleUsersFromManifest(svc.Spec.ResponsibleUsers),
		ReviewCycle:      newReviewCycleFromManifest(svc.Spec.ReviewCycle),
		Status:           status,
	}, nil
}

func newResponsibleUsersFromManifest(users []v1alphaService.ResponsibleUser) []ResponsibleUserModel {
	responsibleUsersModel := make([]ResponsibleUserModel, 0, len(users))
	for _, user := range users {
		responsibleUsersModel = append(responsibleUsersModel, ResponsibleUserModel{Id: stringValue(user.ID)})
	}

	return responsibleUsersModel
}

func newReviewCycleFromManifest(cycle *v1alphaService.ReviewCycle) *ReviewCycleModel {
	if cycle == nil {
		return nil
	}

	return &ReviewCycleModel{
		RRule:     stringValue(cycle.RRule),
		StartTime: stringValue(cycle.StartTime),
		TimeZone:  stringValue(cycle.TimeZone),
	}
}

func (s ServiceResourceModel) ToManifest() v1alphaService.Service {
	return v1alphaService.New(
		v1alphaService.Metadata{
			Name:        s.Name,
			DisplayName: s.DisplayName.ValueString(),
			Project:     s.Project,
			Annotations: s.Annotations,
			Labels:      s.Labels.ToManifest(),
		},
		v1alphaService.Spec{
			Description:      s.Description.ValueString(),
			ResponsibleUsers: s.ResponsibleUsers.ToManifest(),
			ReviewCycle:      getReviewCycleManifest(s.ReviewCycle),
		},
	)
}

func getReviewCycleManifest(cycle *ReviewCycleModel) *v1alphaService.ReviewCycle {
	if cycle == nil {
		return nil
	}
	return cycle.ToManifest()
}
