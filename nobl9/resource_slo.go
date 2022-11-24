package nobl9

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	n9api "github.com/nobl9/nobl9-go"
)

func resourceSLO() *schema.Resource {
	return &schema.Resource{
		Schema:        schemaSLO(),
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

func diffSuppressListStringOrder(attribute string) func(
	_, _, _ string,
	d *schema.ResourceData,
) bool {
	return func(_, _, _ string, d *schema.ResourceData) bool {
		// Ignore the order of elements on alert_policy list
		oldValue, newValue := d.GetChange(attribute)
		if oldValue == nil && newValue == nil {
			return true
		}
		apOld := oldValue.([]interface{})
		apNew := newValue.([]interface{})

		sort.Slice(apOld, func(i, j int) bool {
			return apOld[i].(string) < apOld[j].(string)
		})
		sort.Slice(apNew, func(i, j int) bool {
			return apNew[i].(string) < apNew[j].(string)
		})
		sort.Slice(apOld, func(i, j int) bool {
			return apOld[i].(string) < apOld[j].(string)
		})
		sort.Slice(apNew, func(i, j int) bool {
			return apNew[i].(string) < apNew[j].(string)
		})

		return equalSlices(apOld, apNew)
	}
}

func resourceObjective() *schema.Resource {
	res := &schema.Resource{
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
			"name": {
				Type:        schema.TypeString,
				Description: "Objective's name. This field is computed if not provided.",
				Computed:    true,
				Optional:    true,
			},
		},
	}
	return res
}

func schemaObjective() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "[Objectives documentation](https://docs.nobl9.com/yaml-guide#objective)",
		Elem:        resourceObjective(),
	}
}

func schemaSLO() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":         schemaName(),
		"display_name": schemaDisplayName(),
		"project":      schemaProject(),
		"description":  schemaDescription(),
		"label":        schemaLabels(),
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
		"objective": schemaObjective(),
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
			DiffSuppressFunc: diffSuppressListStringOrder("alert_policies"),
		},
		"attachments": {
			Type:          schema.TypeList,
			Optional:      true,
			Description:   "",
			Deprecated:    "\"attachments\" argument is deprecated use \"attachment\" instead",
			ConflictsWith: []string{"attachment"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name which is displayed for the attachment",
					},
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Url to the attachment",
					},
				},
			},
		},
		"attachment": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name which is displayed for the attachment",
					},
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Url to the attachment",
					},
				},
			},
		},
	}
}

func resourceSLOApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	slo, diags := marshalSLO(d)
	if diags.HasError() {
		return diags
	}

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
	client, ds := getClient(config, project)
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
	client, ds := getClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectSLO, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func schemaMetricSpec() *schema.Schema {
	metricSchema := map[string]*schema.Schema{}

	metricSchemaDefinitions := []map[string]*schema.Schema{
		schemaMetricAmazonPrometheus(),
		schemaMetricAppDynamics(),
		schemaMetricBigQuery(),
		schemaMetricCloudwatch(),
		schemaMetricDatadog(),
		schemaMetricDynatrace(),
		schemaMetricElasticsearch(),
		schemaMetricGCM(),
		schemaMetricGrafanaLoki(),
		schemaMetricGraphite(),
		schemaMetricInfluxDB(),
		schemaMetricInstana(),
		schemaMetricLightstep(),
		schemaMetricNewRelic(),
		schemaMetricOpenTSDB(),
		schemaMetricPingdom(),
		schemaMetricPrometheus(),
		schemaMetricRedshift(),
		schemaMetricSplunk(),
		schemaMetricSplunkObservability(),
		schemaMetricSumologic(),
		schemaMetricThousandEyes(),
	}
	for _, metricSchemaDef := range metricSchemaDefinitions {
		for agentKey, schema := range metricSchemaDef {
			metricSchema[agentKey] = schema
		}
	}

	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Configuration for metric source",
		Elem: &schema.Resource{
			Schema: metricSchema,
		},
	}
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

