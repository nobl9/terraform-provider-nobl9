package nobl9

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"

	"github.com/nobl9/nobl9-go/manifest"
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
				Type:     schema.TypeSet,
				Optional: true,
				Description: "Compares two time series, calculating the ratio of either good or bad values to the" +
					" total number of values. Fill either the 'good' or 'bad' series, but not both.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"good": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Configuration for good time series metrics.",
							Elem:        schemaMetricSpec(),
						},
						"bad": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Configuration for bad time series metrics.",
							Elem:        schemaMetricSpec(),
						},
						"total": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Configuration for metric source.",
							Elem:        schemaMetricSpec(),
						},
						"incremental": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Should the metrics be incrementing or not.",
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
						"query": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Configuration for metric source.",
							Elem:        schemaMetricSpec(),
						},
					},
				},
			},
			"composite": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "An assembly of objectives from different SLOs reflecting their combined performance.",
				Elem:        resourceComposite(),
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name to be displayed.",
			},
			"op": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of logical operation.",
			},
			"target": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Designated value.",
			},
			"time_slice_target": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Designated value for slice.",
			},
			"value": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Value.",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Objective's name. This field is computed if not provided.",
				Computed:    true,
				Optional:    true,
			},
			"primary": {
				Type:        schema.TypeBool,
				Description: "Is objective marked as primary.",
				Optional:    true,
			},
		},
	}
	return res
}

func resourceComposite() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_delay": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Maximum time for your composite SLO to wait for data from objectives.",
			},
			"components": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Objectives to be assembled in your composite SLO.",
				Elem:        resourceCompositeComponents(),
			},
		},
	}
}

func resourceCompositeComponents() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"objectives": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "An additional nesting for the components of your composite SLO.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"composite_objective": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Your composite SLO component.",
							Elem:        resourceCompositeObjective(),
						},
					},
				},
			},
		},
	}
}

func resourceCompositeObjective() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project name.",
			},
			"slo": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SLO name.",
			},
			"objective": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SLO objective name.",
			},
			"weight": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Weights determine each component’s contribution to the composite SLO.",
			},
			"when_delayed": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Defines how to treat missing component data on `max_delay` expiry.",
				ValidateFunc: validation.StringInSlice(v1alphaSLO.WhenDelayedNames(), false),
			},
		},
	}
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
		"annotations":  schemaAnnotations(),
		"composite": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Composite SLO documentation](https://docs.nobl9.com/yaml-guide/#slo)",
			Deprecated:  "\"composite\" is deprecated, use \"objective.composite\" instead.",
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"target": {
						Type:        schema.TypeFloat,
						Required:    true,
						Description: "Designated value.",
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
									Description: "Type of logical operation.",
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
			Description: "Method which will be use to calculate budget.",
		},
		"service": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the service.",
		},
		"indicator": {
			Type:        schema.TypeSet,
			Optional:    true,
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
						Description: "Alert Policies attached to SLO.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"start_time": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Date of the start.",
								},
								"time_zone": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Timezone name in IANA Time Zone Database.",
								},
							},
						},
					},
					"count": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Count of the time unit.",
					},
					"is_rolling": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Is the window moving or not.",
					},
					"period": {
						Type:        schema.TypeMap,
						Computed:    true,
						Description: "Period between start time and added count.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"unit": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Unit of time.",
					},
				},
			},
		},
		"alert_policies": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Alert Policies attached to SLO.",
			Elem: &schema.Schema{
				Type:        schema.TypeString,
				Description: "Alert Policy.",
			},
			DiffSuppressFunc: diffSuppressListStringOrder("alert_policies"),
		},
		"attachments": {
			Type:          schema.TypeList,
			Optional:      true,
			Description:   "",
			MaxItems:      20,
			Deprecated:    "\"attachments\" argument is deprecated use \"attachment\" instead.",
			ConflictsWith: []string{"attachment"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validateMaxLength("display_name", 63),
						Description:      "Name displayed for the attachment. Max. length: 63 characters.",
					},
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "URL to the attachment.",
					},
				},
			},
		},
		"attachment": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "",
			MaxItems:    20,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validateMaxLength("display_name", 63),
						Description:      "Name displayed for the attachment. Max. length: 63 characters.",
					},
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "URL to the attachment.",
					},
				},
			},
		},
		"anomaly_config": schemaAnomalyConfig(),
	}
}

func resourceSLOApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	slo, diags := marshalSLO(d)
	if diags.HasError() {
		return diags
	}
	resultSlo := manifest.SetDefaultProject([]manifest.Object{slo}, config.Project)

	if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *retry.RetryError {
		err := client.Objects().V1().Apply(ctx, resultSlo)
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(slo.Metadata.Name)
	return resourceSLORead(ctx, d, meta)
}

