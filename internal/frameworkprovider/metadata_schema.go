package frameworkprovider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:lll
const dnsRFC1123NamingConventionNotice = "must conform to the [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names) naming convention"

func metadataNameAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Required:            true,
		MarkdownDescription: fmt.Sprintf("Unique name of the resource, %s.", dnsRFC1123NamingConventionNotice),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func metadataDisplayNameAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: "User-friendly display name of the resource.",
	}
}

func metadataProjectAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Required:            true,
		MarkdownDescription: fmt.Sprintf("Name of the Nobl9 project the resource sits in, %s.", dnsRFC1123NamingConventionNotice),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func specDescriptionAttr() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Description: "Optional description of the resource. " +
			"Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.",
	}
}

func metadataAnnotationsAttr() *schema.MapAttribute {
	return &schema.MapAttribute{
		ElementType:         types.StringType,
		Optional:            true,
		MarkdownDescription: "[Metadata annotations](https://docs.nobl9.com/features/labels/#metadata-annotations) attached to the resource.", //nolint:lll
	}
}

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
				"values": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Description: "A list of unique values for a single key.",
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
					},
				},
			},
			// TODO: Add DiffSuppressFunc
			PlanModifiers: []planmodifier.Object{},
		},
	}
}
