package nobl9

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

type directResource struct {
	directSpecResource
}

type directSpecResource interface {
	GetSchema() map[string]*schema.Schema
	GetDescription() string
	MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec
	UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics)
}

func resourceDirectFactory(directSpec directSpecResource) *schema.Resource {
	i := directResource{directSpecResource: directSpec}
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),
			"source_of": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				MaxItems:    2,
				Deprecated:  "'source_of' is deprecated and not used anywhere. You can safely remove it from your configuration file.",
				Description: "This value indicated whether the field was a source of metrics and/or services. 'source_of' is deprecated and not used anywhere; however, it's kept for backward compatibility.",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "This value indicated whether the field was a source of metrics and/or services. 'source_of' is deprecated and not used anywhere; however, it's kept for backward compatibility.",
				},
			},
			releaseChannel:      schemaReleaseChannel(),
			queryDelayConfigKey: schemaQueryDelay(),
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the created direct.",
			},
		},
		CreateContext: i.resourceDirectApply,
		UpdateContext: i.resourceDirectApply,
		DeleteContext: i.resourceDirectDelete,
		ReadContext:   i.resourceDirectRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: directSpec.GetDescription(),
	}

	for k, v := range directSpec.GetSchema() {
		r.Schema[k] = v
	}

	return r
}

func (dr directResource) resourceDirectApply(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client := getClient(config)

	n9Direct, diags := dr.marshalDirect(d)
	if diags.HasError() {
		return diags
	}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		err := client.Objects().V1().Apply(ctx, []manifest.Object{n9Direct})
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.Errorf("could not add direct: %s", err.Error())
	}

	d.SetId(n9Direct.Metadata.Name)

	readDirectDiags := dr.resourceDirectRead(ctx, d, meta)

	return append(diags, readDirectDiags...)
}

func (dr directResource) resourceDirectRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client := getClient(config)

	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	directs, err := client.Objects().V1().GetV1alphaDirects(ctx, v1Objects.GetDirectsRequest{
		Project: project,
		Names:   []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return dr.unmarshalDirect(d, directs)
}

func (dr directResource) resourceDirectDelete(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client := getClient(config)

	project := d.Get("project").(string)
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		err := client.Objects().V1().DeleteByName(ctx, manifest.KindDirect, project, d.Id())
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (dr directResource) marshalDirect(d *schema.ResourceData) (*v1alphaDirect.Direct, diag.Diagnostics) {
	var diags diag.Diagnostics

	spec := dr.MarshalSpec(d)
	spec.Description = d.Get("description").(string)
	spec.HistoricalDataRetrieval = marshalHistoricalDataRetrieval(d)
	spec.QueryDelay = marshalQueryDelay(d)
	spec.ReleaseChannel = marshalReleaseChannel(d, &diags)

	if d.GetRawConfig().Type().HasAttribute(logCollectionConfigKey) &&
		!d.GetRawConfig().GetAttr(logCollectionConfigKey).IsNull() {
		spec.LogCollectionEnabled = marshalLogCollectionEnabled(d)
	}

	var displayName string
	if dn := d.Get("display_name"); dn != nil {
		displayName = dn.(string)
	}

	direct := v1alphaDirect.New(
		v1alphaDirect.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: displayName,
			Project:     d.Get("project").(string),
		},
		spec,
	)
	return &direct, diags
}

func (dr directResource) unmarshalDirect(d *schema.ResourceData, directs []v1alphaDirect.Direct) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(directs) != 1 {
		d.SetId("")
		return nil
	}
	direct := directs[0]

	set(d, "status", direct.Status.DirectType, &diags)
	set(d, "name", direct.Metadata.Name, &diags)
	set(d, "display_name", direct.Metadata.DisplayName, &diags)
	set(d, "project", direct.Metadata.Project, &diags)
	diags = append(diags, dr.UnmarshalSpec(d, direct.Spec)...)
	diags = append(diags, unmarshalHistoricalDataRetrieval(d, direct.Spec.HistoricalDataRetrieval)...)
	diags = append(diags, unmarshalQueryDelay(d, direct.Spec.QueryDelay)...)
	diags = append(diags, unmarshalLogCollectionEnabled(d, direct.Spec.LogCollectionEnabled)...)
	diags = append(diags, unmarshalReleaseChannel(d, direct.Spec.ReleaseChannel)...)

	return diags
}

