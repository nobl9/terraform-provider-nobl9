package frameworkprovider

import (
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
)

// ProjectDataSourceModel describes the [ProjectDataSource] data model.
type ProjectDataSourceModel struct {
	Name string `tfsdk:"name"`
}

func newProjectDataSourceModelFromManifest(project v1alphaProject.Project) *ProjectDataSourceModel {
	return &ProjectDataSourceModel{
		Name: project.Metadata.Name,
	}
}
