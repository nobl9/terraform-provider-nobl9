package nobl9

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"

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
			"composite": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Composite SLO documentation](https://docs.nobl9.com/yaml-guide/#slo)",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target": {
							Type:        schema.TypeFloat,
							Required:    true,
							Description: "Designated value",
						},
						"burn_rate_condition": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Condition when the Composite SLO’s error budget is burning.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"op": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of logical operation",
									},
									"value": {
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Burn rate value.",
									},
								},
							},
						},
					},
				},
			},
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
							Description: "Name of the metric source (agent).",
						},
						"project": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of the metric source project.",
						},
						"kind": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "Agent",
							Description: "Kind of the metric source. One of {Agent, Direct}.",
						},
					},
				},
			},
			"objective": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "[Objectives documentation](https://docs.nobl9.com/yaml-guide#objective)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count_metrics": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Compares two time series, indicating the ratio of the count of good values to total values.",
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
						"raw_metric": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Raw data is used to compare objective values.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query": schemaMetricSpec(),
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
						"time_slice_target": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Description: "Designated value for slice",
						},
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
							Description: "Period between start time and added count",
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
				DiffSuppressFunc: diffSuppressAlertPolicyNames,
			},
			"attachments": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				MaxItems:    1,
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
		Description: "[SLO configuration documentation](https://docs.nobl9.com/yaml-guide#slo)",
	}
}

func diffSuppressAlertPolicyNames(_, _, _ string, d *schema.ResourceData) bool {
	// Ignore the order of elements on alert_policy list
	old, new := d.GetChange("alert_policies")
	if old == nil && new == nil {
		return true
	}
	apOld := old.([]interface{})
	apNew := new.([]interface{})

	sort.Slice(apOld, func(i, j int) bool {
		return apOld[i].(string) < apOld[j].(string)
	})
	sort.Slice(apNew, func(i, j int) bool {
		return apNew[i].(string) < apNew[j].(string)
	})

	return equalSlices(apOld, apNew)
}

func equalSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func marshalSLO(d *schema.ResourceData) *n9api.SLO {
	indicator := marshalIndicator(d)
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
			Composite:       marshalComposite(d),
			Thresholds:      marshalThresholds(d),
			TimeWindows:     marshalTimeWindows(d),
			AlertPolicies:   toStringSlice(d.Get("alert_policies").([]interface{})),
			Attachments:     marshalAttachments(d.Get("attachments").([]interface{})),
		},
	}
}

func marshalComposite(d *schema.ResourceData) *n9api.Composite {
	compositeSet := d.Get("composite").(*schema.Set)

	if compositeSet.Len() > 0 {
		compositeTf := compositeSet.List()[0].(map[string]interface{})

		var burnRateCondition *n9api.CompositeBurnRateCondition
		burnRateConditionSet := compositeTf["burn_rate_condition"].(*schema.Set)

		if burnRateConditionSet.Len() > 0 {
			burnRateConditionTf := burnRateConditionSet.List()[0].(map[string]interface{})

			burnRateCondition = &n9api.CompositeBurnRateCondition{
				Value:    burnRateConditionTf["value"].(float64),
				Operator: burnRateConditionTf["op"].(string),
			}
		}

		return &n9api.Composite{
			BudgetTarget:      compositeTf["target"].(float64),
			BurnRateCondition: burnRateCondition,
		}
	}

	return nil
}

func marshalTimeWindows(d *schema.ResourceData) []n9api.TimeWindow {
	timeWindow := d.Get("time_window").(*schema.Set).List()[0].(map[string]interface{})

	return []n9api.TimeWindow{{
		Unit:      timeWindow["unit"].(string),
		Count:     timeWindow["count"].(int),
		IsRolling: timeWindow["is_rolling"].(bool),
		Calendar:  marshalCalendar(timeWindow),
	}}
}

func marshalAttachments(attachments []interface{}) []n9api.Attachment {
	resultConditions := make([]n9api.Attachment, len(attachments))
	for i, c := range attachments {
		attachments := c.(map[string]interface{})
		displayName := attachments["display_name"].(string)

		resultConditions[i] = n9api.Attachment{
			DisplayName: &displayName,
			URL:         attachments["url"].(string),
		}
	}

	return resultConditions
}