// AppDynamics Direct
// https://docs.nobl9.com/Sources/appdynamics#appdynamics-direct
const appDynamicsDirectType = "appdynamics"

type appDynamicsDirectSpec struct{}

func (s appDynamicsDirectSpec) GetSchema() map[string]*schema.Schema {
	appDynamicsSchema := map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Description: "Base URL to the AppDynamics Controller.",
			Required:    true,
		},
		"account_name": {
			Type:        schema.TypeString,
			Description: "AppDynamics Account Name.",
			Required:    true,
		},
		"client_id": {
			Type:        schema.TypeString,
			Description: "AppDynamics Client ID.",
			Computed:    true,
		},
		"client_secret": {
			Type:        schema.TypeString,
			Description: "[required] | AppDynamics Client Secret.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
		"client_name": {
			Type:        schema.TypeString,
			Description: "AppDynamics Client Name.",
			Required:    true,
		},
	}
	setLogCollectionSchema(appDynamicsSchema)
	setHistoricalDataRetrievalSchema(appDynamicsSchema)

	return appDynamicsSchema
}

func (s appDynamicsDirectSpec) GetDescription() string {
	return "[AppDynamics Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/appdynamics#appdynamics-direct)"
}

func (s appDynamicsDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		AppDynamics: &v1alphaDirect.AppDynamicsConfig{
			URL:          d.Get("url").(string),
			AccountName:  d.Get("account_name").(string),
			ClientID:     d.Get("client_id").(string),
			ClientSecret: d.Get("client_secret").(string),
			ClientName:   d.Get("client_name").(string),
		},
	}
}

