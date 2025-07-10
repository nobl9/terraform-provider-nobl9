package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// SLOResourceModel describes the [SLOResource] data model.
type SLOResourceModel struct {
	Name                       string               `tfsdk:"name"`
	DisplayName                types.String         `tfsdk:"display_name"`
	Project                    string               `tfsdk:"project"`
	Description                types.String         `tfsdk:"description"`
	Annotations                map[string]string    `tfsdk:"annotations"`
	Labels                     Labels               `tfsdk:"label"`
	Service                    string               `tfsdk:"service"`
	BudgetingMethod            string               `tfsdk:"budgeting_method"`
	Tier                       types.String         `tfsdk:"tier"`
	AlertPolicies              []string             `tfsdk:"alert_policies"`
	Indicator                  []IndicatorModel     `tfsdk:"indicator"`
	Objectives                 []ObjectiveModel     `tfsdk:"objective"`
	TimeWindow                 []TimeWindowModel    `tfsdk:"time_window"`
	Attachments                []AttachmentModel    `tfsdk:"attachment"`
	AnomalyConfig              []AnomalyConfigModel `tfsdk:"anomaly_config"`
	Composite                  []CompositeV1Model   `tfsdk:"composite"`
	RetrieveHistoricalDataFrom types.String         `tfsdk:"retrieve_historical_data_from"`
}

// IndicatorModel represents [v1alphaSLO.Indicator].
type IndicatorModel struct {
	Name    string       `tfsdk:"name"`
	Project types.String `tfsdk:"project"`
	Kind    types.String `tfsdk:"kind"`
}

// ObjectiveModel represents [v1alphaSLO.Objective].
type ObjectiveModel struct {
	DisplayName     types.String              `tfsdk:"display_name"`
	Op              types.String              `tfsdk:"op"`
	Target          float64                   `tfsdk:"target"`
	TimeSliceTarget types.Float64             `tfsdk:"time_slice_target"`
	Value           types.Float64             `tfsdk:"value"`
	Name            types.String              `tfsdk:"name"`
	Primary         types.Bool                `tfsdk:"primary"`
	CountMetrics    []CountMetricsModel       `tfsdk:"count_metrics"`
	RawMetric       []RawMetricModel          `tfsdk:"raw_metric"`
	Composite       []CompositeObjectiveModel `tfsdk:"composite"`
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
	AmazonPrometheus    []AmazonPrometheusModel    `tfsdk:"amazon_prometheus"`
	AppDynamics         []AppDynamicsModel         `tfsdk:"appdynamics"`
	AzureMonitor        []AzureMonitorModel        `tfsdk:"azure_monitor"`
	BigQuery            []BigQueryModel            `tfsdk:"bigquery"`
	CloudWatch          []CloudWatchModel          `tfsdk:"cloudwatch"`
	Datadog             []DatadogModel             `tfsdk:"datadog"`
	Dynatrace           []DynatraceModel           `tfsdk:"dynatrace"`
	Elasticsearch       []ElasticsearchModel       `tfsdk:"elasticsearch"`
	GCM                 []GCMModel                 `tfsdk:"gcm"`
	GrafanaLoki         []GrafanaLokiModel         `tfsdk:"grafana_loki"`
	Graphite            []GraphiteModel            `tfsdk:"graphite"`
	Honeycomb           []HoneycombModel           `tfsdk:"honeycomb"`
	InfluxDB            []InfluxDBModel            `tfsdk:"influxdb"`
	Instana             []InstanaModel             `tfsdk:"instana"`
	Lightstep           []LightstepModel           `tfsdk:"lightstep"`
	LogicMonitor        []LogicMonitorModel        `tfsdk:"logic_monitor"`
	NewRelic            []NewRelicModel            `tfsdk:"newrelic"`
	OpenTSDB            []OpenTSDBModel            `tfsdk:"opentsdb"`
	Pingdom             []PingdomModel             `tfsdk:"pingdom"`
	Prometheus          []PrometheusModel          `tfsdk:"prometheus"`
	Redshift            []RedshiftModel            `tfsdk:"redshift"`
	Splunk              []SplunkModel              `tfsdk:"splunk"`
	SplunkObservability []SplunkObservabilityModel `tfsdk:"splunk_observability"`
	SumoLogic           []SumoLogicModel           `tfsdk:"sumologic"`
	ThousandEyes        []ThousandEyesModel        `tfsdk:"thousandeyes"`
	AzurePrometheus     []AzurePrometheusModel     `tfsdk:"azure_prometheus"`
	Coralogix           []CoralogixModel           `tfsdk:"coralogix"`
}

// CompositeObjectiveModel represents the composite block in an objective.
type CompositeObjectiveModel struct {
	MaxDelay   types.String               `tfsdk:"max_delay"`
	Components []CompositeComponentsModel `tfsdk:"components"`
}

// CompositeComponentsModel represents the components block within a composite objective.
type CompositeComponentsModel struct {
	Objectives []CompositeObjectivesModel `tfsdk:"objectives"`
}

// CompositeObjectivesModel represents the objectives block within composite components.
type CompositeObjectivesModel struct {
	CompositeObjective []CompositeObjectiveSpecModel `tfsdk:"composite_objective"`
}

