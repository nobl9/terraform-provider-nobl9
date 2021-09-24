package nobl9

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceSLO() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),

			"slo_spec": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[SLO documentation](https://nobl9.github.io/techdocs_YAML_Guide/#slo)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert_policies": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Alert Policies attached to SLO",
							Elem: &schema.Schema{
								Type:        schema.TypeString,
								Description: "Alert Policy",
							},
						},
						"attachments": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name which is dispalyed for the attachment",
									},
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Url to the attachment",
									},
								},
							},
						},
						"budgeting_method": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Method which will be use to calculate budget",
						},
						"created_at": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Time of creation",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the SLO",
						},
						"indicator": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: " ",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_source_spec": {
										Type:        schema.TypeSet,
										Required:    true,
										Description: "",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Name of the metric source",
												},
												"project": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Name of the metric souce project",
												},
											},
										},
									},
									"metric_spec": schemaMetricSpec(),
									},
								},
							},
						},
						"objectives": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: " ([Objectives documentation] https://nobl9.github.io/techdocs_YAML_Guide/#objectives)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"count_metrics": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Alert Policies attached to SLO",
										Elem: &schema.Schema{
											"good": schemaMetricSpec(),
											"incemental": {
												Type:        schema.TypeBool,
							        			Required:    true,
												Description: "Should the metrics be incrementing or not",
											},
											"metric_spec": schemaMetricSpec(),
										},
									},
									"display_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name to be displayed",
									},
									"op": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Type of logical operation",
									},
									"target": {
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Desiganted value",
									},
									"time_slice_target": {
										Type:        schema.TypeFloat,
										Optional:    true,
										Description: "Designated value for slice",
									},
									"value": {
										Type:        schema.TypeFloat,
										Optional:    true,
										Description: "Value",
									},
								},
							},
						},
						"service": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the service",
						},
						"time_windows": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: " ",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"calendar": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Alert Policies attached to SLO",
										Elem: &schema.Schema{
											"start_time": {
												Type:        schema.TypeString,
							        			Required:    true,
												Description: "Date of the start",
											},
											"time_zone": {
												Type:        schema.TypeString,
							        			Required:    true,
												Description: "Timezone name in IANA Time Zone Database",
											},
										},
									},
									"count": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Count of the time unit",
									},
									"is_rolling": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Is the window moving or not",
									},
									"period": {
										Type:        schema.TypeFloat,
										Optional:    true,
										Description: "Specific time frame",
										Elem: &schema.Schema{
											"begin": {
												Type:        schema.TypeString,
							        			Optional:    true,
												Description: "Beginning of the period",
											},
											"end": {
												Type:        schema.TypeString,
							        			Optional:    true,
												Description: "End of the period",
											},
										},
									},
									"unit": {
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Unit of time",
									},
								},
							},
						},
					},
				},
			},
		},
		CreateContext: resourceSLOApply,
		UpdateContext: resourceSLOApply,
		DeleteContext: resourceSLODelete,
		ReadContext:   resourceSLORead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[SLO configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#SLO)",
	}
}

func marshalSLO(d *schema.ResourceData) *n9api.SLO {
	alertPolicies := d.Get("alert_policies").([]interface{})
	alertPoliciesStr := make([]string, len(sourceOf))
	for i, s := range alertPolicies {
		alertPoliciesStr[i] = s.(string)
	}

	return &n9api.SLO{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindSLO,
			MetadataHolder: marshalMetadata(d),
		},
		Spec: n9api.SLOSpec{
			Description: d.Get("description").(string),
			Indicator: n9api.Indicator{
				MetricSource: &n9api.MetricSourceSpec{
					Project: d.Get("project").(string),
					Name: d.Get("name").(string),
				},
				RawMetric: &n9api.MetricSpec{
					Prometheus:          marshalSLOPrometheus(d),
					Datadog:             marshalSLODatadog(d),
					NewRelic:            marshalSLONewRelic(d),
					AppDynamics:         marshalSLOAppDynamics(d),
					Splunk:              marshalSLOSplunk(d),
					Lightstep:           marshalSLOLightstep(d),
					SplunkObservability: marshalSLOSplunkObservability(d),
					Dynatrace:           marshalDynatrace(d),
					ThousandEyes:        marshalSLOThousandEyes(d),
					Graphite:            marshalSLOGraphite(d),
					BigQuery:            marshalSLOBigQuery(d),
					OpenTSDB:            marshalSLOOpenTSDB(d),
				},
			BudgetingMethod: d.Get("budgeting_method").(string),
			Thresholds: n9api.Threshold{
				ThresholdBase: d.Get("value").(float64),
				BudgetTarget: d.Get("target").(float64),
				TimeSliceTarget: d.Get("time_slice_target").(float64),
				CountMetrics: &n9api.CountMetricsSpec{
					Incremental: d.Get("incremental").(bool),
					GoodMetric: &n9api.MetricSpec{
						Prometheus:          marshalSLOPrometheus(d),
						Datadog:             marshalSLODatadog(d),
						NewRelic:            marshalSLONewRelic(d),
						AppDynamics:         marshalSLOAppDynamics(d),
						Splunk:              marshalSLOSplunk(d),
						Lightstep:           marshalSLOLightstep(d),
						SplunkObservability: marshalSLOSplunkObservability(d),
						Dynatrace:           marshalDynatrace(d),
						ThousandEyes:        marshalSLOThousandEyes(d),
						Graphite:            marshalSLOGraphite(d),
						BigQuery:            marshalSLOBigQuery(d),
						OpenTSDB:            marshalSLOOpenTSDB(d),	
					},
					TotalMetric: &n9api.MetricSpec{
						Prometheus:          marshalSLOPrometheus(d),
						Datadog:             marshalSLODatadog(d),
						NewRelic:            marshalSLONewRelic(d),
						AppDynamics:         marshalSLOAppDynamics(d),
						Splunk:              marshalSLOSplunk(d),
						Lightstep:           marshalSLOLightstep(d),
						SplunkObservability: marshalSLOSplunkObservability(d),
						Dynatrace:           marshalDynatrace(d),
						ThousandEyes:        marshalSLOThousandEyes(d),
						Graphite:            marshalSLOGraphite(d),
						BigQuery:            marshalSLOBigQuery(d),
						OpenTSDB:            marshalSLOOpenTSDB(d),
					},
				},
				Operator: d.Get("op").(string),
			},
			Service: d.Get("service").(string),
			TimeWindows: n9api.TimeWindow{
				Unit: d.Get("unit").(string),
				Count: d.Get(),
				IsRolling: d.Get().(bool),
				Calendar: &n9api.Calendar{
					StartTime: d.Get("start_time").(string),
					TimeZone: d.Get("time_zone").(string),
				},
				Period: &n9api.Period{
					Begin: d.Get("begin").(string),
					End: d.Get("end").(string),
				},
			},
			AlertPolicies: alertPoliciesStr,
			Attachemnts: n9api.Attachment{
				DisplayName: d.Get("display_name").(string),
				Url: d.Get("url").(string)
			},
			CreatedAt: d.Get("created_at").(string),
		},	
	}
}

