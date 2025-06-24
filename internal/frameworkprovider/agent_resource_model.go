package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AgentResourceBaseModel contains the common fields for all agent resources.
type AgentResourceBaseModel struct {
	Name                    string                        `tfsdk:"name"`
	DisplayName             types.String                  `tfsdk:"display_name"`
	Project                 string                        `tfsdk:"project"`
	Description             types.String                  `tfsdk:"description"`
	AgentType               types.String                  `tfsdk:"agent_type"`
	ClientID                types.String                  `tfsdk:"client_id"`
	ClientSecret            types.String                  `tfsdk:"client_secret"`
	ReleaseChannel          types.String                  `tfsdk:"release_channel"`
	QueryDelay              *QueryDelayModel              `tfsdk:"query_delay"`
	HistoricalDataRetrieval *HistoricalDataRetrievalModel `tfsdk:"historical_data_retrieval"`
	Status                  *AgentStatusModel             `tfsdk:"status"`
}

type AmazonPrometheusAgentResourceModel struct {
	AgentResourceBaseModel
	URL    types.String `tfsdk:"url"`
	Region types.String `tfsdk:"region"`
}

func (m AmazonPrometheusAgentResourceModel) GetType() string {
	return "amazon_prometheus"
}

type AppDynamicsAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m AppDynamicsAgentResourceModel) GetType() string {
	return "appdynamics"
}

type AzureMonitorAgentResourceModel struct {
	AgentResourceBaseModel
	TenantID types.String `tfsdk:"tenant_id"`
}

func (m AzureMonitorAgentResourceModel) GetType() string {
	return "azure_monitor"
}

type BigQueryAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m BigQueryAgentResourceModel) GetType() string {
	return "bigquery"
}

type CloudWatchAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m CloudWatchAgentResourceModel) GetType() string {
	return "cloudwatch"
}

type DatadogAgentResourceModel struct {
	AgentResourceBaseModel
	Site types.String `tfsdk:"site"`
}

func (m DatadogAgentResourceModel) GetType() string {
	return "datadog"
}

type DynatraceAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m DynatraceAgentResourceModel) GetType() string {
	return "dynatrace"
}

type ElasticsearchAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m ElasticsearchAgentResourceModel) GetType() string {
	return "elasticsearch"
}

type GCMAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m GCMAgentResourceModel) GetType() string {
	return "gcm"
}

type GrafanaLokiAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m GrafanaLokiAgentResourceModel) GetType() string {
	return "grafana_loki"
}

type GraphiteAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m GraphiteAgentResourceModel) GetType() string {
	return "graphite"
}

type HoneycombAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m HoneycombAgentResourceModel) GetType() string {
	return "honeycomb"
}

type InfluxDBAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m InfluxDBAgentResourceModel) GetType() string {
	return "influxdb"
}

type InstanaAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m InstanaAgentResourceModel) GetType() string {
	return "instana"
}

type LightstepAgentResourceModel struct {
	AgentResourceBaseModel
	Organization types.String `tfsdk:"organization"`
	Project      types.String `tfsdk:"project"`
	URL          types.String `tfsdk:"url"`
}

func (m LightstepAgentResourceModel) GetType() string {
	return "lightstep"
}

type LogicMonitorAgentResourceModel struct {
	AgentResourceBaseModel
	Account types.String `tfsdk:"account"`
}

func (m LogicMonitorAgentResourceModel) GetType() string {
	return "logic_monitor"
}

type NewRelicAgentResourceModel struct {
	AgentResourceBaseModel
	AccountID types.String `tfsdk:"account_id"`
}

func (m NewRelicAgentResourceModel) GetType() string {
	return "newrelic"
}

type OpenTSDBAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m OpenTSDBAgentResourceModel) GetType() string {
	return "opentsdb"
}

type PingdomAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m PingdomAgentResourceModel) GetType() string {
	return "pingdom"
}

type PrometheusAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m PrometheusAgentResourceModel) GetType() string {
	return "prometheus"
}

type RedshiftAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m RedshiftAgentResourceModel) GetType() string {
	return "redshift"
}

type SplunkAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m SplunkAgentResourceModel) GetType() string {
	return "splunk"
}

type SplunkObservabilityAgentResourceModel struct {
	AgentResourceBaseModel
	Realm types.String `tfsdk:"realm"`
}

func (m SplunkObservabilityAgentResourceModel) GetType() string {
	return "splunk_observability"
}

type SumoLogicAgentResourceModel struct {
	AgentResourceBaseModel
	URL types.String `tfsdk:"url"`
}

func (m SumoLogicAgentResourceModel) GetType() string {
	return "sumologic"
}

type ThousandEyesAgentResourceModel struct {
	AgentResourceBaseModel
}

func (m ThousandEyesAgentResourceModel) GetType() string {
	return "thousandeyes"
}

type QueryDelayModel struct {
	Unit  types.String `tfsdk:"unit"`
	Value types.Int64  `tfsdk:"value"`
}

type HistoricalDataRetrievalModel struct {
	DefaultDuration *DurationModel `tfsdk:"default_duration"`
	MaxDuration     *DurationModel `tfsdk:"max_duration"`
}

type DurationModel struct {
	Unit  types.String `tfsdk:"unit"`
	Value types.Int64  `tfsdk:"value"`
}

type AgentStatusModel struct {
	AgentType      types.String `tfsdk:"agent_type"`
	AgentVersion   types.String `tfsdk:"agent_version"`
	LastConnection types.String `tfsdk:"last_connection"`
}