// CompositeObjectiveSpecModel represents an individual composite objective specification.
type CompositeObjectiveSpecModel struct {
	Project     string  `tfsdk:"project"`
	SLO         string  `tfsdk:"slo"`
	Objective   string  `tfsdk:"objective"`
	Weight      float64 `tfsdk:"weight"`
	WhenDelayed string  `tfsdk:"when_delayed"`
}

// TimeWindowModel represents the time_window block in the SLO resource.
type TimeWindowModel struct {
	Count     int64           `tfsdk:"count"`
	IsRolling types.Bool      `tfsdk:"is_rolling"`
	Unit      string          `tfsdk:"unit"`
	Calendar  []CalendarModel `tfsdk:"calendar"`
}

// CalendarModel represents the calendar block in a time_window.
type CalendarModel struct {
	StartTime string `tfsdk:"start_time"`
	TimeZone  string `tfsdk:"time_zone"`
}

// AttachmentModel represents an attachment in the SLO resource.
type AttachmentModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	URL         types.String `tfsdk:"url"`
}

// AnomalyConfigModel represents the anomaly_config block in the SLO resource.
type AnomalyConfigModel struct {
	NoData []AnomalyConfigNoDataModel `tfsdk:"no_data"`
}

type AnomalyConfigNoDataModel struct {
	AlertAfter   types.String                    `tfsdk:"alert_after"`
	AlertMethods []AnomalyConfigAlertMethodModel `tfsdk:"alert_method"`
}

type AnomalyConfigAlertMethodModel struct {
	Name    string `tfsdk:"name"`
	Project string `tfsdk:"project"`
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
	PromQL string `tfsdk:"promql"`
}

type AppDynamicsModel struct {
	ApplicationName string `tfsdk:"application_name"`
	MetricPath      string `tfsdk:"metric_path"`
}

type AzureMonitorModel struct {
	DataType        string                       `tfsdk:"data_type"`
	ResourceID      types.String                 `tfsdk:"resource_id"`
	MetricNamespace types.String                 `tfsdk:"metric_namespace"`
	MetricName      types.String                 `tfsdk:"metric_name"`
	Aggregation     types.String                 `tfsdk:"aggregation"`
	KQLQuery        types.String                 `tfsdk:"kql_query"`
	Dimensions      []AzureMonitorDimensionModel `tfsdk:"dimensions"`
	Workspace       []AzureMonitorWorkspaceModel `tfsdk:"workspace"`
}

type AzureMonitorDimensionModel struct {
	Name  string `tfsdk:"name"`
	Value string `tfsdk:"value"`
}

type AzureMonitorWorkspaceModel struct {
	SubscriptionID string `tfsdk:"subscription_id"`
	ResourceGroup  string `tfsdk:"resource_group"`
	WorkspaceID    string `tfsdk:"workspace_id"`
}

type BigQueryModel struct {
	Location  string `tfsdk:"location"`
	ProjectID string `tfsdk:"project_id"`
	Query     string `tfsdk:"query"`
}

type CloudWatchModel struct {
	AccountID  types.String               `tfsdk:"account_id"`
	Region     string                     `tfsdk:"region"`
	Namespace  types.String               `tfsdk:"namespace"`
	MetricName types.String               `tfsdk:"metric_name"`
	Stat       types.String               `tfsdk:"stat"`
	SQL        types.String               `tfsdk:"sql"`
	JSON       types.String               `tfsdk:"json"`
	Dimensions []CloudWatchDimensionModel `tfsdk:"dimensions"`
}

type CloudWatchDimensionModel struct {
	Name  string `tfsdk:"name"`
	Value string `tfsdk:"value"`
}

type DatadogModel struct {
	Query string `tfsdk:"query"`
}

type DynatraceModel struct {
	MetricSelector string `tfsdk:"metric_selector"`
}

type ElasticsearchModel struct {
	Index string `tfsdk:"index"`
	Query string `tfsdk:"query"`
}

type GCMModel struct {
	ProjectID string       `tfsdk:"project_id"`
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
	Query string `tfsdk:"query"`
}

type InstanaModel struct {
	MetricType     string                       `tfsdk:"metric_type"`
	Infrastructure []InstanaInfrastructureModel `tfsdk:"infrastructure"`
	Application    []InstanaApplicationModel    `tfsdk:"application"`
}

type InstanaInfrastructureModel struct {
	MetricRetrievalMethod string       `tfsdk:"metric_retrieval_method"`
	Query                 types.String `tfsdk:"query"`
	SnapshotID            types.String `tfsdk:"snapshot_id"`
	MetricID              string       `tfsdk:"metric_id"`
	PluginID              string       `tfsdk:"plugin_id"`
}

type InstanaApplicationModel struct {
	MetricID         string                `tfsdk:"metric_id"`
	Aggregation      string                `tfsdk:"aggregation"`
	APIQuery         string                `tfsdk:"api_query"`
	IncludeInternal  types.Bool            `tfsdk:"include_internal"`
	IncludeSynthetic types.Bool            `tfsdk:"include_synthetic"`
	GroupBy          []InstanaGroupByModel `tfsdk:"group_by"`
}

type InstanaGroupByModel struct {
	Tag               string       `tfsdk:"tag"`
	TagEntity         string       `tfsdk:"tag_entity"`
	TagSecondLevelKey types.String `tfsdk:"tag_second_level_key"`
}

