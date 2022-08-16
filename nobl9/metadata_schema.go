package nobl9

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

//nolint:lll
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

//nolint:unused,deadcode
func schemaLabels() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeList,
		Optional:         true,
		Description:      "Labels containing a single key and a list of values.",
		DiffSuppressFunc: diffSuppressLabels,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One key for the label, unique within the associated resource.",
				},
				"values": {
					Type:        schema.TypeList,
					Optional:    true,
					MinItems:    1,
					Description: "A list of unique values for a single key.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func diffSuppressLabels(_, _, _ string, d *schema.ResourceData) bool {
	// the N9 API will return the labels in alphabetical by name order, however users
	// can have them in any order.  So we want to flatten the list into a 2D map and do a DeepEqual
	// comparison to see if we have any actual changes
	oldValue, newValue := d.GetChange("label")
	labelsOld := oldValue.([]interface{})
	labelsNew := newValue.([]interface{})

	oldMap := transformLabelsTo2DMap(labelsOld)
	newMap := transformLabelsTo2DMap(labelsNew)

	return reflect.DeepEqual(oldMap, newMap)
}

func transformLabelsTo2DMap(labels []interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	for _, label := range labels {
		s := label.(map[string]interface{})
		values := make(map[string]interface{})

		values["key"] = s["key"].(string)
		values["values"] = s["values"].([]interface{})
		result[s["key"].(string)] = values
	}
	return result
}

//nolint:lll
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

func marshalMetadata(d *schema.ResourceData) (n9api.MetadataHolder, diag.Diagnostics) {
	var diags diag.Diagnostics

	var labels []interface{}
	if labelsData := d.Get("label"); labelsData != nil {
		labels = labelsData.([]interface{})
	}
	var labelsMarshalled n9api.Labels
	labelsMarshalled, diags = marshalLabels(labels)

	return n9api.MetadataHolder{
		Metadata: n9api.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: d.Get("display_name").(string),
			Project:     d.Get("project").(string),
			Labels:      labelsMarshalled,
		},
	}, diags
}

func unmarshalMetadata(object n9api.AnyJSONObj, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	metadata := object["metadata"].(map[string]interface{})
	err := d.Set("name", metadata["name"])
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata["displayName"])
	diags = appendError(diags, err)

	diags = appendError(diags, err)
	err = d.Set("project", metadata["project"])
	diags = appendError(diags, err)

	err = unmarshalLabels(d, metadata)
	diags = appendError(diags, err)

	return diags
}

func marshalLabels(labels []interface{}) (n9api.Labels, diag.Diagnostics) {
	var diags diag.Diagnostics
	labelsResult := make(n9api.Labels, 0)

	for _, labelRaw := range labels {
		labelMap := labelRaw.(map[string]interface{})

		labelKey := labelMap["key"].(string)
		if labelKey == "" {
			diags = appendError(diags, fmt.Errorf("error creating label because the key is empty"))
		}
		if _, exist := labelsResult[labelKey]; exist {
			diags = appendError(diags, fmt.Errorf(
				"duplicate label key [%s] found - expected only one occurrence of each label key",
				labelKey,
			))
		}

		labelValuesRaw := labelMap["values"].([]interface{})
		labelValuesStr := make([]string, len(labelValuesRaw))
		if len(labelValuesRaw) < 1 {
			diags = appendError(diags, fmt.Errorf("error creating label because there was no value specified"))
		}
		for i, labelValueRaw := range labelValuesRaw {
			labelValuesStr[i] = labelValueRaw.(string)
		}

		labelsResult[labelKey] = labelValuesStr
	}

	return labelsResult, diags
}

func unmarshalLabels(d *schema.ResourceData, metadata map[string]interface{}) error {
	labelsRaw, exist := metadata["labels"].(map[string]interface{})
	if !exist {
		return nil
	}

	resultLabels := make([]map[string]interface{}, 0)

	for labelKey, labelValuesRaw := range labelsRaw {
		var labelValuesStr []string
		for _, labelValueRaw := range labelValuesRaw.([]interface{}) {
			labelValuesStr = append(labelValuesStr, labelValueRaw.(string))
		}
		labelKeyWithValues := make(map[string]interface{})
		labelKeyWithValues["key"] = labelKey
		labelKeyWithValues["values"] = labelValuesStr

		resultLabels = append(resultLabels, labelKeyWithValues)
	}

	return d.Set("label", resultLabels)
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

func toStringSlice(in []interface{}) []string {
	ret := make([]string, len(in))
	for i, v := range in {
		ret[i] = v.(string)
	}
	return ret
}
