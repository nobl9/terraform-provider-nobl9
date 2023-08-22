package nobl9

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk"
)

const agentTypeKey = "agent_type"

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		Schema:        agentSchema(),
		CreateContext: resourceAgentApply,
		UpdateContext: resourceAgentApply,
		DeleteContext: resourceAgentDelete,
		ReadContext:   resourceAgentRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Agent configuration | Nobl9 Documentation](https://docs.nobl9.com/nobl9_agent)",
	}
}

func agentSchema() map[string]*schema.Schema {
	agentSchema := map[string]*schema.Schema{
		"name":         schemaName(),
		"display_name": schemaDisplayName(),
		"project":      schemaProject(),
		"description":  schemaDescription(),
		"source_of": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			MaxItems:    2,
			Description: "Source of Metrics and/or Services.",
			Elem: &schema.Schema{
				Type:        schema.TypeString,
				Description: "Source of Metrics or Services",
			},
		},
		agentTypeKey: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The type of the Agent. Check [Supported Agent types | Nobl9 Documentation](https://docs.nobl9.com/Sources/)",
		},
		"client_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "client_id of created agent.",
		},
		"client_secret": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "client_secret of created agent.",
		},
		"release_channel": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Release channel of the created agent [stable/beta]",
		},
		queryDelayConfigKey: schemaQueryDelay(),
		"status": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Status of the created agent.",
		},
	}

	agentSchemaDefinitions := []map[string]*schema.Schema{
		schemaAgentAmazonPrometheus(),
		schemaAgentAppDynamics(),
		schemaAgentBigQuery(),
		schemaAgentCloudWatch(),
		schemaAgentDatadog(),
		schemaAgentDynatrace(),
		schemaAgentElasticsearch(),
		schemaAgentGCM(),
		schemaAgentGrafanaLoki(),
		schemaAgentGraphite(),
		schemaAgentInfluxDB(),
		schemaAgentInstana(),
		schemaAgentLightstep(),
		schemaAgentNewRelic(),
		schemaAgentOpenTSDB(),
		schemaAgentPingdom(),
		schemaAgentPrometheus(),
		schemaAgentRedshift(),
		schemaAgentSplunk(),
		schemaAgentSplunkObservability(),
		schemaAgentSumoLogic(),
		schemaAgentThousandEyes(),
	}

	for _, agentSchemaDef := range agentSchemaDefinitions {
		for agentKey, schema := range agentSchemaDef {
			agentSchema[agentKey] = schema
		}
	}

	return agentSchema
}

func resourceAgentApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	agent, diags := marshalAgent(d)
	if diags.HasError() {
		return diags
	}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		err := client.ApplyObjects(ctx, []manifest.Object{agent}, false)
		if err != nil {
			if errors.Is(err, sdk.ErrConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		project := d.Get("project").(string)
		agentsData, err := client.GetAgentCredentials(ctx, project, agent.Metadata.Name)
		err = d.Set("client_id", agentsData.ClientID)
		diags = appendError(diags, err)
		err = d.Set("client_secret", agentsData.ClientSecret)
		diags = appendError(diags, err)
		return nil
	}); err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}

	d.SetId(agent.Metadata.Name)

	readAgentDiags := resourceAgentRead(ctx, d, meta)

	return append(diags, readAgentDiags...)
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	objects, err := client.GetObjects(ctx, project, manifest.KindAgent, nil, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	data, err := json.Marshal(objects)
	if err != nil {
		return diag.FromErr(err)
	}
	var converted []v1alpha.Agent
	if err = json.Unmarshal(data, &converted); err != nil {
		return diag.FromErr(err)
	}
	return unmarshalAgent(d, converted)
}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		err := client.DeleteObjectsByName(ctx, project, manifest.KindAgent, false, d.Id())
		if err != nil {
			if errors.Is(err, sdk.ErrConcurrencyIssue) {
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

func marshalAgent(d *schema.ResourceData) (*v1alpha.Agent, diag.Diagnostics) {
	sourceOf := d.Get("source_of").([]interface{})
	sourceOfStr := make([]string, len(sourceOf))
	for i, s := range sourceOf {
		sourceOfStr[i] = s.(string)
	}

	var displayName string
	if dn := d.Get("display_name"); dn != nil {
		displayName = dn.(string)
	}

	labelsMarshalled, diags := getMarshalledLabels(d)
	if diags.HasError() {
		return nil, diags
	}

	return &v1alpha.Agent{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindAgent,
		Metadata: v1alpha.AgentMetadata{
			Name:        d.Get("name").(string),
			DisplayName: displayName,
			Project:     d.Get("project").(string),
			Labels:      labelsMarshalled,
		},
		Spec: v1alpha.AgentSpec{
			Description:         d.Get("description").(string),
			SourceOf:            sourceOfStr,
			AmazonPrometheus:    marshalAgentAmazonPrometheus(d, diags),
			AppDynamics:         marshalAgentAppDynamics(d, diags),
			BigQuery:            marshalAgentBigQuery(d),
			CloudWatch:          marshalAgentCloudWatch(d),
			Datadog:             marshalAgentDatadog(d, diags),
			Dynatrace:           marshalAgentDynatrace(d, diags),
			Elasticsearch:       marshalAgentElasticsearch(d, diags),
			GCM:                 marshalAgentGCM(d),
			GrafanaLoki:         marshalAgentGrafanaLoki(d, diags),
			Graphite:            marshalAgentGraphite(d, diags),
			InfluxDB:            marshalAgentInfluxDB(d, diags),
			Instana:             marshalAgentInstana(d, diags),
			Lightstep:           marshalAgentLightstep(d, diags),
			NewRelic:            marshalAgentNewRelic(d, diags),
			OpenTSDB:            marshalAgentOpenTSDB(d, diags),
			Prometheus:          marshalAgentPrometheus(d, diags),
			Pingdom:             marshalAgentPingdom(d),
			Redshift:            marshalAgentRedshift(d),
			Splunk:              marshalAgentSplunk(d, diags),
			SplunkObservability: marshalAgentSplunkObservability(d, diags),
			SumoLogic:           marshalAgentSumoLogic(d, diags),
			ThousandEyes:        marshalAgentThousandEyes(d),
			QueryDelay:          marshalQueryDelay(d),
		},
	}, diags
}

func unmarshalAgent(d *schema.ResourceData, agents []v1alpha.Agent) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(agents) != 1 {
		d.SetId("")
		return nil
	}
	agent := agents[0]

	status := map[string]interface{}{
		"agent_type":      agent.Status.AgentType,
		"agent_version":   agent.Status.AgentVersion,
		"last_connection": agent.Status.LastConnection,
	}
	err := d.Set("status", status)

	diags = appendError(diags, err)

	set(d, "name", agent.Metadata.Name, &diags)
	set(d, "display_name", agent.Metadata.DisplayName, &diags)
	set(d, "project", agent.Metadata.Project, &diags)
	if agent.Metadata.Labels != nil {
		set(d, "label", agent.Metadata.Labels, &diags)
	}

	diags = append(diags, unmarshalQueryDelay(d, agent.Spec.QueryDelay)...)
	spec := v1alpha.AgentSpec{}
	supportedAgents := []struct {
		hclName  string
		jsonName string
	}{
		{amazonPrometheusAgentConfigKey, agentSpecJSONName(spec.AmazonPrometheus, diags)},
		{appDynamicsAgentConfigKey, agentSpecJSONName(spec.AppDynamics, diags)},
		{bigqueryAgentConfigKey, agentSpecJSONName(spec.BigQuery, diags)},
		{cloudWatchAgentConfigKey, agentSpecJSONName(spec.CloudWatch, diags)},
		{datadogAgentConfigKey, agentSpecJSONName(spec.Datadog, diags)},
		{dynatraceAgentConfigKey, agentSpecJSONName(spec.Dynatrace, diags)},
		{elasticsearchAgentConfigKey, agentSpecJSONName(spec.Elasticsearch, diags)},
		{gcmAgentConfigKey, agentSpecJSONName(spec.GCM, diags)},
		{grafanalokiAgentConfigKey, agentSpecJSONName(spec.GrafanaLoki, diags)},
		{graphiteAgentConfigKey, agentSpecJSONName(spec.Graphite, diags)},
		{influxdbAgentConfigKey, agentSpecJSONName(spec.InfluxDB, diags)},
		{instanaAgentConfigKey, agentSpecJSONName(spec.Instana, diags)},
		{lightstepAgentConfigKey, agentSpecJSONName(spec.Lightstep, diags)},
		{newRelicAgentConfigKey, agentSpecJSONName(spec.NewRelic, diags)},
		{opentsdbAgentConfigKey, agentSpecJSONName(spec.OpenTSDB, diags)},
		{pingdomAgentConfigKey, agentSpecJSONName(spec.Pingdom, diags)},
		{prometheusAgentConfigKey, agentSpecJSONName(spec.Prometheus, diags)},
		{redshiftAgentConfigKey, agentSpecJSONName(spec.Redshift, diags)},
		{splunkAgentConfigKey, agentSpecJSONName(spec.Splunk, diags)},
		{splunkObservabilityAgentConfigKey, agentSpecJSONName(spec.SplunkObservability, diags)},
		{sumologicAgentConfigKey, agentSpecJSONName(spec.SumoLogic, diags)},
		{thousandeyesAgentConfigKey, agentSpecJSONName(spec.ThousandEyes, diags)},
	}

	for _, name := range supportedAgents {
		ok, ds := unmarshalAgentConfig(d, agent, name.hclName, name.jsonName)
		if ds.HasError() {
			diags = append(diags, ds...)
		}
		if ok {
			break
		}
	}

	return diags
}

func unmarshalAgentConfig(
	d *schema.ResourceData,
	agent v1alpha.Agent,
	hclName,
	jsonName string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// err := d.Set("agent_type", spec[""]) TODO
	err := d.Set("source_of", agent.Spec.SourceOf)
	diags = appendError(diags, err)

	spec, err := json.Marshal(&agent.Spec)
	diags = appendError(diags, err)

	var m map[string]interface{}
	err = json.Unmarshal(spec, &m)
	diags = appendError(diags, err)

	switch jsonName {
	case agentSpecJSONName(v1alpha.AgentSpec{}.NewRelic, diags):
		unmarshalDiags := unmarshalNewRelicAgentSpec(d, agent)
		diags = append(diags, unmarshalDiags...)
	default:
		err = d.Set(hclName, schema.NewSet(oneElementSet, []interface{}{m[jsonName]}))
		diags = appendError(diags, err)
	}

	return true, diags
}

func agentSpecJSONName(agentSpecField any, diags diag.Diagnostics) string {
	agentSpec := v1alpha.AgentSpec{}
	getAgentSpecFieldName := func() string {
		var name string
		val := reflect.Indirect(reflect.ValueOf(agentSpec))
		for i := 0; i < val.NumField(); i++ {
			typeField := val.Type().Field(i)

			if typeField.Type == reflect.TypeOf(agentSpecField) {
				name = typeField.Name
			}
		}
		return name
	}

	agentSpecType := reflect.TypeOf(agentSpec)
	name := getAgentSpecFieldName()

	field, _ := agentSpecType.FieldByName(name)
	if tag, tagOk := field.Tag.Lookup("json"); tagOk {
		jsonName := strings.Split(tag, ",")
		if len(jsonName) > 1 {
			return jsonName[0]
		}
	}

	appendError(diags, fmt.Errorf("not supported agent type: %v", reflect.TypeOf(agentSpecField).String()))

	return ""
}

/**
 * Amazon Prometheus Agent
 * https://docs.nobl9.com/Sources/Amazon_Prometheus/#ams-prometheus-agent
 */
const amazonPrometheusAgentType = "amazon_prometheus"
const amazonPrometheusAgentConfigKey = "amazon_prometheus_config"

func schemaAgentAmazonPrometheus() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		amazonPrometheusAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/Amazon_Prometheus/#ams-prometheus-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Base URL to Amazon Prometheus server.",
					},
					"region": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "AWS region e.g., eu-central-1",
					},
				},
			},
		},
	}
}

