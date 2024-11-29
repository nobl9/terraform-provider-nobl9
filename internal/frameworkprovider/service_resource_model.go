package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alpha "github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

// ExampleResourceConfig describes the [ServiceResource] data config.
type ServiceResourceModel struct {
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Project     types.String `tfsdk:"project"`
	Description types.String `tfsdk:"description"`
	Annotations types.Map    `tfsdk:"annotations"`
	// Labels      types.List   `tfsdk:"label"`
}

func newServiceResourceConfigFromManifest(svc v1alphaService.Service) (*ServiceResourceModel, diag.Diagnostics) {
	var annotations types.Map
	if len(svc.Metadata.Annotations) > 0 {
		values := make(map[string]attr.Value, len(svc.Metadata.Annotations))
		var diags diag.Diagnostics
		annotations, diags = types.MapValue(types.StringType, values)
		if diags.HasError() {
			return nil, diags
		}
	}
	return &ServiceResourceModel{
		Name:        types.StringValue(svc.Metadata.Name),
		DisplayName: types.StringValue(svc.Metadata.DisplayName),
		Project:     types.StringValue(svc.Metadata.Project),
		Description: types.StringValue(svc.Spec.Description),
		Annotations: annotations,
		// Labels:      types.ListValue(),
	}, nil
}

func (s ServiceResourceModel) ToManifest() v1alphaService.Service {
	var annotations v1alpha.MetadataAnnotations
	if !s.Annotations.IsNull() {
		elements := s.Annotations.Elements()
		annotations = make(v1alpha.MetadataAnnotations, len(elements))
		for k, v := range elements {
			annotations[k] = v.(types.String).ValueString()
		}
	}
	return v1alphaService.New(
		v1alphaService.Metadata{
			Name:        s.Name.ValueString(),
			DisplayName: s.DisplayName.ValueString(),
			Project:     s.Project.ValueString(),
			Annotations: annotations,
		},
		v1alphaService.Spec{
			Description: s.Description.ValueString(),
		},
	)
}
