package nobl9

import (
	"fmt"
	"sort"
	"strings"

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
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Additional labels for the resource",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		DiffSuppressFunc: diffSuppressListStringOrder("labels"),
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
	if labelsData := d.Get("labels"); labelsData != nil {
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

type label struct {
	key   string
	value string
}

func newLabel(labelRaw string) (*label, error) {
	labelSegments := strings.Split(labelRaw, ":")
	if len(labelSegments) != 2 {
		return nil, fmt.Errorf("wrong label format, expected \"key:value\", got \"%s\" ", labelRaw)
	}
	return &label{
		key:   labelSegments[0],
		value: labelSegments[1],
	}, nil
}

func (l label) toString() string {
	return strings.Join([]string{l.key, l.value}, ":")
}

func marshalLabels(labels []interface{}) (n9api.Labels, diag.Diagnostics) {
	var diags diag.Diagnostics

	labelsResult := make(n9api.Labels, 0)

	for _, labelRaw := range labels {
		l, err := newLabel(labelRaw.(string))
		if err != nil {
			diags = appendError(diags, fmt.Errorf("error creating new l - %w", err))
		} else {
			labelsResult[l.key] = append(labelsResult[l.key], l.value)
		}
	}

	return labelsResult, diags
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
	labelsRaw, exist := metadata["labels"]
	if !exist {
		return nil
	}

	labels := labelsRaw.(map[string]interface{})
	var res []string
	for key, valuesRaw := range labels {
		for _, value := range valuesRaw.([]interface{}) {
			l := label{
				key:   key,
				value: value.(string),
			}
			res = append(res, l.toString())
		}
	}

	return d.Set("labels", res)
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