func marshalAgentAmazonPrometheus(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.AmazonPrometheusAgentConfig {
	data := getAgentResourceData(d, amazonPrometheusAgentType, amazonPrometheusAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.AmazonPrometheusAgentConfig{
		URL:    data["url"].(string),
		Region: data["region"].(string),
	}
}

/**
 * AppDynamics Agent
 * https://docs.nobl9.com/Sources/appdynamics#appdynamics-agent
 */
const appDynamicsAgentType = "appdynamics"
const appDynamicsAgentConfigKey = "appdynamics_config"

func schemaAgentAppDynamics() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		appDynamicsAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/appdynamics#appdynamics-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Base URL to the AppDynamics Controller.",
					},
				},
			},
		},
	}
}

func marshalAgentAppDynamics(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.AppDynamicsAgentConfig {
	data := getAgentResourceData(d, appDynamicsAgentType, appDynamicsAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	url := data["url"].(string)
	return &v1alpha.AppDynamicsAgentConfig{
		URL: url,
	}
}

/**
 * BigQuery Agent
 * https://docs.nobl9.com/Sources/bigquery#bigquery-agent
 */
const bigqueryAgentType = "bigquery"
const bigqueryAgentConfigKey = "bigquery_config"

func schemaAgentBigQuery() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		bigqueryAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/bigquery#bigquery-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Description: "Agent configuration is not required.",
			},
		},
	}
}

