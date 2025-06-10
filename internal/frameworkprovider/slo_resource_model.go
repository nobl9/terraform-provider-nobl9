package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// SLOResourceModel describes the [SLOResource] data model.
type SLOResourceModel struct {
	Name                       string                     `tfsdk:"name"`
	DisplayName                types.String               `tfsdk:"display_name"`
	Project                    string                     `tfsdk:"project"`
	Description                types.String               `tfsdk:"description"`
	Annotations                map[string]string          `tfsdk:"annotations"`
	Labels                     Labels                     `tfsdk:"label"`
	Service                    types.String               `tfsdk:"service"`
	BudgetingMethod            types.String               `tfsdk:"budgeting_method"`
	Tier                       types.String               `tfsdk:"tier"`
	AlertPolicies              []string                   `tfsdk:"alert_policies"`
	RetrieveHistoricalDataFrom types.String               `tfsdk:"retrieve_historical_data_from"`
	Status                     types.Object               `tfsdk:"status"`
	Indicator                  *IndicatorModel            `tfsdk:"indicator"`
	Objectives                 []ObjectiveModel           `tfsdk:"objective"`
	TimeWindow                 *TimeWindowModel           `tfsdk:"time_window"`
	Attachments                []AttachmentModel          `tfsdk:"attachment"`
	AnomalyConfig              *AnomalyConfigModel        `tfsdk:"anomaly_config"`
	Composite                  []DeprecatedCompositeModel `tfsdk:"composite"`
}

// IndicatorModel represents the indicator block in the SLO resource.
type IndicatorModel struct {
	Name    types.String `tfsdk:"name"`
	Project types.String `tfsdk:"project"`
	Kind    types.String `tfsdk:"kind"`
}

// ObjectiveModel represents an objective in the SLO resource.
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

// CountMetricsModel represents the count_metrics block in an objective.
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
	SLOProject        types.String  `tfsdk:"slo_project"`
	SLO               types.String  `tfsdk:"slo"`
	Objective         types.String  `tfsdk:"objective"`
	Weight            types.Float64 `tfsdk:"weight"`
	BurnRateThreshold types.Float64 `tfsdk:"burn_rate_threshold"`
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

// ThresholdModel represents the trigger_threshold or clear_threshold block in an anomaly_config.
type ThresholdModel struct {
	Count      types.Int64   `tfsdk:"count"`
	Percentage types.Float64 `tfsdk:"percentage"`
}

// DeprecatedCompositeModel represents the deprecated composite block in the SLO resource.
type DeprecatedCompositeModel struct {
	Target            types.Float64                      `tfsdk:"target"`
	BurnRateCondition []DeprecatedBurnRateConditionModel `tfsdk:"burn_rate_condition"`
}

// DeprecatedBurnRateConditionModel represents the deprecated burn_rate_condition block in the composite block.
type DeprecatedBurnRateConditionModel struct {
	Op    types.String  `tfsdk:"op"`
	Value types.Float64 `tfsdk:"value"`
}

// Individual metric type models

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

var sloStatusTypes = map[string]attr.Type{
	"slo_count": types.Int64Type,
}

type SLOResourceStatusModel struct {
	SLOCount types.Int64 `tfsdk:"slo_count"`
}

func newSLOResourceConfigFromManifest(
	ctx context.Context,
	svc v1alphaSLO.SLO,
) (*SLOResourceModel, diag.Diagnostics) {
	var status types.Object
	if svc.Status != nil {
		v, diags := types.ObjectValueFrom(ctx, sloStatusTypes, SLOResourceStatusModel{
			// SLOCount: types.Int64Value(int64(svc.Status.SloCount)), // FIXME:
		})
		if diags.HasError() {
			return nil, diags
		}
		status = v
	} else {
		status = types.ObjectNull(sloStatusTypes)
	}

	// Create a basic model with the core fields
	model := &SLOResourceModel{
		Name:            svc.Metadata.Name,
		DisplayName:     stringValue(svc.Metadata.DisplayName),
		Project:         svc.Metadata.Project,
		Description:     stringValue(svc.Spec.Description),
		Annotations:     svc.Metadata.Annotations,
		Labels:          newLabelsFromManifest(svc.Metadata.Labels),
		Status:          status,
		Service:         stringValue(svc.Spec.Service),
		BudgetingMethod: stringValue(svc.Spec.BudgetingMethod),
		Tier:            stringValueFromPointer(svc.Spec.Tier),
	}

	// TODO: Add code to populate the more complex fields from the SLO manifest
	// This includes AlertPolicies, Indicator, Objectives, TimeWindow, Attachments, and AnomalyConfig

	return model, nil
}

func (s SLOResourceModel) ToManifest() v1alphaSLO.SLO {
	// Start with a basic SLO manifest
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
		},
	)

	// Set the Tier field if it's not empty
	if !s.Tier.IsNull() && !s.Tier.IsUnknown() {
		tier := s.Tier.ValueString()
		slo.Spec.Tier = &tier
	}

	// TODO: Add code to populate the more complex fields in the SLO manifest
	// This includes AlertPolicies, Indicator, Objectives, TimeWindow, Attachments, and AnomalyConfig

	return slo
}