type LightstepModel struct {
	Percentile types.Float64 `tfsdk:"percentile"`
	StreamID   types.String  `tfsdk:"stream_id"`
	TypeOfData string        `tfsdk:"type_of_data"`
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
	NRQL string `tfsdk:"nrql"`
}

type OpenTSDBModel struct {
	Query string `tfsdk:"query"`
}

type PingdomModel struct {
	CheckID   string       `tfsdk:"check_id"`
	CheckType types.String `tfsdk:"check_type"`
	Status    types.String `tfsdk:"status"`
}

type PrometheusModel struct {
	PromQL string `tfsdk:"promql"`
}

type RedshiftModel struct {
	Region       string `tfsdk:"region"`
	ClusterID    string `tfsdk:"cluster_id"`
	DatabaseName string `tfsdk:"database_name"`
	Query        string `tfsdk:"query"`
}

type SplunkModel struct {
	Query string `tfsdk:"query"`
}

type SplunkObservabilityModel struct {
	Program string `tfsdk:"program"`
}

type SumoLogicModel struct {
	Type         string       `tfsdk:"type"`
	Query        string       `tfsdk:"query"`
	Quantization types.String `tfsdk:"quantization"`
	Rollup       types.String `tfsdk:"rollup"`
}

type ThousandEyesModel struct {
	TestID   int64        `tfsdk:"test_id"`
	TestType types.String `tfsdk:"test_type"`
}

type AzurePrometheusModel struct {
	PromQL string `tfsdk:"promql"`
}