func marshalAgentBigQuery(d *schema.ResourceData) *v1alpha.BigQueryAgentConfig {
	if !isAgentType(d, bigqueryAgentType) {
		return nil
	}

	return &v1alpha.BigQueryAgentConfig{}
}

/**
 * Amazon CloudWatch Agent
 * https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cloudwatch-agent
 */
const cloudWatchAgentType = "cloudwatch"
const cloudWatchAgentConfigKey = "cloudwatch_config"

func schemaAgentCloudWatch() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		cloudWatchAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cloudwatch-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Description: "Agent configuration is not required.",
			},
		},
	}
}

func marshalAgentCloudWatch(d *schema.ResourceData) *v1alpha.CloudWatchAgentConfig {
	if !isAgentType(d, cloudWatchAgentType) {
		return nil
	}

	return &v1alpha.CloudWatchAgentConfig{}
}

/**
 * Datadog Agent
 * https://docs.nobl9.com/Sources/datadog#datadog-agent
 */
const datadogAgentType = "datadog"
const datadogAgentConfigKey = "datadog_config"

func schemaAgentDatadog() map[string]*schema.Schema {
	return map[string]*schema.Schema{datadogAgentConfigKey: {
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "[Configuration documentation](https://docs.nobl9.com/Sources/datadog#datadog-agent)",
		MinItems:    1,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"site": {
					Type:     schema.TypeString,
					Required: true,
					Description: "`com` or `eu`, Datadog SaaS instance, which corresponds to one of Datadog's " +
						"two locations (https://www.datadoghq.com/ in the U.S. " +
						"or https://datadoghq.eu/ in the European Union)",
				},
			},
		},
	},
	}
}

