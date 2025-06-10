package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// SLOResourceModel describes the [SLOResource] data model.
type SLOResourceModel struct {
	Name                       string              `tfsdk:"name"`
	DisplayName                types.String        `tfsdk:"display_name"`
	Project                    string              `tfsdk:"project"`
	Description                types.String        `tfsdk:"description"`
	Annotations                map[string]string   `tfsdk:"annotations"`
	Labels                     Labels              `tfsdk:"label"`
	Service                    types.String        `tfsdk:"service"`
	BudgetingMethod            types.String        `tfsdk:"budgeting_method"`
	Tier                       types.String        `tfsdk:"tier"`
	AlertPolicies              []string            `tfsdk:"alert_policies"`
	Indicator                  *IndicatorModel     `tfsdk:"indicator"`
	Objectives                 []ObjectiveModel    `tfsdk:"objective"`
	TimeWindow                 *TimeWindowModel    `tfsdk:"time_window"`
	Attachments                []AttachmentModel   `tfsdk:"attachment"`
	AnomalyConfig              *AnomalyConfigModel `tfsdk:"anomaly_config"`
	Composite                  []CompositeV1Model  `tfsdk:"composite"`
	RetrieveHistoricalDataFrom types.String        `tfsdk:"retrieve_historical_data_from"`
}

// IndicatorModel represents [v1alphaSLO.Indicator].
type IndicatorModel struct {
	Name    types.String `tfsdk:"name"`
	Project types.String `tfsdk:"project"`
	Kind    types.String `tfsdk:"kind"`
}

// ObjectiveModel represents [v1alphaSLO.Objective].
type ObjectiveModel struct {
	DisplayName     types.String             `tfsdk:"display_name"`
	Op              types.String             `tfsdk:"op"`
	Target          types.Float64            `tfsdk:"target"`
	TimeSliceTarget types.Float64            `tfsdk:"time_slice_target"`
	Value           types.Float64            `tfsdk:"value"`
	Name            types.String             `tfsdk:"name"`
	Primary         types.Bool               `tfsdk:"primary"`
	CountMetrics    *CountMetricsModel       `tfsdk:"count_metrics"`
	RawMetric       *RawMetricModel          `tfsdk:"raw_metric"`
	Composite       *CompositeObjectiveModel `tfsdk:"composite"`
}

// CountMetricsModel represents [v1alphaSLO.CountMetricsSpec].
type CountMetricsModel struct {
	Incremental types.Bool        `tfsdk:"incremental"`
	Good        []MetricSpecModel `tfsdk:"good"`
	Bad         []MetricSpecModel `tfsdk:"bad"`
	Total       []MetricSpecModel `tfsdk:"total"`
	GoodTotal   []MetricSpecModel `tfsdk:"good_total"`
}

// RawMetricModel represents the raw_metric block in an objective.
type RawMetricModel struct {
	Query []MetricSpecModel `tfsdk:"query"`
}

// MetricSpecModel is a generic model for all metric types.
// The actual metric type is determined by which field is populated.
type MetricSpecModel struct {
	AmazonPrometheus    *AmazonPrometheusModel    `tfsdk:"amazon_prometheus"`
	AppDynamics         *AppDynamicsModel         `tfsdk:"appdynamics"`
	AzureMonitor        *AzureMonitorModel        `tfsdk:"azure_monitor"`
	BigQuery            *BigQueryModel            `tfsdk:"bigquery"`
	CloudWatch          *CloudWatchModel          `tfsdk:"cloudwatch"`
	Datadog             *DatadogModel             `tfsdk:"datadog"`
	Dynatrace           *DynatraceModel           `tfsdk:"dynatrace"`
	Elasticsearch       *ElasticsearchModel       `tfsdk:"elasticsearch"`
	GCM                 *GCMModel                 `tfsdk:"gcm"`
	GrafanaLoki         *GrafanaLokiModel         `tfsdk:"grafana_loki"`
	Graphite            *GraphiteModel            `tfsdk:"graphite"`
	Honeycomb           *HoneycombModel           `tfsdk:"honeycomb"`
	InfluxDB            *InfluxDBModel            `tfsdk:"influxdb"`
	Instana             *InstanaModel             `tfsdk:"instana"`
	Lightstep           *LightstepModel           `tfsdk:"lightstep"`
	LogicMonitor        *LogicMonitorModel        `tfsdk:"logic_monitor"`
	NewRelic            *NewRelicModel            `tfsdk:"newrelic"`
	OpenTSDB            *OpenTSDBModel            `tfsdk:"opentsdb"`
	Pingdom             *PingdomModel             `tfsdk:"pingdom"`
	Prometheus          *PrometheusModel          `tfsdk:"prometheus"`
	Redshift            *RedshiftModel            `tfsdk:"redshift"`
	Splunk              *SplunkModel              `tfsdk:"splunk"`
	SplunkObservability *SplunkObservabilityModel `tfsdk:"splunk_observability"`
	SumoLogic           *SumoLogicModel           `tfsdk:"sumologic"`
	ThousandEyes        *ThousandEyesModel        `tfsdk:"thousandeyes"`
}

// CompositeObjectiveModel represents the composite block in an objective.
type CompositeObjectiveModel struct {
	MaxDelay   types.String              `tfsdk:"max_delay"`
	Components *CompositeComponentsModel `tfsdk:"components"`
}

// CompositeComponentsModel represents the components block within a composite objective.
type CompositeComponentsModel struct {
	Objectives *CompositeObjectivesModel `tfsdk:"objectives"`
}

// CompositeObjectivesModel represents the objectives block within composite components.
type CompositeObjectivesModel struct {
	CompositeObjective []CompositeObjectiveSpecModel `tfsdk:"composite_objective"`
}

// CompositeObjectiveSpecModel represents an individual composite objective specification.
type CompositeObjectiveSpecModel struct {
	Project     types.String  `tfsdk:"project"`
	SLO         types.String  `tfsdk:"slo"`
	Objective   types.String  `tfsdk:"objective"`
	Weight      types.Float64 `tfsdk:"weight"`
	WhenDelayed types.String  `tfsdk:"when_delayed"`
}

// TimeWindowModel represents the time_window block in the SLO resource.
type TimeWindowModel struct {
	Count     types.Int64       `tfsdk:"count"`
	IsRolling types.Bool        `tfsdk:"is_rolling"`
	Unit      types.String      `tfsdk:"unit"`
	Period    map[string]string `tfsdk:"period"`
	Calendar  *CalendarModel    `tfsdk:"calendar"`
}