func (s appDynamicsDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "url", spec.AppDynamics.URL, &diags)
	set(d, "account_name", spec.AppDynamics.AccountName, &diags)
	set(d, "client_id", spec.AppDynamics.ClientID, &diags)
	set(d, "client_name", spec.AppDynamics.ClientName, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Azure Monitor Direct
// https://docs.nobl9.com/Sources/azure-monitor#azure-monitor-direct
const azureMonitorDirectType = "azure_monitor"

type azureMonitorDirectSpec struct{}

func (s azureMonitorDirectSpec) GetSchema() map[string]*schema.Schema {
	azureMonitorSchema := map[string]*schema.Schema{
		"tenant_id": {
			Type:        schema.TypeString,
			Description: "[required] | Azure Tenant ID.",
			Required:    true,
		},
		"client_id": {
			Type:        schema.TypeString,
			Description: "[required] | Azure Application (client) ID.",
			Computed:    true,
			Optional:    true,
			Sensitive:   true,
		},
		"client_secret": {
			Type:        schema.TypeString,
			Description: "[required] | Azure Application (client) Secret.",
			Computed:    true,
			Optional:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(azureMonitorSchema)
	setHistoricalDataRetrievalSchema(azureMonitorSchema)

	return azureMonitorSchema
}

// nolint: lll
func (s azureMonitorDirectSpec) GetDescription() string {
	return "[Azure Monitor Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/azure-monitor#azure-monitor-direct)"
}

func (s azureMonitorDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		AzureMonitor: &v1alphaDirect.AzureMonitorConfig{
			TenantID:     d.Get("tenant_id").(string),
			ClientID:     d.Get("client_id").(string),
			ClientSecret: d.Get("client_secret").(string),
		},
	}
}

func (s azureMonitorDirectSpec) UnmarshalSpec(
	d *schema.ResourceData,
	spec v1alphaDirect.Spec,
) (diags diag.Diagnostics) {
	set(d, "tenant_id", spec.AzureMonitor.TenantID, &diags)
	return
}

// BigQuery Direct
// https://docs.nobl9.com/Sources/bigquery#bigquery-direct
const bigqueryDirectType = "bigquery"

type bigqueryDirectSpec struct{}

func (s bigqueryDirectSpec) GetSchema() map[string]*schema.Schema {
	bigQuerySchema := map[string]*schema.Schema{
		"service_account_key": {
			Type:        schema.TypeString,
			Description: "[required] | Service Account Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(bigQuerySchema)

	return bigQuerySchema
}

func (s bigqueryDirectSpec) GetDescription() string {
	return "[BigQuery Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/bigquery#bigquery-direct)"
}

func (s bigqueryDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		BigQuery: &v1alphaDirect.BigQueryConfig{
			ServiceAccountKey: d.Get("service_account_key").(string),
		},
	}
}

func (s bigqueryDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "description", spec.Description, &diags)
	return
}

// Amazon CloudWatch Direct
// https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cloudwatch-direct
const cloudWatchDirectType = "cloudwatch"

type cloudWatchDirectSpec struct{}

func (s cloudWatchDirectSpec) GetSchema() map[string]*schema.Schema {
	cloudWatchSchema := map[string]*schema.Schema{
		"role_arn": {
			Type:        schema.TypeString,
			Description: "[required] | ARN of the AWS IAM Role to assume.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setHistoricalDataRetrievalSchema(cloudWatchSchema)
	setLogCollectionSchema(cloudWatchSchema)

	return cloudWatchSchema
}

// nolint: lll
func (s cloudWatchDirectSpec) GetDescription() string {
	return "[Amazon CloudWatch Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cloudwatch-direct)"
}

func (s cloudWatchDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		CloudWatch: &v1alphaDirect.CloudWatchConfig{
			RoleARN: d.Get("role_arn").(string),
		},
	}
}

func (s cloudWatchDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "description", spec.Description, &diags)
	return
}

// Datadog Direct
// https://docs.nobl9.com/Sources/datadog#datadog-direct
const datadogDirectType = "datadog"

type datadogDirectSpec struct{}

func (s datadogDirectSpec) GetDescription() string {
	return "[Datadog Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/datadog#datadog-direct)."
}

func (s datadogDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Datadog: &v1alphaDirect.DatadogConfig{
			Site:           d.Get("site").(string),
			APIKey:         d.Get("api_key").(string),
			ApplicationKey: d.Get("application_key").(string),
		},
	}
}

func (s datadogDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "site", spec.Datadog.Site, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

func (s datadogDirectSpec) GetSchema() map[string]*schema.Schema {
	datadogSchema := map[string]*schema.Schema{
		"site": {
			Type: schema.TypeString,
			Description: "`com` or `eu`, Datadog SaaS instance, which corresponds to one of Datadog's " +
				"two locations (https://www.datadoghq.com/ in the U.S. " +
				"or https://datadoghq.eu/ in the European Union).",
			Required: true,
		},
		"api_key": {
			Type:        schema.TypeString,
			Description: "[required] | Datadog API Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
		"application_key": {
			Type:        schema.TypeString,
			Description: "[required] | Datadog Application Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setHistoricalDataRetrievalSchema(datadogSchema)
	setLogCollectionSchema(datadogSchema)

	return datadogSchema
}

// Dynatrace Direct
// https://docs.nobl9.com/Sources/dynatrace#dynatrace-direct
const dynatraceDirectType = "dynatrace"

type dynatraceDirectSpec struct{}

func (s dynatraceDirectSpec) GetDescription() string {
	return "[Dynatrace Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/dynatrace#dynatrace-direct)."
}

func (s dynatraceDirectSpec) GetSchema() map[string]*schema.Schema {
	dynatraceSchema := map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Description: "Dynatrace API URL.",
			Required:    true,
		},
		"dynatrace_token": {
			Type:        schema.TypeString,
			Description: "[required] | Dynatrace Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setHistoricalDataRetrievalSchema(dynatraceSchema)
	setLogCollectionSchema(dynatraceSchema)

	return dynatraceSchema
}

func (s dynatraceDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Dynatrace: &v1alphaDirect.DynatraceConfig{
			URL:            d.Get("url").(string),
			DynatraceToken: d.Get("dynatrace_token").(string),
		},
	}
}

func (s dynatraceDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "url", spec.Dynatrace.URL, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Google Cloud Monitoring (GCM) Direct
// https://docs.nobl9.com/Sources/google-cloud-monitoring#google-cloud-monitoring-direct
const gcmDirectType = "gcm"

type gcmDirectSpec struct{}

func (s gcmDirectSpec) GetSchema() map[string]*schema.Schema {
	gcmSchema := map[string]*schema.Schema{
		"service_account_key": {
			Type:        schema.TypeString,
			Description: "[required] | Service Account Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(gcmSchema)

	return gcmSchema
}

func (s gcmDirectSpec) GetDescription() string {
	return "[Google Cloud Monitoring Direct | Nobl9 Documentation]" +
		"(https://docs.nobl9.com/Sources/google-cloud-monitoring#google-cloud-monitoring-direct)."
}

func (s gcmDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		GCM: &v1alphaDirect.GCMConfig{
			ServiceAccountKey: d.Get("service_account_key").(string),
		},
	}
}

func (s gcmDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "description", spec.Description, &diags)
	return
}

// Honeycomb Direct
// https://docs.nobl9.com/Sources/honeycomba#honeycomb-direct
// To access this integration, contact support@nobl9.com.
const honeycombDirectType = "honeycomb"

type honeycombDirectSpec struct{}

func (h honeycombDirectSpec) GetSchema() map[string]*schema.Schema {
	honeycombSchema := map[string]*schema.Schema{
		"api_key": {
			Type:        schema.TypeString,
			Description: "[required] | Honeycomb API Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(honeycombSchema)
	setHistoricalDataRetrievalSchema(honeycombSchema)
	return honeycombSchema
}

func (h honeycombDirectSpec) GetDescription() string {
	return "[Honeycomb Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/honeycomb#hc-direct)."
}

func (h honeycombDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Honeycomb: &v1alphaDirect.HoneycombConfig{
			APIKey: d.Get("api_key").(string),
		},
	}
}

func (h honeycombDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "description", spec.Description, &diags)
	return
}

// InfluxDB Direct
// https://docs.nobl9.com/Sources/influxdb#influxdb-direct
const influxdbDirectType = "influxdb"

type influxdbDirectSpec struct{}

func (s influxdbDirectSpec) GetDescription() string {
	return "[InfluxDB Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/influxdb#influxdb-direct)."
}

func (s influxdbDirectSpec) GetSchema() map[string]*schema.Schema {
	influxdbSchema := map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "API URL endpoint to the InfluxDB's instance.",
		},
		"api_token": {
			Type:        schema.TypeString,
			Description: "[required] | InfluxDB API Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
		"organization_id": {
			Type:        schema.TypeString,
			Description: "[required] | InfluxDB Organization ID.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(influxdbSchema)

	return influxdbSchema
}

func (s influxdbDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		InfluxDB: &v1alphaDirect.InfluxDBConfig{
			URL:            d.Get("url").(string),
			APIToken:       d.Get("api_token").(string),
			OrganizationID: d.Get("organization_id").(string),
		},
	}
}

func (s influxdbDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "url", spec.InfluxDB.URL, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Instana Direct
// https://docs.nobl9.com/Sources/instana#instana-direct
const instanaDirectType = "instana"

type instanaDirectSpec struct{}

func (s instanaDirectSpec) GetDescription() string {
	return "[Instana Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/instana#instana-direct)."
}

func (s instanaDirectSpec) GetSchema() map[string]*schema.Schema {
	instanaSchema := map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Instana API URL.",
		},
		"api_token": {
			Type:        schema.TypeString,
			Description: "[required] | Instana API Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}

	setLogCollectionSchema(instanaSchema)
	return instanaSchema
}

func (s instanaDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Instana: &v1alphaDirect.InstanaConfig{
			URL:      d.Get("url").(string),
			APIToken: d.Get("api_token").(string),
		},
	}
}

func (s instanaDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "url", spec.Instana.URL, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Lightstep Direct
// https://docs.nobl9.com/Sources/lightstep#lightstep-direct
const lightstepDirectType = "lightstep"

type lightstepDirectSpec struct{}

func (s lightstepDirectSpec) GetDescription() string {
	return "[Lightstep Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/lightstep#lightstep-direct)."
}

func (s lightstepDirectSpec) GetSchema() map[string]*schema.Schema {
	lightstepSchema := map[string]*schema.Schema{
		"lightstep_organization": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Organization name registered in Lightstep.",
		},
		"lightstep_project": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the Lightstep project.",
		},
		"app_token": {
			Type:        schema.TypeString,
			Description: "[required] | Lightstep App Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setHistoricalDataRetrievalSchema(lightstepSchema)
	setLogCollectionSchema(lightstepSchema)

	return lightstepSchema
}

func (s lightstepDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Lightstep: &v1alphaDirect.LightstepConfig{
			AppToken:     d.Get("app_token").(string),
			Organization: d.Get("lightstep_organization").(string),
			Project:      d.Get("lightstep_project").(string),
		},
	}
}

func (s lightstepDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "lightstep_organization", spec.Lightstep.Organization, &diags)
	set(d, "lightstep_project", spec.Lightstep.Project, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// New Relic Direct
// https://docs.nobl9.com/Sources/new-relic#new-relic-direct
const newRelicDirectType = "newrelic"

type newRelicDirectSpec struct{}

func (s newRelicDirectSpec) GetSchema() map[string]*schema.Schema {
	newRelicSchema := map[string]*schema.Schema{
		"account_id": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "ID number assigned to the New Relic user account.",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.IntAtLeast(0),
			),
		},
		"insights_query_key": {
			Type:        schema.TypeString,
			Description: "[required] | New Relic Insights Query Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setHistoricalDataRetrievalSchema(newRelicSchema)
	setLogCollectionSchema(newRelicSchema)

	return newRelicSchema
}

func (s newRelicDirectSpec) GetDescription() string {
	return "[New Relic Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/new-relic#new-relic-direct)."
}

func (s newRelicDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{NewRelic: &v1alphaDirect.NewRelicConfig{
		AccountID:        d.Get("account_id").(int),
		InsightsQueryKey: d.Get("insights_query_key").(string),
	}}
}

func (s newRelicDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "account_id", spec.NewRelic.AccountID, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Pingdom Direct
// https://docs.nobl9.com/Sources/pingdom#pingdom-direct
const pingdomDirectType = "pingdom"

type pingdomDirectSpec struct{}

func (s pingdomDirectSpec) GetDescription() string {
	return "[Pingdom Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/pingdom#pingdom-direct)."
}

func (s pingdomDirectSpec) GetSchema() map[string]*schema.Schema {
	pingdomSchema := map[string]*schema.Schema{
		"api_token": {
			Type:        schema.TypeString,
			Description: "[required] | Pingdom API token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(pingdomSchema)

	return pingdomSchema
}

func (s pingdomDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Pingdom: &v1alphaDirect.PingdomConfig{
			APIToken: d.Get("api_token").(string),
		},
	}
}

func (s pingdomDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "description", spec.Description, &diags)
	return
}

// Amazon Redshift Direct
// https://docs.nobl9.com/Sources/Amazon_Redshift/?_highlight=redshift#amazon-redshift-direct
const redshiftDirectType = "redshift"

type redshiftDirectSpec struct{}

func (s redshiftDirectSpec) GetSchema() map[string]*schema.Schema {
	redshiftSchema := map[string]*schema.Schema{
		"secret_arn": {
			Type:        schema.TypeString,
			Description: "AWS Secret ARN.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
		"role_arn": {
			Type:        schema.TypeString,
			Description: "[required] | ARN of the AWS IAM Role to assume.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(redshiftSchema)

	return redshiftSchema
}

func (s redshiftDirectSpec) GetDescription() string {
	return "[Amazon Redshift Direct | Nobl9 Documentation]" +
		"(https://docs.nobl9.com/Sources/Amazon_Redshift/?_highlight=redshift#amazon-redshift-direct)."
}

func (s redshiftDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{
		Redshift: &v1alphaDirect.RedshiftConfig{
			RoleARN:   d.Get("role_arn").(string),
			SecretARN: d.Get("secret_arn").(string),
		},
	}
}

func (s redshiftDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "secret_arn", spec.Redshift.SecretARN, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Splunk Direct
// https://docs.nobl9.com/Sources/splunk#splunk-direct
const splunkDirectType = "splunk"

type splunkDirectSpec struct{}

func (s splunkDirectSpec) GetDescription() string {
	return "[Splunk Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/splunk#splunk-direct)."
}

func (s splunkDirectSpec) GetSchema() map[string]*schema.Schema {
	splunkSchema := map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Base API URL to the Splunk Search app.",
		},
		"access_token": {
			Type:        schema.TypeString,
			Description: "[required] | Splunk API Access Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setHistoricalDataRetrievalSchema(splunkSchema)
	setLogCollectionSchema(splunkSchema)

	return splunkSchema
}

func (s splunkDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{Splunk: &v1alphaDirect.SplunkConfig{
		URL:         d.Get("url").(string),
		AccessToken: d.Get("access_token").(string),
	}}
}

func (s splunkDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "url", spec.Splunk.URL, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Splunk Observability Direct
// https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-direct
const splunkObservabilityDirectType = "splunk_observability"

type splunkObservabilityDirectSpec struct{}

func (s splunkObservabilityDirectSpec) GetDescription() string {
	return "[Splunk Observability Direct | Nobl9 Documentation]" +
		"(https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-direct)."
}

func (s splunkObservabilityDirectSpec) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"realm": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "SplunkObservability Realm.",
		},
		"access_token": {
			Type:        schema.TypeString,
			Description: "[required] | Splunk API Access Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
}

func (s splunkObservabilityDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{SplunkObservability: &v1alphaDirect.SplunkObservabilityConfig{
		Realm:       d.Get("realm").(string),
		AccessToken: d.Get("access_token").(string),
	}}
}

func (s splunkObservabilityDirectSpec) UnmarshalSpec(
	d *schema.ResourceData,
	spec v1alphaDirect.Spec,
) (diags diag.Diagnostics) {
	set(d, "realm", spec.SplunkObservability.Realm, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// Sumo Logic Direct
// https://docs.nobl9.com/Sources/sumo-logic#sumo-logic-direct
const sumologicDirectType = "sumologic"

type sumologicDirectSpec struct{}

func (s sumologicDirectSpec) GetDescription() string {
	return "[Sumo Logic Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/sumo-logic#sumo-logic-direct)."
}

func (s sumologicDirectSpec) GetSchema() map[string]*schema.Schema {
	sumologicSchema := map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Sumo Logic API URL.",
		},
		"access_id": {
			Type:        schema.TypeString,
			Description: "[required] | Sumo Logic API Access ID.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
		"access_key": {
			Type:        schema.TypeString,
			Description: "[required] | Sumo Logic API Access Key.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}

	setLogCollectionSchema(sumologicSchema)

	return sumologicSchema
}

func (s sumologicDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{SumoLogic: &v1alphaDirect.SumoLogicConfig{
		URL:       d.Get("url").(string),
		AccessID:  d.Get("access_id").(string),
		AccessKey: d.Get("access_key").(string),
	}}
}

func (s sumologicDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) (diags diag.Diagnostics) {
	set(d, "url", spec.SumoLogic.URL, &diags)
	set(d, "description", spec.Description, &diags)
	return
}

// ThousandEyes Direct
// https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-direct
const thousandeyesDirectType = "thousandeyes"

type thousandeyesDirectSpec struct{}

func (s thousandeyesDirectSpec) GetSchema() map[string]*schema.Schema {
	thousandeyesSchema := map[string]*schema.Schema{
		"oauth_bearer_token": {
			Type:        schema.TypeString,
			Description: "[required] | ThousandEyes OAuth Bearer Token.",
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringIsNotEmpty,
			),
		},
	}
	setLogCollectionSchema(thousandeyesSchema)

	return thousandeyesSchema
}

func (s thousandeyesDirectSpec) GetDescription() string {
	return "[ThousandEyes Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-direct)."
}

func (s thousandeyesDirectSpec) MarshalSpec(d *schema.ResourceData) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{ThousandEyes: &v1alphaDirect.ThousandEyesConfig{
		OauthBearerToken: d.Get("oauth_bearer_token").(string),
	}}
}

func (s thousandeyesDirectSpec) UnmarshalSpec(
	d *schema.ResourceData,
	spec v1alphaDirect.Spec,
) (diags diag.Diagnostics) {
	set(d, "description", spec.Description, &diags)
	return
}
