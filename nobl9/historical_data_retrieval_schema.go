package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const historicalDataRetrievalConfigKey = "historical_data_retrieval"

func getHistoricalDataRetrievalSchema() map[string]*schema.Schema {
	durationSchema := map[string]*schema.Schema{
		"unit": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Must be one of Minute, Hour, or Day.",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{"Minute", "Hour", "Day"}, false),
			),
		},
		"value": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Must be an integer greater than or equal to 0.",
		},
	}

	return map[string]*schema.Schema{
		historicalDataRetrievalConfigKey: {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Description: "[Replay configuration documentation](https://docs.nobl9.com/replay)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"default_duration": {
						Type:        schema.TypeList,
						Required:    true,
						Description: "Used by default for any SLOs connected to this data source.",
						Elem:        &schema.Resource{Schema: durationSchema},
					},
					"max_duration": {
						Type:        schema.TypeList,
						Required:    true,
						Description: "Defines the maximum period for which data can be retrieved.",
						Elem:        &schema.Resource{Schema: durationSchema},
					},
				},
			},
		},
	}
}

func setHistoricalDataRetrievalSchema(s map[string]*schema.Schema) {
	s[historicalDataRetrievalConfigKey] = getHistoricalDataRetrievalSchema()[historicalDataRetrievalConfigKey]
}

func marshalHistoricalDataRetrieval(d *schema.ResourceData) *v1alpha.HistoricalDataRetrieval {
	hData, ok := d.GetOk(historicalDataRetrievalConfigKey)
	if !ok {
		return nil
	}
	historicalDataRetrieval := hData.([]interface{})[0].(map[string]interface{})
	defaultDuration := historicalDataRetrieval["default_duration"].([]interface{})[0].(map[string]interface{})
	maxDuration := historicalDataRetrieval["max_duration"].([]interface{})[0].(map[string]interface{})

	valueDefaultDuration := defaultDuration["value"].(int)
	valueMaxDuration := maxDuration["value"].(int)
	return &v1alpha.HistoricalDataRetrieval{
		DefaultDuration: v1alpha.HistoricalRetrievalDuration{
			Value: &valueDefaultDuration,
			Unit:  v1alpha.HistoricalRetrievalDurationUnit(defaultDuration["unit"].(string)),
		},
		MaxDuration: v1alpha.HistoricalRetrievalDuration{
			Value: &valueMaxDuration,
			Unit:  v1alpha.HistoricalRetrievalDurationUnit(maxDuration["unit"].(string)),
		},
	}
}

func unmarshalHistoricalDataRetrieval(
	d *schema.ResourceData,
	h *v1alpha.HistoricalDataRetrieval,
) (diags diag.Diagnostics) {
	if h == nil {
		return
	}
	config := map[string]interface{}{
		"default_duration": []interface{}{
			map[string]interface{}{
				"unit":  h.DefaultDuration.Unit,
				"value": h.DefaultDuration.Value,
			},
		},
		"max_duration": []interface{}{
			map[string]interface{}{
				"unit":  h.MaxDuration.Unit,
				"value": h.MaxDuration.Value,
			},
		},
	}
	set(d, historicalDataRetrievalConfigKey, []interface{}{config}, &diags)

	return
}