func marshalSLO(d *schema.ResourceData) (*n9api.SLO, diag.Diagnostics) {
	metadataHolder, diags := marshalMetadata(d)
	if diags.HasError() {
		return nil, diags
	}
	attachments, ok := d.GetOk("attachment")
	if !ok {
		attachments = d.Get("attachments")
	}

	return &n9api.SLO{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindSLO,
			MetadataHolder: metadataHolder,
		},
		Spec: n9api.SLOSpec{
			Description:     d.Get("description").(string),
			Service:         d.Get("service").(string),
			BudgetingMethod: d.Get("budgeting_method").(string),
			Indicator:       marshalIndicator(d),
			Composite:       marshalComposite(d),
			Thresholds:      marshalThresholds(d),
			TimeWindows:     marshalTimeWindows(d),
			AlertPolicies:   toStringSlice(d.Get("alert_policies").([]interface{})),
			Attachments:     marshalAttachments(attachments.([]interface{})),
		},
	}, diags
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
				Name:        objective["name"].(string),
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
		AmazonPrometheus:    marshalAmazonPrometheusMetric(metric[amazonPrometheusMetric].(*schema.Set)),
		AppDynamics:         marshalAppDynamicsMetric(metric[appDynamicsMetric].(*schema.Set)),
		BigQuery:            marshalBigQueryMetric(metric[bigQueryMetric].(*schema.Set)),
		CloudWatch:          marshalCloudWatchMetric(metric[cloudwatchMetric].(*schema.Set)),
		Datadog:             marshalDatadogMetric(metric[datadogMetric].(*schema.Set)),
		Dynatrace:           marshalDynatraceMetric(metric[dynatraceMetric].(*schema.Set)),
		Elasticsearch:       marshalElasticsearchMetric(metric[elasticsearchMetric].(*schema.Set)),
		GCM:                 marshalGCMMetric(metric[gcmMetric].(*schema.Set)),
		GrafanaLoki:         marshalGrafanaLokiMetric(metric[grafanaLokiMetric].(*schema.Set)),
		Graphite:            marshalGraphiteMetric(metric[graphiteMetric].(*schema.Set)),
		InfluxDB:            marshalInfluxDBMetric(metric[influxdbMetric].(*schema.Set)),
		Instana:             marshalInstanaMetric(metric[instanaMetric].(*schema.Set)),
		Lightstep:           marshalLightstepMetric(metric[lightstepMetric].(*schema.Set)),
		NewRelic:            marshalNewRelicMetric(metric[newrelicMetric].(*schema.Set)),
		OpenTSDB:            marshalOpenTSDBMetric(metric[opentsdbMetric].(*schema.Set)),
		Pingdom:             marshalPingdomMetric(metric[pingdomMetric].(*schema.Set)),
		Prometheus:          marshalPrometheusMetric(metric[prometheusMetric].(*schema.Set)),
		Redshift:            marshalRedshiftMetric(metric[redshiftMetric].(*schema.Set)),
		Splunk:              marshalSplunkMetric(metric[splunkMetric].(*schema.Set)),
		SplunkObservability: marshalSplunkObservabilityMetric(metric[splunkObservabilityMetric].(*schema.Set)),
		SumoLogic:           marshalSumologicMetric(metric[sumologicMetric].(*schema.Set)),
		ThousandEyes:        marshalThousandEyesMetric(metric[thousandeyesMetric].(*schema.Set)),
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

	// Remove this warning once SLO objective unique identifier grace period ends.
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "SLO objective unique identifier warning",
		Detail: "Nobl9 is introducing an SLO objective unique identifier to support the same value for different " +
			"SLIs in the same SLO. As such, Nobl9 is adding a name identifier to each SLO objective. " +
			"Objective names can be set now, and they'll be required once grace period ends. " +
			"For more detailed information, refer to: https://docs.nobl9.com/Features/SLO-objective-unique-identifier",
	})

	return diags
}

func unmarshalAttachments(d *schema.ResourceData, spec map[string]interface{}) error {
	attachmentTagName := getExistingAttachmentTag(spec)
	if attachmentTagName == "" {
		return nil
	}

	attachments := spec[attachmentTagName].([]interface{})
	res := make([]interface{}, len(attachments))
	for i, v := range attachments {
		m := v.(map[string]interface{})
		attachment := map[string]interface{}{
			"display_name": m["displayName"],
			"url":          m["url"],
		}
		res[i] = attachment
	}

	return d.Set("attachment", res)
}

// getExistingAttachmentTag check if used tag was deprecated or not and return one that was used.
func getExistingAttachmentTag(spec map[string]interface{}) string {
	if _, ok := spec["attachment"]; !ok {
		if _, ok = spec["attachments"]; ok {
			return ""
		}
		return "attachments"
	}
	return "attachment"
}