func marshalSLOPrometheus(d *schema.ResourceData) *n9api.PrometheusMetric {
	p := d.Get("prometheus_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	prom := p[0].(map[string]interface{})

	query := prom["promql"].(string)
	return &n9api.PrometheusMetric{
		PromQL: &query,
	}
}

func marshalSLODatadog(d *schema.ResourceData) *n9api.DatadogMetric {
	p := d.Get("datadog_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	ddog := p[0].(map[string]interface{})

	return &n9api.DatadogMetric{
		Query: ddog["query"].(string),
	}
}

func marshalSLONewRelic(d *schema.ResourceData) *n9api.NewRelicMetric {
	p := d.Get("newrelic_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	newrelic := p[0].(map[string]interface{})

	return &n9api.NewRelicMetric{
		NRQL: ddog["nrql"].(string),
	}
}

func marshalSLOAppDynamics(d *schema.ResourceData) *n9api.AppDynamicsMetric {
	p := d.Get("appdynamics_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	appdynamics := p[0].(map[string]interface{})

	return &n9api.AppDynamicsMetric{
		ApplicationName: appdynamics["application_name"].(string),
		MetricPath: appdynamics["metric_path"].(string),
	}
}
func marshalSLOSplunk(d *schema.ResourceData) *n9api.SplunkMetric {
	p := d.Get("splunk_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunk := p[0].(map[string]interface{})

	return &n9api.SplunkMetric{
		Query: appdynamics["query"].(string),
		FieldName: appdynamics["field_name"].(string),
	}
}
func marshalSLOLightstep(d *schema.ResourceData) *n9api.LightstepMetric {
	p := d.Get("lightstep_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	lightstep := p[0].(map[string]interface{})

	return &n9api.LightstepMetric{
		StreamID: lightstep["stream_id"].(string),
		TypeOfData: lightstep["type_of_data"].(string),
		Percentile: lightstep["percentile"].(string),
	}
}
func marshalSLOSplunkObservability(d *schema.ResourceData) *n9api.SplunkObservabilityMetric {
	p := d.Get("splunk_observability_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunkObservability := p[0].(map[string]interface{})

	return &n9api.SplunkObservabilityMetric{
		Query: splunkObservability["query"].(string),
	}
}
func marshalSLODynatrace(d *schema.ResourceData) *n9api.DynatraceMetric {
	p := d.Get("dynatrace_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	dynatrace := p[0].(map[string]interface{})

	return &n9api.DynatraceMetric{
		MetricSelector: dynatrace["metric_selector"].(string),
	}
}
func marshalSLOThousandEyes(d *schema.ResourceData) *n9api.ThousandEyesMetric {
	p := d.Get("thousandeyes_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	thousandeyes := p[0].(map[string]interface{})

	return &n9api.ThousandEyesMetric{
		TestID: thousandeyes["test_id"].(string),
	}
}
func marshalSLOGraphite(d *schema.ResourceData) *n9api.GraphiteMetric {
	p := d.Get("graphite_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	graphite := p[0].(map[string]interface{})

	return &n9api.GraphiteMetric{
		MetricPath: graphite["metric_path"].(string),
	}
}
func marshalSLOBigQuery(d *schema.ResourceData) *n9api.BigQueryMetric {
	p := d.Get("bigquery_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	bigquery := p[0].(map[string]interface{})

	return &n9api.BigQueryMetric{
		Query: bigquery["query"].(string),
		ProjectID: bigquery["project_id"].(string),
		Location: bigquery["location"].(string),
	}
}
func marshalSLOOpenTSDB(d *schema.ResourceData) *n9api. {
	p := d.Get("_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	 := p[0].(map[string]interface{})

	return &n9api.Metric{
		: [""].(string),

	}
}

func unmarshalSLO() {}

func resourceSLOApply() {}

func resourceSLORead() {}

func resourceSLODelete() {}