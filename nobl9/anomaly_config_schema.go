package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
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
					Description: "No data alerts configuration",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"alert_after": {
								Type:     schema.TypeString,
								Optional: true,
								//nolint:lll
								Description: "Specifies the duration to wait after receiving no data before triggering an alert. " +
									"The value must be a valid Go duration string, such as \"1h\" for one hour. " +
									"If not specified, the system defaults to \"15m\" (15 minutes).",
								Default: "15m",
							},
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

func marshalAnomalyConfig(anomalyConfigRaw interface{}) *v1alphaSLO.AnomalyConfig {
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

	alertAfter, ok := noDataAnomalyConfig["alert_after"].(string)
	if !ok || alertAfter == "" {
		alertAfter = "15m"
	}

	return &v1alphaSLO.AnomalyConfig{
		NoData: &v1alphaSLO.AnomalyConfigNoData{
			AlertMethods: marshaledAlertMethods,
			AlertAfter:   &alertAfter,
		},
	}
}

func marshalAnomalyConfigAlertMethods(alertMethodsTF []interface{}) []v1alphaSLO.AnomalyConfigAlertMethod {
	alertMethodsAPI := make([]v1alphaSLO.AnomalyConfigAlertMethod, 0)

	for i := 0; i < len(alertMethodsTF); i++ {
		if alertMethodsTF[i] == nil {
			continue
		}
		alertMethodTF := alertMethodsTF[i].(map[string]interface{})
		alertMethodsAPI = append(alertMethodsAPI, v1alphaSLO.AnomalyConfigAlertMethod{
			Name:    alertMethodTF["name"].(string),
			Project: alertMethodTF["project"].(string),
		})
	}

	return alertMethodsAPI
}

func unmarshalAnomalyConfig(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
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
				"alert_after":  noData.AlertAfter,
			},
		}),
	}

	return d.Set(
		"anomaly_config",
		schema.NewSet(oneElementSet, []interface{}{anomalyConfigTF}),
	)
}