func unmarshalIndicator(d *schema.ResourceData, spec map[string]interface{}) error {
	indicator := spec["indicator"].(map[string]interface{})
	res := make(map[string]interface{})
	metricSource := indicator["metricSource"].(map[string]interface{})
	res["name"] = metricSource["name"]
	res["project"] = metricSource["project"]
	res["kind"] = metricSource["kind"]
	if rawMetric, ok := indicator["rawMetric"]; ok {
		tfMetric := unmarshalSLOMetric(rawMetric.(map[string]interface{}))
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
		objectiveTF["name"] = objective["name"]
		objectiveTF["display_name"] = objective["displayName"]
		objectiveTF["op"] = objective["op"]
		objectiveTF["value"] = objective["value"]
		objectiveTF["target"] = objective["target"]
		objectiveTF["time_slice_target"] = objective["timeSliceTarget"]

		if countMetrics, isCountMetrics := objective["countMetrics"]; isCountMetrics {
			cm := countMetrics.(map[string]interface{})
			countMetricsTF := make(map[string]interface{})
			countMetricsTF["incremental"] = cm["incremental"]
			good := unmarshalSLOMetric(cm["good"].(map[string]interface{}))
			countMetricsTF["good"] = good
			total := unmarshalSLOMetric(cm["total"].(map[string]interface{}))
			countMetricsTF["total"] = total
			objectiveTF["count_metrics"] = schema.NewSet(oneElementSet, []interface{}{countMetricsTF})
		}

		if rawMetric, isRawMetric := objective["rawMetric"]; isRawMetric {
			rm := rawMetric.(map[string]interface{})
			tfMetric := unmarshalSLORawMetric(rm)
			objectiveTF["raw_metric"] = tfMetric
		}

		objectivesTF[i] = objectiveTF
	}
	return d.Set("objective", schema.NewSet(objectiveHash, objectivesTF))
}

func objectiveHash(objective interface{}) int {
	o := objective.(map[string]interface{})
	indicator := fmt.Sprintf("%s_%s_%s_%f_%f_%f",
		o["name"],
		o["display_name"],
		o["op"],
		o["value"],
		o["target"],
		o["time_slice_target"],
	)
	return schema.HashString(indicator)
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

func unmarshalSLORawMetric(rawMetricSource map[string]interface{}) *schema.Set {
	var rawMetricQuery *schema.Set
	if rawMetricQuerySource, isRawMetric := rawMetricSource["query"]; isRawMetric {
		rawMetricQuery = unmarshalSLOMetric(rawMetricQuerySource.(map[string]interface{}))
	}
	return schema.NewSet(oneElementSet, []interface{}{map[string]interface{}{"query": rawMetricQuery}})
}

func unmarshalSLOMetric(spec map[string]interface{}) *schema.Set {
	supportedMetrics := []struct {
		hclName       string
		jsonName      string
		unmarshalFunc func(map[string]interface{}) map[string]interface{}
	}{
		{amazonPrometheusMetric, "amazonPrometheus", unmarshalAmazonPrometheusMetric},
		{appDynamicsMetric, "appDynamics", unmarshalAppdynamicsMetric},
		{bigQueryMetric, "bigQuery", unmarshalBigqueryMetric},
		{cloudwatchMetric, "cloudWatch", unmarshalCloudWatchMetric},
		{datadogMetric, "datadog", unmarshalDatadogMetric},
		{dynatraceMetric, "dynatrace", unmarshalDynatraceMetric},
		{elasticsearchMetric, "elasticsearch", unmarshalElasticsearchMetric},
		{gcmMetric, "gcm", unmarshalGCMMetric},
		{grafanaLokiMetric, "grafanaLoki", unmarshalGrafanaLokiMetric},
		{graphiteMetric, "graphite", unmarshalGraphiteMetric},
		{influxdbMetric, "influxdb", unmarshalInfluxDBMetric},
		{instanaMetric, "instana", unmarshalInstanaMetric},
		{lightstepMetric, "lightstep", unmarshalLightstepMetric},
		{newrelicMetric, "newRelic", unmarshalNewRelicMetric},
		{opentsdbMetric, "opentsdb", unmarshalOpentsdbMetric},
		{pingdomMetric, "pingdom", unmarshalPingdomMetric},
		{prometheusMetric, "prometheus", unmarshalPrometheusMetric},
		{redshiftMetric, "redshift", unmarshalRedshiftMetric},
		{splunkMetric, "splunk", unmarshalSplunkMetric},
		{splunkObservabilityMetric, "splunkObservability", unmarshalSplunkObservabilityMetric},
		{sumologicMetric, "sumoLogic", unmarshalSumologicMetric},
		{thousandeyesMetric, "thousandEyes", unmarshalThousandeyesMetric},
	}

	res := make(map[string]interface{})
	for _, name := range supportedMetrics {
		if metric, ok := spec[name.jsonName]; ok {
			tfMetric := name.unmarshalFunc(metric.(map[string]interface{}))
			res[name.hclName] = schema.NewSet(oneElementSet, []interface{}{tfMetric})
			break
		}
	}

	return schema.NewSet(oneElementSet, []interface{}{res})
}

/**
 * Amazon Prometheus Metric
 * https://docs.nobl9.com/Sources/Amazon_Prometheus/#creating-slos-with-ams-prometheus
 */
const amazonPrometheusMetric = "amazon_prometheus"

func schemaMetricAmazonPrometheus() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		amazonPrometheusMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/Amazon_Prometheus/#creating-slos-with-ams-prometheus)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"promql": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalAmazonPrometheusMetric(s *schema.Set) *n9api.AmazonPrometheusMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["promql"].(string)
	return &n9api.AmazonPrometheusMetric{
		PromQL: &query,
	}
}

func unmarshalAmazonPrometheusMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["promql"] = metric["promql"]

	return res
}

/**
 * AppDynamics Metric
 * https://docs.nobl9.com/Sources/appdynamics#creating-slos-with-appdynamics
 */
const appDynamicsMetric = "appdynamics"

func schemaMetricAppDynamics() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		appDynamicsMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/appdynamics#creating-slos-with-appdynamics)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"application_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the added application",
					},
					"metric_path": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Path to the metrics",
					},
				},
			},
		},
	}
}

func marshalAppDynamicsMetric(s *schema.Set) *n9api.AppDynamicsMetric {
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

func unmarshalAppdynamicsMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["application_name"] = metric["applicationName"]
	res["metric_path"] = metric["metricPath"]

	return res
}

/**
 * BigQuery Metric
 * https://docs.nobl9.com/Sources/bigquery#creating-slos-with-bigquery
 */
const bigQueryMetric = "bigquery"

func schemaMetricBigQuery() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		bigQueryMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/bigquery#creating-slos-with-bigquery)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"location": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Location of you BigQuery",
					},
					"project_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Project ID",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalBigQueryMetric(s *schema.Set) *n9api.BigQueryMetric {
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

func unmarshalBigqueryMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["location"] = metric["location"]
	res["project_id"] = metric["projectId"]
	res["query"] = metric["query"]

	return res
}

/**
 * Amazon CloudWatch Metric
 * https://docs.nobl9.com/Sources/Amazon_CloudWatch/#creating-slos-with-cloudwatch
 */
const cloudwatchMetric = "cloudwatch"

func schemaMetricCloudwatch() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		cloudwatchMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/Amazon_CloudWatch/#creating-slos-with-cloudwatch)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Region of the CloudWatch instance",
					},
					"namespace": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Namespace of the metric",
					},
					"metric_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Metric name",
					},
					"stat": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Metric data aggregations",
					},
					"sql": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "SQL query",
					},
					"json": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "JSON query",
					},
					"dimensions": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Set of name/value pairs that is a part of the identity of a metric",
						MinItems:    1,
						MaxItems:    10,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Name",
								},
								"value": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Value",
								},
							},
						},
					},
				},
			},
		},
	}
}

func marshalCloudWatchMetric(s *schema.Set) *n9api.CloudWatchMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	region := metric["region"].(string)

	var namespace *string
	if value := metric["namespace"].(string); value != "" {
		namespace = &value
	}

	var metricName *string
	if value := metric["metric_name"].(string); value != "" {
		metricName = &value
	}

	var stat *string
	if value := metric["stat"].(string); value != "" {
		stat = &value
	}

	var sql *string
	if value := metric["sql"].(string); value != "" {
		sql = &value
	}

	var json *string
	if value := metric["json"].(string); value != "" {
		json = &value
	}

	dimensions := metric["dimensions"].(*schema.Set)
	var metricDimensions []n9api.CloudWatchMetricDimension

	if dimensions.Len() > 0 {
		metricDimensions = make([]n9api.CloudWatchMetricDimension, dimensions.Len())
	}

	for idx, dimension := range dimensions.List() {
		n9Dimension := dimension.(map[string]interface{})
		name := n9Dimension["name"].(string)
		value := n9Dimension["value"].(string)

		metricDimensions[idx] = n9api.CloudWatchMetricDimension{
			Name:  &name,
			Value: &value,
		}
	}

	return &n9api.CloudWatchMetric{
		Region:     &region,
		Namespace:  namespace,
		MetricName: metricName,
		Stat:       stat,
		Dimensions: metricDimensions,
		SQL:        sql,
		JSON:       json,
	}
}

func unmarshalCloudWatchMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["region"] = metric["region"]
	res["namespace"] = metric["namespace"]
	res["metric_name"] = metric["metricName"]
	res["stat"] = metric["stat"]
	res["sql"] = metric["sql"]
	res["json"] = metric["json"]
	res["dimensions"] = metric["dimensions"]

	return res
}