func resourceSLORead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	slos, err := client.Objects().V1().GetV1alphaSLOs(ctx, v1Objects.GetSLOsRequest{
		Project: project,
		Names:   []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalSLO(d, slos)
}

func resourceSLODelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}

	if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *retry.RetryError {
		err := client.Objects().V1().DeleteByName(ctx, manifest.KindSLO, project, d.Id())
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func schemaMetricSpec() *schema.Resource {
	metricSchemaDefinitions := []map[string]*schema.Schema{
		schemaMetricAmazonPrometheus(),
		schemaMetricAppDynamics(),
		schemaMetricAzureMonitor(),
		schemaMetricBigQuery(),
		schemaMetricCloudwatch(),
		schemaMetricDatadog(),
		schemaMetricDynatrace(),
		schemaMetricElasticsearch(),
		schemaMetricGCM(),
		schemaMetricGrafanaLoki(),
		schemaMetricGraphite(),
		schemaMetricHoneycomb(),
		schemaMetricInfluxDB(),
		schemaMetricInstana(),
		schemaMetricLightstep(),
		schemaLogicMonitorMetric(),
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
	metricSchema := make(map[string]*schema.Schema, len(metricSchemaDefinitions))
	for _, metricSchemaDef := range metricSchemaDefinitions {
		for agentKey, sch := range metricSchemaDef {
			metricSchema[agentKey] = sch
		}
	}

	return &schema.Resource{
		Schema: metricSchema,
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

func marshalSLO(d *schema.ResourceData) (*v1alphaSLO.SLO, diag.Diagnostics) {
	attachments, ok := d.GetOk("attachment")
	if !ok {
		attachments = d.Get("attachments")
	}
	displayName, _ := d.Get("display_name").(string)
	labelsMarshaled, diags := getMarshaledLabels(d)
	if diags.HasError() {
		return nil, diags
	}
	objectiveSpec, diags := marshalObjectives(d)
	if diags.HasError() {
		return nil, diags
	}
	annotationsMarshaled := getMarshaledAnnotations(d)

	slo := v1alphaSLO.New(
		v1alphaSLO.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: displayName,
			Project:     d.Get("project").(string),
			Labels:      labelsMarshaled,
			Annotations: annotationsMarshaled,
		},
		v1alphaSLO.Spec{
			Description:     d.Get("description").(string),
			Service:         d.Get("service").(string),
			BudgetingMethod: d.Get("budgeting_method").(string),
			Indicator:       marshalIndicator(d),
			Composite:       marshalCompositeDeprecated(d),
			Objectives:      objectiveSpec,
			TimeWindows:     marshalTimeWindows(d),
			AlertPolicies:   toStringSlice(d.Get("alert_policies").([]interface{})),
			Attachments:     marshalAttachments(attachments.([]interface{})),
			AnomalyConfig:   marshalAnomalyConfig(d.Get("anomaly_config")),
		})
	return &slo, diags
}

func marshalCompositeDeprecated(d *schema.ResourceData) *v1alphaSLO.Composite {
	compositeSet := d.Get("composite").(*schema.Set)

	if compositeSet.Len() > 0 {
		compositeTf := compositeSet.List()[0].(map[string]interface{})

		var burnRateCondition *v1alphaSLO.CompositeBurnRateCondition
		burnRateConditionSet := compositeTf["burn_rate_condition"].(*schema.Set)

		if burnRateConditionSet.Len() > 0 {
			burnRateConditionTf := burnRateConditionSet.List()[0].(map[string]interface{})

			burnRateCondition = &v1alphaSLO.CompositeBurnRateCondition{
				Value:    burnRateConditionTf["value"].(float64),
				Operator: burnRateConditionTf["op"].(string),
			}
		}

		budgetTarget := compositeTf["target"].(float64)
		return &v1alphaSLO.Composite{
			BudgetTarget:      &budgetTarget,
			BurnRateCondition: burnRateCondition,
		}
	}

	return nil
}

func marshalComposite(sloObjective map[string]interface{}) (*v1alphaSLO.CompositeSpec, error) {
	compositeSet := sloObjective["composite"].(*schema.Set)
	if compositeSet.Len() == 0 {
		return nil, nil
	}
	compositeMap := compositeSet.List()[0].(map[string]interface{})
	maxDelay := compositeMap["max_delay"].(string)

	componentsSet := compositeMap["components"].(*schema.Set)
	resultObjectives := make([]v1alphaSLO.CompositeObjective, 0)
	if len(componentsSet.List()) > 0 {
		componentsMap := componentsSet.List()[0].(map[string]interface{})
		objectivesSet := componentsMap["objectives"].(*schema.Set)
		compositeObjectivesMap := objectivesSet.List()[0].(map[string]interface{})
		compositeObjectives := compositeObjectivesMap["composite_objective"].([]interface{})

		for _, compObjElems := range compositeObjectives {
			compObj := compObjElems.(map[string]interface{})
			whenDelayed, ok := compObj["when_delayed"].(string)
			if !ok || whenDelayed == "" {
				return nil, fmt.Errorf("when_delayed is required for composite objective")
			}
			whenDelayedParsed, err := v1alphaSLO.ParseWhenDelayed(whenDelayed)
			if err != nil {
				return nil, err
			}

			resultObjectives = append(resultObjectives, v1alphaSLO.CompositeObjective{
				Project:     compObj["project"].(string),
				SLO:         compObj["slo"].(string),
				Objective:   compObj["objective"].(string),
				Weight:      compObj["weight"].(float64),
				WhenDelayed: whenDelayedParsed,
			})
		}
	}
	return &v1alphaSLO.CompositeSpec{
		MaxDelay:   maxDelay,
		Components: v1alphaSLO.Components{Objectives: resultObjectives},
	}, nil
}

func marshalTimeWindows(d *schema.ResourceData) []v1alphaSLO.TimeWindow {
	timeWindow := d.Get("time_window").(*schema.Set).List()[0].(map[string]interface{})

	return []v1alphaSLO.TimeWindow{{
		Unit:      timeWindow["unit"].(string),
		Count:     timeWindow["count"].(int),
		IsRolling: timeWindow["is_rolling"].(bool),
		Calendar:  marshalCalendar(timeWindow),
	}}
}

func marshalAttachments(attachments []interface{}) []v1alphaSLO.Attachment {
	resultConditions := make([]v1alphaSLO.Attachment, len(attachments))
	for i, c := range attachments {
		attachments := c.(map[string]interface{})
		displayName := attachments["display_name"].(string)

		resultConditions[i] = v1alphaSLO.Attachment{
			DisplayName: &displayName,
			URL:         attachments["url"].(string),
		}
	}

	return resultConditions
}

func marshalCalendar(c map[string]interface{}) *v1alphaSLO.Calendar {
	calendars := c["calendar"].(*schema.Set).List()
	if len(calendars) == 0 {
		return nil
	}
	calendar := calendars[0].(map[string]interface{})

	return &v1alphaSLO.Calendar{
		StartTime: calendar["start_time"].(string),
		TimeZone:  calendar["time_zone"].(string),
	}
}

func marshalIndicator(d *schema.ResourceData) *v1alphaSLO.Indicator {
	if d.Get("indicator").(*schema.Set).Len() == 0 {
		return nil
	}
	var resultIndicator v1alphaSLO.Indicator
	indicator := d.Get("indicator").(*schema.Set).List()[0].(map[string]interface{})
	kind, err := manifest.ParseKind(indicator["kind"].(string))
	if err != nil {
		return &resultIndicator
	}
	resultIndicator = v1alphaSLO.Indicator{
		MetricSource: v1alphaSLO.MetricSourceSpec{
			Project: indicator["project"].(string),
			Name:    indicator["name"].(string),
			Kind:    kind,
		},
	}
	return &resultIndicator
}

func marshalObjectives(d *schema.ResourceData) ([]v1alphaSLO.Objective, diag.Diagnostics) {
	objectivesSchema := d.Get("objective").(*schema.Set).List()
	objectives := make([]v1alphaSLO.Objective, len(objectivesSchema))
	for i, o := range objectivesSchema {
		objective := o.(map[string]interface{})
		value := objective["value"].(float64)
		target := objective["target"].(float64)
		timeSliceTarget := objective["time_slice_target"].(float64)
		var timeSliceTargetPtr *float64
		if timeSliceTarget != 0 {
			timeSliceTargetPtr = &timeSliceTarget
		}
		operator := objective["op"].(string)
		primary := objective["primary"].(bool)
		compositeSpec, err := marshalComposite(objective)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		objectives[i] = v1alphaSLO.Objective{
			ObjectiveBase: v1alphaSLO.ObjectiveBase{
				DisplayName: objective["display_name"].(string),
				Value:       &value,
				Name:        objective["name"].(string),
			},
			BudgetTarget:    &target,
			TimeSliceTarget: timeSliceTargetPtr,
			Operator:        &operator,
			CountMetrics:    marshalCountMetrics(objective),
			RawMetric:       marshalRawMetric(objective),
			Primary:         &primary,
			Composite:       compositeSpec,
		}
	}
	return objectives, nil
}

func marshalRawMetric(metricRoot map[string]interface{}) *v1alphaSLO.RawMetricSpec {
	rawMetricSet := metricRoot["raw_metric"].(*schema.Set)
	if rawMetricSet.Len() == 0 {
		return nil
	}

	rawMetric := metricRoot["raw_metric"].(*schema.Set).List()[0].(map[string]interface{})
	if _, ok := rawMetric["query"]; !ok {
		return nil
	}

	metric := rawMetric["query"].(*schema.Set).List()[0].(map[string]interface{})

	return &v1alphaSLO.RawMetricSpec{
		MetricQuery: marshalMetric(metric),
	}
}

func marshalCountMetrics(countMetricsTf map[string]interface{}) *v1alphaSLO.CountMetricsSpec {
	countMetricsSet := countMetricsTf["count_metrics"].(*schema.Set)
	if countMetricsSet.Len() == 0 {
		return nil
	}

	countMetrics := countMetricsSet.List()[0].(map[string]interface{})

	incremental := countMetrics["incremental"].(bool)

	total := countMetrics["total"].(*schema.Set).List()[0].(map[string]interface{})
	spec := &v1alphaSLO.CountMetricsSpec{
		Incremental: &incremental,
		TotalMetric: marshalMetric(total),
	}

	if good := countMetrics["good"].(*schema.Set).List(); len(good) > 0 {
		spec.GoodMetric = marshalMetric(good[0].(map[string]interface{}))
	}
	if bad := countMetrics["bad"].(*schema.Set).List(); len(bad) > 0 {
		spec.BadMetric = marshalMetric(bad[0].(map[string]interface{}))
	}

	return spec
}

func marshalMetric(metric map[string]interface{}) *v1alphaSLO.MetricSpec {
	return &v1alphaSLO.MetricSpec{
		AmazonPrometheus:    marshalAmazonPrometheusMetric(metric[amazonPrometheusMetric].(*schema.Set)),
		AppDynamics:         marshalAppDynamicsMetric(metric[appDynamicsMetric].(*schema.Set)),
		AzureMonitor:        marshalAzureMonitorMetric(metric[azureMonitorMetric].(*schema.Set)),
		BigQuery:            marshalBigQueryMetric(metric[bigQueryMetric].(*schema.Set)),
		CloudWatch:          marshalCloudWatchMetric(metric[cloudwatchMetric].(*schema.Set)),
		Datadog:             marshalDatadogMetric(metric[datadogMetric].(*schema.Set)),
		Dynatrace:           marshalDynatraceMetric(metric[dynatraceMetric].(*schema.Set)),
		Elasticsearch:       marshalElasticsearchMetric(metric[elasticsearchMetric].(*schema.Set)),
		GCM:                 marshalGCMMetric(metric[gcmMetric].(*schema.Set)),
		GrafanaLoki:         marshalGrafanaLokiMetric(metric[grafanaLokiMetric].(*schema.Set)),
		Graphite:            marshalGraphiteMetric(metric[graphiteMetric].(*schema.Set)),
		Honeycomb:           marshalHoneycombMetric(metric[honeycombMetric].(*schema.Set)),
		InfluxDB:            marshalInfluxDBMetric(metric[influxdbMetric].(*schema.Set)),
		Instana:             marshalInstanaMetric(metric[instanaMetric].(*schema.Set)),
		Lightstep:           marshalLightstepMetric(metric[lightstepMetric].(*schema.Set)),
		LogicMonitor:        marshalLogicMonitorMetric(metric[logicMonitorMetric].(*schema.Set)),
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

func unmarshalSLO(d *schema.ResourceData, objects []v1alphaSLO.SLO) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics
	var err error

	metadata := object.Metadata
	err = d.Set("name", metadata.Name)
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata.DisplayName)
	diags = appendError(diags, err)

	err = d.Set("project", metadata.Project)
	diags = appendError(diags, err)

	if labelsRaw := metadata.Labels; len(labelsRaw) > 0 {
		err = d.Set("label", unmarshalLabels(labelsRaw))
		diags = appendError(diags, err)
	}

	if len(metadata.Annotations) > 0 {
		err = d.Set("annotations", metadata.Annotations)
		diags = appendError(diags, err)
	}

	spec := object.Spec

	err = d.Set("alert_policies", spec.AlertPolicies)
	diags = appendError(diags, err)

	budgetingMethod := spec.BudgetingMethod
	err = d.Set("budgeting_method", budgetingMethod)
	diags = appendError(diags, err)

	description := spec.Description
	err = d.Set("description", description)
	diags = appendError(diags, err)

	service := spec.Service
	err = d.Set("service", service)
	diags = appendError(diags, err)

	err = unmarshalTimeWindow(d, spec)
	diags = appendError(diags, err)

	err = unmarshalIndicator(d, spec)
	diags = appendError(diags, err)

	err = unmarshalObjectives(d, spec)
	diags = appendError(diags, err)

	err = unmarshalCompositeDeprecated(d, spec)
	diags = appendError(diags, err)

	err = unmarshalAttachments(d, spec)
	diags = appendError(diags, err)

	err = unmarshalAnomalyConfig(d, spec)
	diags = appendError(diags, err)

	err = d.Set("alert_policies", spec.AlertPolicies)
	diags = appendError(diags, err)

	// Remove this warning once SLO objective unique identifier grace period ends.
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "SLO objective unique identifier warning",
		Detail: "Nobl9 is introducing an SLO objective unique identifier to support the same value for different " +
			"SLIs in the same SLO. As such, Nobl9 is adding a name identifier to each SLO objective. " +
			"Objective names can be set now, and they'll be required once grace period ends. " +
			"For more detailed information, refer to: https://docs.nobl9.com/features/slo-objective-unique-identifier",
	})

	return diags
}

func unmarshalAttachments(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
	if len(spec.Attachments) == 0 {
		return nil
	}

	declaredAttachmentTag := getDeclaredAttachmentTag(d)

	attachments := spec.Attachments
	res := make([]interface{}, len(attachments))
	for i, attachment := range attachments {
		attachment := map[string]interface{}{
			"display_name": attachment.DisplayName,
			"url":          attachment.URL,
		}
		res[i] = attachment
	}
	return d.Set(declaredAttachmentTag, res)
}

// getDeclaredAttachmentTag return name of attachments object declared in .tf file
func getDeclaredAttachmentTag(d *schema.ResourceData) string {
	_, newAttachments := d.GetChange("attachments")
	if len(newAttachments.([]interface{})) > 0 {
		return "attachments"
	}
	return "attachment"
}

func unmarshalIndicator(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
	if spec.Indicator == nil {
		return nil
	}
	indicator := spec.Indicator
	res := make(map[string]interface{})
	metricSource := indicator.MetricSource
	res["name"] = metricSource.Name
	res["project"] = metricSource.Project
	res["kind"] = metricSource.Kind.String()
	if rawMetric := indicator.RawMetric; rawMetric != nil {
		tfMetric := unmarshalSLOMetric(rawMetric)
		res["raw_metric"] = tfMetric
	}
	return d.Set("indicator", schema.NewSet(oneElementSet, []interface{}{res}))
}

func unmarshalTimeWindow(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
	timeWindows := spec.TimeWindows
	timeWindow := timeWindows[0]
	timeWindowsTF := make(map[string]interface{})
	timeWindowsTF["count"] = timeWindow.Count
	timeWindowsTF["is_rolling"] = timeWindow.IsRolling
	timeWindowsTF["unit"] = timeWindow.Unit
	timeWindowsTF["period"] = map[string]string{"begin": timeWindow.Period.Begin, "end": timeWindow.Period.End}

	if calendar := timeWindow.Calendar; calendar != nil {
		calendarTF := make(map[string]interface{})
		calendarTF["start_time"] = calendar.StartTime
		calendarTF["time_zone"] = calendar.TimeZone
		timeWindowsTF["calendar"] = schema.NewSet(oneElementSet, []interface{}{calendarTF})
	}
	tw := schema.NewSet(oneElementSet, []interface{}{timeWindowsTF})
	return d.Set("time_window", tw)
}

func unmarshalObjectives(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
	objectives := spec.Objectives
	objectivesTF := make([]interface{}, len(objectives))

	for i, objective := range objectives {
		objectiveTF := make(map[string]interface{})
		objectiveTF["name"] = objective.Name
		objectiveTF["display_name"] = objective.DisplayName
		objectiveTF["op"] = objective.Operator
		objectiveTF["value"] = objective.Value
		objectiveTF["target"] = objective.BudgetTarget
		objectiveTF["time_slice_target"] = objective.TimeSliceTarget
		objectiveTF["primary"] = objective.Primary

		if objective.CountMetrics != nil {
			cm := objective.CountMetrics
			countMetricsTF := make(map[string]interface{})
			countMetricsTF["incremental"] = cm.Incremental

			if cm.GoodMetric != nil {
				countMetricsTF["good"] = unmarshalSLOMetric(cm.GoodMetric)
			}
			if cm.BadMetric != nil {
				countMetricsTF["bad"] = unmarshalSLOMetric(cm.BadMetric)
			}
			total := unmarshalSLOMetric(cm.TotalMetric)
			countMetricsTF["total"] = total
			objectiveTF["count_metrics"] = schema.NewSet(oneElementSet, []interface{}{countMetricsTF})
		}

		if objective.RawMetric != nil {
			tfMetric := unmarshalSLORawMetric(objective.RawMetric)
			objectiveTF["raw_metric"] = tfMetric
		}

		if objective.Composite != nil {
			objectiveTF["composite"] = unmarshalComposite(objective.Composite)
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

func unmarshalCompositeDeprecated(d *schema.ResourceData, spec v1alphaSLO.Spec) error {
	//nolint:staticcheck
	if spec.Composite != nil {
		//nolint:staticcheck
		composite := spec.Composite
		compositeTF := make(map[string]interface{})

		compositeTF["target"] = composite.BudgetTarget

		if composite.BurnRateCondition != nil {
			burnRateCondition := composite.BurnRateCondition
			burnRateConditionTF := make(map[string]interface{})
			burnRateConditionTF["value"] = burnRateCondition.Value
			burnRateConditionTF["op"] = burnRateCondition.Operator
			compositeTF["burn_rate_condition"] = schema.NewSet(oneElementSet, []interface{}{burnRateConditionTF})
		}

		return d.Set("composite", schema.NewSet(oneElementSet, []interface{}{compositeTF}))
	}

	return nil
}

func unmarshalComposite(compositeSpec *v1alphaSLO.CompositeSpec) *schema.Set {
	if compositeSpec == nil {
		return nil
	}

	composite := make(map[string]interface{})
	composite["max_delay"] = compositeSpec.MaxDelay

	compObjList := make([]interface{}, 0, len(compositeSpec.Components.Objectives))
	for _, objective := range compositeSpec.Components.Objectives {
		compositeObjective := make(map[string]interface{})
		compositeObjective["project"] = objective.Project
		compositeObjective["slo"] = objective.SLO
		compositeObjective["objective"] = objective.Objective
		compositeObjective["weight"] = objective.Weight
		compositeObjective["when_delayed"] = objective.WhenDelayed.String()

		compObjList = append(compObjList, compositeObjective)
	}

	objectives := make(map[string]interface{})
	objectives["composite_objective"] = compObjList

	components := make(map[string]interface{})
	components["objectives"] = schema.NewSet(
		schema.HashResource(resourceCompositeComponents()),
		[]interface{}{objectives},
	)
	composite["components"] = schema.NewSet(oneElementSet, []interface{}{components})

	return schema.NewSet(schema.HashResource(resourceComposite()), []interface{}{composite})
}

func unmarshalSLORawMetric(rawMetricSource *v1alphaSLO.RawMetricSpec) *schema.Set {
	var rawMetricQuery *schema.Set
	if rawMetricSource.MetricQuery != nil {
		rawMetricQuery = unmarshalSLOMetric(rawMetricSource.MetricQuery)
	}
	return schema.NewSet(oneElementSet, []interface{}{map[string]interface{}{"query": rawMetricQuery}})
}

func unmarshalSLOMetric(spec *v1alphaSLO.MetricSpec) *schema.Set {
	supportedMetrics := []struct {
		hclName       string
		specFieldName string
		unmarshalFunc func(interface{}) map[string]interface{}
	}{
		{amazonPrometheusMetric, "AmazonPrometheus", unmarshalAmazonPrometheusMetric},
		{appDynamicsMetric, "AppDynamics", unmarshalAppdynamicsMetric},
		{azureMonitorMetric, "AzureMonitor", unmarshalAzureMonitorMetric},
		{bigQueryMetric, "BigQuery", unmarshalBigqueryMetric},
		{cloudwatchMetric, "CloudWatch", unmarshalCloudWatchMetric},
		{datadogMetric, "Datadog", unmarshalDatadogMetric},
		{dynatraceMetric, "Dynatrace", unmarshalDynatraceMetric},
		{elasticsearchMetric, "Elasticsearch", unmarshalElasticsearchMetric},
		{gcmMetric, "GCM", unmarshalGCMMetric},
		{grafanaLokiMetric, "GrafanaLoki", unmarshalGrafanaLokiMetric},
		{graphiteMetric, "Graphite", unmarshalGraphiteMetric},
		{honeycombMetric, "Honeycomb", unmarshalHoneycombMetric},
		{influxdbMetric, "InfluxDB", unmarshalInfluxDBMetric},
		{instanaMetric, "Instana", unmarshalInstanaMetric},
		{lightstepMetric, "Lightstep", unmarshalLightstepMetric},
		{logicMonitorMetric, "LogicMonitor", unmarshalLogicMonitorMetric},
		{newrelicMetric, "NewRelic", unmarshalNewRelicMetric},
		{opentsdbMetric, "OpenTSDB", unmarshalOpentsdbMetric},
		{pingdomMetric, "Pingdom", unmarshalPingdomMetric},
		{prometheusMetric, "Prometheus", unmarshalPrometheusMetric},
		{redshiftMetric, "Redshift", unmarshalRedshiftMetric},
		{splunkMetric, "Splunk", unmarshalSplunkMetric},
		{splunkObservabilityMetric, "SplunkObservability", unmarshalSplunkObservabilityMetric},
		{sumologicMetric, "SumoLogic", unmarshalSumologicMetric},
		{thousandeyesMetric, "ThousandEyes", unmarshalThousandeyesMetric},
	}

	res := make(map[string]interface{})

	v := reflect.ValueOf(spec).Elem()
	for _, name := range supportedMetrics {
		field := v.FieldByName(name.specFieldName)
		if field.IsValid() && !field.IsNil() {
			tfMetric := name.unmarshalFunc(field.Interface())
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

func marshalAmazonPrometheusMetric(s *schema.Set) *v1alphaSLO.AmazonPrometheusMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["promql"].(string)
	return &v1alphaSLO.AmazonPrometheusMetric{
		PromQL: &query,
	}
}

func unmarshalAmazonPrometheusMetric(metric interface{}) map[string]interface{} {
	apMetric, ok := metric.(*v1alphaSLO.AmazonPrometheusMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["promql"] = apMetric.PromQL

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

func marshalAppDynamicsMetric(s *schema.Set) *v1alphaSLO.AppDynamicsMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	applicationName := metric["application_name"].(string)
	metricPath := metric["metric_path"].(string)
	return &v1alphaSLO.AppDynamicsMetric{
		ApplicationName: &applicationName,
		MetricPath:      &metricPath,
	}
}

func unmarshalAppdynamicsMetric(metric interface{}) map[string]interface{} {
	adMetric, ok := metric.(*v1alphaSLO.AppDynamicsMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["application_name"] = adMetric.ApplicationName
	res["metric_path"] = adMetric.MetricPath

	return res
}

/**
 * Azure Monitor Metric
 * https://docs.nobl9.com/Sources/azure-monitor#creating-slos-with-azure-monitor
 */
const azureMonitorMetric = "azure_monitor"

func schemaMetricAzureMonitor() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		azureMonitorMetric: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/azure-monitor#creating-slos-with-azure-monitor)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"data_type": {
						Type:     schema.TypeString,
						Required: true,
						ValidateDiagFunc: validation.ToDiagFunc(
							validation.StringInSlice([]string{
								v1alphaSLO.AzureMonitorDataTypeMetrics,
								v1alphaSLO.AzureMonitorDataTypeLogs,
							}, false),
						),
						Description: "Specifies source: 'metrics' or 'logs'",
					},
					"resource_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Identifier of the Azure Cloud resource [Required for metrics]",
					},
					"metric_namespace": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Namespace of the metric [Optional for metrics]",
					},
					"metric_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the metric [Required for metrics]",
					},
					"aggregation": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Aggregation type [Required for metrics]",
					},
					"dimensions": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Dimensions of the metric [Optional for metrics]",
						MinItems:    1,
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
					"workspace": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Log analytics workspace [Required for logs]",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"subscription_id": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Subscription ID of the workspace",
								},
								"resource_group": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Resource group of the workspace",
								},
								"workspace_id": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "ID of the workspace",
								},
							},
						},
					},
					"kql_query": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Logs query in Kusto Query Language [Required for logs]",
					},
				},
			},
		},
	}
}