// CalendarModel represents the calendar block in a time_window.
type CalendarModel struct {
	StartTime types.String `tfsdk:"start_time"`
	TimeZone  types.String `tfsdk:"time_zone"`
}

// AttachmentModel represents an attachment in the SLO resource.
type AttachmentModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	URL         types.String `tfsdk:"url"`
}

// AnomalyConfigModel represents the anomaly_config block in the SLO resource.
type AnomalyConfigModel struct {
	NoData *AnomalyConfigNoDataModel `tfsdk:"no_data"`
}

type AnomalyConfigNoDataModel struct {
	AlertAfter   types.String                    `tfsdk:"alert_after"`
	AlertMethods []AnomalyConfigAlertMethodModel `tfsdk:"alert_method"`
}

type AnomalyConfigAlertMethodModel struct {
	Name    types.String `tfsdk:"name"`
	Project types.String `tfsdk:"project"`
}

// CompositeV1Model represents the deprecated composite block in the SLO resource.
type CompositeV1Model struct {
	Target            types.Float64                       `tfsdk:"target"`
	BurnRateCondition []CompositeV1BurnRateConditionModel `tfsdk:"burn_rate_condition"`
}

// CompositeV1BurnRateConditionModel represents the deprecated burn_rate_condition block in the composite block.
type CompositeV1BurnRateConditionModel struct {
	Op    types.String  `tfsdk:"op"`
	Value types.Float64 `tfsdk:"value"`
}

type AmazonPrometheusModel struct {
	PromQL types.String `tfsdk:"promql"`
}

type AppDynamicsModel struct {
	ApplicationName types.String `tfsdk:"application_name"`
	MetricPath      types.String `tfsdk:"metric_path"`
}

type AzureMonitorModel struct {
	DataType        types.String                 `tfsdk:"data_type"`
	ResourceID      types.String                 `tfsdk:"resource_id"`
	MetricNamespace types.String                 `tfsdk:"metric_namespace"`
	MetricName      types.String                 `tfsdk:"metric_name"`
	Aggregation     types.String                 `tfsdk:"aggregation"`
	KQLQuery        types.String                 `tfsdk:"kql_query"`
	Dimensions      []AzureMonitorDimensionModel `tfsdk:"dimensions"`
	Workspace       *AzureMonitorWorkspaceModel  `tfsdk:"workspace"`
}

type AzureMonitorDimensionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type AzureMonitorWorkspaceModel struct {
	SubscriptionID types.String `tfsdk:"subscription_id"`
	ResourceGroup  types.String `tfsdk:"resource_group"`
	WorkspaceID    types.String `tfsdk:"workspace_id"`
}

type BigQueryModel struct {
	Location  types.String `tfsdk:"location"`
	ProjectID types.String `tfsdk:"project_id"`
	Query     types.String `tfsdk:"query"`
}

type CloudWatchModel struct {
	AccountID  types.String               `tfsdk:"account_id"`
	Region     types.String               `tfsdk:"region"`
	Namespace  types.String               `tfsdk:"namespace"`
	MetricName types.String               `tfsdk:"metric_name"`
	Stat       types.String               `tfsdk:"stat"`
	SQL        types.String               `tfsdk:"sql"`
	JSON       types.String               `tfsdk:"json"`
	Dimensions []CloudWatchDimensionModel `tfsdk:"dimensions"`
}

type CloudWatchDimensionModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type DatadogModel struct {
	Query types.String `tfsdk:"query"`
}

type DynatraceModel struct {
	MetricSelector types.String `tfsdk:"metric_selector"`
}

type ElasticsearchModel struct {
	Index types.String `tfsdk:"index"`
	Query types.String `tfsdk:"query"`
}

type GCMModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Query     types.String `tfsdk:"query"`
	PromQL    types.String `tfsdk:"promql"`
}

type GrafanaLokiModel struct {
	Logql types.String `tfsdk:"logql"`
}

type GraphiteModel struct {
	MetricPath types.String `tfsdk:"metric_path"`
}

type HoneycombModel struct {
	Attribute types.String `tfsdk:"attribute"`
}

type InfluxDBModel struct {
	Query types.String `tfsdk:"query"`
}

type InstanaModel struct {
	MetricType     types.String                `tfsdk:"metric_type"`
	Infrastructure *InstanaInfrastructureModel `tfsdk:"infrastructure"`
	Application    *InstanaApplicationModel    `tfsdk:"application"`
}

type InstanaInfrastructureModel struct {
	MetricRetrievalMethod types.String `tfsdk:"metric_retrieval_method"`
	Query                 types.String `tfsdk:"query"`
	SnapshotID            types.String `tfsdk:"snapshot_id"`
	MetricID              types.String `tfsdk:"metric_id"`
	PluginID              types.String `tfsdk:"plugin_id"`
}

type InstanaApplicationModel struct {
	MetricID         types.String         `tfsdk:"metric_id"`
	Aggregation      types.String         `tfsdk:"aggregation"`
	APIQuery         types.String         `tfsdk:"api_query"`
	IncludeInternal  types.Bool           `tfsdk:"include_internal"`
	IncludeSynthetic types.Bool           `tfsdk:"include_synthetic"`
	GroupBy          *InstanaGroupByModel `tfsdk:"group_by"`
}

type InstanaGroupByModel struct {
	Tag               types.String `tfsdk:"tag"`
	TagEntity         types.String `tfsdk:"tag_entity"`
	TagSecondLevelKey types.String `tfsdk:"tag_second_level_key"`
}

type LightstepModel struct {
	Percentile types.Float64 `tfsdk:"percentile"`
	StreamID   types.String  `tfsdk:"stream_id"`
	TypeOfData types.String  `tfsdk:"type_of_data"`
	UQL        types.String  `tfsdk:"uql"`
}

type LogicMonitorModel struct {
	QueryType                  types.String `tfsdk:"query_type"`
	DeviceDataSourceInstanceID types.Int64  `tfsdk:"device_data_source_instance_id"`
	GraphID                    types.Int64  `tfsdk:"graph_id"`
	WebsiteID                  types.String `tfsdk:"website_id"`
	CheckpointID               types.String `tfsdk:"checkpoint_id"`
	GraphName                  types.String `tfsdk:"graph_name"`
	Line                       types.String `tfsdk:"line"`
}

type NewRelicModel struct {
	NRQL types.String `tfsdk:"nrql"`
}

type OpenTSDBModel struct {
	Query types.String `tfsdk:"query"`
}

type PingdomModel struct {
	CheckID   types.String `tfsdk:"check_id"`
	CheckType types.String `tfsdk:"check_type"`
	Status    types.String `tfsdk:"status"`
}

