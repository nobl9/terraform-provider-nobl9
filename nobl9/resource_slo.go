package nobl9

import (
	"context"
	"errors"

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
			"budgeting_method": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Method which will be use to calculate budget",
			},
			"service": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the service",
			},
			"indicator": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: " ",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "Name of the metric source (agent).",
						},
						"project": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Name of the metric source project.",
						},
						"kind": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Default:     "Agent",
							Description: "Kind of the metric source. One of {Agent, Direct}.",
						},
						"raw_metric": schemaMetricSpec(),
					},
				},
			},
			"objective": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "[Objectives documentation](https://nobl9.github.io/techdocs_YAML_Guide/#objectives)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count_metrics": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Alert Policies attached to SLO",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"good":  schemaMetricSpec(),
									"total": schemaMetricSpec(),
									"incremental": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Should the metrics be incrementing or not",
									},
								},
							},
						},
						"display_name": {
							Type:        schema.TypeString,
							Required:    true,
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
							Description: "Designated value",
						},
						// TODO enable time_slices back
						//"time_slice_target": {
						//	Type:        schema.TypeFloat,
						//	Optional:    true,
						//	Description: "Designated value for slice",
						//},
						"value": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "Value",
						},
					},
				},
			},
			"time_window": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: " ",
				MaxItems:    1,
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
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "", // TODO docs
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"unit": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unit of time",
						},
					},
				},
			},
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
	alertPoliciesStr := make([]string, len(alertPolicies))
	for i, s := range alertPolicies {
		alertPoliciesStr[i] = s.(string)
	} // TODO use slinceOfSting

	indicator := marshalIndicator(d)
	isRawMetric := indicator.RawMetric != nil
	return &n9api.SLO{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindSLO,
			MetadataHolder: marshalMetadata(d),
		},
		Spec: n9api.SLOSpec{
			Description:     d.Get("description").(string),
			Service:         d.Get("service").(string),
			BudgetingMethod: d.Get("budgeting_method").(string),
			Indicator:       indicator,
			Thresholds:      marshalThresholds(d, isRawMetric),
			TimeWindows:     marshalTimeWindows(d),
			AlertPolicies:   alertPoliciesStr,
			//Attachments: n9api.Attachment{ TODO
			//	DisplayName: d.Get("display_name").(string),
			//	Url:         d.Get("url").(string),
			//},
		},
	}
}

func marshalTimeWindows(d *schema.ResourceData) []n9api.TimeWindow {
	timeWindows := d.Get("time_window").(*schema.Set).List()[0].(map[string]interface{})

	return []n9api.TimeWindow{{
		Unit:      timeWindows["unit"].(string),
		Count:     timeWindows["count"].(int),
		IsRolling: timeWindows["is_rolling"].(bool),
		Calendar:  nil, // TODO impl me
	}}
}

func marshalIndicator(d *schema.ResourceData) n9api.Indicator {
	indicator := d.Get("indicator").(*schema.Set).List()[0].(map[string]interface{})
	var rawMetric *n9api.MetricSpec
	if raw := indicator["raw_metric"].(*schema.Set); raw.Len() > 0 {
		rawMetric = marshalMetric(raw.List()[0].(map[string]interface{}))
	}
	return n9api.Indicator{
		MetricSource: &n9api.MetricSourceSpec{
			Project: indicator["project"].(string),
			Name:    indicator["name"].(string),
			Kind:    indicator["kind"].(string),
		},
		RawMetric: rawMetric,
	}
}

func marshalMetric(metric map[string]interface{}) *n9api.MetricSpec {
	return &n9api.MetricSpec{
		Prometheus:          marshalSLOPrometheus(metric["prometheus_metric"].(*schema.Set)),
		Datadog:             nil,
		NewRelic:            nil,
		AppDynamics:         nil,
		Splunk:              nil,
		Lightstep:           nil,
		SplunkObservability: nil,
		Dynatrace:           nil,
		Elasticsearch:       nil,
		ThousandEyes:        nil,
		Graphite:            nil,
		BigQuery:            nil,
		OpenTSDB:            nil,
		GrafanaLoki:         nil,
	}
}

func marshalThresholds(d *schema.ResourceData, isRawMetric bool) []n9api.Threshold {
	objectives := d.Get("objective").(*schema.Set).List()
	thresholds := make([]n9api.Threshold, len(objectives))
	for i, o := range objectives {
		objective := o.(map[string]interface{})
		target := objective["target"].(float64)
		operator := objective["op"].(string)
		var countMetrics *n9api.CountMetricsSpec
		if !isRawMetric {
			cm := objective["count_metrics"].(*schema.Set).List()[0].(map[string]interface{})
			countMetrics = marshalCountMetrics(cm)
		}

		thresholds[i] = n9api.Threshold{
			ThresholdBase: n9api.ThresholdBase{
				DisplayName: objective["display_name"].(string),
				Value:       objective["value"].(float64),
			},
			BudgetTarget: &target,
			Operator:     &operator,
			CountMetrics: countMetrics,
		}
	}

	return thresholds
}

func marshalCountMetrics(countMetrics map[string]interface{}) *n9api.CountMetricsSpec {
	incremental := countMetrics["incremental"].(bool)
	good := countMetrics["good"].(*schema.Set).List()[0].(map[string]interface{})
	total := countMetrics["total"].(*schema.Set).List()[0].(map[string]interface{})
	return &n9api.CountMetricsSpec{
		Incremental: &incremental,
		GoodMetric:  marshalMetric(good),
		TotalMetric: marshalMetric(total),
	}
}

