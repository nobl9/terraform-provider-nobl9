package nobl9

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
	"reflect"
)

//const anomalyConfigKey = "anomaly_config"

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
					Optional:    true,
					Description: "Alert Policies attached to SLO",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alert_method": {
								Type:             schema.TypeList,
								Optional:         true,
								Description:      "Alert methods attached to Anomaly Config",
								MaxItems:         5,
								DiffSuppressFunc: diffSuppressAnomalyConfig,
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
											Optional: true,
											Description: "Project name the Alert Method is in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)." +
												" If not defined, Nobl9 returns a default value for this field.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func transformAnomalyConfigAlertMethodsTo2DMap(alertMethods []n9api.AnomalyConfigAlertMethod) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, method := range alertMethods {
		values := make(map[string]string)

		values["name"] = method.Name
		values["project"] = method.Project
		result[method.Name] = values
	}
	return result
}

func diffSuppressAnomalyConfig(_, _, _ string, d *schema.ResourceData) bool {
	oldValue, newValue := d.GetChange("anomaly_config")

	oldAnomalyConfig := marshalAnomalyConfig(oldValue)
	newAnomalyConfig := marshalAnomalyConfig(newValue)

	return reflect.DeepEqual(
		transformAnomalyConfigAlertMethodsTo2DMap(
			oldAnomalyConfig.NoData.AlertMethods,
		),
		transformAnomalyConfigAlertMethodsTo2DMap(
			newAnomalyConfig.NoData.AlertMethods,
		),
	)
}

func marshalAnomalyConfig(anomalyConfigRaw interface{}) *n9api.AnomalyConfig {
	anomalyConfigSet := anomalyConfigRaw.(*schema.Set)
	if anomalyConfigSet.Len() == 0 {
		return nil
	}

	anomalyConfig := anomalyConfigSet.List()[0].(map[string]interface{})
	noDataAnomalyConfig := anomalyConfig["no_data"].(*schema.Set).List()[0].(map[string]interface{})
	noDataAlertMethods := noDataAnomalyConfig["alert_method"].([]interface{})

	fmt.Println(noDataAlertMethods)
	return &n9api.AnomalyConfig{
		NoData: &n9api.AnomalyConfigNoData{
			AlertMethods: marshalAnomalyConfigAlertMethods(noDataAlertMethods),
		},
	}
}

func marshalAnomalyConfigAlertMethods(alertMethodsTF []interface{}) []n9api.AnomalyConfigAlertMethod {
	alertMethodsAPI := make([]n9api.AnomalyConfigAlertMethod, len(alertMethodsTF))
	for i := 0; i < len(alertMethodsTF); i++ {
		alertMethodTF := alertMethodsTF[i].(map[string]interface{})
		alertMethodsAPI[i] = n9api.AnomalyConfigAlertMethod{
			Name:    alertMethodTF["name"].(string),
			Project: alertMethodTF["project"].(string),
		}
	}

	return alertMethodsAPI
}

func unmarshalAnomalyConfig(d *schema.ResourceData, spec map[string]interface{}) error {
	anomalyConfigRaw, _ := spec["anomalyConfig"]
	//if !ok {
	//	return nil
	//}
	anomalyConfig := anomalyConfigRaw.(map[string]interface{})

	noData := anomalyConfig["noData"].(map[string]interface{})
	noDataMethods := noData["alertMethods"].([]interface{})
	resNoDataMethods := make([]map[string]interface{}, len(noDataMethods))
	for i, amRaw := range noDataMethods {
		am := amRaw.(map[string]interface{})

		resNoDataMethods[i] = map[string]interface{}{
			"name":    am["name"],
			"project": am["project"],
		}
	}

	noData2 := map[string]interface{}{
		"alert_method": resNoDataMethods,
	}

	//anomalyConfigTF := make(map[string]interface{})
	anomalyConfigTF := map[string]interface{}{
		"no_data": schema.NewSet(oneElementSet, []interface{}{noData2}),
	}

	x := schema.NewSet(oneElementSet, []interface{}{anomalyConfigTF})
	y := d.Set(
		"anomaly_config",
		x,
	)
	return y
}

func convertMapToSet(data map[string]interface{}) *schema.Set {
	set := &schema.Set{
		F: schema.HashString,
	}

	for k, v := range data {
		set.Add(map[string]interface{}{
			"key":   k,
			"value": v,
		})
	}

	return set
}