/**
 * Datadog Metric
 * https://docs.nobl9.com/Sources/datadog#creating-slos-with-datadog
 */
const datadogMetric = "datadog"

func schemaMetricDatadog() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		datadogMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/datadog#creating-slos-with-datadog)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalDatadogMetric(s *schema.Set) *n9api.DatadogMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["query"].(string)
	return &n9api.DatadogMetric{
		Query: &query,
	}
}

func unmarshalDatadogMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

/**
 * Dynatrace Metric
 * https://docs.nobl9.com/Sources/dynatrace#creating-slos-with-dynatrace)
 */
const dynatraceMetric = "dynatrace"

func schemaMetricDynatrace() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		dynatraceMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/dynatrace#creating-slos-with-dynatrace)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"metric_selector": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Selector for the metrics",
					},
				},
			},
		},
	}
}

func marshalDynatraceMetric(s *schema.Set) *n9api.DynatraceMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	selector := metric["metric_selector"].(string)
	return &n9api.DynatraceMetric{
		MetricSelector: &selector,
	}
}

func unmarshalDynatraceMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["metric_selector"] = metric["metricSelector"]

	return res
}

/**
 * Elasticsearch Metric
 * https://docs.nobl9.com/Sources/elasticsearch#creating-slos-with-elasticsearch
 */
const elasticsearchMetric = "elasticsearch"

func schemaMetricElasticsearch() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		elasticsearchMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/elasticsearch#creating-slos-with-elasticsearch)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"index": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Index of metrics we want to query",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalElasticsearchMetric(s *schema.Set) *n9api.ElasticsearchMetric {
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

func unmarshalElasticsearchMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["index"] = metric["index"]
	res["query"] = metric["query"]

	return res
}

/**
 * Google Cloud Monitoring (GCM) Metric
 * https://docs.nobl9.com/Sources/google-cloud-monitoring#creating-slos-with-google-cloud-monitoring
 */
const gcmMetric = "gcm"

func schemaMetricGCM() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		gcmMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/google-cloud-monitoring#creating-slos-with-google-cloud-monitoring)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"project_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Project ID",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalGCMMetric(s *schema.Set) *n9api.GCMMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	return &n9api.GCMMetric{
		ProjectID: metric["project_id"].(string),
		Query:     metric["query"].(string),
	}
}

func unmarshalGCMMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["project_id"] = metric["projectId"]
	res["query"] = metric["query"]

	return res
}

/**
 * Grafana Loki Metric
 * https://docs.nobl9.com/Sources/grafana-loki#creating-slos-with-grafana-loki
 */
const grafanaLokiMetric = "grafana_loki"

func schemaMetricGrafanaLoki() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		grafanaLokiMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/grafana-loki#creating-slos-with-grafana-loki)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"logql": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the logs",
					},
				},
			},
		},
	}
}

func marshalGrafanaLokiMetric(s *schema.Set) *n9api.GrafanaLokiMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	logql := metric["logql"].(string)
	return &n9api.GrafanaLokiMetric{
		Logql: &logql,
	}
}

func unmarshalGrafanaLokiMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["logql"] = metric["logql"]

	return res
}

/**
 * Graphite Metric
 * https://docs.nobl9.com/Sources/graphite#creating-slos-with-graphite
 */
const graphiteMetric = "graphite"

func schemaMetricGraphite() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		graphiteMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/graphite#creating-slos-with-graphite)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"metric_path": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Path to the metrics",
					},
				},
			},
		},
	}
}

func marshalGraphiteMetric(s *schema.Set) *n9api.GraphiteMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	metricPath := metric["metric_path"].(string)
	return &n9api.GraphiteMetric{
		MetricPath: &metricPath,
	}
}

func unmarshalGraphiteMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["metric_path"] = metric["metricPath"]

	return res
}

/**
 * InfluxDB Metric
 * https://docs.nobl9.com/Sources/influxdb#creating-slos-with-influxdb
 */
const influxdbMetric = "influxdb"

func schemaMetricInfluxDB() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		influxdbMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/influxdb#creating-slos-with-influxdb)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalInfluxDBMetric(s *schema.Set) *n9api.InfluxDBMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &n9api.InfluxDBMetric{
		Query: &query,
	}
}

func unmarshalInfluxDBMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

/**
 * Instana Metric
 * https://docs.nobl9.com/Sources/instana#creating-slos-with-instana
 */
const instanaMetric = "instana"