func marshalCalendar(c map[string]interface{}) *n9api.Calendar {
	calendars := c["calendar"].(*schema.Set).List()
	if len(calendars) == 0 {
		return nil
	}
	calendar := calendars[0].(map[string]interface{})

	return &n9api.Calendar{
		StartTime: calendar["start_time"].(string),
		TimeZone:  calendar["time_zone"].(string),
	}
}

func marshalIndicator(d *schema.ResourceData) n9api.Indicator {
	indicator := d.Get("indicator").(*schema.Set).List()[0].(map[string]interface{})

	return n9api.Indicator{
		MetricSource: &n9api.MetricSourceSpec{
			Project: indicator["project"].(string),
			Name:    indicator["name"].(string),
			Kind:    indicator["kind"].(string),
		},
	}
}

func marshalThresholds(d *schema.ResourceData) []n9api.Threshold {
	objectives := d.Get("objective").(*schema.Set).List()
	thresholds := make([]n9api.Threshold, len(objectives))
	for i, o := range objectives {
		objective := o.(map[string]interface{})
		target := objective["target"].(float64)
		timeSliceTarget := objective["time_slice_target"].(float64)
		var timeSliceTargetPtr *float64
		if timeSliceTarget != 0 {
			timeSliceTargetPtr = &timeSliceTarget
		}
		operator := objective["op"].(string)

		thresholds[i] = n9api.Threshold{
			ThresholdBase: n9api.ThresholdBase{
				DisplayName: objective["display_name"].(string),
				Value:       objective["value"].(float64),
			},
			BudgetTarget:    &target,
			TimeSliceTarget: timeSliceTargetPtr,
			Operator:        &operator,
			CountMetrics:    marshalCountMetrics(objective),
			RawMetric:       marshalRawMetric(objective),
		}
	}

	return thresholds
}

func marshalRawMetric(metricRoot map[string]interface{}) *n9api.RawMetricSpec {
	rawMetricSet := metricRoot["raw_metric"].(*schema.Set)
	if rawMetricSet.Len() == 0 {
		return nil
	}

	rawMetric := metricRoot["raw_metric"].(*schema.Set).List()[0].(map[string]interface{})
	if _, ok := rawMetric["query"]; !ok {
		return nil
	}

	metric := rawMetric["query"].(*schema.Set).List()[0].(map[string]interface{})

	return &n9api.RawMetricSpec{
		MetricQuery: marshalMetric(metric),
	}
}

func marshalCountMetrics(countMetricsTf map[string]interface{}) *n9api.CountMetricsSpec {
	countMetricsSet := countMetricsTf["count_metrics"].(*schema.Set)
	if countMetricsSet.Len() == 0 {
		return nil
	}

	countMetrics := countMetricsSet.List()[0].(map[string]interface{})

	incremental := countMetrics["incremental"].(bool)
	good := countMetrics["good"].(*schema.Set).List()[0].(map[string]interface{})
	total := countMetrics["total"].(*schema.Set).List()[0].(map[string]interface{})
	return &n9api.CountMetricsSpec{
		Incremental: &incremental,
		GoodMetric:  marshalMetric(good),
		TotalMetric: marshalMetric(total),
	}
}