type PrometheusModel struct {
	PromQL types.String `tfsdk:"promql"`
}

type RedshiftModel struct {
	Region       types.String `tfsdk:"region"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	Query        types.String `tfsdk:"query"`
}

type SplunkModel struct {
	Query types.String `tfsdk:"query"`
}

type SplunkObservabilityModel struct {
	Program types.String `tfsdk:"program"`
}

type SumoLogicModel struct {
	Type         types.String `tfsdk:"type"`
	Query        types.String `tfsdk:"query"`
	Quantization types.String `tfsdk:"quantization"`
	Rollup       types.String `tfsdk:"rollup"`
}

type ThousandEyesModel struct {
	TestID   types.Int64  `tfsdk:"test_id"`
	TestType types.String `tfsdk:"test_type"`
}

// newSLOResourceConfigFromManifest creates a new [SLOResourceModel] from a [v1alphaSLO.SLO] manifest.
//
// [SLOResourceModel.RetrieveHistoricalDataFrom] - this field is not part of the manifest.
// It's handled separately in the Create operation.
func newSLOResourceConfigFromManifest(slo v1alphaSLO.SLO) *SLOResourceModel {
	model := &SLOResourceModel{
		Name:            slo.Metadata.Name,
		DisplayName:     stringValue(slo.Metadata.DisplayName),
		Project:         slo.Metadata.Project,
		Description:     stringValue(slo.Spec.Description),
		Annotations:     slo.Metadata.Annotations,
		Labels:          newLabelsFromManifest(slo.Metadata.Labels),
		Service:         stringValue(slo.Spec.Service),
		BudgetingMethod: stringValue(slo.Spec.BudgetingMethod),
		Tier:            stringValueFromPointer(slo.Spec.Tier),
		AlertPolicies:   slo.Spec.AlertPolicies,
	}
	if slo.Spec.Indicator != nil {
		model.Indicator = &IndicatorModel{
			Name:    types.StringValue(slo.Spec.Indicator.MetricSource.Name),
			Project: types.StringValue(slo.Spec.Indicator.MetricSource.Project),
			Kind:    types.StringValue(slo.Spec.Indicator.MetricSource.Kind.String()),
		}
	}
	if len(slo.Spec.Objectives) > 0 {
		objectives := make([]ObjectiveModel, len(slo.Spec.Objectives))
		for i, o := range slo.Spec.Objectives {
			objectives[i] = ObjectiveModel{
				DisplayName:     types.StringValue(o.DisplayName),
				Op:              types.StringPointerValue(o.Operator),
				Target:          types.Float64PointerValue(o.BudgetTarget),
				TimeSliceTarget: types.Float64PointerValue(o.TimeSliceTarget),
				Value:           types.Float64PointerValue(o.Value),
				Name:            types.StringValue(o.Name),
				Primary:         types.BoolPointerValue(o.Primary),
				CountMetrics:    countMetricsToModel(o.CountMetrics),
				RawMetric:       rawMetricToModel(o.RawMetric),
				Composite:       compositeObjectiveToModel(o.Composite),
			}
		}
		model.Objectives = objectives
	}
	if len(slo.Spec.TimeWindows) > 0 {
		tw := slo.Spec.TimeWindows[0]
		model.TimeWindow = &TimeWindowModel{
			Count:     types.Int64Value(int64(tw.Count)),
			IsRolling: types.BoolValue(tw.IsRolling),
			Unit:      types.StringValue(tw.Unit),
			Period:    map[string]string{"begin": tw.Period.Begin, "end": tw.Period.End},
		}
		if tw.Calendar != nil {
			model.TimeWindow.Calendar = &CalendarModel{
				StartTime: types.StringValue(tw.Calendar.StartTime),
				TimeZone:  types.StringValue(tw.Calendar.TimeZone),
			}
		}
	}
	if len(slo.Spec.Attachments) > 0 {
		attachments := make([]AttachmentModel, len(slo.Spec.Attachments))
		for i, a := range slo.Spec.Attachments {
			attachments[i] = AttachmentModel{
				DisplayName: types.StringPointerValue(a.DisplayName),
				URL:         types.StringValue(a.URL),
			}
		}
		model.Attachments = attachments
	}
	if slo.Spec.AnomalyConfig != nil && slo.Spec.AnomalyConfig.NoData != nil {
		ac := slo.Spec.AnomalyConfig.NoData
		methods := make([]AnomalyConfigAlertMethodModel, len(ac.AlertMethods))
		for i, m := range ac.AlertMethods {
			methods[i] = AnomalyConfigAlertMethodModel{
				Name:    types.StringValue(m.Name),
				Project: types.StringValue(m.Project),
			}
		}
		model.AnomalyConfig = &AnomalyConfigModel{
			NoData: &AnomalyConfigNoDataModel{
				AlertAfter:   types.StringPointerValue(ac.AlertAfter),
				AlertMethods: methods,
			},
		}
	}
	model.Composite = compositeV1ToModel(slo.Spec.Composite)
	return model
}

func (s SLOResourceModel) ToManifest() v1alphaSLO.SLO {
	slo := v1alphaSLO.New(
		v1alphaSLO.Metadata{
			Name:        s.Name,
			DisplayName: s.DisplayName.ValueString(),
			Project:     s.Project,
			Annotations: s.Annotations,
			Labels:      s.Labels.ToManifest(),
		},
		v1alphaSLO.Spec{
			Description:     s.Description.ValueString(),
			Service:         s.Service.ValueString(),
			BudgetingMethod: s.BudgetingMethod.ValueString(),
			AlertPolicies:   s.AlertPolicies,
		},
	)
	if !isNullOrUnknown(s.Tier) {
		tier := s.Tier.ValueString()
		slo.Spec.Tier = &tier
	}
	if s.Indicator != nil {
		kind, _ := manifest.ParseKind(s.Indicator.Kind.ValueString())
		slo.Spec.Indicator = &v1alphaSLO.Indicator{
			MetricSource: v1alphaSLO.MetricSourceSpec{
				Name:    s.Indicator.Name.ValueString(),
				Project: s.Indicator.Project.ValueString(),
				Kind:    kind,
			},
		}
	}
	if len(s.Objectives) > 0 {
		objectives := make([]v1alphaSLO.Objective, len(s.Objectives))
		for i, o := range s.Objectives {
			objectives[i] = v1alphaSLO.Objective{
				ObjectiveBase: v1alphaSLO.ObjectiveBase{
					DisplayName: o.DisplayName.ValueString(),
					Value:       float64Pointer(o.Value),
					Name:        o.Name.ValueString(),
				},
				Operator:        stringPointer(o.Op),
				BudgetTarget:    float64Pointer(o.Target),
				TimeSliceTarget: float64Pointer(o.TimeSliceTarget),
				Primary:         boolPointer(o.Primary),
				CountMetrics:    o.CountMetrics.ToManifest(),
				RawMetric:       o.RawMetric.ToManifest(),
				Composite:       o.Composite.ToManifest(),
			}
		}
		slo.Spec.Objectives = objectives
	}
	if s.TimeWindow != nil {
		tw := s.TimeWindow
		var calendar *v1alphaSLO.Calendar
		if tw.Calendar != nil {
			calendar = &v1alphaSLO.Calendar{
				StartTime: tw.Calendar.StartTime.ValueString(),
				TimeZone:  tw.Calendar.TimeZone.ValueString(),
			}
		}
		slo.Spec.TimeWindows = []v1alphaSLO.TimeWindow{{
			Count:     int(tw.Count.ValueInt64()),
			IsRolling: tw.IsRolling.ValueBool(),
			Unit:      tw.Unit.ValueString(),
			Period:    &v1alphaSLO.Period{Begin: tw.Period["begin"], End: tw.Period["end"]},
			Calendar:  calendar,
		}}
	}
	if len(s.Attachments) > 0 {
		attachments := make([]v1alphaSLO.Attachment, len(s.Attachments))
		for i, a := range s.Attachments {
			var displayName *string
			if !isNullOrUnknown(a.DisplayName) {
				dn := a.DisplayName.ValueString()
				displayName = &dn
			}
			attachments[i] = v1alphaSLO.Attachment{
				DisplayName: displayName,
				URL:         a.URL.ValueString(),
			}
		}
		slo.Spec.Attachments = attachments
	}
	if s.AnomalyConfig != nil && s.AnomalyConfig.NoData != nil {
		ac := s.AnomalyConfig.NoData
		methods := make([]v1alphaSLO.AnomalyConfigAlertMethod, len(ac.AlertMethods))
		for i, m := range ac.AlertMethods {
			methods[i] = v1alphaSLO.AnomalyConfigAlertMethod{
				Name:    m.Name.ValueString(),
				Project: m.Project.ValueString(),
			}
		}
		var alertAfter *string
		if !isNullOrUnknown(ac.AlertAfter) {
			aa := ac.AlertAfter.ValueString()
			alertAfter = &aa
		}
		slo.Spec.AnomalyConfig = &v1alphaSLO.AnomalyConfig{
			NoData: &v1alphaSLO.AnomalyConfigNoData{
				AlertAfter:   alertAfter,
				AlertMethods: methods,
			},
		}
	}
	slo.Spec.Composite = s.toV1CompositeModel()
	return slo
}

func countMetricsToModel(src *v1alphaSLO.CountMetricsSpec) *CountMetricsModel {
	if src == nil {
		return nil
	}
	model := &CountMetricsModel{}
	if src.Incremental != nil {
		model.Incremental = types.BoolValue(*src.Incremental)
	}
	if src.GoodMetric != nil {
		model.Good = []MetricSpecModel{metricSpecToModel(src.GoodMetric)}
	}
	if src.BadMetric != nil {
		model.Bad = []MetricSpecModel{metricSpecToModel(src.BadMetric)}
	}
	if src.TotalMetric != nil {
		model.Total = []MetricSpecModel{metricSpecToModel(src.TotalMetric)}
	}
	if src.GoodTotalMetric != nil {
		model.GoodTotal = []MetricSpecModel{metricSpecToModel(src.GoodTotalMetric)}
	}
	return model
}

func rawMetricToModel(src *v1alphaSLO.RawMetricSpec) *RawMetricModel {
	if src == nil {
		return nil
	}
	return &RawMetricModel{
		Query: []MetricSpecModel{metricSpecToModel(src.MetricQuery)},
	}
}

func compositeObjectiveToModel(src *v1alphaSLO.CompositeSpec) *CompositeObjectiveModel {
	if src == nil {
		return nil
	}
	model := &CompositeObjectiveModel{
		MaxDelay: types.StringValue(src.MaxDelay),
	}
	if len(src.Components.Objectives) > 0 {
		compositeObjectives := make([]CompositeObjectiveSpecModel, len(src.Components.Objectives))
		for i, obj := range src.Components.Objectives {
			compositeObjectives[i] = CompositeObjectiveSpecModel{
				Project:     types.StringValue(obj.Project),
				SLO:         types.StringValue(obj.SLO),
				Objective:   types.StringValue(obj.Objective),
				Weight:      types.Float64Value(obj.Weight),
				WhenDelayed: types.StringValue(obj.WhenDelayed.String()),
			}
		}
		model.Components = &CompositeComponentsModel{
			Objectives: &CompositeObjectivesModel{
				CompositeObjective: compositeObjectives,
			},
		}
	}

	return model
}

func compositeV1ToModel(src *v1alphaSLO.Composite) []CompositeV1Model {
	if src == nil {
		return nil
	}
	result := []CompositeV1Model{}
	if src.BudgetTarget != nil {
		model := CompositeV1Model{
			Target: types.Float64Value(*src.BudgetTarget),
		}
		if src.BurnRateCondition != nil {
			model.BurnRateCondition = []CompositeV1BurnRateConditionModel{
				{
					Op:    types.StringValue(src.BurnRateCondition.Operator),
					Value: types.Float64Value(src.BurnRateCondition.Value),
				},
			}
		}
		result = append(result, model)
	}
	return result
}

// ToManifest methods for model types
func (c *CountMetricsModel) ToManifest() *v1alphaSLO.CountMetricsSpec {
	if c == nil {
		return nil
	}
	spec := &v1alphaSLO.CountMetricsSpec{}
	if !isNullOrUnknown(c.Incremental) {
		incremental := c.Incremental.ValueBool()
		spec.Incremental = &incremental
	}
	if len(c.Good) > 0 {
		spec.GoodMetric = c.Good[0].ToManifest()
	}
	if len(c.Bad) > 0 {
		spec.BadMetric = c.Bad[0].ToManifest()
	}
	if len(c.Total) > 0 {
		spec.TotalMetric = c.Total[0].ToManifest()
	}
	if len(c.GoodTotal) > 0 {
		spec.GoodTotalMetric = c.GoodTotal[0].ToManifest()
	}
	return spec
}

func (r *RawMetricModel) ToManifest() *v1alphaSLO.RawMetricSpec {
	if r == nil || len(r.Query) == 0 {
		return nil
	}
	return &v1alphaSLO.RawMetricSpec{
		MetricQuery: r.Query[0].ToManifest(),
	}
}

func (c *CompositeObjectiveModel) ToManifest() *v1alphaSLO.CompositeSpec {
	if c == nil {
		return nil
	}
	spec := &v1alphaSLO.CompositeSpec{
		MaxDelay: c.MaxDelay.ValueString(),
	}
	if c.Components != nil && c.Components.Objectives != nil && len(c.Components.Objectives.CompositeObjective) > 0 {
		objectives := c.Components.Objectives.CompositeObjective
		compositeObjectives := make([]v1alphaSLO.CompositeObjective, len(objectives))
		for i, obj := range objectives {
			whenDelayed, _ := v1alphaSLO.ParseWhenDelayed(obj.WhenDelayed.ValueString())
			compositeObjectives[i] = v1alphaSLO.CompositeObjective{
				Project:     obj.Project.ValueString(),
				SLO:         obj.SLO.ValueString(),
				Objective:   obj.Objective.ValueString(),
				Weight:      obj.Weight.ValueFloat64(),
				WhenDelayed: whenDelayed,
			}
		}
		spec.Components = v1alphaSLO.Components{Objectives: compositeObjectives}
	}

	return spec
}

func (s SLOResourceModel) toV1CompositeModel() *v1alphaSLO.Composite {
	if len(s.Composite) == 0 {
		return nil
	}
	// Map deprecated composite fields appropriately
	composite := s.Composite[0]
	target := composite.Target.ValueFloat64()
	result := &v1alphaSLO.Composite{
		BudgetTarget: &target,
	}
	// Map burn rate conditions if present
	if len(composite.BurnRateCondition) > 0 {
		brc := composite.BurnRateCondition[0]
		result.BurnRateCondition = &v1alphaSLO.CompositeBurnRateCondition{
			Value:    brc.Value.ValueFloat64(),
			Operator: brc.Op.ValueString(),
		}
	}
	return result
}

func metricSpecToModel(spec *v1alphaSLO.MetricSpec) MetricSpecModel {
	if spec == nil {
		return MetricSpecModel{}
	}
	return MetricSpecModel{
		AmazonPrometheus:    amazonPrometheusToModel(spec.AmazonPrometheus),
		AppDynamics:         appDynamicsToModel(spec.AppDynamics),
		AzureMonitor:        azureMonitorToModel(spec.AzureMonitor),
		BigQuery:            bigQueryToModel(spec.BigQuery),
		CloudWatch:          cloudWatchToModel(spec.CloudWatch),
		Datadog:             datadogToModel(spec.Datadog),
		Dynatrace:           dynatraceToModel(spec.Dynatrace),
		Elasticsearch:       elasticsearchToModel(spec.Elasticsearch),
		GCM:                 gcmToModel(spec.GCM),
		GrafanaLoki:         grafanaLokiToModel(spec.GrafanaLoki),
		Graphite:            graphiteToModel(spec.Graphite),
		Honeycomb:           honeycombToModel(spec.Honeycomb),
		InfluxDB:            influxDBToModel(spec.InfluxDB),
		Instana:             instanaToModel(spec.Instana),
		Lightstep:           lightstepToModel(spec.Lightstep),
		LogicMonitor:        logicMonitorToModel(spec.LogicMonitor),
		NewRelic:            newRelicToModel(spec.NewRelic),
		OpenTSDB:            openTSDBToModel(spec.OpenTSDB),
		Pingdom:             pingdomToModel(spec.Pingdom),
		Prometheus:          prometheusToModel(spec.Prometheus),
		Redshift:            redshiftToModel(spec.Redshift),
		Splunk:              splunkToModel(spec.Splunk),
		SplunkObservability: splunkObservabilityToModel(spec.SplunkObservability),
		SumoLogic:           sumoLogicToModel(spec.SumoLogic),
		ThousandEyes:        thousandEyesToModel(spec.ThousandEyes),
	}
}

func (m MetricSpecModel) ToManifest() *v1alphaSLO.MetricSpec {
	spec := &v1alphaSLO.MetricSpec{
		AmazonPrometheus:    modelToAmazonPrometheus(m.AmazonPrometheus),
		AppDynamics:         modelToAppDynamics(m.AppDynamics),
		AzureMonitor:        modelToAzureMonitor(m.AzureMonitor),
		BigQuery:            modelToBigQuery(m.BigQuery),
		CloudWatch:          modelToCloudWatch(m.CloudWatch),
		Datadog:             modelToDatadog(m.Datadog),
		Dynatrace:           modelToDynatrace(m.Dynatrace),
		Elasticsearch:       modelToElasticsearch(m.Elasticsearch),
		GCM:                 modelToGCM(m.GCM),
		GrafanaLoki:         modelToGrafanaLoki(m.GrafanaLoki),
		Graphite:            modelToGraphite(m.Graphite),
		Honeycomb:           modelToHoneycomb(m.Honeycomb),
		InfluxDB:            modelToInfluxDB(m.InfluxDB),
		Instana:             modelToInstana(m.Instana),
		Lightstep:           modelToLightstep(m.Lightstep),
		LogicMonitor:        modelToLogicMonitor(m.LogicMonitor),
		NewRelic:            modelToNewRelic(m.NewRelic),
		OpenTSDB:            modelToOpenTSDB(m.OpenTSDB),
		Pingdom:             modelToPingdom(m.Pingdom),
		Prometheus:          modelToPrometheus(m.Prometheus),
		Redshift:            modelToRedshift(m.Redshift),
		Splunk:              modelToSplunk(m.Splunk),
		SplunkObservability: modelToSplunkObservability(m.SplunkObservability),
		SumoLogic:           modelToSumoLogic(m.SumoLogic),
		ThousandEyes:        modelToThousandEyes(m.ThousandEyes),
	}
	return spec
}

// Helper functions for converting from SDK types to model types
func amazonPrometheusToModel(src *v1alphaSLO.AmazonPrometheusMetric) *AmazonPrometheusModel {
	if src == nil {
		return nil
	}
	return &AmazonPrometheusModel{
		PromQL: types.StringValue(*src.PromQL),
	}
}

func appDynamicsToModel(src *v1alphaSLO.AppDynamicsMetric) *AppDynamicsModel {
	if src == nil {
		return nil
	}
	return &AppDynamicsModel{
		ApplicationName: types.StringValue(*src.ApplicationName),
		MetricPath:      types.StringValue(*src.MetricPath),
	}
}

func azureMonitorToModel(src *v1alphaSLO.AzureMonitorMetric) *AzureMonitorModel {
	if src == nil {
		return nil
	}
	model := &AzureMonitorModel{
		DataType:        stringValue(src.DataType),
		ResourceID:      stringValue(src.ResourceID),
		MetricNamespace: stringValue(src.MetricNamespace),
		MetricName:      stringValue(src.MetricName),
		Aggregation:     stringValue(src.Aggregation),
		KQLQuery:        stringValue(src.KQLQuery),
	}
	if len(src.Dimensions) > 0 {
		dimensions := make([]AzureMonitorDimensionModel, len(src.Dimensions))
		for i, d := range src.Dimensions {
			dimensions[i] = AzureMonitorDimensionModel{
				Name:  types.StringValue(*d.Name),
				Value: types.StringValue(*d.Value),
			}
		}
		model.Dimensions = dimensions
	}
	if src.Workspace != nil {
		model.Workspace = &AzureMonitorWorkspaceModel{
			SubscriptionID: stringValue(src.Workspace.SubscriptionID),
			ResourceGroup:  stringValue(src.Workspace.ResourceGroup),
			WorkspaceID:    stringValue(src.Workspace.WorkspaceID),
		}
	}
	return model
}

func bigQueryToModel(src *v1alphaSLO.BigQueryMetric) *BigQueryModel {
	if src == nil {
		return nil
	}
	return &BigQueryModel{
		Location:  stringValue(src.Location),
		ProjectID: stringValue(src.ProjectID),
		Query:     stringValue(src.Query),
	}
}

func cloudWatchToModel(src *v1alphaSLO.CloudWatchMetric) *CloudWatchModel {
	if src == nil {
		return nil
	}
	model := &CloudWatchModel{
		Region:     types.StringPointerValue(src.Region),
		Namespace:  types.StringPointerValue(src.Namespace),
		MetricName: types.StringPointerValue(src.MetricName),
		Stat:       types.StringPointerValue(src.Stat),
		SQL:        types.StringPointerValue(src.SQL),
		JSON:       types.StringPointerValue(src.JSON),
	}
	if src.AccountID != nil {
		model.AccountID = types.StringValue(*src.AccountID)
	}
	if len(src.Dimensions) > 0 {
		dimensions := make([]CloudWatchDimensionModel, len(src.Dimensions))
		for i, d := range src.Dimensions {
			dimensions[i] = CloudWatchDimensionModel{
				Name:  types.StringPointerValue(d.Name),
				Value: types.StringPointerValue(d.Value),
			}
		}
		model.Dimensions = dimensions
	}
	return model
}

func datadogToModel(src *v1alphaSLO.DatadogMetric) *DatadogModel {
	if src == nil {
		return nil
	}
	return &DatadogModel{
		Query: stringValueFromPointer(src.Query),
	}
}

func dynatraceToModel(src *v1alphaSLO.DynatraceMetric) *DynatraceModel {
	if src == nil {
		return nil
	}
	return &DynatraceModel{
		MetricSelector: stringValueFromPointer(src.MetricSelector),
	}
}

func elasticsearchToModel(src *v1alphaSLO.ElasticsearchMetric) *ElasticsearchModel {
	if src == nil {
		return nil
	}
	return &ElasticsearchModel{
		Index: stringValueFromPointer(src.Index),
		Query: stringValueFromPointer(src.Query),
	}
}

func gcmToModel(src *v1alphaSLO.GCMMetric) *GCMModel {
	if src == nil {
		return nil
	}
	return &GCMModel{
		ProjectID: stringValue(src.ProjectID),
		Query:     stringValue(src.Query),
		PromQL:    stringValue(src.PromQL),
	}
}

func grafanaLokiToModel(src *v1alphaSLO.GrafanaLokiMetric) *GrafanaLokiModel {
	if src == nil {
		return nil
	}
	return &GrafanaLokiModel{
		Logql: stringValueFromPointer(src.Logql),
	}
}

func graphiteToModel(src *v1alphaSLO.GraphiteMetric) *GraphiteModel {
	if src == nil {
		return nil
	}
	return &GraphiteModel{
		MetricPath: stringValueFromPointer(src.MetricPath),
	}
}

func honeycombToModel(src *v1alphaSLO.HoneycombMetric) *HoneycombModel {
	if src == nil {
		return nil
	}
	return &HoneycombModel{
		Attribute: stringValue(src.Attribute),
	}
}

func influxDBToModel(src *v1alphaSLO.InfluxDBMetric) *InfluxDBModel {
	if src == nil {
		return nil
	}
	return &InfluxDBModel{
		Query: stringValueFromPointer(src.Query),
	}
}

func instanaToModel(src *v1alphaSLO.InstanaMetric) *InstanaModel {
	if src == nil {
		return nil
	}
	model := &InstanaModel{
		MetricType: stringValue(src.MetricType),
	}
	if src.Infrastructure != nil {
		model.Infrastructure = &InstanaInfrastructureModel{
			MetricRetrievalMethod: stringValue(src.Infrastructure.MetricRetrievalMethod),
			Query:                 stringValueFromPointer(src.Infrastructure.Query),
			SnapshotID:            stringValueFromPointer(src.Infrastructure.SnapshotID),
			MetricID:              stringValue(src.Infrastructure.MetricID),
			PluginID:              stringValue(src.Infrastructure.PluginID),
		}
	}
	if src.Application != nil {
		app := &InstanaApplicationModel{
			MetricID:         stringValue(src.Application.MetricID),
			Aggregation:      stringValue(src.Application.Aggregation),
			APIQuery:         stringValue(src.Application.APIQuery),
			IncludeInternal:  types.BoolValue(src.Application.IncludeInternal),
			IncludeSynthetic: types.BoolValue(src.Application.IncludeSynthetic),
		}

		app.GroupBy = &InstanaGroupByModel{
			Tag:               stringValue(src.Application.GroupBy.Tag),
			TagEntity:         stringValue(src.Application.GroupBy.TagEntity),
			TagSecondLevelKey: stringValueFromPointer(src.Application.GroupBy.TagSecondLevelKey),
		}
		model.Application = app
	}
	return model
}

func lightstepToModel(src *v1alphaSLO.LightstepMetric) *LightstepModel {
	if src == nil {
		return nil
	}
	model := &LightstepModel{
		TypeOfData: stringValueFromPointer(src.TypeOfData),
		StreamID:   stringValueFromPointer(src.StreamID),
		UQL:        stringValueFromPointer(src.UQL),
	}
	if src.Percentile != nil {
		model.Percentile = types.Float64Value(*src.Percentile)
	}
	return model
}

func logicMonitorToModel(src *v1alphaSLO.LogicMonitorMetric) *LogicMonitorModel {
	if src == nil {
		return nil
	}
	model := &LogicMonitorModel{
		QueryType: stringValue(src.QueryType),
		Line:      stringValue(src.Line),
	}
	model.DeviceDataSourceInstanceID = types.Int64Value(int64(src.DeviceDataSourceInstanceID))
	model.GraphID = types.Int64Value(int64(src.GraphID))
	model.WebsiteID = stringValue(src.WebsiteID)
	model.CheckpointID = stringValue(src.CheckpointID)
	model.GraphName = stringValue(src.GraphName)
	return model
}

func newRelicToModel(src *v1alphaSLO.NewRelicMetric) *NewRelicModel {
	if src == nil {
		return nil
	}
	return &NewRelicModel{
		NRQL: stringValueFromPointer(src.NRQL),
	}
}

func openTSDBToModel(src *v1alphaSLO.OpenTSDBMetric) *OpenTSDBModel {
	if src == nil {
		return nil
	}
	return &OpenTSDBModel{
		Query: stringValueFromPointer(src.Query),
	}
}

func pingdomToModel(src *v1alphaSLO.PingdomMetric) *PingdomModel {
	if src == nil {
		return nil
	}
	return &PingdomModel{
		CheckID:   stringValueFromPointer(src.CheckID),
		CheckType: stringValueFromPointer(src.CheckType),
		Status:    stringValueFromPointer(src.Status),
	}
}

func prometheusToModel(src *v1alphaSLO.PrometheusMetric) *PrometheusModel {
	if src == nil {
		return nil
	}
	return &PrometheusModel{
		PromQL: stringValueFromPointer(src.PromQL),
	}
}

func redshiftToModel(src *v1alphaSLO.RedshiftMetric) *RedshiftModel {
	if src == nil {
		return nil
	}
	return &RedshiftModel{
		Region:       stringValueFromPointer(src.Region),
		ClusterID:    stringValueFromPointer(src.ClusterID),
		DatabaseName: stringValueFromPointer(src.DatabaseName),
		Query:        stringValueFromPointer(src.Query),
	}
}

func splunkToModel(src *v1alphaSLO.SplunkMetric) *SplunkModel {
	if src == nil {
		return nil
	}
	return &SplunkModel{
		Query: stringValueFromPointer(src.Query),
	}
}

func splunkObservabilityToModel(src *v1alphaSLO.SplunkObservabilityMetric) *SplunkObservabilityModel {
	if src == nil {
		return nil
	}
	return &SplunkObservabilityModel{
		Program: stringValueFromPointer(src.Program),
	}
}

func sumoLogicToModel(src *v1alphaSLO.SumoLogicMetric) *SumoLogicModel {
	if src == nil {
		return nil
	}
	return &SumoLogicModel{
		Type:         stringValueFromPointer(src.Type),
		Query:        stringValueFromPointer(src.Query),
		Rollup:       stringValueFromPointer(src.Rollup),
		Quantization: stringValueFromPointer(src.Quantization),
	}
}

func thousandEyesToModel(src *v1alphaSLO.ThousandEyesMetric) *ThousandEyesModel {
	if src == nil {
		return nil
	}
	model := &ThousandEyesModel{}
	if src.TestID != nil {
		model.TestID = types.Int64Value(int64(*src.TestID))
	}
	if src.TestType != nil {
		model.TestType = types.StringValue(*src.TestType)
	}
	return model
}

// Helper functions for converting from model types to SDK types
func modelToAmazonPrometheus(model *AmazonPrometheusModel) *v1alphaSLO.AmazonPrometheusMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.AmazonPrometheusMetric{
		PromQL: stringPointer(model.PromQL),
	}
}

func modelToAppDynamics(model *AppDynamicsModel) *v1alphaSLO.AppDynamicsMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.AppDynamicsMetric{
		ApplicationName: stringPointer(model.ApplicationName),
		MetricPath:      stringPointer(model.MetricPath),
	}
}

func modelToAzureMonitor(model *AzureMonitorModel) *v1alphaSLO.AzureMonitorMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.AzureMonitorMetric{
		DataType:        model.DataType.ValueString(),
		ResourceID:      model.ResourceID.ValueString(),
		MetricNamespace: model.MetricNamespace.ValueString(),
		MetricName:      model.MetricName.ValueString(),
		Aggregation:     model.Aggregation.ValueString(),
		KQLQuery:        model.KQLQuery.ValueString(),
	}
	if len(model.Dimensions) > 0 {
		dimensions := make([]v1alphaSLO.AzureMonitorMetricDimension, len(model.Dimensions))
		for i, d := range model.Dimensions {
			dimensions[i] = v1alphaSLO.AzureMonitorMetricDimension{
				Name:  stringPointer(d.Name),
				Value: stringPointer(d.Value),
			}
		}
		spec.Dimensions = dimensions
	}
	if model.Workspace != nil {
		spec.Workspace = &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
			SubscriptionID: model.Workspace.SubscriptionID.ValueString(),
			ResourceGroup:  model.Workspace.ResourceGroup.ValueString(),
			WorkspaceID:    model.Workspace.WorkspaceID.ValueString(),
		}
	}
	return spec
}

func modelToBigQuery(model *BigQueryModel) *v1alphaSLO.BigQueryMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.BigQueryMetric{
		Location:  model.Location.ValueString(),
		ProjectID: model.ProjectID.ValueString(),
		Query:     model.Query.ValueString(),
	}
}

func modelToCloudWatch(model *CloudWatchModel) *v1alphaSLO.CloudWatchMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.CloudWatchMetric{
		Region:     stringPointer(model.Region),
		Namespace:  stringPointer(model.Namespace),
		MetricName: stringPointer(model.MetricName),
		Stat:       stringPointer(model.Stat),
		SQL:        stringPointer(model.SQL),
		JSON:       stringPointer(model.JSON),
	}
	if !isNullOrUnknown(model.AccountID) {
		accountID := model.AccountID.ValueString()
		spec.AccountID = &accountID
	}
	if len(model.Dimensions) > 0 {
		dimensions := make([]v1alphaSLO.CloudWatchMetricDimension, len(model.Dimensions))
		for i, d := range model.Dimensions {
			dimensions[i] = v1alphaSLO.CloudWatchMetricDimension{
				Name:  stringPointer(d.Name),
				Value: stringPointer(d.Value),
			}
		}
		spec.Dimensions = dimensions
	}
	return spec
}

func modelToDatadog(model *DatadogModel) *v1alphaSLO.DatadogMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.DatadogMetric{
		Query: stringPointer(model.Query),
	}
}

func modelToDynatrace(model *DynatraceModel) *v1alphaSLO.DynatraceMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.DynatraceMetric{
		MetricSelector: stringPointer(model.MetricSelector),
	}
}

func modelToElasticsearch(model *ElasticsearchModel) *v1alphaSLO.ElasticsearchMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.ElasticsearchMetric{
		Index: stringPointer(model.Index),
		Query: stringPointer(model.Query),
	}
}

func modelToGCM(model *GCMModel) *v1alphaSLO.GCMMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GCMMetric{
		ProjectID: model.ProjectID.ValueString(),
		Query:     model.Query.ValueString(),
		PromQL:    model.PromQL.ValueString(),
	}
}

func modelToGrafanaLoki(model *GrafanaLokiModel) *v1alphaSLO.GrafanaLokiMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GrafanaLokiMetric{
		Logql: stringPointer(model.Logql),
	}
}

func modelToGraphite(model *GraphiteModel) *v1alphaSLO.GraphiteMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GraphiteMetric{
		MetricPath: stringPointer(model.MetricPath),
	}
}

func modelToHoneycomb(model *HoneycombModel) *v1alphaSLO.HoneycombMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.HoneycombMetric{
		Attribute: model.Attribute.ValueString(),
	}
}

func modelToInfluxDB(model *InfluxDBModel) *v1alphaSLO.InfluxDBMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.InfluxDBMetric{
		Query: stringPointer(model.Query),
	}
}

func modelToInstana(model *InstanaModel) *v1alphaSLO.InstanaMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.InstanaMetric{
		MetricType: model.MetricType.ValueString(),
	}
	if model.Infrastructure != nil {
		spec.Infrastructure = &v1alphaSLO.InstanaInfrastructureMetricType{
			MetricRetrievalMethod: model.Infrastructure.MetricRetrievalMethod.ValueString(),
			Query:                 stringPointer(model.Infrastructure.Query),
			SnapshotID:            stringPointer(model.Infrastructure.SnapshotID),
			MetricID:              model.Infrastructure.MetricID.ValueString(),
			PluginID:              model.Infrastructure.PluginID.ValueString(),
		}
	}
	if model.Application != nil {
		app := &v1alphaSLO.InstanaApplicationMetricType{
			MetricID:         model.Application.MetricID.ValueString(),
			Aggregation:      model.Application.Aggregation.ValueString(),
			APIQuery:         model.Application.APIQuery.ValueString(),
			IncludeInternal:  model.Application.IncludeInternal.ValueBool(),
			IncludeSynthetic: model.Application.IncludeSynthetic.ValueBool(),
		}
		if model.Application.GroupBy != nil {
			app.GroupBy = v1alphaSLO.InstanaApplicationMetricGroupBy{
				Tag:               model.Application.GroupBy.Tag.ValueString(),
				TagEntity:         model.Application.GroupBy.TagEntity.ValueString(),
				TagSecondLevelKey: stringPointer(model.Application.GroupBy.TagSecondLevelKey),
			}
		}
		spec.Application = app
	}
	return spec
}

func modelToLightstep(model *LightstepModel) *v1alphaSLO.LightstepMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.LightstepMetric{
		TypeOfData: stringPointer(model.TypeOfData),
		StreamID:   stringPointer(model.StreamID),
		UQL:        stringPointer(model.UQL),
	}
	if !isNullOrUnknown(model.Percentile) {
		percentile := model.Percentile.ValueFloat64()
		spec.Percentile = &percentile
	}
	return spec
}

func modelToLogicMonitor(model *LogicMonitorModel) *v1alphaSLO.LogicMonitorMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.LogicMonitorMetric{
		QueryType: model.QueryType.ValueString(),
		Line:      model.Line.ValueString(),
	}
	if !isNullOrUnknown(model.DeviceDataSourceInstanceID) {
		id := int(model.DeviceDataSourceInstanceID.ValueInt64())
		spec.DeviceDataSourceInstanceID = id
	}
	if !isNullOrUnknown(model.GraphID) {
		id := int(model.GraphID.ValueInt64())
		spec.GraphID = id
	}
	if !isNullOrUnknown(model.WebsiteID) {
		id := model.WebsiteID.ValueString()
		spec.WebsiteID = id
	}
	if !isNullOrUnknown(model.CheckpointID) {
		id := model.CheckpointID.ValueString()
		spec.CheckpointID = id
	}
	if !isNullOrUnknown(model.GraphName) {
		name := model.GraphName.ValueString()
		spec.GraphName = name
	}
	return spec
}

func modelToNewRelic(model *NewRelicModel) *v1alphaSLO.NewRelicMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.NewRelicMetric{
		NRQL: stringPointer(model.NRQL),
	}
}

func modelToOpenTSDB(model *OpenTSDBModel) *v1alphaSLO.OpenTSDBMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.OpenTSDBMetric{
		Query: stringPointer(model.Query),
	}
}

func modelToPingdom(model *PingdomModel) *v1alphaSLO.PingdomMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.PingdomMetric{
		CheckID:   stringPointer(model.CheckID),
		CheckType: stringPointer(model.CheckType),
		Status:    stringPointer(model.Status),
	}
}

func modelToPrometheus(model *PrometheusModel) *v1alphaSLO.PrometheusMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.PrometheusMetric{
		PromQL: stringPointer(model.PromQL),
	}
}

func modelToRedshift(model *RedshiftModel) *v1alphaSLO.RedshiftMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.RedshiftMetric{
		Region:       stringPointer(model.Region),
		ClusterID:    stringPointer(model.ClusterID),
		DatabaseName: stringPointer(model.DatabaseName),
		Query:        stringPointer(model.Query),
	}
}

func modelToSplunk(model *SplunkModel) *v1alphaSLO.SplunkMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SplunkMetric{
		Query: stringPointer(model.Query),
	}
}

func modelToSplunkObservability(model *SplunkObservabilityModel) *v1alphaSLO.SplunkObservabilityMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SplunkObservabilityMetric{
		Program: stringPointer(model.Program),
	}
}

func modelToSumoLogic(model *SumoLogicModel) *v1alphaSLO.SumoLogicMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SumoLogicMetric{
		Type:         stringPointer(model.Type),
		Query:        stringPointer(model.Query),
		Rollup:       stringPointer(model.Rollup),
		Quantization: stringPointer(model.Quantization),
	}
}

func modelToThousandEyes(model *ThousandEyesModel) *v1alphaSLO.ThousandEyesMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.ThousandEyesMetric{}
	if !isNullOrUnknown(model.TestID) {
		id := model.TestID.ValueInt64()
		spec.TestID = &id
	}
	if !isNullOrUnknown(model.TestType) {
		testType := model.TestType.ValueString()
		spec.TestType = &testType
	}
	return spec
}
