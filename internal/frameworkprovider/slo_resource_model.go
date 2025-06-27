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
	Service                    string              `tfsdk:"service"`
	BudgetingMethod            string              `tfsdk:"budgeting_method"`
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
	Name    string       `tfsdk:"name"`    // Required
	Project types.String `tfsdk:"project"` // Optional
	Kind    types.String `tfsdk:"kind"`    // Optional (computed)
}

// ObjectiveModel represents [v1alphaSLO.Objective].
type ObjectiveModel struct {
	DisplayName     types.String             `tfsdk:"display_name"`
	Op              types.String             `tfsdk:"op"`
	Target          float64                  `tfsdk:"target"`
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
	Project     string  `tfsdk:"project"`      // Required
	SLO         string  `tfsdk:"slo"`          // Required
	Objective   string  `tfsdk:"objective"`    // Required
	Weight      float64 `tfsdk:"weight"`       // Required
	WhenDelayed string  `tfsdk:"when_delayed"` // Required
}

// TimeWindowModel represents the time_window block in the SLO resource.
type TimeWindowModel struct {
	Count     int64             `tfsdk:"count"`      // Required
	IsRolling types.Bool        `tfsdk:"is_rolling"` // Optional
	Unit      string            `tfsdk:"unit"`       // Required
	Period    map[string]string `tfsdk:"period"`     // Computed
	Calendar  *CalendarModel    `tfsdk:"calendar"`   // Optional
}

// CalendarModel represents the calendar block in a time_window.
type CalendarModel struct {
	StartTime string `tfsdk:"start_time"` // Required
	TimeZone  string `tfsdk:"time_zone"`  // Required
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
	Name    string `tfsdk:"name"`    // Required
	Project string `tfsdk:"project"` // Required
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
	PromQL string `tfsdk:"promql"` // Required
}

type AppDynamicsModel struct {
	ApplicationName string `tfsdk:"application_name"` // Required
	MetricPath      string `tfsdk:"metric_path"`      // Required
}

type AzureMonitorModel struct {
	DataType        string                       `tfsdk:"data_type"`        // Required
	ResourceID      types.String                 `tfsdk:"resource_id"`      // Optional (Required for metrics)
	MetricNamespace types.String                 `tfsdk:"metric_namespace"` // Optional
	MetricName      types.String                 `tfsdk:"metric_name"`      // Optional (Required for metrics)
	Aggregation     types.String                 `tfsdk:"aggregation"`      // Optional (Required for metrics)
	KQLQuery        types.String                 `tfsdk:"kql_query"`        // Optional (Required for logs)
	Dimensions      []AzureMonitorDimensionModel `tfsdk:"dimensions"`       // Optional
	Workspace       *AzureMonitorWorkspaceModel  `tfsdk:"workspace"`        // Optional (Required for logs)
}

type AzureMonitorDimensionModel struct {
	Name  string `tfsdk:"name"`  // Required
	Value string `tfsdk:"value"` // Required
}

type AzureMonitorWorkspaceModel struct {
	SubscriptionID string `tfsdk:"subscription_id"` // Required
	ResourceGroup  string `tfsdk:"resource_group"`  // Required
	WorkspaceID    string `tfsdk:"workspace_id"`    // Required
}

type BigQueryModel struct {
	Location  string `tfsdk:"location"`   // Required
	ProjectID string `tfsdk:"project_id"` // Required
	Query     string `tfsdk:"query"`      // Required
}

type CloudWatchModel struct {
	AccountID  types.String               `tfsdk:"account_id"`  // Optional
	Region     string                     `tfsdk:"region"`      // Required
	Namespace  types.String               `tfsdk:"namespace"`   // Optional
	MetricName types.String               `tfsdk:"metric_name"` // Optional
	Stat       types.String               `tfsdk:"stat"`        // Optional
	SQL        types.String               `tfsdk:"sql"`         // Optional
	JSON       types.String               `tfsdk:"json"`        // Optional
	Dimensions []CloudWatchDimensionModel `tfsdk:"dimensions"`  // Optional
}

type CloudWatchDimensionModel struct {
	Name  string `tfsdk:"name"`  // Required
	Value string `tfsdk:"value"` // Required
}

type DatadogModel struct {
	Query string `tfsdk:"query"` // Required
}

type DynatraceModel struct {
	MetricSelector string `tfsdk:"metric_selector"` // Required
}

type ElasticsearchModel struct {
	Index string `tfsdk:"index"` // Required
	Query string `tfsdk:"query"` // Required
}