func marshalSLOPrometheus(s *schema.Set) *n9api.PrometheusMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["promql"].(string)
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

	var err error
	if alertPolicies, ok := object["alertPolicies"]; ok {
		err = d.Set("alert_policies", alertPolicies.([]interface{}))
		diags = appendError(diags, err)
	}

	if attachments, ok := object["attachments"]; ok {
		err = d.Set("attachments", attachments.([]interface{}))
		diags = appendError(diags, err)
	}

	spec := object["spec"].(map[string]interface{})

	budgetingMethod := spec["budgetingMethod"].(string)
	err = d.Set("budgeting_method", budgetingMethod)
	diags = appendError(diags, err)

	description := spec["description"].(string)
	err = d.Set("description", description)
	diags = appendError(diags, err)

	service := spec["service"].(string)
	err = d.Set("service", service)
	diags = appendError(diags, err)

	err = unmarshalTimeWindow(d, spec, err)
	diags = appendError(diags, err)
	isRawMetric, err := unmarshalIndicator(d, spec, diags, err)
	diags = appendError(diags, err)

	err = unmarshalObjectives(d, spec, isRawMetric)
	diags = appendError(diags, err)

	err = d.Set("alert_policies", spec["alertPolicies"])
	diags = appendError(diags, err)

	return diags
}

func unmarshalIndicator(d *schema.ResourceData, spec map[string]interface{}, diags diag.Diagnostics, err error) (bool, error) {
	indicator := spec["indicator"].(map[string]interface{})
	res := make(map[string]interface{})
	metricSource := indicator["metricSource"].(map[string]interface{})
	res["name"] = metricSource["name"]
	res["project"] = metricSource["project"]
	res["kind"] = metricSource["kind"]
	isRawMetric := false
	if rawMetric, ok := indicator["rawMetric"]; ok {
		isRawMetric = true
		tfMetric, err := unmarshalSLOMetric(rawMetric.(map[string]interface{}))
		if err != nil {
			return false, err
		}
		res["raw_metric"] = tfMetric
	}
	return isRawMetric, d.Set("indicator", schema.NewSet(oneElementSet, []interface{}{res}))
}

func unmarshalTimeWindow(d *schema.ResourceData, spec map[string]interface{}, err error) error {
	timeWindows := spec["timeWindows"].([]interface{})
	timeWindow := timeWindows[0].(map[string]interface{})
	timeWindowsTF := make(map[string]interface{})
	timeWindowsTF["count"] = timeWindow["count"]
	timeWindowsTF["is_rolling"] = timeWindow["isRolling"]
	timeWindowsTF["unit"] = timeWindow["unit"]
	timeWindowsTF["period"] = timeWindow["period"]
	// TODO handle calendar
	err = d.Set("time_window", schema.NewSet(oneElementSet, []interface{}{timeWindowsTF}))
	return err
}

func unmarshalObjectives(d *schema.ResourceData, spec map[string]interface{}, isRawMetric bool) error {
	objectives := spec["objectives"].([]interface{})
	// TODO support multiple objectives
	objective := objectives[0].(map[string]interface{})
	objectiveTF := make(map[string]interface{})
	objectiveTF["display_name"] = objective["displayName"]
	objectiveTF["op"] = objective["op"]
	objectiveTF["value"] = objective["value"]
	objectiveTF["target"] = objective["target"]
	countMetrics, ok := objective["countMetrics"]
	if isRawMetric && ok {
		return errors.New("cannot be rawMetric and countMetric at the same time")
	}
	if !isRawMetric {
		cm := countMetrics.(map[string]interface{})
		countMetricsTF := make(map[string]interface{})
		countMetricsTF["incremental"] = cm["incremental"]
		good, err := unmarshalSLOMetric(cm["good"].(map[string]interface{}))
		if err != nil {
			return err
		}
		countMetricsTF["good"] = good
		total, err := unmarshalSLOMetric(cm["total"].(map[string]interface{}))
		if err != nil {
			return err
		}
		countMetricsTF["total"] = total
		objectiveTF["count_metrics"] = schema.NewSet(oneElementSet, []interface{}{countMetricsTF})
	}

	return d.Set("objective", schema.NewSet(oneElementSet, []interface{}{objectiveTF}))
}

func unmarshalSLOMetric(spec map[string]interface{}) (*schema.Set, error) {
	supportedMetrics := []struct {
		hclName       string
		jsonName      string
		unmarshalFunc func(map[string]interface{}) map[string]interface{}
	}{
		{"prometheus_metric", "prometheus", unmarshalPrometheusMetric},
		//{"datadog_metric", "datadog"},
		//{"newrelic_metric", "newRelic"},
		//{"appdynamics_metric", "appDynamics"},
		//{"splunk_metric", "splunk"},
		//{"lightstep_metric", "lightstep"},
		//{"splunk_observability_metric", "splunkObservability"},
		//{"dynatrace_metric", "dynatrace"},
		//{"thousandeyes_metric", "thousandEyes"},
		//{"graphite_metric", "graphite"},
		//{"bigquery_metric", "bigQuery"},
		//{"opentsdb_metric", "opentsdb"},
		//{"elasticsearch_metric", "elasticsearch"},
		//{"grafana_loki_metric", "grafanaLoki"},
	}

	res := make(map[string]interface{})
	for _, name := range supportedMetrics {
		if metric, ok := spec[name.jsonName]; ok {
			tfMetric := name.unmarshalFunc(metric.(map[string]interface{}))
			res[name.hclName] = schema.NewSet(oneElementSet, []interface{}{tfMetric})
			break
		}
	}

	return schema.NewSet(oneElementSet, []interface{}{res}), nil
}

func unmarshalPrometheusMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["promql"] = metric["promql"]

	return res
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
