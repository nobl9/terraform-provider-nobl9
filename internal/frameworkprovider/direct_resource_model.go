package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DirectResourceBaseModel contains the common fields for all direct resources.
type DirectResourceBaseModel struct {
	Name                    string                        `tfsdk:"name"`
	DisplayName             types.String                  `tfsdk:"display_name"`
	Project                 string                        `tfsdk:"project"`
	Description             types.String                  `tfsdk:"description"`
	ReleaseChannel          types.String                  `tfsdk:"release_channel"`
	QueryDelay              *QueryDelayModel              `tfsdk:"query_delay"`
	HistoricalDataRetrieval *HistoricalDataRetrievalModel `tfsdk:"historical_data_retrieval"`
	LogCollectionEnabled    types.Bool                    `tfsdk:"log_collection_enabled"`
	Status                  types.String                  `tfsdk:"status"`
}

type AppDynamicsDirectResourceModel struct {
	DirectResourceBaseModel
	URL          types.String `tfsdk:"url"`
	AccountName  types.String `tfsdk:"account_name"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	ClientName   types.String `tfsdk:"client_name"`
}

func (m AppDynamicsDirectResourceModel) GetType() string {
	return "appdynamics"
}

type AzureMonitorDirectResourceModel struct {
	DirectResourceBaseModel
	TenantID     types.String `tfsdk:"tenant_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (m AzureMonitorDirectResourceModel) GetType() string {
	return "azure_monitor"
}

type BigQueryDirectResourceModel struct {
	DirectResourceBaseModel
	ServiceAccountKey types.String `tfsdk:"service_account_key"`
}

func (m BigQueryDirectResourceModel) GetType() string {
	return "bigquery"
}

type CloudWatchDirectResourceModel struct {
	DirectResourceBaseModel
	RoleARN types.String `tfsdk:"role_arn"`
}

func (m CloudWatchDirectResourceModel) GetType() string {
	return "cloudwatch"
}

type DatadogDirectResourceModel struct {
	DirectResourceBaseModel
	Site           types.String `tfsdk:"site"`
	APIKey         types.String `tfsdk:"api_key"`
	ApplicationKey types.String `tfsdk:"application_key"`
}

func (m DatadogDirectResourceModel) GetType() string {
	return "datadog"
}

type DynatraceDirectResourceModel struct {
	DirectResourceBaseModel
	URL            types.String `tfsdk:"url"`
	DynatraceToken types.String `tfsdk:"dynatrace_token"`
}

func (m DynatraceDirectResourceModel) GetType() string {
	return "dynatrace"
}

type GCMDirectResourceModel struct {
	DirectResourceBaseModel
	ServiceAccountKey types.String `tfsdk:"service_account_key"`
}

func (m GCMDirectResourceModel) GetType() string {
	return "gcm"
}

type HoneycombDirectResourceModel struct {
	DirectResourceBaseModel
	APIKey types.String `tfsdk:"api_key"`
}

func (m HoneycombDirectResourceModel) GetType() string {
	return "honeycomb"
}

type InfluxDBDirectResourceModel struct {
	DirectResourceBaseModel
	URL            types.String `tfsdk:"url"`
	APIToken       types.String `tfsdk:"api_token"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (m InfluxDBDirectResourceModel) GetType() string {
	return "influxdb"
}

type InstanaDirectResourceModel struct {
	DirectResourceBaseModel
	URL      types.String `tfsdk:"url"`
	APIToken types.String `tfsdk:"api_token"`
}

func (m InstanaDirectResourceModel) GetType() string {
	return "instana"
}

type LightstepDirectResourceModel struct {
	DirectResourceBaseModel
	Organization types.String `tfsdk:"lightstep_organization"`
	Project      types.String `tfsdk:"lightstep_project"`
	URL          types.String `tfsdk:"url"`
	AppToken     types.String `tfsdk:"app_token"`
}

func (m LightstepDirectResourceModel) GetType() string {
	return "lightstep"
}

type LogicMonitorDirectResourceModel struct {
	DirectResourceBaseModel
	Account   types.String `tfsdk:"account"`
	AccountID types.String `tfsdk:"account_id"`
	AccessKey types.String `tfsdk:"access_key"`
}

func (m LogicMonitorDirectResourceModel) GetType() string {
	return "logic_monitor"
}

type NewRelicDirectResourceModel struct {
	DirectResourceBaseModel
	AccountID        types.Int64  `tfsdk:"account_id"`
	InsightsQueryKey types.String `tfsdk:"insights_query_key"`
}

func (m NewRelicDirectResourceModel) GetType() string {
	return "newrelic"
}

type PingdomDirectResourceModel struct {
	DirectResourceBaseModel
	APIToken types.String `tfsdk:"api_token"`
}

func (m PingdomDirectResourceModel) GetType() string {
	return "pingdom"
}

type RedshiftDirectResourceModel struct {
	DirectResourceBaseModel
	RoleARN   types.String `tfsdk:"role_arn"`
	SecretARN types.String `tfsdk:"secret_arn"`
}

func (m RedshiftDirectResourceModel) GetType() string {
	return "redshift"
}

type SplunkDirectResourceModel struct {
	DirectResourceBaseModel
	URL         types.String `tfsdk:"url"`
	AccessToken types.String `tfsdk:"access_token"`
}

func (m SplunkDirectResourceModel) GetType() string {
	return "splunk"
}

type SplunkObservabilityDirectResourceModel struct {
	DirectResourceBaseModel
	Realm       types.String `tfsdk:"realm"`
	AccessToken types.String `tfsdk:"access_token"`
}

func (m SplunkObservabilityDirectResourceModel) GetType() string {
	return "splunk_observability"
}

type SumoLogicDirectResourceModel struct {
	DirectResourceBaseModel
	URL       types.String `tfsdk:"url"`
	AccessID  types.String `tfsdk:"access_id"`
	AccessKey types.String `tfsdk:"access_key"`
}

func (m SumoLogicDirectResourceModel) GetType() string {
	return "sumologic"
}

type ThousandEyesDirectResourceModel struct {
	DirectResourceBaseModel
	OauthBearerToken types.String `tfsdk:"oauth_bearer_token"`
}

func (m ThousandEyesDirectResourceModel) GetType() string {
	return "thousandeyes"
}
