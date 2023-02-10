package nobl9

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	n9api "github.com/nobl9/nobl9-go"
)

const directTypeKey = "direct_type"

func resourceDirect() *schema.Resource {
	return &schema.Resource{
		Schema:        directSchema(),
		CreateContext: resourceDirectApply,
		UpdateContext: resourceDirectApply,
		DeleteContext: resourceDirectDelete,
		ReadContext:   resourceDirectRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Direct configuration | Nobl9 Documentation](https://docs.nobl9.com/nobl9_direct)",
	}
}

func directSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name":         schemaName(),
		"display_name": schemaDisplayName(),
		"project":      schemaProject(),
		"description":  schemaDescription(),
		"source_of": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			MaxItems:    2,
			Description: "Source of Metrics and/or Services",
			Elem: &schema.Schema{
				Type:        schema.TypeString,
				Description: "Source of Metrics or Services",
			},
		},
		directTypeKey: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The type of the Direct. Check [Supported Direct types | Nobl9 Documentation](https://docs.nobl9.com/Sources/)",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Status of the created direct.",
		},
	}

	directSchemaDefinitions := []map[string]*schema.Schema{
		schemaDirectAppDynamics(),
		schemaDirectBigQuery(),
		schemaDirectCloudWatch(),
		schemaDirectDatadog(),
		schemaDirectDynatrace(),
		schemaDirectGCM(),
		schemaDirectInfluxDB(),
		schemaDirectInstana(),
		schemaDirectLightstep(),
		schemaDirectNewRelic(),
		schemaDirectPingdom(),
		schemaDirectRedshift(),
		schemaDirectSplunk(),
		schemaDirectSplunkObservability(),
		schemaDirectSumoLogic(),
		schemaDirectThousandEyes(),
	}

	for _, directSchemaDef := range directSchemaDefinitions {
		for directKey, schema := range directSchemaDef {
			s[directKey] = schema
		}
	}

	return s
}

func resourceDirectApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}
	direct, diags := marshalDirect(d)
	if diags.HasError() {
		return diags
	}

	var p n9api.Payload
	p.AddObject(direct)

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		err := client.ApplyObjects(p.GetObjects())
		if err != nil {
			if errors.Is(err, n9api.ErrConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.Errorf("could not add direct: %s", err.Error())
	}

	d.SetId(direct.Metadata.Name)

	readDirectDiags := resourceDirectRead(ctx, d, meta)

	return append(diags, readDirectDiags...)
}

func resourceDirectRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	objects, err := client.GetDirects("", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalDirect(d, objects)
}

func resourceDirectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		err := client.DeleteObjectsByName(n9api.ObjectDirect, d.Id())
		if err != nil {
			if errors.Is(err, n9api.ErrConcurrencyIssue) {
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

func marshalDirect(d *schema.ResourceData) (*n9api.Direct, diag.Diagnostics) {
	var diags diag.Diagnostics
	metadataHolder, diags := marshalMetadata(d)
	if diags.HasError() {
		return nil, diags
	}

	sourceOf := d.Get("source_of").([]interface{})
	sourceOfStr := make([]string, len(sourceOf))
	for i, s := range sourceOf {
		sourceOfStr[i] = s.(string)
	}

	return &n9api.Direct{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindDirect,
			MetadataHolder: metadataHolder,
		},
		Spec: n9api.DirectSpec{
			Description:         d.Get("description").(string),
			SourceOf:            sourceOfStr,
			AppDynamics:         marshalDirectAppDynamics(d, diags),
			BigQuery:            marshalDirectBigQuery(d, diags),
			CloudWatch:          marshalDirectCloudWatch(d, diags),
			Datadog:             marshalDirectDatadog(d, diags),
			Dynatrace:           marshalDirectDynatrace(d, diags),
			GCM:                 marshalDirectGCM(d, diags),
			InfluxDB:            marshalDirectInfluxDB(d, diags),
			Instana:             marshalDirectInstana(d, diags),
			Lightstep:           marshalDirectLightstep(d, diags),
			NewRelic:            marshalDirectNewRelic(d, diags),
			Pingdom:             marshalDirectPingdom(d, diags),
			Redshift:            marshalDirectRedshift(d, diags),
			Splunk:              marshalDirectSplunk(d, diags),
			SplunkObservability: marshalDirectSplunkObservability(d, diags),
			SumoLogic:           marshalDirectSumoLogic(d, diags),
			ThousandEyes:        marshalDirectThousandEyes(d, diags),
		},
	}, diags
}

func unmarshalDirect(d *schema.ResourceData, directs []n9api.Direct) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(directs) != 1 {
		d.SetId("")
		return nil
	}
	direct := directs[0]

	if ds := unmarshalMetadata(direct.MetadataHolder, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	set(d, "status", direct.Status.DirectType, &diags)

	return diags
}

/**
 * AppDynamics Direct
 * https://docs.nobl9.com/Sources/appdynamics#appdynamics-direct
 */
const appDynamicsDirectType = "appdynamics"
const appDynamicsDirectConfigKey = "appdynamics_config"

func schemaDirectAppDynamics() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		appDynamicsDirectConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/appdynamics#appdynamics-direct)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Base URL to the AppDynamics Controller.",
					},
					"account_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "AppDynamics account name.",
					},
					"client_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "AppDynamics client ID.",
					},
					"client_secret": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "AppDynamics client secret.",
					},
					"client_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "AppDynamics client name.",
					},
				},
			},
		},
	}
}