func marshalMetric(metric map[string]interface{}) *n9api.MetricSpec {
	return &n9api.MetricSpec{
		Prometheus:          marshalSLOPrometheus(metric["prometheus"].(*schema.Set)),
		Datadog:             marshalSLODatadog(metric["datadog"].(*schema.Set)),
		NewRelic:            marshalSLONewRelic(metric["newrelic"].(*schema.Set)),
		AppDynamics:         marshalSLOAppDynamics(metric["appdynamics"].(*schema.Set)),
		Splunk:              marshalSLOSplunk(metric["splunk"].(*schema.Set)),
		Lightstep:           marshalSLOLightstep(metric["lightstep"].(*schema.Set)),
		SplunkObservability: marshalSLOSplunkObservability(metric["splunk_observability"].(*schema.Set)),
		Dynatrace:           marshalSLODynatrace(metric["dynatrace"].(*schema.Set)),
		Elasticsearch:       marshalSLOElasticsearch(metric["elasticsearch"].(*schema.Set)),
		ThousandEyes:        marshalSLOThousandEyes(metric["thousandeyes"].(*schema.Set)),
		Graphite:            marshalSLOGraphite(metric["graphite"].(*schema.Set)),
		BigQuery:            marshalSLOBigQuery(metric["bigquery"].(*schema.Set)),
		OpenTSDB:            marshalSLOOpenTSDB(metric["opentsdb"].(*schema.Set)),
		GrafanaLoki:         marshalSLOGrafanaLoki(metric["grafana_loki"].(*schema.Set)),
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

func marshalSLODatadog(s *schema.Set) *n9api.DatadogMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["query"].(string)
	return &n9api.DatadogMetric{
		Query: &query,
	}
}

func marshalSLONewRelic(s *schema.Set) *n9api.NewRelicMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	nrql := metric["nrql"].(string)
	return &n9api.NewRelicMetric{
		NRQL: &nrql,
	}
}

func marshalSLOAppDynamics(s *schema.Set) *n9api.AppDynamicsMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	applicationName := metric["application_name"].(string)
	metricPath := metric["metric_path"].(string)
	return &n9api.AppDynamicsMetric{
		ApplicationName: &applicationName,
		MetricPath:      &metricPath,
	}
}

func marshalSLOSplunk(s *schema.Set) *n9api.SplunkMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &n9api.SplunkMetric{
		Query: &query,
	}
}

func marshalSLOLightstep(s *schema.Set) *n9api.LightstepMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	streamID := metric["stream_id"].(string)
	typeOfData := metric["type_of_data"].(string)
	var percentile *float64
	if p := metric["percentile"].(float64); p != 0 {
		// the API does not accept percentile = 0
		// terraform sets it to 0 even if it was omitted in the .tf
		percentile = &p
	}

	return &n9api.LightstepMetric{
		StreamID:   &streamID,
		TypeOfData: &typeOfData,
		Percentile: percentile,
	}
}

func marshalSLOSplunkObservability(s *schema.Set) *n9api.SplunkObservabilityMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	program := metric["program"].(string)
	return &n9api.SplunkObservabilityMetric{
		Program: &program,
	}
}

func marshalSLODynatrace(s *schema.Set) *n9api.DynatraceMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	selector := metric["metric_selector"].(string)
	return &n9api.DynatraceMetric{
		MetricSelector: &selector,
	}
}

func marshalSLOThousandEyes(s *schema.Set) *n9api.ThousandEyesMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	testID := int64(metric["test_id"].(int))
	return &n9api.ThousandEyesMetric{
		TestID: &testID,
	}
}

func marshalSLOGraphite(s *schema.Set) *n9api.GraphiteMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	metricPath := metric["metric_path"].(string)
	return &n9api.GraphiteMetric{
		MetricPath: &metricPath,
	}
}

func marshalSLOBigQuery(s *schema.Set) *n9api.BigQueryMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	return &n9api.BigQueryMetric{
		Query:     metric["query"].(string),
		ProjectID: metric["project_id"].(string),
		Location:  metric["location"].(string),
	}
}

func marshalSLOOpenTSDB(s *schema.Set) *n9api.OpenTSDBMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &n9api.OpenTSDBMetric{
		Query: &query,
	}
}

func marshalSLOGrafanaLoki(s *schema.Set) *n9api.GrafanaLokiMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	logql := metric["logql"].(string)
	return &n9api.GrafanaLokiMetric{
		Logql: &logql,
	}
}

func marshalSLOElasticsearch(s *schema.Set) *n9api.ElasticsearchMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	index := metric["index"].(string)
	query := metric["query"].(string)
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

	err = unmarshalTimeWindow(d, spec)
	diags = appendError(diags, err)
	err = unmarshalIndicator(d, spec)
	diags = appendError(diags, err)

	err = unmarshalObjectives(d, spec)
	diags = appendError(diags, err)

	err = unmarshalComposite(d, spec)
	diags = appendError(diags, err)

	err = unmarshalAttachments(d, spec)
	diags = appendError(diags, err)

	err = d.Set("alert_policies", spec["alertPolicies"].([]interface{}))
	diags = appendError(diags, err)

	return diags
}

