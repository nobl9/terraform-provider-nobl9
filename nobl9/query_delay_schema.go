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
		Type:        schema.TypeList,
		Optional:    true,
		Description: "[Query delay configuration documentation](https://docs.nobl9.com/Features/query-delay)",
		MinItems:    1,
		MaxItems:    1,
		Elem:        &schema.Resource{Schema: durationSchema},
	}
}

func setQueryDelaySchema(s map[string]*schema.Schema) {
	s[queryDelayConfigKey] = schemaQueryDelay()
}

func marshalQueryDelay(d *schema.ResourceData) *n9api.QueryDelayDuration {
	hData, ok := d.GetOk(queryDelayConfigKey)
	if !ok {
		return nil
	}
	queryDelay := hData.([]interface{})[0].(map[string]interface{})

	return &n9api.QueryDelayDuration{
		Unit:  queryDelay["unit"].(string),
		Value: json.Number(strconv.Itoa(queryDelay["value"].(int))),
	}
}

func unmarshalQueryDelay(d *schema.ResourceData, qd *n9api.QueryDelayDuration) (diags diag.Diagnostics) {
	if qd == nil {
		return
	}
	config := map[string]interface{}{
		"unit":  qd.Unit,
		"value": qd.Value,
	}
	set(d, queryDelayConfigKey, []interface{}{config}, &diags)

	return
}