type CoralogixModel struct {
	PromQL string `tfsdk:"promql"`
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
		Tier:            types.StringPointerValue(slo.Spec.Tier),
		AlertPolicies:   slo.Spec.AlertPolicies,
	}
	if slo.Spec.Indicator != nil {
		model.Indicator = []IndicatorModel{{
			Name:    slo.Spec.Indicator.MetricSource.Name,
			Project: types.StringValue(slo.Spec.Indicator.MetricSource.Project),
			Kind:    types.StringValue(slo.Spec.Indicator.MetricSource.Kind.String()),
		}}
	}
	if len(slo.Spec.Objectives) > 0 {
		objectives := make([]ObjectiveModel, len(slo.Spec.Objectives))
		for i, o := range slo.Spec.Objectives {
			obj := ObjectiveModel{
				DisplayName:     stringValue(o.DisplayName),
				Op:              types.StringPointerValue(o.Operator),
				Target:          *o.BudgetTarget,
				TimeSliceTarget: types.Float64PointerValue(o.TimeSliceTarget),
				Value:           types.Float64PointerValue(o.Value),
				Name:            types.StringValue(o.Name),
				Primary:         types.BoolPointerValue(o.Primary),
			}
			if countMetrics := countMetricsToModel(o.CountMetrics); countMetrics != nil {
				obj.CountMetrics = []CountMetricsModel{*countMetrics}
			}
			if rawMetric := rawMetricToModel(o.RawMetric); rawMetric != nil {
				obj.RawMetric = []RawMetricModel{*rawMetric}
			}
			if composite := compositeObjectiveToModel(o.Composite); composite != nil {
				obj.Composite = []CompositeObjectiveModel{*composite}
			}
			objectives[i] = obj
		}
		model.Objectives = objectives
	}
	if len(slo.Spec.TimeWindows) > 0 {
		tw := slo.Spec.TimeWindows[0]
		twModel := TimeWindowModel{
			Count:     int64(tw.Count),
			IsRolling: types.BoolValue(tw.IsRolling),
			Unit:      tw.Unit,
		}
		if tw.Calendar != nil {
			twModel.Calendar = []CalendarModel{{
				StartTime: tw.Calendar.StartTime,
				TimeZone:  tw.Calendar.TimeZone,
			}}
		}
		model.TimeWindow = []TimeWindowModel{twModel}
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
				Name:    m.Name,
				Project: m.Project,
			}
		}
		model.AnomalyConfig = []AnomalyConfigModel{{
			NoData: []AnomalyConfigNoDataModel{{
				AlertAfter:   types.StringPointerValue(ac.AlertAfter),
				AlertMethods: methods,
			}},
		}}
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
	if len(s.Indicator) > 0 {
		indicator := &s.Indicator[0]
		kind, _ := manifest.ParseKind(indicator.Kind.ValueString())
		slo.Spec.Indicator = &v1alphaSLO.Indicator{
			MetricSource: v1alphaSLO.MetricSourceSpec{
				Name:    indicator.Name,
				Project: indicator.Project.ValueString(),
				Kind:    kind,
			},
		}
	}
	if len(s.Objectives) > 0 {
		objectives := make([]v1alphaSLO.Objective, len(s.Objectives))
		for i, o := range s.Objectives {
			obj := v1alphaSLO.Objective{
				ObjectiveBase: v1alphaSLO.ObjectiveBase{
					DisplayName: o.DisplayName.ValueString(),
					Value:       o.Value.ValueFloat64Pointer(),
					Name:        o.Name.ValueString(),
				},
				Operator:        o.Op.ValueStringPointer(),
				BudgetTarget:    &o.Target,
				TimeSliceTarget: o.TimeSliceTarget.ValueFloat64Pointer(),
				Primary:         o.Primary.ValueBoolPointer(),
			}
			if len(o.CountMetrics) > 0 {
				obj.CountMetrics = o.CountMetrics[0].ToManifest()
			}
			if len(o.RawMetric) > 0 {
				obj.RawMetric = o.RawMetric[0].ToManifest()
			}
			if len(o.Composite) > 0 {
				obj.Composite = o.Composite[0].ToManifest()
			}
			objectives[i] = obj
		}
		slo.Spec.Objectives = objectives
	}
	if len(s.TimeWindow) > 0 {
		tw := &s.TimeWindow[0]
		var calendar *v1alphaSLO.Calendar
		if len(tw.Calendar) > 0 {
			cal := &tw.Calendar[0]
			calendar = &v1alphaSLO.Calendar{
				StartTime: cal.StartTime,
				TimeZone:  cal.TimeZone,
			}
		}
		slo.Spec.TimeWindows = []v1alphaSLO.TimeWindow{{
			Count:     int(tw.Count),
			IsRolling: tw.IsRolling.ValueBool(),
			Unit:      tw.Unit,
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
	if len(s.AnomalyConfig) > 0 && len(s.AnomalyConfig[0].NoData) > 0 {
		ac := &s.AnomalyConfig[0].NoData[0]
		methods := make([]v1alphaSLO.AnomalyConfigAlertMethod, len(ac.AlertMethods))
		for i, m := range ac.AlertMethods {
			methods[i] = v1alphaSLO.AnomalyConfigAlertMethod{
				Name:    m.Name,
				Project: m.Project,
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
				Project:     obj.Project,
				SLO:         obj.SLO,
				Objective:   obj.Objective,
				Weight:      obj.Weight,
				WhenDelayed: obj.WhenDelayed.String(),
			}
		}
		model.Components = []CompositeComponentsModel{{
			Objectives: []CompositeObjectivesModel{{
				CompositeObjective: compositeObjectives,
			}},
		}}
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
		MaxDelay:   c.MaxDelay.ValueString(),
		Components: v1alphaSLO.Components{Objectives: make([]v1alphaSLO.CompositeObjective, 0)},
	}
	if len(c.Components) > 0 && len(c.Components[0].Objectives) > 0 &&
		len(c.Components[0].Objectives[0].CompositeObjective) > 0 {
		objectives := c.Components[0].Objectives[0].CompositeObjective
		compositeObjectives := make([]v1alphaSLO.CompositeObjective, len(objectives))
		for i, obj := range objectives {
			whenDelayed, _ := v1alphaSLO.ParseWhenDelayed(obj.WhenDelayed)
			compositeObjectives[i] = v1alphaSLO.CompositeObjective{
				Project:     obj.Project,
				SLO:         obj.SLO,
				Objective:   obj.Objective,
				Weight:      obj.Weight,
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
	composite := s.Composite[0]
	target := composite.Target.ValueFloat64()
	result := &v1alphaSLO.Composite{
		BudgetTarget: &target,
	}
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

	model := MetricSpecModel{}

	if amazonPrometheus := amazonPrometheusToModel(spec.AmazonPrometheus); amazonPrometheus != nil {
		model.AmazonPrometheus = []AmazonPrometheusModel{*amazonPrometheus}
	}
	if appDynamics := appDynamicsToModel(spec.AppDynamics); appDynamics != nil {
		model.AppDynamics = []AppDynamicsModel{*appDynamics}
	}
	if azureMonitor := azureMonitorToModel(spec.AzureMonitor); azureMonitor != nil {
		model.AzureMonitor = []AzureMonitorModel{*azureMonitor}
	}
	if bigQuery := bigQueryToModel(spec.BigQuery); bigQuery != nil {
		model.BigQuery = []BigQueryModel{*bigQuery}
	}
	if cloudWatch := cloudWatchToModel(spec.CloudWatch); cloudWatch != nil {
		model.CloudWatch = []CloudWatchModel{*cloudWatch}
	}
	if datadog := datadogToModel(spec.Datadog); datadog != nil {
		model.Datadog = []DatadogModel{*datadog}
	}
	if dynatrace := dynatraceToModel(spec.Dynatrace); dynatrace != nil {
		model.Dynatrace = []DynatraceModel{*dynatrace}
	}
	if elasticsearch := elasticsearchToModel(spec.Elasticsearch); elasticsearch != nil {
		model.Elasticsearch = []ElasticsearchModel{*elasticsearch}
	}
	if gcm := gcmToModel(spec.GCM); gcm != nil {
		model.GCM = []GCMModel{*gcm}
	}
	if grafanaLoki := grafanaLokiToModel(spec.GrafanaLoki); grafanaLoki != nil {
		model.GrafanaLoki = []GrafanaLokiModel{*grafanaLoki}
	}
	if graphite := graphiteToModel(spec.Graphite); graphite != nil {
		model.Graphite = []GraphiteModel{*graphite}
	}
	if honeycomb := honeycombToModel(spec.Honeycomb); honeycomb != nil {
		model.Honeycomb = []HoneycombModel{*honeycomb}
	}
	if influxDB := influxDBToModel(spec.InfluxDB); influxDB != nil {
		model.InfluxDB = []InfluxDBModel{*influxDB}
	}
	if instana := instanaToModel(spec.Instana); instana != nil {
		model.Instana = []InstanaModel{*instana}
	}
	if lightstep := lightstepToModel(spec.Lightstep); lightstep != nil {
		model.Lightstep = []LightstepModel{*lightstep}
	}
	if logicMonitor := logicMonitorToModel(spec.LogicMonitor); logicMonitor != nil {
		model.LogicMonitor = []LogicMonitorModel{*logicMonitor}
	}
	if newRelic := newRelicToModel(spec.NewRelic); newRelic != nil {
		model.NewRelic = []NewRelicModel{*newRelic}
	}
	if openTSDB := openTSDBToModel(spec.OpenTSDB); openTSDB != nil {
		model.OpenTSDB = []OpenTSDBModel{*openTSDB}
	}
	if pingdom := pingdomToModel(spec.Pingdom); pingdom != nil {
		model.Pingdom = []PingdomModel{*pingdom}
	}
	if prometheus := prometheusToModel(spec.Prometheus); prometheus != nil {
		model.Prometheus = []PrometheusModel{*prometheus}
	}
	if redshift := redshiftToModel(spec.Redshift); redshift != nil {
		model.Redshift = []RedshiftModel{*redshift}
	}
	if splunk := splunkToModel(spec.Splunk); splunk != nil {
		model.Splunk = []SplunkModel{*splunk}
	}
	if splunkObservability := splunkObservabilityToModel(spec.SplunkObservability); splunkObservability != nil {
		model.SplunkObservability = []SplunkObservabilityModel{*splunkObservability}
	}
	if sumoLogic := sumoLogicToModel(spec.SumoLogic); sumoLogic != nil {
		model.SumoLogic = []SumoLogicModel{*sumoLogic}
	}
	if thousandEyes := thousandEyesToModel(spec.ThousandEyes); thousandEyes != nil {
		model.ThousandEyes = []ThousandEyesModel{*thousandEyes}
	}
	if azurePrometheus := azurePrometheusToModel(spec.AzurePrometheus); azurePrometheus != nil {
		model.AzurePrometheus = []AzurePrometheusModel{*azurePrometheus}
	}
	if coralogix := coralogixToModel(spec.Coralogix); coralogix != nil {
		model.Coralogix = []CoralogixModel{*coralogix}
	}

	return model
}

func (m MetricSpecModel) ToManifest() *v1alphaSLO.MetricSpec {
	spec := &v1alphaSLO.MetricSpec{}

	if len(m.AmazonPrometheus) > 0 {
		spec.AmazonPrometheus = modelToAmazonPrometheus(&m.AmazonPrometheus[0])
	}
	if len(m.AppDynamics) > 0 {
		spec.AppDynamics = modelToAppDynamics(&m.AppDynamics[0])
	}
	if len(m.AzureMonitor) > 0 {
		spec.AzureMonitor = modelToAzureMonitor(&m.AzureMonitor[0])
	}
	if len(m.BigQuery) > 0 {
		spec.BigQuery = modelToBigQuery(&m.BigQuery[0])
	}
	if len(m.CloudWatch) > 0 {
		spec.CloudWatch = modelToCloudWatch(&m.CloudWatch[0])
	}
	if len(m.Datadog) > 0 {
		spec.Datadog = modelToDatadog(&m.Datadog[0])
	}
	if len(m.Dynatrace) > 0 {
		spec.Dynatrace = modelToDynatrace(&m.Dynatrace[0])
	}
	if len(m.Elasticsearch) > 0 {
		spec.Elasticsearch = modelToElasticsearch(&m.Elasticsearch[0])
	}
	if len(m.GCM) > 0 {
		spec.GCM = modelToGCM(&m.GCM[0])
	}
	if len(m.GrafanaLoki) > 0 {
		spec.GrafanaLoki = modelToGrafanaLoki(&m.GrafanaLoki[0])
	}
	if len(m.Graphite) > 0 {
		spec.Graphite = modelToGraphite(&m.Graphite[0])
	}
	if len(m.Honeycomb) > 0 {
		spec.Honeycomb = modelToHoneycomb(&m.Honeycomb[0])
	}
	if len(m.InfluxDB) > 0 {
		spec.InfluxDB = modelToInfluxDB(&m.InfluxDB[0])
	}
	if len(m.Instana) > 0 {
		spec.Instana = modelToInstana(&m.Instana[0])
	}
	if len(m.Lightstep) > 0 {
		spec.Lightstep = modelToLightstep(&m.Lightstep[0])
	}
	if len(m.LogicMonitor) > 0 {
		spec.LogicMonitor = modelToLogicMonitor(&m.LogicMonitor[0])
	}
	if len(m.NewRelic) > 0 {
		spec.NewRelic = modelToNewRelic(&m.NewRelic[0])
	}
	if len(m.OpenTSDB) > 0 {
		spec.OpenTSDB = modelToOpenTSDB(&m.OpenTSDB[0])
	}
	if len(m.Pingdom) > 0 {
		spec.Pingdom = modelToPingdom(&m.Pingdom[0])
	}
	if len(m.Prometheus) > 0 {
		spec.Prometheus = modelToPrometheus(&m.Prometheus[0])
	}
	if len(m.Redshift) > 0 {
		spec.Redshift = modelToRedshift(&m.Redshift[0])
	}
	if len(m.Splunk) > 0 {
		spec.Splunk = modelToSplunk(&m.Splunk[0])
	}
	if len(m.SplunkObservability) > 0 {
		spec.SplunkObservability = modelToSplunkObservability(&m.SplunkObservability[0])
	}
	if len(m.SumoLogic) > 0 {
		spec.SumoLogic = modelToSumoLogic(&m.SumoLogic[0])
	}
	if len(m.ThousandEyes) > 0 {
		spec.ThousandEyes = modelToThousandEyes(&m.ThousandEyes[0])
	}
	if len(m.AzurePrometheus) > 0 {
		spec.AzurePrometheus = modelToAzurePrometheus(&m.AzurePrometheus[0])
	}
	if len(m.Coralogix) > 0 {
		spec.Coralogix = modelToCoralogix(&m.Coralogix[0])
	}

	return spec
}

func amazonPrometheusToModel(src *v1alphaSLO.AmazonPrometheusMetric) *AmazonPrometheusModel {
	if src == nil {
		return nil
	}
	return &AmazonPrometheusModel{
		PromQL: *src.PromQL,
	}
}

func appDynamicsToModel(src *v1alphaSLO.AppDynamicsMetric) *AppDynamicsModel {
	if src == nil {
		return nil
	}
	return &AppDynamicsModel{
		ApplicationName: *src.ApplicationName,
		MetricPath:      *src.MetricPath,
	}
}

func azureMonitorToModel(src *v1alphaSLO.AzureMonitorMetric) *AzureMonitorModel {
	if src == nil {
		return nil
	}
	model := &AzureMonitorModel{
		DataType:        src.DataType,
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
				Name:  *d.Name,
				Value: *d.Value,
			}
		}
		model.Dimensions = dimensions
	}
	if src.Workspace != nil {
		model.Workspace = []AzureMonitorWorkspaceModel{{
			SubscriptionID: src.Workspace.SubscriptionID,
			ResourceGroup:  src.Workspace.ResourceGroup,
			WorkspaceID:    src.Workspace.WorkspaceID,
		}}
	}
	return model
}

func bigQueryToModel(src *v1alphaSLO.BigQueryMetric) *BigQueryModel {
	if src == nil {
		return nil
	}
	return &BigQueryModel{
		Location:  src.Location,
		ProjectID: src.ProjectID,
		Query:     src.Query,
	}
}

func cloudWatchToModel(src *v1alphaSLO.CloudWatchMetric) *CloudWatchModel {
	if src == nil {
		return nil
	}
	model := &CloudWatchModel{
		Region:     *src.Region,
		Namespace:  types.StringPointerValue(src.Namespace),
		MetricName: types.StringPointerValue(src.MetricName),
		Stat:       types.StringPointerValue(src.Stat),
		SQL:        types.StringPointerValue(src.SQL),
		JSON:       types.StringPointerValue(src.JSON),
		AccountID:  types.StringPointerValue(src.AccountID),
	}
	if len(src.Dimensions) > 0 {
		dimensions := make([]CloudWatchDimensionModel, len(src.Dimensions))
		for i, d := range src.Dimensions {
			dimensions[i] = CloudWatchDimensionModel{
				Name:  *d.Name,
				Value: *d.Value,
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
		Query: *src.Query,
	}
}

func dynatraceToModel(src *v1alphaSLO.DynatraceMetric) *DynatraceModel {
	if src == nil {
		return nil
	}
	return &DynatraceModel{
		MetricSelector: *src.MetricSelector,
	}
}

func elasticsearchToModel(src *v1alphaSLO.ElasticsearchMetric) *ElasticsearchModel {
	if src == nil {
		return nil
	}
	return &ElasticsearchModel{
		Index: *src.Index,
		Query: *src.Query,
	}
}

func gcmToModel(src *v1alphaSLO.GCMMetric) *GCMModel {
	if src == nil {
		return nil
	}
	model := &GCMModel{
		ProjectID: src.ProjectID,
		Query:     stringValue(src.Query),
		PromQL:    stringValue(src.PromQL),
	}
	return model
}

func grafanaLokiToModel(src *v1alphaSLO.GrafanaLokiMetric) *GrafanaLokiModel {
	if src == nil {
		return nil
	}
	return &GrafanaLokiModel{
		Logql: types.StringPointerValue(src.Logql),
	}
}

func graphiteToModel(src *v1alphaSLO.GraphiteMetric) *GraphiteModel {
	if src == nil {
		return nil
	}
	return &GraphiteModel{
		MetricPath: types.StringPointerValue(src.MetricPath),
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
		Query: *src.Query,
	}
}

func instanaToModel(src *v1alphaSLO.InstanaMetric) *InstanaModel {
	if src == nil {
		return nil
	}
	model := &InstanaModel{
		MetricType: src.MetricType,
	}
	if src.Infrastructure != nil {
		model.Infrastructure = []InstanaInfrastructureModel{{
			MetricRetrievalMethod: src.Infrastructure.MetricRetrievalMethod,
			Query:                 types.StringPointerValue(src.Infrastructure.Query),
			SnapshotID:            types.StringPointerValue(src.Infrastructure.SnapshotID),
			MetricID:              src.Infrastructure.MetricID,
			PluginID:              src.Infrastructure.PluginID,
		}}
	}
	if src.Application != nil {
		app := InstanaApplicationModel{
			MetricID:         src.Application.MetricID,
			Aggregation:      src.Application.Aggregation,
			APIQuery:         src.Application.APIQuery,
			IncludeInternal:  types.BoolValue(src.Application.IncludeInternal),
			IncludeSynthetic: types.BoolValue(src.Application.IncludeSynthetic),
		}

		if src.Application.GroupBy.Tag != "" || src.Application.GroupBy.TagEntity != "" {
			app.GroupBy = []InstanaGroupByModel{{
				Tag:               src.Application.GroupBy.Tag,
				TagEntity:         src.Application.GroupBy.TagEntity,
				TagSecondLevelKey: types.StringPointerValue(src.Application.GroupBy.TagSecondLevelKey),
			}}
		}
		model.Application = []InstanaApplicationModel{app}
	}
	return model
}

func lightstepToModel(src *v1alphaSLO.LightstepMetric) *LightstepModel {
	if src == nil {
		return nil
	}
	model := &LightstepModel{
		TypeOfData: *src.TypeOfData,
		StreamID:   types.StringPointerValue(src.StreamID),
		UQL:        types.StringPointerValue(src.UQL),
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
		NRQL: *src.NRQL,
	}
}

func openTSDBToModel(src *v1alphaSLO.OpenTSDBMetric) *OpenTSDBModel {
	if src == nil {
		return nil
	}
	return &OpenTSDBModel{
		Query: *src.Query,
	}
}

func pingdomToModel(src *v1alphaSLO.PingdomMetric) *PingdomModel {
	if src == nil {
		return nil
	}
	return &PingdomModel{
		CheckID:   *src.CheckID,
		CheckType: types.StringPointerValue(src.CheckType),
		Status:    types.StringPointerValue(src.Status),
	}
}

func prometheusToModel(src *v1alphaSLO.PrometheusMetric) *PrometheusModel {
	if src == nil {
		return nil
	}
	return &PrometheusModel{
		PromQL: *src.PromQL,
	}
}

func redshiftToModel(src *v1alphaSLO.RedshiftMetric) *RedshiftModel {
	if src == nil {
		return nil
	}
	return &RedshiftModel{
		Region:       *src.Region,
		ClusterID:    *src.ClusterID,
		DatabaseName: *src.DatabaseName,
		Query:        *src.Query,
	}
}

func splunkToModel(src *v1alphaSLO.SplunkMetric) *SplunkModel {
	if src == nil {
		return nil
	}
	return &SplunkModel{
		Query: *src.Query,
	}
}

func splunkObservabilityToModel(src *v1alphaSLO.SplunkObservabilityMetric) *SplunkObservabilityModel {
	if src == nil {
		return nil
	}
	return &SplunkObservabilityModel{
		Program: *src.Program,
	}
}

func sumoLogicToModel(src *v1alphaSLO.SumoLogicMetric) *SumoLogicModel {
	if src == nil {
		return nil
	}
	return &SumoLogicModel{
		Type:         *src.Type,
		Query:        *src.Query,
		Rollup:       types.StringPointerValue(src.Rollup),
		Quantization: types.StringPointerValue(src.Quantization),
	}
}

func thousandEyesToModel(src *v1alphaSLO.ThousandEyesMetric) *ThousandEyesModel {
	if src == nil {
		return nil
	}
	model := &ThousandEyesModel{
		TestType: types.StringPointerValue(src.TestType),
	}
	if src.TestID != nil {
		model.TestID = *src.TestID
	}
	return model
}

func azurePrometheusToModel(src *v1alphaSLO.AzurePrometheusMetric) *AzurePrometheusModel {
	if src == nil {
		return nil
	}
	return &AzurePrometheusModel{
		PromQL: src.PromQL,
	}
}

func coralogixToModel(src *v1alphaSLO.CoralogixMetric) *CoralogixModel {
	if src == nil {
		return nil
	}
	return &CoralogixModel{
		PromQL: src.PromQL,
	}
}

func modelToAmazonPrometheus(model *AmazonPrometheusModel) *v1alphaSLO.AmazonPrometheusMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.AmazonPrometheusMetric{
		PromQL: &model.PromQL,
	}
}

func modelToAppDynamics(model *AppDynamicsModel) *v1alphaSLO.AppDynamicsMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.AppDynamicsMetric{
		ApplicationName: &model.ApplicationName,
		MetricPath:      &model.MetricPath,
	}
}

func modelToAzureMonitor(model *AzureMonitorModel) *v1alphaSLO.AzureMonitorMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.AzureMonitorMetric{
		DataType:        model.DataType,
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
				Name:  &d.Name,
				Value: &d.Value,
			}
		}
		spec.Dimensions = dimensions
	}
	if len(model.Workspace) > 0 {
		workspace := &model.Workspace[0]
		spec.Workspace = &v1alphaSLO.AzureMonitorMetricLogAnalyticsWorkspace{
			SubscriptionID: workspace.SubscriptionID,
			ResourceGroup:  workspace.ResourceGroup,
			WorkspaceID:    workspace.WorkspaceID,
		}
	}
	return spec
}

