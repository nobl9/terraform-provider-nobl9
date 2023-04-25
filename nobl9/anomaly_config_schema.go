package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
	"reflect"
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

// diffSuppressAnomalyConfig takes the old and new value of anomaly_config and searches for diff.
// If the result is true, it means that there's no need to reapply (the diff is suppressed).
// Example input:
//
//	oldValue: {
//	  NoData: {
//	    AlertMethods: [
//	      { Name: "method1", Project: "project1" },
//	      { Name: "method2", Project: "project2" },
//	    ],
//	  },
//	}
//
//	newValue: {
//	  NoData: {
//	    AlertMethods: [
//	      { Name: "method2", Project: "project2" },
//	      { Name: "method1", Project: "project1" },
//	    ],
//	  },
//	}
//
// Example output:
// true
func diffSuppressAnomalyConfig(_, _, _ string, d *schema.ResourceData) bool {
	oldValue, newValue := d.GetChange("anomaly_config")

	oldAnomalyConfig := marshalAnomalyConfig(oldValue)
	newAnomalyConfig := marshalAnomalyConfig(newValue)

	if oldAnomalyConfig == nil && newAnomalyConfig != nil ||
		oldAnomalyConfig != nil && newAnomalyConfig == nil {
		return false
	}

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

// transformAnomalyConfigAlertMethodsTo2DMap transforms a slice of
// AnomalyConfigAlertMethod values into a 2D map, where each row of
// the map corresponds to an alert method and contains the method's
// name and project as key-value pairs.
// Example input:
// [
//
//	{ Name: "method1", Project: "project1" },
//	{ Name: "method2", Project: "project2" },
//	{ Name: "method3", Project: "project1" },
//
// ]
//
// Example output:
//
//	{
//	  "method1": { "name": "method1", "project": "project1" },
//	  "method2": { "name": "method2", "project": "project2" },
//	  "method3": { "name": "method3", "project": "project1" },
//	}
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
