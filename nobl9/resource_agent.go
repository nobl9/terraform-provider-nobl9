package nobl9

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	n9api "github.com/nobl9/nobl9-go"
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
		Description: "[Agent configuration documentation](https://docs.nobl9.com/the-nobl9-agent)",
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
			Description: "Source of Metrics and/or Services",
			Elem: &schema.Schema{
				Type:        schema.TypeString,
				Description: "Source of Metrics or Services",
			},
		},
		agentTypeKey: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Type of an agent. [Supported agent types](https://docs.nobl9.com/the-nobl9-agent)",
		},
		"status": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Status of created agent.",
		},
	}

	agentSchemaDefinitions := []map[string]*schema.Schema{
		schemaAgentAmazonPrometheus(),
		schemaAgentAppDynamics(),
		schemaAgentBigQuery(),
		schemaAgentDatadog(),
		schemaAgentDynatrace(),
		schemaAgentGraphite(),
		schemaAgentLightstep(),
		schemaAgentNewRelic(),
		schemaAgentOpenTSDB(),
		schemaAgentPrometheus(),
		schemaAgentSplunk(),
		schemaAgentSplunkObservability(),
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
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}
	service, diags := marshalAgent(d)
	if diags.HasError() {
		return diags
	}

	var p n9api.Payload
	p.AddObject(service)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add agent: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	readAgentDiags := resourceAgentRead(ctx, d, meta)

	return append(diags, readAgentDiags...)
}

func resourceAgentRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	client, ds := newClient(config, project)
	if ds.HasError() {
		return ds
	}

	objects, err := client.GetObject(n9api.ObjectAgent, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalAgent(d, objects)
}

func resourceAgentDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectAgent, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func unmarshalAgent(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]

	if ds := unmarshalMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	status := object["status"].(map[string]interface{})
	err := d.Set("status", status)
	diags = appendError(diags, err)

	supportedAgents := []struct {
		hclName  string
		jsonName string
	}{
		{"prometheus_config", "prometheus"},
		{"datadog_config", "datadog"},
		{"newrelic_config", "newrelic"},
		{"appdynamics_config", "appDynamics"},
		{"splunk_config", "splunk"},
		{"lightstep_config", "lightstep"},
		{"splunk_observability_config", "splunkObservability"},
		{"dynatrace_config", "dynatrace"},
		{"thousandeyes_config", "thousandEyes"},
		{"graphite_config", "graphite"},
		{"bigquery_config", "bigQuery"},
		{"opentsdb_config", "opentsdb"},
	}

	for _, name := range supportedAgents {
		ok, ds := unmarshalAgentConfig(d, object, name.hclName, name.jsonName)
		if ds.HasError() {
			diags = append(diags, ds...)
		}
		if ok {
			break
		}
	}

	return diags
}

func unmarshalAgentConfig(d *schema.ResourceData, object n9api.AnyJSONObj, hclName, jsonName string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	spec := object["spec"].(map[string]interface{})
	if spec[jsonName] == nil {
		return false, nil
	}

	// err := d.Set("agent_type", spec[""]) TODO
	err := d.Set("source_of", spec["sourceOf"])
	diags = appendError(diags, err)
	err = d.Set(hclName, schema.NewSet(oneElementSet, []interface{}{spec[jsonName]}))
	diags = appendError(diags, err)

	return true, diags
}

func marshalAgent(d *schema.ResourceData) (*n9api.Agent, diag.Diagnostics) {
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

	return &n9api.Agent{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindAgent,
			MetadataHolder: metadataHolder,
		},
		Spec: n9api.AgentSpec{
			Description:         d.Get("description").(string),
			SourceOf:            sourceOfStr,
			AmazonPrometheus:    marshalAgentAmazonPrometheus(d, diags),
			AppDynamics:         marshalAgentAppDynamics(d, diags),
			BigQuery:            marshalAgentBigQuery(d),
			Datadog:             marshalAgentDatadog(d, diags),
			Dynatrace:           marshalAgentDynatrace(d, diags),
			Graphite:            marshalAgentGraphite(d, diags),
			Lightstep:           marshalAgentLightstep(d, diags),
			NewRelic:            marshalAgentNewRelic(d, diags),
			OpenTSDB:            marshalAgentOpenTSDB(d, diags),
			Prometheus:          marshalAgentPrometheus(d, diags),
			Splunk:              marshalAgentSplunk(d, diags),
			SplunkObservability: marshalAgentSplunkObservability(d, diags),
			ThousandEyes:        marshalAgentThousandEyes(d),
			//CloudWatch:          marshalAgentCloudWatch(d),
			//Pingdom:             marshalAgentPingdom(d),
		},
	}, diags
}

/**
 * Amazon Prometheus Agent
 * https://docs.nobl9.com/Sources/Amazon_Prometheus/#ams-prometheus-agent
 */
const amazonPrometheusAgentType = "amazonprometheus"
const amazonPrometheusAgentConfigKey = "amazonprometheus_config"

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
						Description: "AWS region ex. eu-central-1",
					},
				},
			},
		},
	}
}

func marshalAgentAmazonPrometheus(d *schema.ResourceData, diags diag.Diagnostics) *n9api.AmazonPrometheusAgentConfig {
	data := getAgentResourceData(d, amazonPrometheusAgentType, amazonPrometheusAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.AmazonPrometheusAgentConfig{
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
						Description: "Base URL to a AppDynamics Controller.",
					},
				},
			},
		},
	}
}