type GCMModel struct {
	ProjectID string       `tfsdk:"project_id"` // Required
	Query     types.String `tfsdk:"query"`      // Optional
	PromQL    types.String `tfsdk:"promql"`     // Optional
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
	NRQL string `tfsdk:"nrql"` // Required
}

type OpenTSDBModel struct {
	Query string `tfsdk:"query"` // Required
}

type PingdomModel struct {
	CheckID   string       `tfsdk:"check_id"`   // Required
	CheckType types.String `tfsdk:"check_type"` // Optional
	Status    types.String `tfsdk:"status"`     // Optional
}

type PrometheusModel struct {
	PromQL string `tfsdk:"promql"` // Required
}

type RedshiftModel struct {
	Region       string `tfsdk:"region"`        // Required
	ClusterID    string `tfsdk:"cluster_id"`    // Required
	DatabaseName string `tfsdk:"database_name"` // Required
	Query        string `tfsdk:"query"`         // Required
}

type SplunkModel struct {
	Query string `tfsdk:"query"` // Required
}

type SplunkObservabilityModel struct {
	Program string `tfsdk:"program"` // Required
}

type SumoLogicModel struct {
	Type         string       `tfsdk:"type"`         // Required
	Query        string       `tfsdk:"query"`        // Required
	Quantization types.String `tfsdk:"quantization"` // Optional
	Rollup       types.String `tfsdk:"rollup"`       // Optional
}

type ThousandEyesModel struct {
	TestID   int64        `tfsdk:"test_id"`   // Required
	TestType types.String `tfsdk:"test_type"` // Optional
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
		Service:         slo.Spec.Service,
		BudgetingMethod: slo.Spec.BudgetingMethod,
		Tier:            stringValueFromPointer(slo.Spec.Tier),
		AlertPolicies:   slo.Spec.AlertPolicies,
	}
	if slo.Spec.Indicator != nil {
		model.Indicator = &IndicatorModel{
			Name:    slo.Spec.Indicator.MetricSource.Name, // Required field, use string directly
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
				Target:          *o.BudgetTarget,
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
			Count:     int64(tw.Count), // Required field, use int64 directly
			IsRolling: types.BoolValue(tw.IsRolling),
			Unit:      tw.Unit, // Required field, use string directly
			Period:    map[string]string{"begin": tw.Period.Begin, "end": tw.Period.End},
		}
		if tw.Calendar != nil {
			model.TimeWindow.Calendar = &CalendarModel{
				StartTime: tw.Calendar.StartTime, // Required field, use string directly
				TimeZone:  tw.Calendar.TimeZone,  // Required field, use string directly
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
				Name:    m.Name,    // Required field, use string directly
				Project: m.Project, // Required field, use string directly
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
			Service:         s.Service,
			BudgetingMethod: s.BudgetingMethod,
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
				Name:    s.Indicator.Name, // Required field, use string directly
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
				BudgetTarget:    &o.Target,
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
				StartTime: tw.Calendar.StartTime, // Required field, use string directly
				TimeZone:  tw.Calendar.TimeZone,  // Required field, use string directly
			}
		}
		slo.Spec.TimeWindows = []v1alphaSLO.TimeWindow{{
			Count:     int(tw.Count), // Required field, use int directly
			IsRolling: tw.IsRolling.ValueBool(),
			Unit:      tw.Unit, // Required field, use string directly
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
				Name:    m.Name,    // Required field, use string directly
				Project: m.Project, // Required field, use string directly
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
				Project:     obj.Project,              // Required field, use string directly
				SLO:         obj.SLO,                  // Required field, use string directly
				Objective:   obj.Objective,            // Required field, use string directly
				Weight:      obj.Weight,               // Required field, use float64 directly
				WhenDelayed: obj.WhenDelayed.String(), // Required field, use string directly
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
			whenDelayed, _ := v1alphaSLO.ParseWhenDelayed(obj.WhenDelayed) // Required field, use string directly
			compositeObjectives[i] = v1alphaSLO.CompositeObjective{
				Project:     obj.Project,   // Required field, use string directly
				SLO:         obj.SLO,       // Required field, use string directly
				Objective:   obj.Objective, // Required field, use string directly
				Weight:      obj.Weight,    // Required field, use float64 directly
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
		PromQL: *src.PromQL, // Required field, directly dereference
	}
}

func appDynamicsToModel(src *v1alphaSLO.AppDynamicsMetric) *AppDynamicsModel {
	if src == nil {
		return nil
	}
	return &AppDynamicsModel{
		ApplicationName: *src.ApplicationName, // Required field, directly dereference
		MetricPath:      *src.MetricPath,      // Required field, directly dereference
	}
}

func azureMonitorToModel(src *v1alphaSLO.AzureMonitorMetric) *AzureMonitorModel {
	if src == nil {
		return nil
	}
	model := &AzureMonitorModel{
		DataType:        src.DataType, // Required field, use directly
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
				Name:  *d.Name,  // Required field, directly dereference
				Value: *d.Value, // Required field, directly dereference
			}
		}
		model.Dimensions = dimensions
	}
	if src.Workspace != nil {
		model.Workspace = &AzureMonitorWorkspaceModel{
			SubscriptionID: src.Workspace.SubscriptionID, // Required field, use directly
			ResourceGroup:  src.Workspace.ResourceGroup,  // Required field, use directly
			WorkspaceID:    src.Workspace.WorkspaceID,    // Required field, use directly
		}
	}
	return model
}

