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
					"triggered_by_slo_creation": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Description: "(Block List) Defines the timeframe Nobl9 will reach back to fetch historical " +
							"data for SLOs based on this data source once they're created ",
						Elem: &schema.Resource{Schema: durationSchema},
					},
					"triggered_by_slo_edit": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Description: "(Block List) Defines the timeframe Nobl9 will reach back to fetch historical " +
							"data for SLOs based on this data source after modifying their budget-sensitive fields ",
						Elem: &schema.Resource{Schema: durationSchema},
					},
				},
			},
		},
	}
}

func setHistoricalDataRetrievalSchema(s map[string]*schema.Schema) {
	s[historicalDataRetrievalConfigKey] = getHistoricalDataRetrievalSchema()[historicalDataRetrievalConfigKey]
}

func marshalHistoricalDataRetrieval(d resourceInterface) *v1alpha.HistoricalDataRetrieval {
	hData, ok := d.GetOk(historicalDataRetrievalConfigKey)
	if !ok {
		return nil
	}
	historicalDataRetrieval := hData.([]interface{})[0].(map[string]interface{})
	defaultDuration := historicalDataRetrieval["default_duration"].([]interface{})[0].(map[string]interface{})
	maxDuration := historicalDataRetrieval["max_duration"].([]interface{})[0].(map[string]interface{})

	valueDefaultDuration := defaultDuration["value"].(int)
	valueMaxDuration := maxDuration["value"].(int)
	v1alphaHistoricalDataRetrieval := &v1alpha.HistoricalDataRetrieval{
		DefaultDuration: v1alpha.HistoricalRetrievalDuration{
			Value: &valueDefaultDuration,
			Unit:  v1alpha.HistoricalRetrievalDurationUnit(defaultDuration["unit"].(string)),
		},
		MaxDuration: v1alpha.HistoricalRetrievalDuration{
			Value: &valueMaxDuration,
			Unit:  v1alpha.HistoricalRetrievalDurationUnit(maxDuration["unit"].(string)),
		},
	}
	if len(historicalDataRetrieval["triggered_by_slo_creation"].([]interface{})) > 0 {
		triggeredBySloCreation :=
			historicalDataRetrieval["triggered_by_slo_creation"].([]interface{})[0].(map[string]interface{})

		valueTriggeredBySloCreation := triggeredBySloCreation["value"].(int)
		v1alphaHistoricalDataRetrieval.TriggeredBySloCreation = &v1alpha.HistoricalRetrievalDuration{
			Value: &valueTriggeredBySloCreation,
			Unit:  v1alpha.HistoricalRetrievalDurationUnit(triggeredBySloCreation["unit"].(string)),
		}
	}
	if len(historicalDataRetrieval["triggered_by_slo_edit"].([]interface{})) > 0 {
		triggeredBySloEdit :=
			historicalDataRetrieval["triggered_by_slo_edit"].([]interface{})[0].(map[string]interface{})
		valueTriggeredBySloEdit := triggeredBySloEdit["value"].(int)
		v1alphaHistoricalDataRetrieval.TriggeredBySloEdit = &v1alpha.HistoricalRetrievalDuration{
			Value: &valueTriggeredBySloEdit,
			Unit:  v1alpha.HistoricalRetrievalDurationUnit(triggeredBySloEdit["unit"].(string)),
		}
	}
	return v1alphaHistoricalDataRetrieval
}

func unmarshalHistoricalDataRetrieval(
	d *schema.ResourceData,
	h *v1alpha.HistoricalDataRetrieval,
) (diags diag.Diagnostics) {
	if h == nil {
		return diags
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

	if h.TriggeredBySloCreation != nil {
		config["triggered_by_slo_creation"] = []interface{}{
			map[string]interface{}{
				"unit":  h.TriggeredBySloCreation.Unit,
				"value": h.TriggeredBySloCreation.Value,
			},
		}
	} else {
		delete(config, "triggered_by_slo_creation")
	}

	if h.TriggeredBySloEdit != nil {
		config["triggered_by_slo_edit"] = []interface{}{
			map[string]interface{}{
				"unit":  h.TriggeredBySloEdit.Unit,
				"value": h.TriggeredBySloEdit.Value,
			},
		}
	} else {
		delete(config, "triggered_by_slo_edit")
	}

	set(d, historicalDataRetrievalConfigKey, []interface{}{config}, &diags)

	return diags
}