func unmarshalAttachments(d *schema.ResourceData, spec map[string]interface{}) error {
	if _, ok := spec["attachments"]; !ok {
		return nil
	}

	attachments := spec["attachments"].([]interface{})
	res := make([]interface{}, len(attachments))
	for i, v := range attachments {
		m := v.(map[string]interface{})
		attachment := map[string]interface{}{
			"display_name": m["displayName"],
			"url":          m["url"],
		}
		res[i] = attachment
	}

	return d.Set("attachments", res)
}

func unmarshalIndicator(d *schema.ResourceData, spec map[string]interface{}) error {
	indicator := spec["indicator"].(map[string]interface{})
	res := make(map[string]interface{})
	metricSource := indicator["metricSource"].(map[string]interface{})
	res["name"] = metricSource["name"]
	res["project"] = metricSource["project"]
	res["kind"] = metricSource["kind"]
	if rawMetric, ok := indicator["rawMetric"]; ok {
		tfMetric, err := unmarshalSLOMetric(rawMetric.(map[string]interface{}))
		if err != nil {
			return err
		}
		res["raw_metric"] = tfMetric
	}
	return d.Set("indicator", schema.NewSet(oneElementSet, []interface{}{res}))
}

func unmarshalTimeWindow(d *schema.ResourceData, spec map[string]interface{}) error {
	timeWindows := spec["timeWindows"].([]interface{})
	timeWindow := timeWindows[0].(map[string]interface{})
	timeWindowsTF := make(map[string]interface{})
	timeWindowsTF["count"] = timeWindow["count"]
	timeWindowsTF["is_rolling"] = timeWindow["isRolling"]
	timeWindowsTF["unit"] = timeWindow["unit"]
	timeWindowsTF["period"] = timeWindow["period"]
	if c, ok := timeWindow["calendar"]; ok {
		calendar := c.(map[string]interface{})
		calendarTF := make(map[string]interface{})
		calendarTF["start_time"] = calendar["startTime"].(string)
		calendarTF["time_zone"] = calendar["timeZone"].(string)
		timeWindowsTF["calendar"] = schema.NewSet(oneElementSet, []interface{}{calendarTF})
	}
	return d.Set("time_window", schema.NewSet(oneElementSet, []interface{}{timeWindowsTF}))
}

