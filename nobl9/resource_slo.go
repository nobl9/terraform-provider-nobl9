package nobl9

import (
	"context"

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
													Description: "Name of the metric source project",
												},
											},
										},
									},
									"metric_spec": schemaMetricSpec(),
								},
							},
						},
						"objectives": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: " [Objectives documentation](https://nobl9.github.io/techdocs_YAML_Guide/#objectives)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"count_metrics": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Alert Policies attached to SLO",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"good": schemaMetricSpec(),
												"incemental": {
													Type:        schema.TypeBool,
													Required:    true,
													Description: "Should the metrics be incrementing or not",
												},
												"metric_spec": schemaMetricSpec(),
											},
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
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
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
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
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
	//alertPolicies := d.Get("alert_policies").([]interface{})
	//alertPoliciesStr := make([]string, len(sourceOf))
	//for i, s := range alertPolicies {
	//	alertPoliciesStr[i] = s.(string)
	//}

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
					Name:    d.Get("name").(string),
				},
				RawMetric: &n9api.MetricSpec{
					Prometheus:          marshalSLOPrometheus(d),
					Datadog:             marshalSLODatadog(d),
					NewRelic:            marshalSLONewRelic(d),
					AppDynamics:         marshalSLOAppDynamics(d),
					Splunk:              marshalSLOSplunk(d),
					Lightstep:           marshalSLOLightstep(d),
					SplunkObservability: marshalSLOSplunkObservability(d),
					Dynatrace:           marshalSLODynatrace(d),
					ThousandEyes:        marshalSLOThousandEyes(d),
					Graphite:            marshalSLOGraphite(d),
					BigQuery:            marshalSLOBigQuery(d),
					OpenTSDB:            marshalSLOOpenTSDB(d),
					GrafanaLoki:         marshalSLOGrafanaLoki(d),
					Elasticsearch:       marshalSLOElasticsearch(d),
				},
			},
			BudgetingMethod: d.Get("budgeting_method").(string),
			Thresholds:      marshalThresholds(d),
			Service:         d.Get("service").(string),
			//TimeWindows: n9api.TimeWindow{ TODO
			//	Unit:      d.Get("unit").(string),
			//	Count:     d.Get(),
			//	IsRolling: d.Get().(bool),
			//	Calendar: &n9api.Calendar{
			//		StartTime: d.Get("start_time").(string),
			//		TimeZone:  d.Get("time_zone").(string),
			//	},
			//	Period: &n9api.Period{
			//		Begin: d.Get("begin").(string),
			//		End:   d.Get("end").(string),
			//	},
			//},
			//AlertPolicies: alertPoliciesStr, TODO
			//Attachments: n9api.Attachment{ TODO
			//	DisplayName: d.Get("display_name").(string),
			//	Url:         d.Get("url").(string),
			//},
			CreatedAt: d.Get("created_at").(string),
		},
	}
}

func marshalThresholds(d *schema.ResourceData) []n9api.Threshold {
	return nil
	//n9api.Threshold{
	//	ThresholdBase:   d.Get("value").(float64),
	//	BudgetTarget:    d.Get("target").(float64),
	//	TimeSliceTarget: d.Get("time_slice_target").(float64),
	//	CountMetrics: &n9api.CountMetricsSpec{
	//		Incremental: d.Get("incremental").(bool),
	//		GoodMetric: &n9api.MetricSpec{
	//			Prometheus:          marshalSLOPrometheus(d),
	//			Datadog:             marshalSLODatadog(d),
	//			NewRelic:            marshalSLONewRelic(d),
	//			AppDynamics:         marshalSLOAppDynamics(d),
	//			Splunk:              marshalSLOSplunk(d),
	//			Lightstep:           marshalSLOLightstep(d),
	//			SplunkObservability: marshalSLOSplunkObservability(d),
	//			Dynatrace:           marshalSLODynatrace(d),
	//			ThousandEyes:        marshalSLOThousandEyes(d),
	//			Graphite:            marshalSLOGraphite(d),
	//			BigQuery:            marshalSLOBigQuery(d),
	//			OpenTSDB:            marshalSLOOpenTSDB(d),
	//			GrafanaLoki:         marshalSLOGrafanaLoki(d),
	//			Elasticsearch:       marshalSLOElasticsearch(d),
	//		},
	//		TotalMetric: &n9api.MetricSpec{
	//			Prometheus:          marshalSLOPrometheus(d),
	//			Datadog:             marshalSLODatadog(d),
	//			NewRelic:            marshalSLONewRelic(d),
	//			AppDynamics:         marshalSLOAppDynamics(d),
	//			Splunk:              marshalSLOSplunk(d),
	//			Lightstep:           marshalSLOLightstep(d),
	//			SplunkObservability: marshalSLOSplunkObservability(d),
	//			Dynatrace:           marshalSLODynatrace(d),
	//			ThousandEyes:        marshalSLOThousandEyes(d),
	//			Graphite:            marshalSLOGraphite(d),
	//			BigQuery:            marshalSLOBigQuery(d),
	//			OpenTSDB:            marshalSLOOpenTSDB(d),
	//			GrafanaLoki:         marshalSLOGrafanaLoki(d),
	//			Elasticsearch:       marshalSLOElasticsearch(d),
	//		},
	//	},
	//	Operator: d.Get("op").(string),
	//}
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

	query := ddog["query"].(string)
	return &n9api.DatadogMetric{
		Query: &query,
	}
}