func modelToBigQuery(model *BigQueryModel) *v1alphaSLO.BigQueryMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.BigQueryMetric{
		Location:  model.Location,
		ProjectID: model.ProjectID,
		Query:     model.Query,
	}
}

func modelToCloudWatch(model *CloudWatchModel) *v1alphaSLO.CloudWatchMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.CloudWatchMetric{
		Region:     &model.Region,
		Namespace:  model.Namespace.ValueStringPointer(),
		MetricName: model.MetricName.ValueStringPointer(),
		Stat:       model.Stat.ValueStringPointer(),
		SQL:        model.SQL.ValueStringPointer(),
		JSON:       model.JSON.ValueStringPointer(),
	}
	if !isNullOrUnknown(model.AccountID) {
		accountID := model.AccountID.ValueString()
		spec.AccountID = &accountID
	}
	if len(model.Dimensions) > 0 {
		dimensions := make([]v1alphaSLO.CloudWatchMetricDimension, len(model.Dimensions))
		for i, d := range model.Dimensions {
			dimensions[i] = v1alphaSLO.CloudWatchMetricDimension{
				Name:  &d.Name,
				Value: &d.Value,
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
		Query: &model.Query,
	}
}

func modelToDynatrace(model *DynatraceModel) *v1alphaSLO.DynatraceMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.DynatraceMetric{
		MetricSelector: &model.MetricSelector,
	}
}