func marshalAgentDatadog(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.DatadogAgentConfig {
	data := getAgentResourceData(d, datadogAgentType, datadogAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.DatadogAgentConfig{
		Site: data["site"].(string),
	}
}

/**
 * Dynatrace Agent
 * https://docs.nobl9.com/Sources/dynatrace#dynatrace-agent
 */
const dynatraceAgentType = "dynatrace"
const dynatraceAgentConfigKey = "dynatrace_config"

func schemaAgentDynatrace() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		dynatraceAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/dynatrace#dynatrace-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentDynatrace(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.DynatraceAgentConfig {
	data := getAgentResourceData(d, dynatraceAgentType, dynatraceAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.DynatraceAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Elasticsearch Agent
 * https://docs.nobl9.com/Sources/elasticsearch#elasticsearch-agent
 */
const elasticsearchAgentType = "elasticsearch"
const elasticsearchAgentConfigKey = "elasticsearch_config"

func schemaAgentElasticsearch() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		elasticsearchAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/elasticsearch#elasticsearch-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "API URL endpoint to the Elasticsearch's instance.",
					},
				},
			},
		},
	}
}

func marshalAgentElasticsearch(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.ElasticsearchAgentConfig {
	data := getAgentResourceData(d, elasticsearchAgentType, elasticsearchAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.ElasticsearchAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Google Cloud Monitoring (GCM) Agent
 * https://docs.nobl9.com/Sources/google-cloud-monitoring#google-cloud-monitoring-agent
 */
const gcmAgentType = "gcm"
const gcmAgentConfigKey = "gcm_config"

func schemaAgentGCM() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		gcmAgentConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/google-cloud-monitoring#google-cloud-monitoring-agent)",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Description: "Agent configuration is not required.",
			},
		},
	}
}

func marshalAgentGCM(d *schema.ResourceData) *v1alpha.GCMAgentConfig {
	if !isAgentType(d, gcmAgentType) {
		return nil
	}

	return &v1alpha.GCMAgentConfig{}
}

/**
 * Grafana Loki Agent
 * https://docs.nobl9.com/Sources/grafana-loki#grafana-loki-agent
 */
const grafanalokiAgentType = "grafana_loki"
const grafanalokiAgentConfigKey = "grafana_loki_config"

func schemaAgentGrafanaLoki() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		grafanalokiAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/grafana-loki#grafana-loki-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "API URL endpoint to the Grafana Loki instance.",
					},
				},
			},
		},
	}
}

func marshalAgentGrafanaLoki(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.GrafanaLokiAgentConfig {
	data := getAgentResourceData(d, grafanalokiAgentType, grafanalokiAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.GrafanaLokiAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Graphite Agent
 * https://docs.nobl9.com/Sources/graphite#graphite-agent
 */
const graphiteAgentType = "graphite"
const graphiteAgentConfigKey = "graphite_config"

func schemaAgentGraphite() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		graphiteAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/graphite#graphite-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "API URL endpoint to the Graphite's instance.",
					},
				},
			},
		},
	}
}