func marshalAgentAppDynamics(d *schema.ResourceData, diags diag.Diagnostics) *n9api.AppDynamicsAgentConfig {
	data := getAgentResourceData(d, appDynamicsAgentType, appDynamicsAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	url := data["url"].(string)
	return &n9api.AppDynamicsAgentConfig{
		URL: &url,
	}
}

/**
 * BigQuery Agent
 * https://docs.nobl9.com/Sources/bigquery#bigquery-agent
 */
const bigqueryAgentType = "bigquery"

func schemaAgentBigQuery() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		bigqueryAgentType: {
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

func marshalAgentBigQuery(d *schema.ResourceData) *n9api.BigQueryAgentConfig {
	if !isAgentType(d, bigqueryAgentType) {
		return nil
	}

	return &n9api.BigQueryAgentConfig{}
}

/**
 * Datadog Agent
 * https://docs.nobl9.com/Sources/prometheus#prometheus-agent
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
					Type:        schema.TypeString,
					Required:    true,
					Description: "`com` or `eu`, Datadog SaaS instance, which corresponds to one of their two locations (https://www.datadoghq.com/ in the U.S. or https://datadoghq.eu/ in the European Union)",
				},
			},
		},
	},
	}
}

func marshalAgentDatadog(d *schema.ResourceData, diags diag.Diagnostics) *n9api.DatadogAgentConfig {
	data := getAgentResourceData(d, datadogAgentType, datadogAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.DatadogAgentConfig{
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

func marshalAgentDynatrace(d *schema.ResourceData, diags diag.Diagnostics) *n9api.DynatraceAgentConfig {
	data := getAgentResourceData(d, dynatraceAgentType, dynatraceAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.DynatraceAgentConfig{
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
						Description: "API URL endpoint of Graphite's instance.",
					},
				},
			},
		},
	}
}

func marshalAgentGraphite(d *schema.ResourceData, diags diag.Diagnostics) *n9api.GraphiteAgentConfig {
	data := getAgentResourceData(d, graphiteAgentType, graphiteAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.GraphiteAgentConfig{
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

func marshalAgentLightstep(d *schema.ResourceData, diags diag.Diagnostics) *n9api.LightstepAgentConfig {
	data := getAgentResourceData(d, lightstepAgentType, lightstepAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.LightstepAgentConfig{
		Organization: data["organization"].(string),
		Project:      data["project"].(string),
	}
}

/**
 * New Relic Agent
 * https://docs.nobl9.com/Sources/new-relic#new-relic-agent)
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
						Description: "ID number assigned to the New Relic user account",
					},
				},
			},
		},
	}
}

func marshalAgentNewRelic(d *schema.ResourceData, diags diag.Diagnostics) *n9api.NewRelicAgentConfig {
	data := getAgentResourceData(d, newRelicAgentType, newRelicAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.NewRelicAgentConfig{
		AccountID: json.Number(data["account_id"].(string)),
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

func marshalAgentOpenTSDB(d *schema.ResourceData, diags diag.Diagnostics) *n9api.OpenTSDBAgentConfig {
	data := getAgentResourceData(d, opentsdbAgentType, opentsdbAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.OpenTSDBAgentConfig{
		URL: data["url"].(string),
	}
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

func marshalAgentPrometheus(d *schema.ResourceData, diags diag.Diagnostics) *n9api.PrometheusAgentConfig {
	data := getAgentResourceData(d, prometheusAgentType, prometheusAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	url := data["url"].(string)
	return &n9api.PrometheusAgentConfig{
		URL: &url,
	}
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
						Description: "Base API URL of the Splunk Search app.",
					},
				},
			},
		},
	}
}

func marshalAgentSplunk(d *schema.ResourceData, diags diag.Diagnostics) *n9api.SplunkAgentConfig {
	data := getAgentResourceData(d, splunkAgentType, splunkAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.SplunkAgentConfig{
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
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk-observability/#splunk-observability-agent)",
			MinItems:    1,
			MaxItems:    1,
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

func marshalAgentSplunkObservability(d *schema.ResourceData, diags diag.Diagnostics) *n9api.SplunkObservabilityAgentConfig {
	data := getAgentResourceData(d, splunkObservabilityAgentType, splunkObservabilityAgentConfigKey, diags)

	if data == nil {
		return nil
	}

	return &n9api.SplunkObservabilityAgentConfig{
		Realm: data["realm"].(string),
	}
}

/**
 * ThousandEyes Agent
 * https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-agent
 */
const thousandeyesAgentType = "thousandeyes"

func schemaAgentThousandEyes() map[string]*schema.Schema {
	return map[string]*schema.Schema{"thousandeyes_config": {
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "[Configuration documentation](https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-agent)",
		MinItems:    1,
		MaxItems:    1,
		Elem: &schema.Resource{
			Description: "Agent configuration is not required.",
		},
	}}
}

func marshalAgentThousandEyes(d *schema.ResourceData) *n9api.ThousandEyesAgentConfig {
	if !isAgentType(d, thousandeyesAgentType) {
		return nil
	}

	return &n9api.ThousandEyesAgentConfig{}
}

func getAgentResourceData(d *schema.ResourceData, agentType, agentConfigKey string, diags diag.Diagnostics) map[string]interface{} {
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