func marshalAzureMonitorMetric(s *schema.Set) *v1alphaSLO.AzureMonitorMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	dataType := metric["data_type"].(string)
	result := v1alphaSLO.AzureMonitorMetric{DataType: dataType}
	if dataType == v1alphaSLO.AzureMonitorDataTypeMetrics {
		result.ResourceID = metric["resource_id"].(string)
		result.MetricNamespace = metric["metric_namespace"].(string)
		result.MetricName = metric["metric_name"].(string)
		result.Aggregation = metric["aggregation"].(string)
		dimensions := metric["dimensions"].(*schema.Set)
		var metricDimensions []v1alphaSLO.AzureMonitorMetricDimension
		if dimensions.Len() > 0 {
			metricDimensions = make([]v1alphaSLO.AzureMonitorMetricDimension, dimensions.Len())
		}
		for idx, dimension := range dimensions.List() {
			n9Dimension := dimension.(map[string]interface{})
			name := n9Dimension["name"].(string)
			value := n9Dimension["value"].(string)
			metricDimensions[idx] = v1alphaSLO.AzureMonitorMetricDimension{
				Name:  &name,
				Value: &value,
			}
		}
		result.Dimensions = metricDimensions
	} else if dataType == v1alphaSLO.AzureMonitorDataTypeLogs {
		result.KQLQuery = metric["kql_query"].(string)
		wsList := metric["workspace"].(*schema.Set).List()
		if len(wsList) > 0 {
			workspace := wsList[0].(map[string]interface{})
			result.Workspace = &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
				SubscriptionID: workspace["subscription_id"].(string),
				ResourceGroup:  workspace["resource_group"].(string),
				WorkspaceID:    workspace["workspace_id"].(string),
			}
		}
	}
	return &result
}