func marshalAgentGraphite(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.GraphiteAgentConfig {
	data := getAgentResourceData(d, graphiteAgentType, graphiteAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.GraphiteAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * InfluxDB Agent
 * https://docs.nobl9.com/Sources/influxdb#influxdb-agent
 */
const influxdbAgentType = "influxdb"
const influxdbAgentConfigKey = "influxdb_config"

func schemaAgentInfluxDB() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		influxdbAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/influxdb#influxdb-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentInfluxDB(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.InfluxDBAgentConfig {
	data := getAgentResourceData(d, influxdbAgentType, influxdbAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.InfluxDBAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Instana Agent
 * https://docs.nobl9.com/Sources/instana#instana-agent
 */
const instanaAgentType = "instana"
const instanaAgentConfigKey = "instana_config"

func schemaAgentInstana() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		instanaAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/instana#instana-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentInstana(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.InstanaAgentConfig {
	data := getAgentResourceData(d, instanaAgentType, instanaAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.InstanaAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Lightstep Agent
 * https://docs.nobl9.com/Sources/lightstep#lightstep-agent
 */
const lightstepAgentType = "lightstep"
const lightstepAgentConfigKey = "lightstep_config"

func schemaAgentLightstep() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		lightstepAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/lightstep#lightstep-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentLightstep(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.LightstepAgentConfig {
	data := getAgentResourceData(d, lightstepAgentType, lightstepAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.LightstepAgentConfig{
		Organization: data["organization"].(string),
		Project:      data["project"].(string),
	}
}

/**
 * New Relic Agent
 * https://docs.nobl9.com/Sources/new-relic#new-relic-agent
 */
const newRelicAgentType = "newrelic"
const newRelicAgentConfigKey = "newrelic_config"

func schemaAgentNewRelic() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		newRelicAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/new-relic#new-relic-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentNewRelic(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.NewRelicAgentConfig {
	data := getAgentResourceData(d, newRelicAgentType, newRelicAgentConfigKey, diags)
	if data == nil {
		return nil
	}

	accID, err := strconv.Atoi(data["account_id"].(string))
	if err != nil {
		return nil
	}
	return &v1alpha.NewRelicAgentConfig{
		AccountID: accID,
	}
}

/**
 * OpenTSDB Agent
 * https://docs.nobl9.com/Sources/opentsdb#opentsdb-agent
 */
const opentsdbAgentType = "opentsdb"
const opentsdbAgentConfigKey = "opentsdb_config"

func schemaAgentOpenTSDB() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		opentsdbAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/opentsdb#opentsdb-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "OpenTSDB cluster URL.",
					},
				},
			},
		}}
}

func marshalAgentOpenTSDB(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.OpenTSDBAgentConfig {
	data := getAgentResourceData(d, opentsdbAgentType, opentsdbAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.OpenTSDBAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Pingdom Agent
 * https://docs.nobl9.com/Sources/pingdom#pingdom-agent
 */
const pingdomAgentType = "pingdom"
const pingdomAgentConfigKey = "pingdom_config"

func schemaAgentPingdom() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		pingdomAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/pingdom#pingdom-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Description: "Agent configuration is not required.",
			},
		}}
}

func marshalAgentPingdom(d *schema.ResourceData) *v1alpha.PingdomAgentConfig {
	if !isAgentType(d, pingdomAgentType) {
		return nil
	}

	return &v1alpha.PingdomAgentConfig{}
}

/**
 * Prometheus Agent
 * https://docs.nobl9.com/Sources/prometheus#prometheus-agent
 */
const prometheusAgentType = "prometheus"
const prometheusAgentConfigKey = "prometheus_config"

func schemaAgentPrometheus() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		prometheusAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/prometheus#prometheus-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Base URL to Prometheus server.",
					},
				},
			},
		},
	}
}

func marshalAgentPrometheus(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.PrometheusAgentConfig {
	data := getAgentResourceData(d, prometheusAgentType, prometheusAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	url := data["url"].(string)
	return &v1alpha.PrometheusAgentConfig{
		URL: &url,
	}
}

/**
 * Amazon Redshift Agent
 * https://docs.nobl9.com/Sources/Amazon_Redshift/?_highlight=redshift#amazon-redshift-agent
 */
const redshiftAgentType = "redshift"
const redshiftAgentConfigKey = "redshift_config"

func schemaAgentRedshift() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		redshiftAgentConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/Amazon_Redshift/?_highlight=redshift#amazon-redshift-agent)",
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Description: "Agent configuration is not required.",
			},
		},
	}
}

func marshalAgentRedshift(d *schema.ResourceData) *v1alpha.RedshiftAgentConfig {
	if !isAgentType(d, redshiftAgentType) {
		return nil
	}

	return &v1alpha.RedshiftAgentConfig{}
}

