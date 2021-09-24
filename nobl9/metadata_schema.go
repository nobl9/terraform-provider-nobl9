package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func schemaName() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Unique name of the resource. Must match [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
	}
}

func schemaDisplayName() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Display name of the resource.",
	}
}

func schemaLabels() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Additional labels for the resource",
		Elem:        &schema.Schema{Type: schema.TypeString},
	}
}

func schemaProject() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Name of the project the resource is in. Must match [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)."}
}

func schemaDescription() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Optional description of the resource.",
	}
}

func marshalMetadata(d *schema.ResourceData) n9api.MetadataHolder {
	return n9api.MetadataHolder{
		Metadata: n9api.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: d.Get("display_name").(string),
			Project:     d.Get("project").(string),
			// TODO Metadata should also support labels - SDK is outdated
		},
	}
}

func unmarshalMetadata(object n9api.AnyJSONObj, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	metadata := object["metadata"].(map[string]interface{})
	err := d.Set("name", metadata["name"])
	appendError(diags, err)
	err = d.Set("display_name", metadata["displayName"])
	appendError(diags, err)
	// err = d.Set("labels", metadata["labels"]) // TODO labels are not supported yet
	appendError(diags, err)
	err = d.Set("project", metadata["project"])
	appendError(diags, err)

	return diags
}

// oneElementSet implements schema.SchemaSetFunc and created only one element set.
// Never use it for sets with more elements as new elements will override the old ones.
func oneElementSet(_ interface{}) int {
	return 0
}

func appendError(d diag.Diagnostics, err error) diag.Diagnostics {
	if err != nil {
		return append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Error(),
		})
	}

	return d
}
