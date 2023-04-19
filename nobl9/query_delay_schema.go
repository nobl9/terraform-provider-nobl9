package nobl9

import (
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	n9api "github.com/nobl9/nobl9-go"
)

const queryDelayConfigKey = "query_delay"

func schemaQueryDelay() *schema.Schema {
	durationSchema := map[string]*schema.Schema{
		"unit": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Must be one of Minute or Second.",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"Minute", "Second"}, false),
			),
		},
		"value": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Must be an integer greater than or equal to 0.",
		},
	}

	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "[Query delay configuration documentation](https://docs.nobl9.com/Features/query-delay). Computed if not provided.",
		MinItems:    1,
		MaxItems:    1,
		Elem:        &schema.Resource{Schema: durationSchema},
	}
}

func marshalQueryDelay(d *schema.ResourceData) *n9api.QueryDelayDuration {
	queryDelay := d.Get(queryDelayConfigKey).(*schema.Set)
	if queryDelay.Len() > 0 {
		qd := queryDelay.List()[0].(map[string]interface{})
		return &n9api.QueryDelayDuration{
			Unit:  qd["unit"].(string),
			Value: json.Number(strconv.Itoa(qd["value"].(int))),
		}
	}
	return nil
}

func unmarshalQueryDelay(d *schema.ResourceData, qd *n9api.QueryDelayDuration) (diags diag.Diagnostics) {
	if qd == nil {
		return
	}
	config := map[string]interface{}{
		"unit":  qd.Unit,
		"value": qd.Value,
	}
	err := d.Set(queryDelayConfigKey, schema.NewSet(oneElementSet, []interface{}{config}))
	diags = appendError(diags, err)
	return
}
