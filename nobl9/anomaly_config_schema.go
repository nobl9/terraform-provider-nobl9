package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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

func marshalAnomalyConfig(anomalyConfigRaw interface{}) *v1alpha.AnomalyConfig {
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
	marshaledAlertMethods := marshalAnomalyConfigAlertMethods(noDataAlertMethods)

	return &v1alpha.AnomalyConfig{
		NoData: &v1alpha.AnomalyConfigNoData{
			AlertMethods: marshaledAlertMethods,
		},
	}
}

func marshalAnomalyConfigAlertMethods(alertMethodsTF []interface{}) []v1alpha.AnomalyConfigAlertMethod {
	alertMethodsAPI := make([]v1alpha.AnomalyConfigAlertMethod, 0)

	for i := 0; i < len(alertMethodsTF); i++ {
		if alertMethodsTF[i] == nil {
			continue
		}
		alertMethodTF := alertMethodsTF[i].(map[string]interface{})
		alertMethodsAPI = append(alertMethodsAPI, v1alpha.AnomalyConfigAlertMethod{
			Name:    alertMethodTF["name"].(string),
			Project: alertMethodTF["project"].(string),
		})
	}

	return alertMethodsAPI
}

func unmarshalAnomalyConfig(d *schema.ResourceData, spec v1alpha.SLOSpec) error {
	if spec.AnomalyConfig == nil {
		return nil
	}

	noData := spec.AnomalyConfig.NoData
	noDataMethods := noData.AlertMethods
	resNoDataMethods := make([]map[string]interface{}, 0)

	if len(noDataMethods) == 0 {
		return nil
	}

	for _, am := range noDataMethods {
		if am.Name == "" || am.Project == "" {
			continue
		}

		resNoDataMethods = append(resNoDataMethods, map[string]interface{}{
			"name":    am.Name,
			"project": am.Project,
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
