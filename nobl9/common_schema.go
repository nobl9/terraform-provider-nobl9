package nobl9

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const (
	apiVersion = "n9/v1alpha"
)

func schemaName() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "",
	}
}

func schemaDisplayName() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "",
	}
}

func schemaLabels() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "",
		Elem:        &schema.Schema{Type: schema.TypeString},
	}
}

func schemaProject() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "",
	}
}

func schemaDescription() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "",
	}
}