func unmarshalAzureMonitorMetric(metric interface{}) map[string]interface{} {
	amMetric, ok := metric.(*v1alphaSLO.AzureMonitorMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["data_type"] = amMetric.DataType
	res["resource_id"] = amMetric.ResourceID
	res["metric_namespace"] = amMetric.MetricNamespace
	res["metric_name"] = amMetric.MetricName
	res["aggregation"] = amMetric.Aggregation
	dim, _ := json.Marshal(amMetric.Dimensions)
	var dimensions any
	_ = json.Unmarshal(dim, &dimensions)
	res["dimensions"] = dimensions
	res["kql_query"] = amMetric.KQLQuery
	if amMetric.Workspace != nil {
		res["workspace"] = schema.NewSet(oneElementSet, []interface{}{map[string]interface{}{
			"subscription_id": amMetric.Workspace.SubscriptionID,
			"resource_group":  amMetric.Workspace.ResourceGroup,
			"workspace_id":    amMetric.Workspace.WorkspaceID,
		}})
	}
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

func marshalBigQueryMetric(s *schema.Set) *v1alphaSLO.BigQueryMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	return &v1alphaSLO.BigQueryMetric{
		Query:     metric["query"].(string),
		ProjectID: metric["project_id"].(string),
		Location:  metric["location"].(string),
	}
}

func unmarshalBigqueryMetric(metric interface{}) map[string]interface{} {
	bqMetric, ok := metric.(*v1alphaSLO.BigQueryMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["location"] = bqMetric.Location
	res["project_id"] = bqMetric.ProjectID
	res["query"] = bqMetric.Query

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
					"account_id": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
							value := i.(string)
							var diags diag.Diagnostics
							if m, err := regexp.MatchString(`^[0-9]{12}$`, value); !m || err != nil {
								diags = append(diags, diag.Diagnostic{
									Severity: diag.Error,
									Summary:  "account_id must be 12-digit identifier",
								})
							}

							return diags
						},
						Description: "AccountID used with cross-account observability feature",
					},
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

func marshalCloudWatchMetric(s *schema.Set) *v1alphaSLO.CloudWatchMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	region := metric["region"].(string)

	var namespace *string
	if value := metric["namespace"].(string); value != "" {
		namespace = &value
	}

	var accountID *string
	if value := metric["account_id"].(string); value != "" {
		accountID = &value
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

	var jsonStrP *string
	if value := metric["json"].(string); value != "" {
		jsonStrP = &value
	}

	dimensions := metric["dimensions"].(*schema.Set)
	var metricDimensions []v1alphaSLO.CloudWatchMetricDimension

	if dimensions.Len() > 0 {
		metricDimensions = make([]v1alphaSLO.CloudWatchMetricDimension, dimensions.Len())
	}

	for idx, dimension := range dimensions.List() {
		n9Dimension := dimension.(map[string]interface{})
		name := n9Dimension["name"].(string)
		value := n9Dimension["value"].(string)

		metricDimensions[idx] = v1alphaSLO.CloudWatchMetricDimension{
			Name:  &name,
			Value: &value,
		}
	}

	return &v1alphaSLO.CloudWatchMetric{
		Region:     &region,
		AccountID:  accountID,
		Namespace:  namespace,
		MetricName: metricName,
		Stat:       stat,
		Dimensions: metricDimensions,
		SQL:        sql,
		JSON:       jsonStrP,
	}
}

func unmarshalCloudWatchMetric(metric interface{}) map[string]interface{} {
	cwMetric, ok := metric.(*v1alphaSLO.CloudWatchMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["region"] = cwMetric.Region
	res["account_id"] = cwMetric.AccountID
	res["namespace"] = cwMetric.Namespace
	res["metric_name"] = cwMetric.MetricName
	res["stat"] = cwMetric.Stat
	res["sql"] = cwMetric.SQL
	res["json"] = cwMetric.JSON
	dim, _ := json.Marshal(cwMetric.Dimensions)
	var dimensions any
	_ = json.Unmarshal(dim, &dimensions)
	res["dimensions"] = dimensions
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

func marshalDatadogMetric(s *schema.Set) *v1alphaSLO.DatadogMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["query"].(string)
	return &v1alphaSLO.DatadogMetric{
		Query: &query,
	}
}

func unmarshalDatadogMetric(metric interface{}) map[string]interface{} {
	ddMetric, ok := metric.(*v1alphaSLO.DatadogMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["query"] = ddMetric.Query

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

func marshalDynatraceMetric(s *schema.Set) *v1alphaSLO.DynatraceMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	selector := metric["metric_selector"].(string)
	return &v1alphaSLO.DynatraceMetric{
		MetricSelector: &selector,
	}
}

func unmarshalDynatraceMetric(metric interface{}) map[string]interface{} {
	dMetric, ok := metric.(*v1alphaSLO.DynatraceMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["metric_selector"] = dMetric.MetricSelector

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

func marshalElasticsearchMetric(s *schema.Set) *v1alphaSLO.ElasticsearchMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	index := metric["index"].(string)
	query := metric["query"].(string)
	return &v1alphaSLO.ElasticsearchMetric{
		Index: &index,
		Query: &query,
	}
}

func unmarshalElasticsearchMetric(metric interface{}) map[string]interface{} {
	esMetric, ok := metric.(*v1alphaSLO.ElasticsearchMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["index"] = esMetric.Index
	res["query"] = esMetric.Query

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

func marshalGCMMetric(s *schema.Set) *v1alphaSLO.GCMMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	return &v1alphaSLO.GCMMetric{
		ProjectID: metric["project_id"].(string),
		Query:     metric["query"].(string),
	}
}

func unmarshalGCMMetric(metric interface{}) map[string]interface{} {
	gMetric, ok := metric.(*v1alphaSLO.GCMMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["project_id"] = gMetric.ProjectID
	res["query"] = gMetric.Query

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

func marshalGrafanaLokiMetric(s *schema.Set) *v1alphaSLO.GrafanaLokiMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	logql := metric["logql"].(string)
	return &v1alphaSLO.GrafanaLokiMetric{
		Logql: &logql,
	}
}

func unmarshalGrafanaLokiMetric(metric interface{}) map[string]interface{} {
	glMetric, ok := metric.(*v1alphaSLO.GrafanaLokiMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["logql"] = glMetric.Logql

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

func marshalGraphiteMetric(s *schema.Set) *v1alphaSLO.GraphiteMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	metricPath := metric["metric_path"].(string)
	return &v1alphaSLO.GraphiteMetric{
		MetricPath: &metricPath,
	}
}

func unmarshalGraphiteMetric(metric interface{}) map[string]interface{} {
	gMetric, ok := metric.(*v1alphaSLO.GraphiteMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["metric_path"] = gMetric.MetricPath

	return res
}

/**
 * Honeycomb Metric
 * https://docs.nobl9.com/Sources/honeycomb#creating-slos-with-honeycomb
 */
const honeycombMetric = "honeycomb"

func schemaMetricHoneycomb() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		honeycombMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/honeycomb#creating-slos-with-honeycomb)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"calculation": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Calculation type",
					},
					"attribute": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Column name - required for all calculation types besides 'CONCURRENCY' and 'COUNT'",
					},
				},
			},
		},
	}
}

func marshalHoneycombMetric(s *schema.Set) *v1alphaSLO.HoneycombMetric {
	if s.Len() == 0 {
		return nil
	}
	metric := s.List()[0].(map[string]interface{})
	calculation := metric["calculation"].(string)
	attribute := metric["attribute"].(string)
	return &v1alphaSLO.HoneycombMetric{
		Calculation: calculation,
		Attribute:   attribute,
	}
}

func unmarshalHoneycombMetric(metric interface{}) map[string]interface{} {
	hMetric, ok := metric.(*v1alphaSLO.HoneycombMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["calculation"] = hMetric.Calculation
	res["attribute"] = hMetric.Attribute
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

func marshalInfluxDBMetric(s *schema.Set) *v1alphaSLO.InfluxDBMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &v1alphaSLO.InfluxDBMetric{
		Query: &query,
	}
}

func unmarshalInfluxDBMetric(metric interface{}) map[string]interface{} {
	idbMetric, ok := metric.(*v1alphaSLO.InfluxDBMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["query"] = idbMetric.Query

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
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "wrong value",
				Detail:   fmt.Sprintf("%q is not %q or %q", value, appType, infraType),
			}
			diags = append(diags, diagnostic)
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

func marshalInstanaMetric(s *schema.Set) *v1alphaSLO.InstanaMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	return &v1alphaSLO.InstanaMetric{
		MetricType:     metric["metric_type"].(string),
		Infrastructure: marshalInstanaInfrastructureMetric(metric["infrastructure"].(*schema.Set)),
		Application:    marshalInstanaApplicationMetric(metric["application"].(*schema.Set)),
	}
}

func marshalInstanaInfrastructureMetric(s *schema.Set) *v1alphaSLO.InstanaInfrastructureMetricType {
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

	return &v1alphaSLO.InstanaInfrastructureMetricType{
		MetricRetrievalMethod: infrastructure["metric_retrieval_method"].(string),
		Query:                 query,
		SnapshotID:            snapshotID,
		MetricID:              infrastructure["metric_id"].(string),
		PluginID:              infrastructure["plugin_id"].(string),
	}
}

func marshalInstanaApplicationMetric(s *schema.Set) *v1alphaSLO.InstanaApplicationMetricType {
	if s.Len() == 0 {
		return nil
	}
	application := s.List()[0].(map[string]interface{})

	var includeInternal bool
	if value, ok := application["include_internal"].(bool); ok {
		includeInternal = value
	}

	var includeSynthetic bool
	if value, ok := application["include_synthetic"].(bool); ok {
		includeSynthetic = value
	}

	var groupBy = application["group_by"].(*schema.Set).List()[0].(map[string]interface{})
	var tagSecondLevelKey *string
	if value, ok := groupBy["tag_second_level_key"].(string); ok && value != "" {
		tagSecondLevelKey = &value
	}

	return &v1alphaSLO.InstanaApplicationMetricType{
		MetricID:    application["metric_id"].(string),
		Aggregation: application["aggregation"].(string),
		GroupBy: v1alphaSLO.InstanaApplicationMetricGroupBy{
			Tag:               groupBy["tag"].(string),
			TagEntity:         groupBy["tag_entity"].(string),
			TagSecondLevelKey: tagSecondLevelKey,
		},
		APIQuery:         application["api_query"].(string),
		IncludeInternal:  includeInternal,
		IncludeSynthetic: includeSynthetic,
	}
}

func unmarshalInstanaMetric(metric interface{}) map[string]interface{} {
	iMetric, ok := metric.(*v1alphaSLO.InstanaMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["metric_type"] = iMetric.MetricType
	res["infrastructure"] = unmarshalInstanaInfrastructureMetric(iMetric)
	res["application"] = unmarshalInstanaApplicationMetric(iMetric)

	return res
}

func unmarshalInstanaInfrastructureMetric(metric *v1alphaSLO.InstanaMetric) *schema.Set {
	if infrastructure := metric.Infrastructure; infrastructure != nil {
		infrastructureTF := map[string]interface{}{
			"metric_retrieval_method": infrastructure.MetricRetrievalMethod,
			"query":                   infrastructure.Query,
			"snapshot_id":             infrastructure.SnapshotID,
			"metric_id":               infrastructure.MetricID,
			"plugin_id":               infrastructure.PluginID,
		}
		return schema.NewSet(oneElementSet, []interface{}{infrastructureTF})
	}
	return nil
}

func unmarshalInstanaApplicationMetric(metric *v1alphaSLO.InstanaMetric) *schema.Set {
	if application := metric.Application; application != nil {
		applicationTF := map[string]interface{}{
			"metric_id":   application.MetricID,
			"aggregation": application.Aggregation,
			"group_by": schema.NewSet(oneElementSet, []interface{}{map[string]interface{}{
				"tag":                  application.GroupBy.Tag,
				"tag_entity":           application.GroupBy.TagEntity,
				"tag_second_level_key": application.GroupBy.TagSecondLevelKey,
			}}),
			"api_query":         application.APIQuery,
			"include_internal":  application.IncludeInternal,
			"include_synthetic": application.IncludeSynthetic,
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
						Optional:    true,
						Description: "ID of the metrics stream",
					},
					"type_of_data": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Type of data to filter by",
					},
					"uql": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "UQL query",
					},
				},
			},
		},
	}
}