func unmarshalObjectives(d *schema.ResourceData, spec map[string]interface{}) error {
	objectives := spec["objectives"].([]interface{})
	objectivesTF := make([]interface{}, len(objectives))

	for i, o := range objectives {
		objective := o.(map[string]interface{})
		objectiveTF := make(map[string]interface{})
		objectiveTF["display_name"] = objective["displayName"]
		objectiveTF["op"] = objective["op"]
		objectiveTF["value"] = objective["value"]
		objectiveTF["target"] = objective["target"]
		objectiveTF["time_slice_target"] = objective["timeSliceTarget"]

		if countMetrics, isCountMetrics := objective["countMetrics"]; isCountMetrics {
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

		if rawMetric, isRawMetric := objective["rawMetric"]; isRawMetric {
			rm := rawMetric.(map[string]interface{})
			tfMetric, err := unmarshalSLORawMetric(rm)
			if err != nil {
				return err
			}
			objectiveTF["raw_metric"] = tfMetric
		}

		objectivesTF[i] = objectiveTF
	}
	return d.Set("objective", schema.NewSet(objectiveHash, objectivesTF))
}

func unmarshalComposite(d *schema.ResourceData, spec map[string]interface{}) error {
	if compositeSpec, isCompositeSLO := spec["composite"]; isCompositeSLO {
		composite := compositeSpec.(map[string]interface{})
		compositeTF := make(map[string]interface{})

		compositeTF["target"] = composite["target"]

		if burnRateConditionRaw, isBurnRateConditionSet := composite["burnRateCondition"]; isBurnRateConditionSet {
			burnRateCondition := burnRateConditionRaw.(map[string]interface{})
			burnRateConditionTF := make(map[string]interface{})
			burnRateConditionTF["value"] = burnRateCondition["value"]
			burnRateConditionTF["op"] = burnRateCondition["op"]
			compositeTF["burn_rate_condition"] = schema.NewSet(oneElementSet, []interface{}{burnRateConditionTF})
		}

		return d.Set("composite", schema.NewSet(oneElementSet, []interface{}{compositeTF}))
	}

	return nil
}

func objectiveHash(objective interface{}) int {
	o := objective.(map[string]interface{})
	hash := fnv.New32()
	indicator := fmt.Sprintf("%s_%s_%f_%f", o["display_name"], o["op"], o["target"], o["value"])
	_, err := hash.Write([]byte(indicator))
	if err != nil {
		panic(err)
	}
	return int(hash.Sum32())
}

func unmarshalSLORawMetric(rawMetricSource map[string]interface{}) (*schema.Set, error) {
	var rawMetricQuery *schema.Set
	var err error
	if rawMetricQuerySource, isRawMetric := rawMetricSource["query"]; isRawMetric {
		rawMetricQuery, err = unmarshalSLOMetric(rawMetricQuerySource.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}
	return schema.NewSet(oneElementSet, []interface{}{map[string]interface{}{"query": rawMetricQuery}}), nil
}

func unmarshalSLOMetric(spec map[string]interface{}) (*schema.Set, error) {
	supportedMetrics := []struct {
		hclName       string
		jsonName      string
		unmarshalFunc func(map[string]interface{}) map[string]interface{}
	}{
		{"prometheus", "prometheus", unmarshalPrometheusMetric},
		{"datadog", "datadog", unmarshalDatadogMetric},
		{"newrelic", "newRelic", unmarshalNewRelicMetric},
		{"appdynamics", "appDynamics", unmarshalAppdynamicsMetric},
		{"splunk", "splunk", unmarshalSplunkMetric},
		{"lightstep", "lightstep", unmarshalLightstepMetric},
		{"splunk_observability", "splunkObservability", unmarshalSplunkObservabilityMetric},
		{"dynatrace", "dynatrace", unmarshalDynatraceMetric},
		{"thousandeyes", "thousandEyes", unmarshalThousandeyesMetric},
		{"graphite", "graphite", unmarshalGraphiteMetric},
		{"bigquery", "bigQuery", unmarshalBigqueryMetric},
		{"opentsdb", "opentsdb", unmarshalOpentsdbMetric},
		{"elasticsearch", "elasticsearch", unmarshalElasticsearchMetric},
		{"grafana_loki", "grafanaLoki", unmarshalGrafanaLokiMetric},
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

func unmarshalDatadogMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

func unmarshalNewRelicMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["nrql"] = metric["nrql"]

	return res
}

func unmarshalAppdynamicsMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["application_name"] = metric["applicationName"]
	res["metric_path"] = metric["metricPath"]

	return res
}

func unmarshalSplunkMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

func unmarshalLightstepMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["percentile"] = metric["percentile"]
	res["stream_id"] = metric["streamId"]
	res["type_of_data"] = metric["typeOfData"]

	return res
}

func unmarshalSplunkObservabilityMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["program"] = metric["program"]

	return res
}

func unmarshalDynatraceMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["metric_selector"] = metric["metricSelector"]

	return res
}

func unmarshalThousandeyesMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["test_id"] = metric["testID"]

	return res
}

func unmarshalGraphiteMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["metric_path"] = metric["metricPath"]

	return res
}

func unmarshalBigqueryMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["location"] = metric["location"]
	res["project_id"] = metric["projectId"]
	res["query"] = metric["query"]

	return res
}

func unmarshalOpentsdbMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

func unmarshalElasticsearchMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["index"] = metric["index"]
	res["query"] = metric["query"]

	return res
}

func unmarshalGrafanaLokiMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["logql"] = metric["logql"]

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
