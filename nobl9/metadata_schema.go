package nobl9

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	fieldLabel       = "label"
	fieldLabelKey    = "key"
	fieldLabelValues = "values"
)

//nolint:lll
func schemaName() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
	}
}

func schemaDisplayName() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "User-friendly display name of the resource.",
	}
}

//nolint:unused,deadcode
func schemaLabels() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeList,
		Optional:         true,
		Description:      "[Labels](https://docs.nobl9.com/features/labels/) containing a single key and a list of values.",
		DiffSuppressFunc: diffSuppressLabels,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				fieldLabelKey: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateNotEmptyString(fieldLabelKey),
					Description:  "A key for the label, unique within the associated resource.",
				},
				fieldLabelValues: {
					Type:        schema.TypeList,
					Required:    true,
					MinItems:    1,
					Description: "A list of unique values for a single key.",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateNotEmptyString(fieldLabelValues),
					},
				},
			},
		},
	}
}

func schemaAnnotations() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Optional:    true,
		Description: "[Metadata annotations](https://docs.nobl9.com/features/labels/#metadata-annotations) attached to the resource.",
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

type Data interface {
	Get(key string) any
	GetOk(key string) (any, bool)
}

func validateNotEmptyString(variableName string) func(interface{}, string) ([]string, []error) {
	return func(valueRaw interface{}, _ string) ([]string, []error) {
		if valueRaw.(string) == "" {
			return nil, []error{fmt.Errorf("%s must not be empty", variableName)}
		}
		return nil, nil
	}
}

func diffSuppressLabels(fieldPath, oldValueStr, newValueStr string, d *schema.ResourceData) bool {
	fieldPathSegments := strings.Split(fieldPath, ".")
	if len(fieldPathSegments) > 1 {
		fieldName := fieldPathSegments[len(fieldPathSegments)-1]
		if fieldName == fieldLabelKey {
			// Terraform's GetChange function will fail to notice if user reapplied the resource
			// with all the labels removed from the file.
			// This is the situation in which one of the values in the label's schema is set and the other one isn't.
			if exactlyOneStringEmpty(oldValueStr, newValueStr) {
				return false
			}
		}
	}

	// the N9 API will return the labels in alphabetical order for keys and values.
	// Users should be able to declare label keys and values in any order
	// and changing order should force recreating the resource.
	// In order to achieve that, we're flattening the initial label struct to 2D map
	// and check if the label values inside that 2D map are deeply equal.
	// A simple reflect.DeepEqual change is not enough for the whole 2D map
	// because it omits the values order inside the array.
	// ---------------------------------
	// Example of (deeply) equal labels:
	//   label {
	//    key    = "team"
	//    values = ["sapphire", "green"]
	//  }
	//  label {
	//    key    = "team"
	//    values = ["green", "sapphire"]
	//  }
	oldValue, newValue := d.GetChange(fieldLabel)
	labelsOld := oldValue.([]interface{})
	labelsNew := newValue.([]interface{})
	if len(labelsOld) != len(labelsNew) {
		return false
	}

	oldMap := transformLabelsTo2DMap(labelsOld)
	newMap := transformLabelsTo2DMap(labelsNew)

	isDeepEqual := true
	for labelKey := range newMap {
		if _, exist := oldMap[labelKey][fieldLabelValues]; !exist {
			return false
		}

		var oldValues = oldMap[labelKey][fieldLabelValues].([]interface{})
		var newValues = newMap[labelKey][fieldLabelValues].([]interface{})

		sort.Slice(oldValues, func(i, j int) bool {
			return oldValues[i].(string) < oldValues[j].(string)
		})
		sort.Slice(newValues, func(i, j int) bool {
			return newValues[i].(string) < newValues[j].(string)
		})

		if !reflect.DeepEqual(oldValues, newValues) {
			isDeepEqual = false
		}
	}

	return isDeepEqual
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
		Description: "Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)."}
}

func schemaDescription() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Optional description of the resource. Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.",
	}
}

func getMarshaledLabels(d Data) (v1alpha.Labels, diag.Diagnostics) {
	var labels []interface{}
	if labelsData := d.Get("label"); labelsData != nil {
		labels = labelsData.([]interface{})
	}
	return marshalLabels(labels)
}

func getMarshaledAnnotations(d Data) v1alpha.MetadataAnnotations {
	rawAnnotations := d.Get("annotations").(map[string]interface{})
	annotations := make(map[string]string, len(rawAnnotations))

	for k, v := range rawAnnotations {
		annotations[k] = v.(string)
	}

	return annotations
}

func marshalLabels(labels []interface{}) (v1alpha.Labels, diag.Diagnostics) {
	var diags diag.Diagnostics
	labelsResult := make(v1alpha.Labels, len(labels))

labelsLoop:
	for _, labelRaw := range labels {
		labelMap := labelRaw.(map[string]interface{})

		labelKey := labelMap["key"].(string)
		if labelKey == "" {
			// This continue is needed because a label with empty key will be applied
			// as a result of deleting all labels in .tf file and reapplying it.
			// This does not break the validation because of the validation schema of label resource.
			continue labelsLoop
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
			if labelValueRaw.(string) == "" {
				// This continue is needed because a label with empty value will be applied
				// as a result of deleting all labels in .tf file and reapplying it.
				// This does not break the validation because of the validation schema of label resource.
				continue labelsLoop
			}
			labelValuesStr[i] = labelValueRaw.(string)
		}

		labelsResult[labelKey] = labelValuesStr
	}

	return labelsResult, diags
}

func unmarshalLabels(labelsRaw v1alpha.Labels) interface{} {
	resultLabels := make([]map[string]interface{}, 0)

	for labelKey, labelValuesRaw := range labelsRaw {
		var labelValuesStr []string
		labelValuesStr = append(labelValuesStr, labelValuesRaw...)
		labelKeyWithValues := make(map[string]interface{})
		labelKeyWithValues["key"] = labelKey
		labelKeyWithValues["values"] = labelValuesStr

		resultLabels = append(resultLabels, labelKeyWithValues)
	}

	return resultLabels
}

// oneElementSet implements schema.SchemaSetFunc and created only one element set.
// Never use it for sets with more elements as new elements will override the old ones.
func oneElementSet(_ interface{}) int {
	return 0
}

func set(d *schema.ResourceData, key string, value interface{}, diags *diag.Diagnostics) {
	appendError(*diags, d.Set(key, value))
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

func diagsToSingleError(diags diag.Diagnostics) error {
	if len(diags) == 0 {
		return nil
	}

	var errsStrings []string
	for _, d := range diags {
		errsStrings = append(errsStrings, fmt.Sprintf("%s: %s", d.Summary, d.Detail))
	}
	combinedErrs := strings.Join(errsStrings, "; ")
	return fmt.Errorf("validation failed: %s", combinedErrs)
}

func formatErrorsAsSingleError(errs []error) error {
	var errsStrings []string
	for _, err := range errs {
		errsStrings = append(errsStrings, err.Error())
	}
	combinedErrs := strings.Join(errsStrings, "; ")
	return fmt.Errorf("validation failed: %s", strings.Trim(combinedErrs, "[]"))
}

func toStringSlice(in []interface{}) []string {
	ret := make([]string, len(in))
	for i, v := range in {
		ret[i] = v.(string)
	}
	return ret
}
