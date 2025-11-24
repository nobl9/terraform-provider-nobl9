package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

// ServiceResourceModel describes the [ServiceResource] data model.
type ServiceResourceModel struct {
	Name             string                 `tfsdk:"name"`
	DisplayName      types.String           `tfsdk:"display_name"`
	Project          string                 `tfsdk:"project"`
	Description      types.String           `tfsdk:"description"`
	Annotations      map[string]string      `tfsdk:"annotations"`
	Labels           Labels                 `tfsdk:"label"`
	ResponsibleUsers []ResponsibleUserModel `tfsdk:"responsible_users"`
	ReviewCycle      *ReviewCycleModel      `tfsdk:"review_cycle"`
}

// ResponsibleUserModel represents [v1alphaService.ResponsibleUser].
type ResponsibleUserModel struct {
	ID types.String `tfsdk:"id"`
}

func (r ResponsibleUserModel) ToManifest() v1alphaService.ResponsibleUser {
	return v1alphaService.ResponsibleUser{ID: r.ID.ValueString()}
}

// sortResponsibleUsers sorts the API returned list based on the user-defined list as a reference for sorting order.
func sortResponsibleUsers(userDefinedResponsibleUsers, apiReturnedList []ResponsibleUserModel) []ResponsibleUserModel {
	// Preserve nil when the user-defined list is nil and API returns an empty list.
	// This ensures consistency between null and empty list in Terraform state.
	if userDefinedResponsibleUsers == nil && len(apiReturnedList) == 0 {
		return nil
	}
	return sortListBasedOnReferenceList(
		apiReturnedList,
		userDefinedResponsibleUsers,
		func(a, b ResponsibleUserModel) bool {
			return a.ID == b.ID
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

func newServiceResourceConfigFromManifest(svc v1alphaService.Service) *ServiceResourceModel {
	return &ServiceResourceModel{
		Name:             svc.Metadata.Name,
		DisplayName:      stringValue(svc.Metadata.DisplayName),
		Project:          svc.Metadata.Project,
		Description:      stringValue(svc.Spec.Description),
		Annotations:      svc.Metadata.Annotations,
		Labels:           newLabelsFromManifest(svc.Metadata.Labels),
		ResponsibleUsers: newResponsibleUsersFromManifest(svc.Spec.ResponsibleUsers),
		ReviewCycle:      newReviewCycleFromManifest(svc.Spec.ReviewCycle),
	}
}

func newResponsibleUsersFromManifest(users []v1alphaService.ResponsibleUser) []ResponsibleUserModel {
	responsibleUsersModel := make([]ResponsibleUserModel, 0, len(users))
	for _, user := range users {
		responsibleUsersModel = append(responsibleUsersModel, ResponsibleUserModel{ID: stringValue(user.ID)})
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
	var responsibleUsersManifest []v1alphaService.ResponsibleUser
	if s.ResponsibleUsers != nil {
		responsibleUsersManifest = make([]v1alphaService.ResponsibleUser, 0, len(s.ResponsibleUsers))
		for _, model := range s.ResponsibleUsers {
			responsibleUsersManifest = append(responsibleUsersManifest, model.ToManifest())
		}
	}

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
			ResponsibleUsers: responsibleUsersManifest,
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
