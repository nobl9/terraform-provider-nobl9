package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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
		Computed:    true,
		Description: "[Query delay configuration documentation](https://docs.nobl9.com/Features/query-delay). Computed if not provided.",
		MinItems:    1,
		MaxItems:    1,
		Elem:        &schema.Resource{Schema: durationSchema},
	}
}

func marshalQueryDelay(d *schema.ResourceData) *v1alpha.QueryDelay {
	queryDelay := d.Get(queryDelayConfigKey).(*schema.Set)
	if queryDelay.Len() > 0 {
		qd := queryDelay.List()[0].(map[string]interface{})

		valueQueryDelayDuration := qd["value"].(int)
		return &v1alpha.QueryDelay{
			Duration: v1alpha.Duration{
				Value: &valueQueryDelayDuration,
				Unit:  v1alpha.DurationUnit(qd["unit"].(string)),
			},
		}
	}
	return nil
}

func unmarshalQueryDelay(d *schema.ResourceData, qd *v1alpha.QueryDelay) diag.Diagnostics {
	if qd == nil {
		return nil
	}
	config := map[string]interface{}{
		"value": qd.Value,
		"unit":  qd.Unit,
	}
	err := d.Set(queryDelayConfigKey, schema.NewSet(oneElementSet, []interface{}{config}))
	if err != nil {
		return appendError(diag.Diagnostics{}, err)
	}
	return nil
}
