package frameworkprovider

import (
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

// ServiceDataSourceModel describes the [ServiceDataSource] data model.
type ServiceDataSourceModel struct {
	Name    string `tfsdk:"name"`
	Project string `tfsdk:"project"`
}

func newServiceDataSourceModelFromManifest(service v1alphaService.Service) *ServiceDataSourceModel {
	return &ServiceDataSourceModel{
		Name:    service.GetName(),
		Project: service.GetProject(),
	}
}
