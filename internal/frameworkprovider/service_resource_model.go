package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"

	"github.com/nobl9/terraform-provider-nobl9/internal/reflectionutils"
)

// ServiceResourceModel describes the [ServiceResource] data model.
type ServiceResourceModel struct {
	Name        string            `tfsdk:"name"`
	DisplayName types.String      `tfsdk:"display_name"`
	Project     string            `tfsdk:"project"`
	Description types.String      `tfsdk:"description"`
	Annotations map[string]string `tfsdk:"annotations"`
	Labels      Labels            `tfsdk:"label"`
	Status      types.Object      `tfsdk:"status"`
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
		statusModel := ServiceResourceStatusModel{
			SLOCount: types.Int64Value(int64(svc.Status.SloCount)),
		}
		v, diags := types.ObjectValueFrom(ctx, reflectionutils.GetAttributeTypes(statusModel), statusModel)
		if diags.HasError() {
			return nil, diags
		}
		status = v
	} else {
		status = types.ObjectNull(reflectionutils.GetAttributeTypes(ServiceResourceStatusModel{}))
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
			Description: s.Description.ValueString(),
		},
	)
}