func schemaMetricInstana() map[string]*schema.Schema {
	validateMetricType := func(v any, p cty.Path) diag.Diagnostics {
		const appType = "application"
		const infraType = "infrastructure"
		value := v.(string)
		var diags diag.Diagnostics
		if value != appType && value != infraType {
			diag := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "wrong value",
				Detail:   fmt.Sprintf("%q is not %q or %q", value, appType, infraType),
			}
			diags = append(diags, diag)
		}
		return diags
	}

	return map[string]*schema.Schema{
		instanaMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/instana#creating-slos-with-instana)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"metric_type": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "Instana metric type 'application' or 'infrastructure'",
						ValidateDiagFunc: validateMetricType,
					},
					"infrastructure": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Infrastructure metric type",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"metric_retrieval_method": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Metric retrieval method 'query' or 'snapshot'",
							},
							"query": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Query for the metrics",
							},
							"snapshot_id": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Snapshot ID",
							},
							"metric_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Metric ID",
							},
							"plugin_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Plugin ID",
							},
						}},
					},
					"application": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Infrastructure metric type",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"metric_id": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Metric ID one of 'calls', 'erroneousCalls', 'errors', 'latency'",
								},
								"aggregation": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Depends on the value specified for 'metric_id'- more info in N9 docs",
								},
								"group_by": {
									Type:        schema.TypeSet,
									Required:    true,
									Description: "Group by method",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"tag": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Group by tag",
											},
											"tag_entity": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Tag entity - one of 'DESTINATION', 'SOURCE', 'NOT_APPLICABLE'",
											},
											"tag_second_level_key": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
								"api_query": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "API query user passes in a JSON format",
								},
								"include_internal": {
									Type:        schema.TypeBool,
									Optional:    true,
									Description: "Include internal",
								},
								"include_synthetic": {
									Type:        schema.TypeBool,
									Optional:    true,
									Description: "Include synthetic",
								},
							}},
					},
				},
			},
		},
	}
}

func marshalInstanaMetric(s *schema.Set) *n9api.InstanaMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	return &n9api.InstanaMetric{
		MetricType:     metric["metric_type"].(string),
		Infrastructure: marshalInstanaInfrastructureMetric(metric["infrastructure"].(*schema.Set)),
		Application:    marshalInstanaApplicationMetric(metric["application"].(*schema.Set)),
	}
}

func marshalInstanaInfrastructureMetric(s *schema.Set) *n9api.InstanaInfrastructureMetricType {
	if s.Len() == 0 {
		return nil
	}
	infrastructure := s.List()[0].(map[string]interface{})

	var query *string
	if value := infrastructure["query"].(string); value != "" {
		query = &value
	}
	var snapshotID *string
	if value := infrastructure["snapshot_id"].(string); value != "" {
		snapshotID = &value
	}

	return &n9api.InstanaInfrastructureMetricType{
		MetricRetrievalMethod: infrastructure["metric_retrieval_method"].(string),
		Query:                 query,
		SnapshotID:            snapshotID,
		MetricID:              infrastructure["metric_id"].(string),
		PluginID:              infrastructure["plugin_id"].(string),
	}
}

func marshalInstanaApplicationMetric(s *schema.Set) *n9api.InstanaApplicationMetricType {
	if s.Len() == 0 {
		return nil
	}
	application := s.List()[0].(map[string]interface{})

	var includeInternal *bool
	if value, ok := application["include_internal"].(bool); ok {
		includeInternal = &value
	}

	var includeSynthetic *bool
	if value, ok := application["include_synthetic"].(bool); ok {
		includeSynthetic = &value
	}

	var groupBy = application["group_by"].(*schema.Set).List()[0].(map[string]interface{})
	var tagSecondLevelKey *string
	if value, ok := groupBy["tag_second_level_key"].(string); ok && value != "" {
		tagSecondLevelKey = &value
	}

	return &n9api.InstanaApplicationMetricType{
		MetricID:    application["metric_id"].(string),
		Aggregation: application["aggregation"].(string),
		GroupBy: n9api.InstanaApplicationMetricGroupBy{
			Tag:               groupBy["tag"].(string),
			TagEntity:         groupBy["tag_entity"].(string),
			TagSecondLevelKey: tagSecondLevelKey,
		},
		APIQuery:         application["api_query"].(string),
		IncludeInternal:  includeInternal,
		IncludeSynthetic: includeSynthetic,
	}
}

func unmarshalInstanaMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["metric_type"] = metric["metricType"]
	res["infrastructure"] = unmarshalInstanaInfrastructureMetric(metric)
	res["application"] = unmarshalInstanaApplicationMetric(metric)

	return res
}