func modelToElasticsearch(model *ElasticsearchModel) *v1alphaSLO.ElasticsearchMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.ElasticsearchMetric{
		Index: &model.Index,
		Query: &model.Query,
	}
}

func modelToGCM(model *GCMModel) *v1alphaSLO.GCMMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GCMMetric{
		ProjectID: model.ProjectID,
		Query:     model.Query.ValueString(),
		PromQL:    model.PromQL.ValueString(),
	}
}

func modelToGrafanaLoki(model *GrafanaLokiModel) *v1alphaSLO.GrafanaLokiMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GrafanaLokiMetric{
		Logql: model.Logql.ValueStringPointer(),
	}
}

func modelToGraphite(model *GraphiteModel) *v1alphaSLO.GraphiteMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.GraphiteMetric{
		MetricPath: model.MetricPath.ValueStringPointer(),
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
		Query: &model.Query,
	}
}

func modelToInstana(model *InstanaModel) *v1alphaSLO.InstanaMetric {
	if model == nil {
		return nil
	}
	spec := &v1alphaSLO.InstanaMetric{
		MetricType: model.MetricType,
	}
	if len(model.Infrastructure) > 0 {
		infra := &model.Infrastructure[0]
		spec.Infrastructure = &v1alphaSLO.InstanaInfrastructureMetricType{
			MetricRetrievalMethod: infra.MetricRetrievalMethod,
			Query:                 infra.Query.ValueStringPointer(),
			SnapshotID:            infra.SnapshotID.ValueStringPointer(),
			MetricID:              infra.MetricID,
			PluginID:              infra.PluginID,
		}
	}
	if len(model.Application) > 0 {
		appModel := &model.Application[0]
		app := &v1alphaSLO.InstanaApplicationMetricType{
			MetricID:         appModel.MetricID,
			Aggregation:      appModel.Aggregation,
			APIQuery:         appModel.APIQuery,
			IncludeInternal:  appModel.IncludeInternal.ValueBool(),
			IncludeSynthetic: appModel.IncludeSynthetic.ValueBool(),
		}
		if len(appModel.GroupBy) > 0 {
			groupBy := &appModel.GroupBy[0]
			app.GroupBy = v1alphaSLO.InstanaApplicationMetricGroupBy{
				Tag:               groupBy.Tag,
				TagEntity:         groupBy.TagEntity,
				TagSecondLevelKey: groupBy.TagSecondLevelKey.ValueStringPointer(),
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
		TypeOfData: &model.TypeOfData,
		StreamID:   model.StreamID.ValueStringPointer(),
		UQL:        model.UQL.ValueStringPointer(),
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
		NRQL: &model.NRQL,
	}
}

func modelToOpenTSDB(model *OpenTSDBModel) *v1alphaSLO.OpenTSDBMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.OpenTSDBMetric{
		Query: &model.Query,
	}
}

func modelToPingdom(model *PingdomModel) *v1alphaSLO.PingdomMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.PingdomMetric{
		CheckID:   &model.CheckID,
		CheckType: model.CheckType.ValueStringPointer(),
		Status:    model.Status.ValueStringPointer(),
	}
}

