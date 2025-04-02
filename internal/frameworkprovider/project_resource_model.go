package frameworkprovider

import (
	"context"

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

func newProjectResourceConfigFromManifest(project v1alphaProject.Project) *ProjectResourceModel {
	return &ProjectResourceModel{
		Name:        project.Metadata.Name,
		DisplayName: stringValue(project.Metadata.DisplayName),
		Description: stringValue(project.Spec.Description),
		Annotations: project.Metadata.Annotations,
		Labels:      newLabelsFromManifest(project.Metadata.Labels),
	}
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
