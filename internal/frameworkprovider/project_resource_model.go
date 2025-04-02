package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
)

// ProjectResourceModel describes the [ProjectResource] data model.
type ProjectResourceModel struct {
	Name        string            `tfsdk:"name"`
	DisplayName types.String      `tfsdk:"display_name"`
	Description types.String      `tfsdk:"description"`
	Annotations map[string]string `tfsdk:"annotations"`
	Labels      Labels            `tfsdk:"label"`
}

func newProjectResourceConfigFromManifest(
	ctx context.Context,
	svc v1alphaProject.Project,
) (*ProjectResourceModel, diag.Diagnostics) {
	return &ProjectResourceModel{
		Name:        svc.Metadata.Name,
		DisplayName: stringValue(svc.Metadata.DisplayName),
		Description: stringValue(svc.Spec.Description),
		Annotations: svc.Metadata.Annotations,
		Labels:      newLabelsFromManifest(svc.Metadata.Labels),
	}, nil
}

func (s ProjectResourceModel) ToManifest(ctx context.Context) v1alphaProject.Project {
	return v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:        s.Name,
			DisplayName: s.DisplayName.ValueString(),
			Annotations: s.Annotations,
			Labels:      s.Labels.ToManifest(),
		},
		v1alphaProject.Spec{
			Description: s.Description.ValueString(),
		},
	)
}