func bigQueryToModel(src *v1alphaSLO.BigQueryMetric) *BigQueryModel {
	if src == nil {
		return nil
	}
	return &BigQueryModel{
		Location:  src.Location,  // Required field, use directly
		ProjectID: src.ProjectID, // Required field, use directly
		Query:     src.Query,     // Required field, use directly
	}
}

func cloudWatchToModel(src *v1alphaSLO.CloudWatchMetric) *CloudWatchModel {
	if src == nil {
		return nil
	}
	model := &CloudWatchModel{
		Region:     *src.Region, // Required field, directly dereference
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
				Name:  *d.Name,  // Required field
				Value: *d.Value, // Required field
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
		Query: *src.Query, // Required field
	}
}

func dynatraceToModel(src *v1alphaSLO.DynatraceMetric) *DynatraceModel {
	if src == nil {
		return nil
	}
	return &DynatraceModel{
		MetricSelector: *src.MetricSelector, // Required field
	}
}

func elasticsearchToModel(src *v1alphaSLO.ElasticsearchMetric) *ElasticsearchModel {
	if src == nil {
		return nil
	}
	return &ElasticsearchModel{
		Index: *src.Index, // Required field
		Query: *src.Query, // Required field
	}
}

func gcmToModel(src *v1alphaSLO.GCMMetric) *GCMModel {
	if src == nil {
		return nil
	}
	model := &GCMModel{
		ProjectID: src.ProjectID, // Required field
	}
	if len(src.Query) > 0 {
		model.Query = types.StringValue(src.Query)
	}
	if len(src.PromQL) > 0 {
		model.PromQL = types.StringValue(src.PromQL)
	}
	return model
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
		NRQL: *src.NRQL, // Required field
	}
}

func openTSDBToModel(src *v1alphaSLO.OpenTSDBMetric) *OpenTSDBModel {
	if src == nil {
		return nil
	}
	return &OpenTSDBModel{
		Query: *src.Query, // Required field
	}
}

func pingdomToModel(src *v1alphaSLO.PingdomMetric) *PingdomModel {
	if src == nil {
		return nil
	}
	return &PingdomModel{
		CheckID:   *src.CheckID, // Required field
		CheckType: stringValueFromPointer(src.CheckType),
		Status:    stringValueFromPointer(src.Status),
	}
}

func prometheusToModel(src *v1alphaSLO.PrometheusMetric) *PrometheusModel {
	if src == nil {
		return nil
	}
	return &PrometheusModel{
		PromQL: *src.PromQL, // Required field
	}
}

func redshiftToModel(src *v1alphaSLO.RedshiftMetric) *RedshiftModel {
	if src == nil {
		return nil
	}
	return &RedshiftModel{
		Region:       *src.Region,       // Required field
		ClusterID:    *src.ClusterID,    // Required field
		DatabaseName: *src.DatabaseName, // Required field
		Query:        *src.Query,        // Required field
	}
}

func splunkToModel(src *v1alphaSLO.SplunkMetric) *SplunkModel {
	if src == nil {
		return nil
	}
	return &SplunkModel{
		Query: *src.Query, // Required field
	}
}

func splunkObservabilityToModel(src *v1alphaSLO.SplunkObservabilityMetric) *SplunkObservabilityModel {
	if src == nil {
		return nil
	}
	return &SplunkObservabilityModel{
		Program: *src.Program, // Required field
	}
}

func sumoLogicToModel(src *v1alphaSLO.SumoLogicMetric) *SumoLogicModel {
	if src == nil {
		return nil
	}
	return &SumoLogicModel{
		Type:         *src.Type,  // Required field
		Query:        *src.Query, // Required field
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
		model.TestID = int64(*src.TestID) // Required field, use plain int64
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
		PromQL: &model.PromQL, // Required field, use pointer to string
	}
}

