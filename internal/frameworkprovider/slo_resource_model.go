package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// SLOResourceModel describes the [SLOResource] data model.
type SLOResourceModel struct {
	Name        string            `tfsdk:"name"`
	DisplayName types.String      `tfsdk:"display_name"`
	Project     string            `tfsdk:"project"`
	Description types.String      `tfsdk:"description"`
	Annotations map[string]string `tfsdk:"annotations"`
	Labels      Labels            `tfsdk:"label"`
	Status      types.Object      `tfsdk:"status"`
}

var sloStatusTypes = map[string]attr.Type{
	"slo_count": types.Int64Type,
}

type SLOResourceStatusModel struct {
	SLOCount types.Int64 `tfsdk:"slo_count"`
}

func newSLOResourceConfigFromManifest(
	ctx context.Context,
	svc v1alphaSLO.SLO,
) (*SLOResourceModel, diag.Diagnostics) {
	var status types.Object
	if svc.Status != nil {
		v, diags := types.ObjectValueFrom(ctx, sloStatusTypes, SLOResourceStatusModel{
			// SLOCount: types.Int64Value(int64(svc.Status.SloCount)), // FIXME:
		})
		if diags.HasError() {
			return nil, diags
		}
		status = v
	} else {
		status = types.ObjectNull(sloStatusTypes)
	}
	return &SLOResourceModel{
		Name:        svc.Metadata.Name,
		DisplayName: stringValue(svc.Metadata.DisplayName),
		Project:     svc.Metadata.Project,
		Description: stringValue(svc.Spec.Description),
		Annotations: svc.Metadata.Annotations,
		Labels:      newLabelsFromManifest(svc.Metadata.Labels),
		Status:      status,
	}, nil
}

func (s SLOResourceModel) ToManifest(ctx context.Context) v1alphaSLO.SLO {
	return v1alphaSLO.New(
		v1alphaSLO.Metadata{
			Name:        s.Name,
			DisplayName: s.DisplayName.ValueString(),
			Project:     s.Project,
			Annotations: s.Annotations,
			Labels:      s.Labels.ToManifest(),
		},
		v1alphaSLO.Spec{
			Description: s.Description.ValueString(),
		},
	)
}
