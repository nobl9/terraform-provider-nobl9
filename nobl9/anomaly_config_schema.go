package nobl9

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
	"reflect"
	"sort"
	"strings"
)

const anomalyConfigKey = "anomaly_config"

func schemaAnomalyConfig() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeSet,
		Optional:         true,
		Description:      "Configuration for Anomalies. Currently supported Anomaly Type is NoData",
		MaxItems:         1,
		DiffSuppressFunc: diffSuppressAnomalyConfig,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"no_data": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Alert Policies attached to SLO",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alert_method": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Alert methods attached to Anomaly Config",
								MaxItems:    5,
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

func diffSuppressAnomalyConfig(fieldPath, oldValueStr, newValueStr string, d *schema.ResourceData) bool {
	return true
	fieldPathSegments := strings.Split(fieldPath, ".")
	if len(fieldPathSegments) > 1 {
		fieldName := fieldPathSegments[len(fieldPathSegments)-1]
		if fieldName == fieldLabelKey {
			// Terraform's GetChange function will fail to notice if user reapplied the resource
			// with all the labels removed from the file.
			// This is the situation in which one of the values in the label's schema is set and the other one isn't.
			if exactlyOneStringEmpty(oldValueStr, newValueStr) {
				return false
			}
		}
	}

	// the N9 API will return the labels in alphabetical order for keys and values.
	// Users should be able to declare label keys and values in any order
	// and changing order should force recreating the resource.
	// In order to achieve that, we're flattening the initial label struct to 2D map
	// and check if the label values inside that 2D map are deeply equal.
	// A simple reflect.DeepEqual change is not enough for the whole 2D map
	// because it omits the values order inside the array.
	// ---------------------------------
	// Example of (deeply) equal labels:
	//   label {
	//    key    = "team"
	//    values = ["sapphire", "green"]
	//  }
	//  label {
	//    key    = "team"
	//    values = ["green", "sapphire"]
	//  }
	oldValue, newValue := d.GetChange(fieldLabel)
	labelsOld := oldValue.([]interface{})
	labelsNew := newValue.([]interface{})
	if len(labelsOld) != len(labelsNew) {
		return false
	}

	oldMap := transformLabelsTo2DMap(labelsOld)
	newMap := transformLabelsTo2DMap(labelsNew)

	isDeepEqual := true
	for labelKey := range newMap {
		if _, exist := oldMap[labelKey][fieldLabelValues]; !exist {
			return false
		}

		var oldValues = oldMap[labelKey][fieldLabelValues].([]interface{})
		var newValues = newMap[labelKey][fieldLabelValues].([]interface{})

		sort.Slice(oldValues, func(i, j int) bool {
			return oldValues[i].(string) < oldValues[j].(string)
		})
		sort.Slice(newValues, func(i, j int) bool {
			return newValues[i].(string) < newValues[j].(string)
		})

		if !reflect.DeepEqual(oldValues, newValues) {
			isDeepEqual = false
		}
	}

	return isDeepEqual
}

func marshalAnomalyConfig(d *schema.ResourceData) *n9api.AnomalyConfig {
	fmt.Println("d", d)

	anomalyConfigSet := d.Get("anomaly_config").(*schema.Set)
	if anomalyConfigSet.Len() == 0 {
		return nil
	}

	anomalyConfig := anomalyConfigSet.List()[0].(map[string]interface{})
	noDataAnomalyConfig := anomalyConfig["no_data"].(*schema.Set).List()[0].(map[string]interface{})
	if _, ok := noDataAnomalyConfig["alert_method"]; !ok {
		panic("no alert method")
	}

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
	anomalyConfig := spec["anomalyConfig"].(map[string]interface{})

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