func marshalDirectAppDynamics(d *schema.ResourceData, diags diag.Diagnostics) *n9api.AppDynamicsDirectConfig {
	data := getDirectResourceData(d, appDynamicsDirectType, appDynamicsDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.AppDynamicsDirectConfig{
		URL:          data["url"].(string),
		AccountName:  data["account_name"].(string),
		ClientID:     data["client_id"].(string),
		ClientSecret: data["client_secret"].(string),
		ClientName:   data["client_name"].(string),
	}
}

/**
 * BigQuery Direct
 * https://docs.nobl9.com/Sources/bigquery#bigquery-direct
 */
const bigqueryDirectType = "bigquery"
const bigqueryDirectConfigKey = "bigquery_config"

func schemaDirectBigQuery() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		bigqueryDirectConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/bigquery#bigquery-direct)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Description: "Direct configuration is not required.",
			},
		},
	}
}

func marshalDirectBigQuery(d *schema.ResourceData, diags diag.Diagnostics) *n9api.BigQueryDirectConfig {
	data := getDirectResourceData(d, bigqueryDirectType, bigqueryDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.BigQueryDirectConfig{
		ServiceAccountKey: "",
	}
}

/**
 * Amazon CloudWatch Direct
 * https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cloudwatch-direct
 */
const cloudWatchDirectType = "cloudwatch"
const cloudWatchDirectConfigKey = "cloudwatch_config"

func schemaDirectCloudWatch() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		cloudWatchDirectConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cloudwatch-direct)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"access_key_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "",
						Sensitive:   true,
					},
					"secret_access_key": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "",
						Sensitive:   true,
					},
				},
			},
		},
	}
}

func marshalDirectCloudWatch(d *schema.ResourceData, diags diag.Diagnostics) *n9api.CloudWatchDirectConfig {
	data := getDirectResourceData(d, cloudWatchDirectType, cloudWatchDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.CloudWatchDirectConfig{
		AccessKeyID:     "",
		SecretAccessKey: "",
	}
}

/**
 * Datadog Direct
 * https://docs.nobl9.com/Sources/datadog#datadog-direct
 */
const datadogDirectType = "datadog"
const datadogDirectConfigKey = "datadog_config"

func schemaDirectDatadog() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		datadogDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/datadog#datadog-direct). " +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			Elem: &schema.Resource{
				ReadContext: resourceDirectRead,
				Schema: map[string]*schema.Schema{
					"site": {
						Type:     schema.TypeString,
						Required: true,
						Description: "`com` or `eu`, Datadog SaaS instance, which corresponds to one of Datadog's " +
							"two locations (https://www.datadoghq.com/ in the U.S. " +
							"or https://datadoghq.eu/ in the European Union).",
					},
					"api_key": {
						Type:        schema.TypeString,
						Description: "Datadog API key.",
						Required:    true,
						Sensitive:   true,
					},
					"application_key": {
						Type:        schema.TypeString,
						Description: "Datadog Application key.",
						Required:    true,
						Sensitive:   true,
					},
				},
			},
		},
	}
}

func marshalDirectDatadog(d *schema.ResourceData, diags diag.Diagnostics) *n9api.DatadogDirectConfig {
	data := getDirectResourceData(d, datadogDirectType, datadogDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.DatadogDirectConfig{
		Site:           data["site"].(string),
		APIKey:         data["api_key"].(string),
		ApplicationKey: data["application_key"].(string),
	}
}

/**
 * Dynatrace Direct
 * https://docs.nobl9.com/Sources/dynatrace#dynatrace-direct
 */
const dynatraceDirectType = "dynatrace"
const dynatraceDirectConfigKey = "dynatrace_config"

func schemaDirectDynatrace() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		dynatraceDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/dynatrace#dynatrace-direct). " +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Dynatrace API URL.",
					},
				},
			},
		},
	}
}

