package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

// ExampleResourceConfig describes the [ServiceResource] data model.
type ServiceResourceModel struct {
	Name        string            `tfsdk:"name"`
	DisplayName types.String      `tfsdk:"display_name"`
	Project     string            `tfsdk:"project"`
	Description types.String      `tfsdk:"description"`
	Annotations map[string]string `tfsdk:"annotations"`
	Labels      Labels            `tfsdk:"label"`
	Status      types.Object      `tfsdk:"status"`
}

var serviceStatusTypes = map[string]attr.Type{
	"slo_count": types.Int64Type,
}

type ServiceResourceStatusModel struct {
	SLOCount types.Int64 `tfsdk:"slo_count"`
}

func newServiceResourceConfigFromManifest(
	ctx context.Context,
	svc v1alphaService.Service,
) (*ServiceResourceModel, diag.Diagnostics) {
	var status types.Object
	if svc.Status != nil {
		v, diags := types.ObjectValueFrom(ctx, serviceStatusTypes, ServiceResourceStatusModel{
			SLOCount: types.Int64Value(int64(svc.Status.SloCount)),
		})
		if diags.HasError() {
			return nil, diags
		}
		status = v
	} else {
		status = types.ObjectNull(serviceStatusTypes)
	}
	return &ServiceResourceModel{
		Name:        svc.Metadata.Name,
		DisplayName: stringValue(svc.Metadata.DisplayName),
		Project:     svc.Metadata.Project,
		Description: stringValue(svc.Spec.Description),
		Annotations: svc.Metadata.Annotations,
		Labels:      newLabelsFromManifest(svc.Metadata.Labels),
		Status:      status,
	}, nil
}

func (s ServiceResourceModel) ToManifest(ctx context.Context) v1alphaService.Service {
	return v1alphaService.New(
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
}
