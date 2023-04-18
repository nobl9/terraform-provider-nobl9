package nobl9

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

const anomalyConfigKey = "anomaly_config"

func schemaAnomalyConfig() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    false,
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
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Alert methods attached to Anomaly Config",
								MaxItems:    5,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validateMaxLength("display_name", 63),
										},
										"project": {
											Type: schema.TypeString,
											//Required: false,
											Optional: true,
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
	anomalyConfigTF := make(map[string]interface{})

	noData := anomalyConfig["noData"].(map[string]interface{})
	//alertMethods := noData["alertMethods"]

	anomalyConfigTF["no_data"] = noData
	//err = d.Set("alert_method", unmarshalAlertMethods(alertMethods))
	//diags = appendError(diags, err)

	return d.Set("anomaly_config", schema.NewSet(oneElementSet, []interface{}{anomalyConfigTF}))
}