func marshalDirectDynatrace(d *schema.ResourceData, diags diag.Diagnostics) *n9api.DynatraceDirectConfig {
	data := getDirectResourceData(d, dynatraceDirectType, dynatraceDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.DynatraceDirectConfig{
		URL:            data["url"].(string),
		DynatraceToken: "",
	}
}

/**
 * Google Cloud Monitoring (GCM) Direct
 * https://docs.nobl9.com/Sources/google-cloud-monitoring#google-cloud-monitoring-direct
 */
const gcmDirectType = "gcm"
const gcmDirectConfigKey = "gcm_config"

func schemaDirectGCM() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		gcmDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/google-cloud-monitoring#google-cloud-monitoring-direct). " +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Description: "Direct configuration is not required.",
			},
		},
	}
}

func marshalDirectGCM(d *schema.ResourceData, diags diag.Diagnostics) *n9api.GCMDirectConfig {
	data := getDirectResourceData(d, gcmDirectType, gcmDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.GCMDirectConfig{}
}

/**
 * InfluxDB Direct
 * https://docs.nobl9.com/Sources/influxdb#influxdb-direct
 */
const influxdbDirectType = "influxdb"
const influxdbDirectConfigKey = "influxdb_config"

func schemaDirectInfluxDB() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		influxdbDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/influxdb#influxdb-direct). " +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "API URL endpoint to the InfluxDB's instance.",
					},
				},
			},
		},
	}
}

func marshalDirectInfluxDB(d *schema.ResourceData, diags diag.Diagnostics) *n9api.InfluxDBDirectConfig {
	data := getDirectResourceData(d, influxdbDirectType, influxdbDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.InfluxDBDirectConfig{
		URL:            data["url"].(string),
		APIToken:       data[""],
		OrganizationID: "",
	}
}

/**
 * Instana Direct
 * https://docs.nobl9.com/Sources/instana#instana-direct
 */
const instanaDirectType = "instana"
const instanaDirectConfigKey = "instana_config"

func schemaDirectInstana() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		instanaDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/instana#instana-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "API URL endpoint to the InfluxDB's instance.",
					},
					"api_token": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "API URL endpoint to the InfluxDB's instance.",
					},
				},
			},
		},
	}
}

func marshalDirectInstana(d *schema.ResourceData, diags diag.Diagnostics) *n9api.InstanaDirectConfig {
	data := getDirectResourceData(d, instanaDirectType, instanaDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.InstanaDirectConfig{
		URL:      data["url"].(string),
		APIToken: data["api_token"].(string),
	}
}

/**
 * Lightstep Direct
 * https://docs.nobl9.com/Sources/lightstep#lightstep-direct
 */
const lightstepDirectType = "lightstep"
const lightstepDirectConfigKey = "lightstep_config"

func schemaDirectLightstep() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		lightstepDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/lightstep#lightstep-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"organization": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Organization name registered in Lightstep.",
					},
					"project": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the Lightstep project.",
					},
				},
			},
		},
	}
}

func marshalDirectLightstep(d *schema.ResourceData, diags diag.Diagnostics) *n9api.LightstepDirectConfig {
	data := getDirectResourceData(d, lightstepDirectType, lightstepDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.LightstepDirectConfig{
		Organization: data["organization"].(string),
		Project:      data["project"].(string),
	}
}

/**
 * New Relic Direct
 * https://docs.nobl9.com/Sources/new-relic#new-relic-direct
 */
const newRelicDirectType = "newrelic"
const newRelicDirectConfigKey = "newrelic_config"

func schemaDirectNewRelic() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		newRelicDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/new-relic#new-relic-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"account_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "ID number assigned to the New Relic user account.",
					},
				},
			},
		},
	}
}

func marshalDirectNewRelic(d *schema.ResourceData, diags diag.Diagnostics) *n9api.NewRelicDirectConfig {
	data := getDirectResourceData(d, newRelicDirectType, newRelicDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.NewRelicDirectConfig{
		AccountID: data["account_id"].(string),
	}
}

/**
 * Pingdom Direct
 * https://docs.nobl9.com/Sources/pingdom#pingdom-direct
 */
const pingdomDirectType = "pingdom"
const pingdomDirectConfigKey = "pingdom_config"

func schemaDirectPingdom() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		pingdomDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/pingdom#pingdom-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Description: "Direct configuration is not required.",
			},
		}}
}

