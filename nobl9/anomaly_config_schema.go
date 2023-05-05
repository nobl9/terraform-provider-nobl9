package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
	"reflect"
	"strings"
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
					Type:             schema.TypeSet,
					Required:         true,
					DiffSuppressFunc: diffSuppressAnomalyConfig,
					Description:      "Alert Policies attached to SLO",
					MinItems:         1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alert_method": {
								Type:        schema.TypeList,
								Required:    true,
								Description: "Alert methods attached to Anomaly Config",
								MaxItems:    5,
								MinItems:    1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:             schema.TypeString,
											Optional:         true,
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
func diffSuppressAnomalyConfig(fieldPath, oldValueStr, newValueStr string, d *schema.ResourceData) bool {
	oldValue, newValue := d.GetChange("anomaly_config")
	oldAnomalyConfig := marshalAnomalyConfig(oldValue)
	newAnomalyConfig := marshalAnomalyConfig(newValue)

	oldMethods := make(map[string]map[string]string)
	if oldAnomalyConfig != nil {
		oldMethods = transformAnomalyConfigAlertMethodsTo2DMap(
			oldAnomalyConfig.NoData.AlertMethods,
		)
	}

	newMethods := make(map[string]map[string]string)
	if newAnomalyConfig != nil {
		newMethods = transformAnomalyConfigAlertMethodsTo2DMap(
			newAnomalyConfig.NoData.AlertMethods,
		)
	}

	fieldPathSegments := strings.Split(fieldPath, ".")
	if len(fieldPathSegments) > 1 {
		fieldName := fieldPathSegments[len(fieldPathSegments)-1]
		if fieldName == "alert_method" || fieldName == "name" || fieldName == "project" {
			// Terraform's GetChange function will fail to notice if user reapplied the resource
			// with all the labels removed from the file.
			// This is the situation in which one of the values in the label's schema is set and the other one isn't.
			if exactlyOneStringEmpty(oldValueStr, newValueStr) {
				return false
			}
		}
	}
	return reflect.DeepEqual(oldMethods, newMethods)
}

func marshalAnomalyConfig(anomalyConfigRaw interface{}) *n9api.AnomalyConfig {
	anomalyConfigSet := anomalyConfigRaw.(*schema.Set)
	if anomalyConfigSet.Len() == 0 || anomalyConfigSet.List()[0] == nil {
		return nil
	}
	anomalyConfig := anomalyConfigSet.List()[0].(map[string]interface{})
	noDataAnomalyConfigList := anomalyConfig["no_data"].(*schema.Set).List()
	if noDataAnomalyConfigList[0] == nil {
		return nil
	}
	noDataAnomalyConfig := noDataAnomalyConfigList[0].(map[string]interface{})
	noDataAlertMethods := noDataAnomalyConfig["alert_method"].([]interface{})

	if len(noDataAlertMethods) == 0 {
		return nil
	}

	marshalledAlertMethods, isEmpty := marshalAnomalyConfigAlertMethods(noDataAlertMethods)

	if isEmpty {
		return nil
	}

	return &n9api.AnomalyConfig{
		NoData: &n9api.AnomalyConfigNoData{
			AlertMethods: marshalledAlertMethods,
		},
	}
}

func marshalAnomalyConfigAlertMethods(alertMethodsTF []interface{}) ([]n9api.AnomalyConfigAlertMethod, bool) {
	alertMethodsAPI := make([]n9api.AnomalyConfigAlertMethod, 0)

	isEmpty := true
	for i := 0; i < len(alertMethodsTF); i++ {
		if alertMethodsTF[i] == nil {
			continue
		}
		alertMethodTF := alertMethodsTF[i].(map[string]interface{})
		if alertMethodTF["name"].(string) == "" || alertMethodTF["project"].(string) == "" {
			continue
		} else {
			isEmpty = false
		}

		alertMethodsAPI = append(alertMethodsAPI, n9api.AnomalyConfigAlertMethod{
			Name:    alertMethodTF["name"].(string),
			Project: alertMethodTF["project"].(string),
		})
	}

	return alertMethodsAPI, isEmpty
}

func unmarshalAnomalyConfig(d *schema.ResourceData, spec map[string]interface{}) error {
	anomalyConfigRaw, ok := spec["anomalyConfig"]
	if !ok {
		return d.Set(
			"anomaly_config",
			nil,
		)
	}
	anomalyConfig := anomalyConfigRaw.(map[string]interface{})

	noData := anomalyConfig["noData"].(map[string]interface{})
	noDataMethods := noData["alertMethods"].([]interface{})
	resNoDataMethods := make([]map[string]interface{}, 0)

	if len(noDataMethods) == 0 {
		return d.Set(
			"anomaly_config",
			nil,
		)
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