func unmarshalNewRelicAgentSpec(d *schema.ResourceData, agent v1alpha.Agent) diag.Diagnostics {
	var diags diag.Diagnostics
	if agent.Spec.NewRelic != nil {
		accountID := agent.Spec.NewRelic.AccountID
		accountIDVal := map[string]interface{}{"account_id": fmt.Sprint(accountID)}
		err := d.Set(newRelicAgentConfigKey, schema.NewSet(oneElementSet, []interface{}{accountIDVal}))
		diags = appendError(diags, err)
		return diags
	}
	diags = appendError(diags, fmt.Errorf("missing newrelic agent spec"))
	return diags
}

/**
 * Splunk Agent
 * https://docs.nobl9.com/Sources/splunk#splunk-agent
 */
const splunkAgentType = "splunk"
const splunkAgentConfigKey = "splunk_config"

func schemaAgentSplunk() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		splunkAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#splunk-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentSplunk(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.SplunkAgentConfig {
	data := getAgentResourceData(d, splunkAgentType, splunkAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.SplunkAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * Splunk Observability Agent
 * https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-agent
 */
const splunkObservabilityAgentType = "splunk_observability"
const splunkObservabilityAgentConfigKey = "splunk_observability_config"

func schemaAgentSplunkObservability() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		splunkObservabilityAgentConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Description: "[Configuration documentation]" +
				"(https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-agent)",
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

func marshalAgentSplunkObservability(
	d *schema.ResourceData,
	diags diag.Diagnostics) *v1alpha.SplunkObservabilityAgentConfig {
	data := getAgentResourceData(d, splunkObservabilityAgentType, splunkObservabilityAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.SplunkObservabilityAgentConfig{
		Realm: data["realm"].(string),
	}
}

/**
 * Sumo Logic Agent
 * https://docs.nobl9.com/Sources/sumo-logic#sumo-logic-agent
 */
const sumologicAgentType = "sumologic"
const sumologicAgentConfigKey = "sumologic_config"

func schemaAgentSumoLogic() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		sumologicAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/sumo-logic#sumo-logic-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Sumo Logic API URL.",
					},
				},
			},
		},
	}
}

func marshalAgentSumoLogic(d *schema.ResourceData, diags diag.Diagnostics) *v1alpha.SumoLogicAgentConfig {
	data := getAgentResourceData(d, sumologicAgentType, sumologicAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &v1alpha.SumoLogicAgentConfig{
		URL: data["url"].(string),
	}
}

/**
 * ThousandEyes Agent
 * https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-agent
 */
const thousandeyesAgentType = "thousandeyes"
const thousandeyesAgentConfigKey = "thousandeyes_config"

func schemaAgentThousandEyes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		thousandeyesAgentConfigKey: {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-agent)",
			MinItems:    1,
			MaxItems:    1,
			Elem: &schema.Resource{
				Description: "Agent configuration is not required.",
			},
		},
	}
}

func marshalAgentThousandEyes(d *schema.ResourceData) *v1alpha.ThousandEyesAgentConfig {
	if !isAgentType(d, thousandeyesAgentType) {
		return nil
	}

	return &v1alpha.ThousandEyesAgentConfig{}
}

func getAgentResourceData(
	d *schema.ResourceData,
	agentType,
	agentConfigKey string,
	diags diag.Diagnostics) map[string]interface{} {
	if !isAgentType(d, agentType) {
		return nil
	}
	p := d.Get(agentConfigKey).(*schema.Set).List()
	if len(p) == 0 {
		appendError(diags, fmt.Errorf("no resource data '%s' for agent type '%s'", agentConfigKey, agentType))
		return nil
	}
	resourceData := p[0].(map[string]interface{})

	return resourceData
}

func isAgentType(d *schema.ResourceData, agentType string) bool {
	agentTypeResource := d.Get(agentTypeKey).(string)
	return agentTypeResource == agentType
}
