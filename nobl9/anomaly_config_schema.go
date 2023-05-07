package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func schemaAnomalyConfig() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Configuration for Anomalies. Currently supported Anomaly Type is NoData",
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"no_data": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "Alert Policies attached to SLO",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alert_method": {
								Type:        schema.TypeList,
								Required:    true,
								Description: "Alert methods attached to Anomaly Config",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validateMaxLength("name", 63),
											Description:      "The name of the previously defined alert method.",
										},
										"project": {
											Type:     schema.TypeString,
											Required: true,
											Description: "Project name the Alert Method is in, " +
												" must conform to the naming convention from [DNS RFC1123] " +
												"(https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)." +
												" If not defined, Nobl9 returns a default value for this field.",
										},
									},
								},
								MaxItems: 5,
								MinItems: 1,
							},
						},
					},
				},
			},
		},
	}
}

func marshalAnomalyConfig(anomalyConfigRaw interface{}) *n9api.AnomalyConfig {
	anomalyConfigSet := anomalyConfigRaw.(*schema.Set)
	if anomalyConfigSet.Len() == 0 {
		return nil
	}
	anomalyConfig := anomalyConfigSet.List()[0].(map[string]interface{})

	noDataAnomalyConfigSet := anomalyConfig["no_data"].(*schema.Set)
	if noDataAnomalyConfigSet.Len() == 0 {
		return nil
	}
	noDataAnomalyConfig := noDataAnomalyConfigSet.List()[0].(map[string]interface{})
	noDataAlertMethods := noDataAnomalyConfig["alert_method"].([]interface{})
	marshalledAlertMethods := marshalAnomalyConfigAlertMethods(noDataAlertMethods)

	return &n9api.AnomalyConfig{
		NoData: &n9api.AnomalyConfigNoData{
			AlertMethods: marshalledAlertMethods,
		},
	}
}

func marshalAnomalyConfigAlertMethods(alertMethodsTF []interface{}) []n9api.AnomalyConfigAlertMethod {
	alertMethodsAPI := make([]n9api.AnomalyConfigAlertMethod, 0)

	for i := 0; i < len(alertMethodsTF); i++ {
		if alertMethodsTF[i] == nil {
			continue
		}
		alertMethodTF := alertMethodsTF[i].(map[string]interface{})
		alertMethodsAPI = append(alertMethodsAPI, n9api.AnomalyConfigAlertMethod{
			Name:    alertMethodTF["name"].(string),
			Project: alertMethodTF["project"].(string),
		})
	}

	return alertMethodsAPI
}

func unmarshalAnomalyConfig(d *schema.ResourceData, spec map[string]interface{}) error {
	anomalyConfigRaw, anomalyConfigExists := spec["anomalyConfig"]
	if !anomalyConfigExists {
		return nil
	}
	anomalyConfig := anomalyConfigRaw.(map[string]interface{})

	noData := anomalyConfig["noData"].(map[string]interface{})
	noDataMethods := noData["alertMethods"].([]interface{})
	resNoDataMethods := make([]map[string]interface{}, 0)

	if len(noDataMethods) == 0 {
		return nil
	}

	for _, amRaw := range noDataMethods {
		am := amRaw.(map[string]interface{})
		if am["name"] == nil || am["project"] == nil {
			continue
		}

		resNoDataMethods = append(resNoDataMethods, map[string]interface{}{
			"name":    am["name"],
			"project": am["project"],
		})
	}

	anomalyConfigTF := map[string]interface{}{
		"no_data": schema.NewSet(oneElementSet, []interface{}{
			map[string]interface{}{
				"alert_method": resNoDataMethods,
			},
		}),
	}

	return d.Set(
		"anomaly_config",
		schema.NewSet(oneElementSet, []interface{}{anomalyConfigTF}),
	)
}
