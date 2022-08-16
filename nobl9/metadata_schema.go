package nobl9

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
	"sort"
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
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Labels containing a single key and a list of values.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "One key for the label, unique within the associated resource.",
					ValidateFunc: validateUniqueLabelKeys,
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

func diffSuppressListStringOrder(attribute string) func(
	_, _, _ string,
	d *schema.ResourceData,
) bool {
	return func(_, _, _ string, d *schema.ResourceData) bool {
		// Ignore the order of elements on alert_policy list
		oldValue, newValue := d.GetChange(attribute)
		if oldValue == nil && newValue == nil {
			return true
		}
		apOld := oldValue.([]interface{})
		apNew := newValue.([]interface{})

		sort.Slice(apOld, func(i, j int) bool {
			return apOld[i].(string) < apOld[j].(string)
		})
		sort.Slice(apNew, func(i, j int) bool {
			return apNew[i].(string) < apNew[j].(string)
		})

		return equalSlices(apOld, apNew)
	}
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

func unmarshalLabels(d *schema.ResourceData, metadata map[string]interface{}) error {
	labelsRaw, exist := metadata["labels"].([]interface{})
	if !exist {
		return nil
	}

	fmt.Println("unmarshala")

	resultLabels := make([]map[string]interface{}, len(labelsRaw))

	for i, labelRaw := range labelsRaw {
		labelMap := labelRaw.(map[string]interface{})

		resultLabels[i] = map[string]interface{}{
			"key":    labelMap["key"].(string),
			"values": labelMap["values"].([]string),
		}
	}

	fmt.Println("result Labels")

	return d.Set("label", resultLabels)
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