func marshalLightstepMetric(s *schema.Set) *v1alphaSLO.LightstepMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	var streamID *string
	if value := metric["stream_id"].(string); value != "" {
		streamID = &value
	}

	var uql *string
	if value := metric["uql"].(string); value != "" {
		uql = &value
	}

	typeOfData := metric["type_of_data"].(string)

	var percentile *float64
	if p := metric["percentile"].(float64); p != 0 {
		// the API does not accept percentile = 0
		// terraform sets it to 0 even if it was omitted in the .tf
		percentile = &p
	}

	return &v1alphaSLO.LightstepMetric{
		StreamID:   streamID,
		TypeOfData: &typeOfData,
		Percentile: percentile,
		UQL:        uql,
	}
}

func unmarshalLightstepMetric(metric interface{}) map[string]interface{} {
	lMetric, ok := metric.(*v1alphaSLO.LightstepMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["percentile"] = lMetric.Percentile
	res["stream_id"] = lMetric.StreamID
	res["type_of_data"] = lMetric.TypeOfData
	res["uql"] = lMetric.UQL

	return res
}

/**
 * LogicMonitor Metric
 * https://docs.nobl9.com/Sources/logic-monitor#creating-slos-with-logic-monitor
 */
const logicMonitorMetric = "logic_monitor"

func schemaLogicMonitorMetric() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		logicMonitorMetric: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/logic-monitor#creating-slos-with-logic-monitor)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query_type": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query type",
					},
					"device_data_source_instance_id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Device Datasource Instance ID",
					},
					"graph_id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Graph ID",
					},
					"line": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Line",
					},
				},
			},
		},
	}
}

