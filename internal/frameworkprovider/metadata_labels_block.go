package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// Labels is a list of [LabelBlockModel].
// It is the Terraform equivalent of [v1alpha.Labels].
type Labels []LabelBlockModel

// LabelBlockModel represents a single label block definition.
// Example:
//
//	```hcl
//	label {
//	  key    = "env"
//	  values = ["prod", "dev"]
//	}
//	```
type LabelBlockModel struct {
	Key    string   `tfsdk:"key"`
	Values []string `tfsdk:"values"`
}

// newLabelsFromManifest converts [v1alpha.Labels] to [Labels].
func newLabelsFromManifest(sdkLabels v1alpha.Labels) Labels {
	labels := make(Labels, 0, len(sdkLabels))
	for key, values := range sdkLabels {
		labels = append(labels, LabelBlockModel{
			Key:    key,
			Values: values,
		})
	}
	return labels
}

// ToManifest converts [Labels] to [v1alpha.Labels].
func (l Labels) ToManifest() v1alpha.Labels {
	labels := make(v1alpha.Labels, len(l))
	for _, label := range l {
		labels[label.Key] = label.Values
	}
	return labels
}

// metadataLabelsBlock returns a nested block for metadata labels.
// Every resource which supports labels can reuse it.
func metadataLabelsBlock() *schema.ListNestedBlock {
	return &schema.ListNestedBlock{
		Description: "[Labels](https://docs.nobl9.com/features/labels/) containing a single key and a list of values.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"key": schema.StringAttribute{
					Required:    true,
					Description: "A key for the label, unique within the associated resource.",
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"values": schema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					Description: "A set of values for a single key.",
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
				},
			},
		},
	}
}

// sortLabels sorts the API returned list based on the user-defined list as a reference for sorting order.
func sortLabels(userDefinedLabels, apiReturnedList Labels) Labels {
	return sortListBasedOnReferenceList(
		apiReturnedList,
		userDefinedLabels,
		func(a, b LabelBlockModel) bool {
			return a.Key == b.Key
		},
	)
}
