package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

// ExampleResourceConfig describes the [ServiceResource] data model.
type ServiceResourceModel struct {
	Name        string                      `tfsdk:"name"`
	DisplayName types.String                `tfsdk:"display_name"`
	Project     string                      `tfsdk:"project"`
	Description types.String                `tfsdk:"description"`
	Annotations map[string]string           `tfsdk:"annotations"`
	Labels      Labels                      `tfsdk:"label"`
	Status      *ServiceResourceStatusModel `tfsdk:"status"`
}

type ServiceResourceStatusModel struct {
	SLOCount int `tfsdk:"slo_count"`
}

func newServiceResourceConfigFromManifest(svc v1alphaService.Service) *ServiceResourceModel {
	var status *ServiceResourceStatusModel
	if svc.Status != nil {
		status = &ServiceResourceStatusModel{
			SLOCount: svc.Status.SloCount,
		}
	}
	return &ServiceResourceModel{
		Name:        svc.Metadata.Name,
		DisplayName: types.StringValue(svc.Metadata.DisplayName),
		Project:     svc.Metadata.Project,
		Description: types.StringValue(svc.Spec.Description),
		Annotations: svc.Metadata.Annotations,
		Labels:      newLabelsFromManifest(svc.Metadata.Labels),
		Status:      status,
	}
}

func (s ServiceResourceModel) ToManifest() v1alphaService.Service {
	svc := v1alphaService.New(
		v1alphaService.Metadata{
			Name:        s.Name,
			DisplayName: s.DisplayName.ValueString(),
			Project:     s.Project,
			Annotations: s.Annotations,
			Labels:      s.Labels.ToManifest(),
		},
		v1alphaService.Spec{
			Description: s.Description.ValueString(),
		},
	)
	if s.Status != nil {
		svc.Status = &v1alphaService.Status{
			SloCount: s.Status.SLOCount,
		}
	}
	return svc
}