func marshalLogicMonitorMetric(s *schema.Set) *v1alphaSLO.LogicMonitorMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	var QueryType string
	if value := metric["query_type"].(string); value != "" {
		QueryType = value
	}

	deviceDataSourceInstanceID := metric["device_data_source_instance_id"].(int)

	graphId := metric["graph_id"].(int)

	line := metric["line"].(string)

	return &v1alphaSLO.LogicMonitorMetric{
		QueryType:                  QueryType,
		DeviceDataSourceInstanceID: deviceDataSourceInstanceID,
		GraphID:                    graphId,
		Line:                       line,
	}
}

func unmarshalLogicMonitorMetric(metric interface{}) map[string]interface{} {
	lMetric, ok := metric.(*v1alphaSLO.LogicMonitorMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["query_type"] = lMetric.QueryType
	res["device_data_source_instance_id"] = lMetric.DeviceDataSourceInstanceID
	res["graph_id"] = lMetric.GraphID
	res["line"] = lMetric.Line

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

func marshalNewRelicMetric(s *schema.Set) *v1alphaSLO.NewRelicMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	nrql := metric["nrql"].(string)
	return &v1alphaSLO.NewRelicMetric{
		NRQL: &nrql,
	}
}

func unmarshalNewRelicMetric(metric interface{}) map[string]interface{} {
	nrMetric, ok := metric.(*v1alphaSLO.NewRelicMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["nrql"] = nrMetric.NRQL

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

func marshalOpenTSDBMetric(s *schema.Set) *v1alphaSLO.OpenTSDBMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &v1alphaSLO.OpenTSDBMetric{
		Query: &query,
	}
}

func unmarshalOpentsdbMetric(metric interface{}) map[string]interface{} {
	oMetric, ok := metric.(*v1alphaSLO.OpenTSDBMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["query"] = oMetric.Query

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

func marshalPingdomMetric(s *schema.Set) *v1alphaSLO.PingdomMetric {
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
	return &v1alphaSLO.PingdomMetric{
		CheckID:   checkID,
		CheckType: checkType,
		Status:    status,
	}
}

func unmarshalPingdomMetric(metric interface{}) map[string]interface{} {
	pMetric, ok := metric.(*v1alphaSLO.PingdomMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["check_id"] = pMetric.CheckID
	res["check_type"] = pMetric.CheckType
	res["status"] = pMetric.Status

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

func marshalPrometheusMetric(s *schema.Set) *v1alphaSLO.PrometheusMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	query := metric["promql"].(string)
	return &v1alphaSLO.PrometheusMetric{
		PromQL: &query,
	}
}

func unmarshalPrometheusMetric(metric interface{}) map[string]interface{} {
	pMetric, ok := metric.(*v1alphaSLO.PrometheusMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["promql"] = pMetric.PromQL

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

func marshalRedshiftMetric(s *schema.Set) *v1alphaSLO.RedshiftMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})
	region := metric["region"].(string)
	clusterID := metric["cluster_id"].(string)
	databaseName := metric["database_name"].(string)
	query := metric["query"].(string)

	return &v1alphaSLO.RedshiftMetric{
		Region:       &region,
		ClusterID:    &clusterID,
		DatabaseName: &databaseName,
		Query:        &query,
	}
}

func unmarshalRedshiftMetric(metric interface{}) map[string]interface{} {
	rMetric, ok := metric.(*v1alphaSLO.RedshiftMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["region"] = rMetric.Region
	res["cluster_id"] = rMetric.ClusterID
	res["database_name"] = rMetric.DatabaseName
	res["query"] = rMetric.Query

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

func marshalSplunkMetric(s *schema.Set) *v1alphaSLO.SplunkMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	query := metric["query"].(string)
	return &v1alphaSLO.SplunkMetric{
		Query: &query,
	}
}

func unmarshalSplunkMetric(metric interface{}) map[string]interface{} {
	sMetric, ok := metric.(*v1alphaSLO.SplunkMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["query"] = sMetric.Query

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

func marshalSplunkObservabilityMetric(s *schema.Set) *v1alphaSLO.SplunkObservabilityMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	program := metric["program"].(string)
	return &v1alphaSLO.SplunkObservabilityMetric{
		Program: &program,
	}
}

func unmarshalSplunkObservabilityMetric(metric interface{}) map[string]interface{} {
	soMetric, ok := metric.(*v1alphaSLO.SplunkObservabilityMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["program"] = soMetric.Program

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

func marshalSumologicMetric(s *schema.Set) *v1alphaSLO.SumoLogicMetric {
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
	return &v1alphaSLO.SumoLogicMetric{
		Type:         &metricType,
		Query:        &query,
		Quantization: quantization,
		Rollup:       rollup,
	}
}

func unmarshalSumologicMetric(metric interface{}) map[string]interface{} {
	sMetric, ok := metric.(*v1alphaSLO.SumoLogicMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["type"] = sMetric.Type
	res["query"] = sMetric.Query
	res["quantization"] = sMetric.Quantization
	res["rollup"] = sMetric.Rollup

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

func marshalThousandEyesMetric(s *schema.Set) *v1alphaSLO.ThousandEyesMetric {
	if s.Len() == 0 {
		return nil
	}

	metric := s.List()[0].(map[string]interface{})

	testID := int64(metric["test_id"].(int))
	return &v1alphaSLO.ThousandEyesMetric{
		TestID: &testID,
	}
}

func unmarshalThousandeyesMetric(metric interface{}) map[string]interface{} {
	teMetric, ok := metric.(*v1alphaSLO.ThousandEyesMetric)
	if !ok {
		return nil
	}
	res := make(map[string]interface{})
	res["test_id"] = teMetric.TestID

	return res
}

func validateMaxLength(fieldName string, maxLength int) func(interface{}, cty.Path) diag.Diagnostics {
	return func(v any, _ cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		if len(v.(string)) > 63 {
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("%s is too long", fieldName),
				Detail:   fmt.Sprintf("%s cannot be longer than %d characters", fieldName, maxLength),
			}
			diags = append(diags, diagnostic)
		}
		return diags
	}
}
