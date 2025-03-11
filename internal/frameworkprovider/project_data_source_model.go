package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
)

// ProjectDataSourceModel describes the [ProjectDataSource] data model.
type ProjectDataSourceModel struct {
	Name        string       `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
}

func newProjectDataSourceModelFromManifest(project v1alphaProject.Project) *ProjectDataSourceModel {
	return &ProjectDataSourceModel{
		Name:        project.Metadata.Name,
		DisplayName: stringValue(project.Metadata.DisplayName),
		Description: stringValue(project.Spec.Description),
	}
}
