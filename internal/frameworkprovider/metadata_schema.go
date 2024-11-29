package frameworkprovider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//
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
		Required: true,
		MarkdownDescription: fmt.Sprintf(
			"Name of the Nobl9 project the resource sits in, %s.",
			dnsRFC1123NamingConventionNotice,
		),
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