func modelToPrometheus(model *PrometheusModel) *v1alphaSLO.PrometheusMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.PrometheusMetric{
		PromQL: &model.PromQL,
	}
}

func modelToRedshift(model *RedshiftModel) *v1alphaSLO.RedshiftMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.RedshiftMetric{
		Region:       &model.Region,
		ClusterID:    &model.ClusterID,
		DatabaseName: &model.DatabaseName,
		Query:        &model.Query,
	}
}

func modelToSplunk(model *SplunkModel) *v1alphaSLO.SplunkMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SplunkMetric{
		Query: &model.Query,
	}
}

func modelToSplunkObservability(model *SplunkObservabilityModel) *v1alphaSLO.SplunkObservabilityMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SplunkObservabilityMetric{
		Program: &model.Program,
	}
}

func modelToSumoLogic(model *SumoLogicModel) *v1alphaSLO.SumoLogicMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.SumoLogicMetric{
		Type:         &model.Type,
		Query:        &model.Query,
		Rollup:       model.Rollup.ValueStringPointer(),
		Quantization: model.Quantization.ValueStringPointer(),
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

func modelToAzurePrometheus(model *AzurePrometheusModel) *v1alphaSLO.AzurePrometheusMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.AzurePrometheusMetric{
		PromQL: model.PromQL,
	}
}

func modelToCoralogix(model *CoralogixModel) *v1alphaSLO.CoralogixMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.CoralogixMetric{
		PromQL: model.PromQL,
	}
}