func marshalDirectPingdom(d *schema.ResourceData, diags diag.Diagnostics) *n9api.PingdomDirectConfig {
	data := getDirectResourceData(d, pingdomDirectType, pingdomDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.PingdomDirectConfig{}
}

/**
 * Amazon Redshift Direct
 * https://docs.nobl9.com/Sources/Amazon_Redshift/?_highlight=redshift#amazon-redshift-direct
 */
const redshiftDirectType = "redshift"
const redshiftDirectConfigKey = "redshift_config"

func schemaDirectRedshift() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		redshiftDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/Amazon_Redshift/?_highlight=redshift#amazon-redshift-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Description: "Direct configuration is not required.",
			},
		},
	}
}

func marshalDirectRedshift(d *schema.ResourceData, diags diag.Diagnostics) *n9api.RedshiftDirectConfig {
	data := getDirectResourceData(d, redshiftDirectType, redshiftDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.RedshiftDirectConfig{}
}

/**
 * Splunk Direct
 * https://docs.nobl9.com/Sources/splunk#splunk-direct
 */
const splunkDirectType = "splunk"
const splunkDirectConfigKey = "splunk_config"

func schemaDirectSplunk() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		splunkDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#splunk-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Base API URL to the Splunk Search app.",
					},
				},
			},
		},
	}
}

func marshalDirectSplunk(d *schema.ResourceData, diags diag.Diagnostics) *n9api.SplunkDirectConfig {
	data := getDirectResourceData(d, splunkDirectType, splunkDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.SplunkDirectConfig{
		URL: data["url"].(string),
	}
}

/**
 * Splunk Observability Direct
 * https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-direct
 */
const splunkObservabilityDirectType = "splunk_observability"
const splunkObservabilityDirectConfigKey = "splunk_observability_config"

func schemaDirectSplunkObservability() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		splunkObservabilityDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"realm": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "SplunkObservability Realm.",
					},
				},
			},
		},
	}
}

func marshalDirectSplunkObservability(
	d *schema.ResourceData,
	diags diag.Diagnostics,
) *n9api.SplunkObservabilityDirectConfig {
	data := getDirectResourceData(d, splunkObservabilityDirectType, splunkObservabilityDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.SplunkObservabilityDirectConfig{
		Realm: data["realm"].(string),
	}
}

/**
 * Sumo Logic Direct
 * https://docs.nobl9.com/Sources/sumo-logic#sumo-logic-direct
 */
const sumologicDirectType = "sumologic"
const sumologicDirectConfigKey = "sumologic_config"

func schemaDirectSumoLogic() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		sumologicDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/sumo-logic#sumo-logic-direct)." +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Base API URL to the Splunk Search app.",
					},
				},
			},
		}}
}

func marshalDirectSumoLogic(d *schema.ResourceData, diags diag.Diagnostics) *n9api.SumoLogicDirectConfig {
	data := getDirectResourceData(d, sumologicDirectType, sumologicDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.SumoLogicDirectConfig{
		URL: data["url"].(string),
	}
}

/**
 * ThousandEyes Direct
 * https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-direct
 */
const thousandeyesDirectType = "thousandeyes"
const thousandeyesDirectConfigKey = "thousandeyes_config"

func schemaDirectThousandEyes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		thousandeyesDirectConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-direct). " +
				"The configuration is not refreshed, out-of-band changes are not tracked.",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Description: "Direct configuration is not required.",
			},
		}}
}

func marshalDirectThousandEyes(d *schema.ResourceData, diags diag.Diagnostics) *n9api.ThousandEyesDirectConfig {
	data := getDirectResourceData(d, thousandeyesDirectType, thousandeyesDirectConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.ThousandEyesDirectConfig{}
}

func getDirectResourceData(
	d *schema.ResourceData,
	directType,
	directConfigKey string,
	diags diag.Diagnostics) map[string]interface{} {
	if !isDirectType(d, directType) {
		return nil
	}
	p := d.Get(directConfigKey).(*schema.Set).List()
	if len(p) == 0 {
		appendError(diags, fmt.Errorf("no resource data '%s' for direct type '%s'", directConfigKey, directType))
		return nil
	}
	resourceData := p[0].(map[string]interface{})

	return resourceData
}

func isDirectType(d *schema.ResourceData, directType string) bool {
	directTypeResource := d.Get(directTypeKey).(string)
	return directTypeResource == directType
}