func modelToAppDynamics(model *AppDynamicsModel) *v1alphaSLO.AppDynamicsMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.AppDynamicsMetric{
		ApplicationName: &model.ApplicationName, // Required field, use pointer to string
		MetricPath:      &model.MetricPath,      // Required field, use pointer to string
	}
}

func modelToAzureMonitor(model *AzureMonitorModel) *v1alphaSLO.AzureMonitorMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.AzureMonitorMetric{
		DataType:        model.DataType, // Required field, use string directly
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
				Name:  &d.Name,  // Required field, use pointer to string
				Value: &d.Value, // Required field, use pointer to string
			}
		}
		spec.Dimensions = dimensions
	}
	if model.Workspace != nil {
		spec.Workspace = &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
			SubscriptionID: model.Workspace.SubscriptionID, // Required field, use string directly
			ResourceGroup:  model.Workspace.ResourceGroup,  // Required field, use string directly
			WorkspaceID:    model.Workspace.WorkspaceID,    // Required field, use string directly
		}
	}
	return spec
}

func modelToBigQuery(model *BigQueryModel) *v1alphaSLO.BigQueryMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.BigQueryMetric{
		Location:  model.Location,  // Required field, use string directly
		ProjectID: model.ProjectID, // Required field, use string directly
		Query:     model.Query,     // Required field, use string directly
	}
}

func modelToCloudWatch(model *CloudWatchModel) *v1alphaSLO.CloudWatchMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.CloudWatchMetric{
		Region:     &model.Region, // Required field, use pointer to string
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
				Name:  &d.Name,  // Required field, use pointer to string
				Value: &d.Value, // Required field, use pointer to string
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
		Query: &model.Query, // Required field, use pointer to string
	}
}

func modelToDynatrace(model *DynatraceModel) *v1alphaSLO.DynatraceMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.DynatraceMetric{
		MetricSelector: &model.MetricSelector, // Required field, use pointer to string
	}
}

func modelToElasticsearch(model *ElasticsearchModel) *v1alphaSLO.ElasticsearchMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.ElasticsearchMetric{
		Index: &model.Index, // Required field, use pointer to string
		Query: &model.Query, // Required field, use pointer to string
	}
}

func modelToGCM(model *GCMModel) *v1alphaSLO.GCMMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GCMMetric{
		ProjectID: model.ProjectID, // Required field, use string directly
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
		NRQL: &model.NRQL, // Required field, use pointer to string
	}
}

func modelToOpenTSDB(model *OpenTSDBModel) *v1alphaSLO.OpenTSDBMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.OpenTSDBMetric{
		Query: &model.Query, // Required field, use pointer to string
	}
}

func modelToPingdom(model *PingdomModel) *v1alphaSLO.PingdomMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.PingdomMetric{
		CheckID:   &model.CheckID, // Required field, use pointer to string
		CheckType: stringPointer(model.CheckType),
		Status:    stringPointer(model.Status),
	}
}

func modelToPrometheus(model *PrometheusModel) *v1alphaSLO.PrometheusMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.PrometheusMetric{
		PromQL: &model.PromQL, // Required field, use pointer to string
	}
}

func modelToRedshift(model *RedshiftModel) *v1alphaSLO.RedshiftMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.RedshiftMetric{
		Region:       &model.Region,       // Required field, use pointer to string
		ClusterID:    &model.ClusterID,    // Required field, use pointer to string
		DatabaseName: &model.DatabaseName, // Required field, use pointer to string
		Query:        &model.Query,        // Required field, use pointer to string
	}
}

func modelToSplunk(model *SplunkModel) *v1alphaSLO.SplunkMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SplunkMetric{
		Query: &model.Query, // Required field, use pointer to string
	}
}

func modelToSplunkObservability(model *SplunkObservabilityModel) *v1alphaSLO.SplunkObservabilityMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SplunkObservabilityMetric{
		Program: &model.Program, // Required field, use pointer to string
	}
}

func modelToSumoLogic(model *SumoLogicModel) *v1alphaSLO.SumoLogicMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SumoLogicMetric{
		Type:         &model.Type,  // Required field, use pointer to string
		Query:        &model.Query, // Required field, use pointer to string
		Rollup:       stringPointer(model.Rollup),
		Quantization: stringPointer(model.Quantization),
	}
}

func modelToThousandEyes(model *ThousandEyesModel) *v1alphaSLO.ThousandEyesMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.ThousandEyesMetric{}
	// TestID is required, use direct assignment
	id := model.TestID
	spec.TestID = &id

	if !isNullOrUnknown(model.TestType) {
		testType := model.TestType.ValueString()
		spec.TestType = &testType
	}
	return spec
}