func marshalSLONewRelic(d *schema.ResourceData) *n9api.NewRelicMetric {
	p := d.Get("newrelic_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	newrelic := p[0].(map[string]interface{})

	nrql := newrelic["nrql"].(string)
	return &n9api.NewRelicMetric{
		NRQL: &nrql,
	}
}

func marshalSLOAppDynamics(d *schema.ResourceData) *n9api.AppDynamicsMetric {
	p := d.Get("appdynamics_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	appdynamics := p[0].(map[string]interface{})

	applicationName := appdynamics["application_name"].(string)
	metricPath := appdynamics["metric_path"].(string)
	return &n9api.AppDynamicsMetric{
		ApplicationName: &applicationName,
		MetricPath:      &metricPath,
	}
}
func marshalSLOSplunk(d *schema.ResourceData) *n9api.SplunkMetric {
	p := d.Get("splunk_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunk := p[0].(map[string]interface{})

	query := splunk["query"].(string)
	fieldName := splunk["field_name"].(string)
	return &n9api.SplunkMetric{
		Query:     &query,
		FieldName: &fieldName,
	}
}
func marshalSLOLightstep(d *schema.ResourceData) *n9api.LightstepMetric {
	p := d.Get("lightstep_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	lightstep := p[0].(map[string]interface{})

	streamID := lightstep["stream_id"].(string)
	typeOfData := lightstep["type_of_data"].(string)
	percentile := lightstep["percentile"].(float64)
	return &n9api.LightstepMetric{
		StreamID:   &streamID,
		TypeOfData: &typeOfData,
		Percentile: &percentile,
	}
}
func marshalSLOSplunkObservability(d *schema.ResourceData) *n9api.SplunkObservabilityMetric {
	p := d.Get("splunk_observability_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunkObservability := p[0].(map[string]interface{})

	program := splunkObservability["program"].(string)
	return &n9api.SplunkObservabilityMetric{
		Program: &program,
	}
}
func marshalSLODynatrace(d *schema.ResourceData) *n9api.DynatraceMetric {
	p := d.Get("dynatrace_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	dynatrace := p[0].(map[string]interface{})

	selector := dynatrace["metric_selector"].(string)
	return &n9api.DynatraceMetric{
		MetricSelector: &selector,
	}
}
func marshalSLOThousandEyes(d *schema.ResourceData) *n9api.ThousandEyesMetric {
	p := d.Get("thousandeyes_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	thousandeyes := p[0].(map[string]interface{})

	testID := thousandeyes["test_id"].(int64)
	return &n9api.ThousandEyesMetric{
		TestID: &testID,
	}
}
func marshalSLOGraphite(d *schema.ResourceData) *n9api.GraphiteMetric {
	p := d.Get("graphite_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	graphite := p[0].(map[string]interface{})

	metricPath := graphite["metric_path"].(string)
	return &n9api.GraphiteMetric{
		MetricPath: &metricPath,
	}
}
func marshalSLOBigQuery(d *schema.ResourceData) *n9api.BigQueryMetric {
	p := d.Get("bigquery_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	bigquery := p[0].(map[string]interface{})

	return &n9api.BigQueryMetric{
		Query:     bigquery["query"].(string),
		ProjectID: bigquery["project_id"].(string),
		Location:  bigquery["location"].(string),
	}
}
func marshalSLOOpenTSDB(d *schema.ResourceData) *n9api.OpenTSDBMetric {
	p := d.Get("opentsdb_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	opentsdb := p[0].(map[string]interface{})

	query := opentsdb["query"].(string)
	return &n9api.OpenTSDBMetric{
		Query: &query,
	}
}

func marshalSLOGrafanaLoki(d *schema.ResourceData) *n9api.GrafanaLokiMetric {
	p := d.Get("grafana_loki_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	grafanaloki := p[0].(map[string]interface{})

	logql := grafanaloki["logql"].(string)
	return &n9api.GrafanaLokiMetric{
		Logql: &logql,
	}
}

func marshalSLOElasticsearch(d *schema.ResourceData) *n9api.ElasticsearchMetric {
	p := d.Get("elasticsearch_metric").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	elasticsearch := p[0].(map[string]interface{})

	index := elasticsearch["index"].(string)
	query := elasticsearch["query"].(string)
	return &n9api.ElasticsearchMetric{
		Index: &index,
		Query: &query,
	}
}

func unmarshalSLO(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	alertPolicies := object["alertPolicies"].(map[string]interface{})
	err := d.Set("alert_policies", alertPolicies)
	diags = appendError(diags, err)

	attachments := object["attachments"].(map[string]interface{})
	err = d.Set("attachments", attachments)
	diags = appendError(diags, err)

	budgetingMethod := object["budgetingMethod"].(map[string]interface{})
	err = d.Set("budgeting_method", budgetingMethod)
	diags = appendError(diags, err)

	createdAt := object["createdAt"].(map[string]interface{})
	err = d.Set("created_at", createdAt)
	diags = appendError(diags, err)

	description := object["description"].(map[string]interface{})
	err = d.Set("description", description)
	diags = appendError(diags, err)

	supportedMetrics := []struct {
		hclName  string
		jsonName string
	}{
		{"prometheus_metric", "prometheus"},
		{"datadog_metric", "datadog"},
		{"newrelic_metric", "newRelic"},
		{"appdynamics_metric", "appDynamics"},
		{"splunk_metric", "splunk"},
		{"lightstep_metric", "lightstep"},
		{"splunk_observability_metric", "splunkObservability"},
		{"dynatrace_metric", "dynatrace"},
		{"thousandeyes_metric", "thousandEyes"},
		{"graphite_metric", "graphite"},
		{"bigquery_metric", "bigQuery"},
		{"opentsdb_metric", "opentsdb"},
		{"elasticsearch_metric", "elasticsearch"},
		{"grafana_loki_metric", "grafanaLoki"},
	}

	for _, name := range supportedMetrics {
		ok, ds := unmarshalSLOMetric(d, object, name.hclName, name.jsonName)
		if ds.HasError() {
			diags = append(diags, ds...)
		}
		if ok {
			break
		}
	}

	objectives := object["objectives"].(map[string]interface{})
	err = d.Set("objectives", objectives)
	diags = appendError(diags, err)

	service := object["service"].(map[string]interface{})
	err = d.Set("service", service)
	diags = appendError(diags, err)

	timeWindows := object["timeWindows"].(map[string]interface{})
	err = d.Set("timeWindows", timeWindows)
	diags = appendError(diags, err)

	return diags
}

func unmarshalSLOMetric(d *schema.ResourceData, object n9api.AnyJSONObj, hclName, jsonName string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	spec := object["spec"].(map[string]interface{})
	indicator := spec["indicator"].(map[string]interface{})
	rawmetric := indicator["rawmetric"].(map[string]interface{})
	if rawmetric[jsonName] == nil {
		return false, nil
	}

	err := d.Set("alert_policies", spec["alertPolicies"])
	appendError(diags, err)
	err = d.Set(hclName, schema.NewSet(oneElementSet, []interface{}{rawmetric[jsonName]}))
	appendError(diags, err)

	return true, diags
}

func resourceSLOApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	slo := marshalSLO(d)

	var p n9api.Payload
	p.AddObject(slo)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add SLO: %s", err.Error())
	}

	d.SetId(slo.Metadata.Name)

	return resourceSLORead(ctx, d, meta)
}

func resourceSLORead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	client, ds := newClient(config, project)
	if ds.HasError() {
		return ds
	}

	objects, err := client.GetObject(n9api.ObjectSLO, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalSLO(d, objects)
}

func resourceSLODelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectSLO, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