func unmarshalInstanaInfrastructureMetric(metric map[string]interface{}) *schema.Set {
	if infrastructure, ok := metric["infrastructure"].(map[string]interface{}); ok {
		infrastructureTF := map[string]interface{}{
			"metric_retrieval_method": infrastructure["metricRetrievalMethod"],
			"query":                   infrastructure["query"],
			"snapshot_id":             infrastructure["snapshotId"],
			"metric_id":               infrastructure["metricId"],
			"plugin_id":               infrastructure["pluginId"],
		}
		return schema.NewSet(oneElementSet, []interface{}{infrastructureTF})
	}

	return nil
}

func unmarshalInstanaApplicationMetric(metric map[string]interface{}) *schema.Set {
	if application, ok := metric["application"].(map[string]interface{}); ok {
		applicationTF := map[string]interface{}{
			"metric_id":   application["metricId"],
			"aggregation": application["aggregation"],
			"group_by": schema.NewSet(oneElementSet, []interface{}{map[string]interface{}{
				"tag":                  application["groupBy"].(map[string]interface{})["tag"],
				"tag_entity":           application["groupBy"].(map[string]interface{})["tagEntity"],
				"tag_second_level_key": application["groupBy"].(map[string]interface{})["tagSecondLevelKey"]}}),
			"api_query":         application["apiQuery"],
			"include_internal":  application["includeInternal"],
			"include_synthetic": application["includeSynthetic"],
		}
		return schema.NewSet(oneElementSet, []interface{}{applicationTF})
	}

	return nil
}

/**
 * Lightstep Metric
 * https://docs.nobl9.com/Sources/lightstep#creating-slos-with-lightstep
 */
const lightstepMetric = "lightstep"

func schemaMetricLightstep() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		lightstepMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/lightstep#creating-slos-with-lightstep)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"percentile": {
						Type:        schema.TypeFloat,
						Optional:    true,
						Description: "Optional value to filter by percentiles",
					},
					"stream_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "ID of the metrics stream",
					},
					"type_of_data": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Type of data to filter by",
					},
				},
			},
		},
	}
}

func marshalLightstepMetric(s *schema.Set) *n9api.LightstepMetric {
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

func unmarshalLightstepMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["percentile"] = metric["percentile"]
	res["stream_id"] = metric["streamId"]
	res["type_of_data"] = metric["typeOfData"]

	return res
}

/**
 * New Relic Metric
 * https://docs.nobl9.com/Sources/new-relic#creating-slos-with-new-relic
 */
const newrelicMetric = "newrelic"

func schemaMetricNewRelic() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		newrelicMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/new-relic#creating-slos-with-new-relic)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"nrql": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalNewRelicMetric(s *schema.Set) *n9api.NewRelicMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	nrql := metric["nrql"].(string)
	return &n9api.NewRelicMetric{
		NRQL: &nrql,
	}
}

func unmarshalNewRelicMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["nrql"] = metric["nrql"]

	return res
}

/**
 * OpenTSDB Metric
 * https://docs.nobl9.com/Sources/opentsdb#creating-slos-with-opentsdb
 */
const opentsdbMetric = "opentsdb"

func schemaMetricOpenTSDB() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		opentsdbMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/opentsdb#creating-slos-with-opentsdb)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalOpenTSDBMetric(s *schema.Set) *n9api.OpenTSDBMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &n9api.OpenTSDBMetric{
		Query: &query,
	}
}

func unmarshalOpentsdbMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

/**
 * Pingdom Metric
 * https://docs.nobl9.com/Sources/pingdom#creating-slos-with-pingdom
 */
const pingdomMetric = "pingdom"

func schemaMetricPingdom() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		pingdomMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/pingdom#creating-slos-with-pingdom)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"check_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Pingdom uptime or transaction check's ID",
					},
					"check_type": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Pingdom check type - uptime or transaction",
					},
					"status": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional for the Uptime checks. Use it to filter the Pingdom check results by status",
					},
				},
			},
		},
	}
}

func marshalPingdomMetric(s *schema.Set) *n9api.PingdomMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	var checkID *string
	if value, ok := metric["check_id"].(string); ok && value != "" {
		checkID = &value
	}
	var checkType *string
	if value, ok := metric["check_type"].(string); ok && value != "" {
		checkType = &value
	}
	var status *string
	if value, ok := metric["status"].(string); ok && value != "" {
		status = &value
	}
	return &n9api.PingdomMetric{
		CheckID:   checkID,
		CheckType: checkType,
		Status:    status,
	}
}

func unmarshalPingdomMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["check_id"] = metric["checkId"]
	res["check_type"] = metric["checkType"]
	res["status"] = metric["status"]

	return res
}

/**
 * Prometheus Metric
 * https://docs.nobl9.com/Sources/prometheus#creating-slos-with-prometheus
 */
const prometheusMetric = "prometheus"

func schemaMetricPrometheus() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		prometheusMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/prometheus#creating-slos-with-prometheus)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"promql": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalPrometheusMetric(s *schema.Set) *n9api.PrometheusMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["promql"].(string)
	return &n9api.PrometheusMetric{
		PromQL: &query,
	}
}

func unmarshalPrometheusMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["promql"] = metric["promql"]

	return res
}

/**
 * Amazon Redshift Metric
 * https://docs.nobl9.com/Sources/Amazon_Redshift/#creating-slos-with-amazon-redshift
 */
const redshiftMetric = "redshift"

func schemaMetricRedshift() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		redshiftMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/Amazon_Redshift/#creating-slos-with-amazon-redshift)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Region of the Redshift instance",
					},
					"cluster_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Redshift custer ID",
					},
					"database_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Database name",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalRedshiftMetric(s *schema.Set) *n9api.RedshiftMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	region := metric["region"].(string)
	clusterID := metric["cluster_id"].(string)
	databaseName := metric["database_name"].(string)
	query := metric["query"].(string)

	return &n9api.RedshiftMetric{
		Region:       &region,
		ClusterID:    &clusterID,
		DatabaseName: &databaseName,
		Query:        &query,
	}
}

func unmarshalRedshiftMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["region"] = metric["region"]
	res["cluster_id"] = metric["clusterId"]
	res["database_name"] = metric["databaseName"]
	res["query"] = metric["query"]

	return res
}

/**
 * Splunk Metric
 * https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk
 */
const splunkMetric = "splunk"

func schemaMetricSplunk() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		splunkMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalSplunkMetric(s *schema.Set) *n9api.SplunkMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &n9api.SplunkMetric{
		Query: &query,
	}
}

func unmarshalSplunkMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["query"] = metric["query"]

	return res
}

/**
 * Splunk Observability Metric
 * https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk-observability
 */
const splunkObservabilityMetric = "splunk_observability"

func schemaMetricSplunkObservability() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		splunkObservabilityMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk-observability)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"program": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func marshalSplunkObservabilityMetric(s *schema.Set) *n9api.SplunkObservabilityMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	program := metric["program"].(string)
	return &n9api.SplunkObservabilityMetric{
		Program: &program,
	}
}

func unmarshalSplunkObservabilityMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["program"] = metric["program"]

	return res
}

/**
 * Sumo Logic Metric
 * https://docs.nobl9.com/Sources/sumo-logic#creating-slos-with-sumo-logic
 */
const sumologicMetric = "sumologic"

func schemaMetricSumologic() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		sumologicMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/sumo-logic#creating-slos-with-sumo-logic)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Sumologic source - metrics or logs",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for the metrics",
					},
					"rollup": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Aggregation function - avg, sum, min, max, count, none",
					},
					"quantization": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Period of data aggregation",
					},
				},
			},
		},
	}
}

func marshalSumologicMetric(s *schema.Set) *n9api.SumoLogicMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	metricType := metric["type"].(string)
	query := metric["query"].(string)
	var quantization *string
	if value, ok := metric["quantization"].(string); ok && value != "" {
		quantization = &value
	}
	var rollup *string
	if value, ok := metric["rollup"].(string); ok && value != "" {
		rollup = &value
	}
	return &n9api.SumoLogicMetric{
		Type:         &metricType,
		Query:        &query,
		Quantization: quantization,
		Rollup:       rollup,
	}
}

func unmarshalSumologicMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["type"] = metric["type"]
	res["query"] = metric["query"]
	res["quantization"] = metric["quantization"]
	res["rollup"] = metric["rollup"]

	return res
}

/**
 * ThousandEyes Metric
 * https://docs.nobl9.com/Sources/thousandeyes#creating-slos-with-thousandeyes
 */
const thousandeyesMetric = "thousandeyes"

func schemaMetricThousandEyes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		thousandeyesMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/thousandeyes#creating-slos-with-thousandeyes)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"test_id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "ID of the test",
					},
				},
			},
		},
	}
}

func marshalThousandEyesMetric(s *schema.Set) *n9api.ThousandEyesMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	testID := int64(metric["test_id"].(int))
	return &n9api.ThousandEyesMetric{
		TestID: &testID,
	}
}

func unmarshalThousandeyesMetric(metric map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	res["test_id"] = metric["testID"]

	return res
}
