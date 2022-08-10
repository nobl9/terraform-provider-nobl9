package nobl9

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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

			"agent_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of an agent. [Supported agent types](https://docs.nobl9.com/the-nobl9-agent)",
			},

			"prometheus_config": {
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

			"datadog_config": {
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

			"newrelic_config": {
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

			"appdynamics_config": {
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

			"splunk_config": {
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

			"lightstep_config": {
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

			"splunk_observability_config": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk-observability)",
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

			"dynatrace_config": {
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

			"thousandeyes_config": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-agent)",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Description: "Agent configuration is not required.",
				},
			},

			"graphite_config": {
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

			"bigquery_config": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://docs.nobl9.com/Sources/bigquery#bigquery-agent)",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Description: "Agent configuration is not required.",
				},
			},

			"opentsdb_config": {
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
			},

			"status": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Status of created agent.",
			},
		},
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

func marshalAgent(d *schema.ResourceData) (*n9api.Agent, diag.Diagnostics) {
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
			Prometheus:          marshalAgentPrometheus(d),
			Datadog:             marshalAgentDatadog(d),
			NewRelic:            marshalAgentNewRelic(d),
			AppDynamics:         marshalAgentAppDynamics(d),
			Splunk:              marshalAgentSplunk(d),
			Lightstep:           marshalAgentLightstep(d),
			SplunkObservability: marshalAgentSplunkObservability(d),
			Dynatrace:           marshalDynatrace(d),
			ThousandEyes:        marshalAgentThousandEyes(d),
			Graphite:            marshalAgentGraphite(d),
			BigQuery:            marshalAgentBigQuery(d),
			OpenTSDB:            marshalAgentOpenTSDB(d),
		},
	}, diags
}

func marshalAgentPrometheus(d *schema.ResourceData) *n9api.PrometheusConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "prometheus" {
		return nil
	}
	p := d.Get("prometheus_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	prom := p[0].(map[string]interface{})

	url := prom["url"].(string)
	return &n9api.PrometheusConfig{
		URL: &url,
	}
}

func marshalAgentDatadog(d *schema.ResourceData) *n9api.DatadogAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "datadog" {
		return nil
	}
	p := d.Get("datadog_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	ddog := p[0].(map[string]interface{})

	return &n9api.DatadogAgentConfig{
		Site: ddog["site"].(string),
	}
}

func marshalAgentNewRelic(d *schema.ResourceData) *n9api.NewRelicAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "newrelic" {
		return nil
	}
	p := d.Get("newrelic_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	newrelic := p[0].(map[string]interface{})

	accountID := newrelic["account_id"].(string)
	return &n9api.NewRelicAgentConfig{
		AccountID: json.Number(accountID),
	}
}

func marshalAgentAppDynamics(d *schema.ResourceData) *n9api.AppDynamicsAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "appdynamics" {
		return nil
	}
	p := d.Get("appdynamics_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	appdynamics := p[0].(map[string]interface{})

	url := appdynamics["url"].(string)
	return &n9api.AppDynamicsAgentConfig{
		URL: &url,
	}
}

func marshalAgentSplunk(d *schema.ResourceData) *n9api.SplunkAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "splunk" {
		return nil
	}
	p := d.Get("splunk_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunk := p[0].(map[string]interface{})

	return &n9api.SplunkAgentConfig{
		URL: splunk["url"].(string),
	}
}

func marshalAgentLightstep(d *schema.ResourceData) *n9api.LightstepAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "lightstep" {
		return nil
	}
	p := d.Get("lightstep_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	lightstep := p[0].(map[string]interface{})

	return &n9api.LightstepAgentConfig{
		Organization: lightstep["organization"].(string),
		Project:      lightstep["project"].(string),
	}
}

func marshalAgentSplunkObservability(d *schema.ResourceData) *n9api.SplunkObservabilityAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "splunk_observability" {
		return nil
	}
	p := d.Get("splunk_observability_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunk := p[0].(map[string]interface{})

	return &n9api.SplunkObservabilityAgentConfig{
		Realm: splunk["realm"].(string),
	}
}

func marshalDynatrace(d *schema.ResourceData) *n9api.DynatraceAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "dynatrace" {
		return nil
	}
	p := d.Get("dynatrace_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	dynatrace := p[0].(map[string]interface{})

	return &n9api.DynatraceAgentConfig{
		URL: dynatrace["url"].(string),
	}
}

func marshalAgentThousandEyes(d *schema.ResourceData) *n9api.ThousandEyesAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "thousandeyes" {
		return nil
	}

	return &n9api.ThousandEyesAgentConfig{}
}

func marshalAgentGraphite(d *schema.ResourceData) *n9api.GraphiteAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "graphite" {
		return nil
	}
	p := d.Get("graphite_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	graphite := p[0].(map[string]interface{})

	return &n9api.GraphiteAgentConfig{
		URL: graphite["url"].(string),
	}
}

func marshalAgentBigQuery(d *schema.ResourceData) *n9api.BigQueryAgentConfig {
	agentType := d.Get("agent_type").(string)
	if agentType != "bigquery" {
		return nil
	}

	return &n9api.BigQueryAgentConfig{}
}

func marshalAgentOpenTSDB(d *schema.ResourceData) *n9api.OpenTSDBAgentConfig {
	p := d.Get("opentsdb_config").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	graphite := p[0].(map[string]interface{})

	return &n9api.OpenTSDBAgentConfig{
		URL: graphite["url"].(string),
	}
}

func unmarshalAgent(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

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

	//err := d.Set("agent_type", spec[""]) TODO
	err := d.Set("source_of", spec["sourceOf"])
	diags = appendError(diags, err)
	err = d.Set(hclName, schema.NewSet(oneElementSet, []interface{}{spec[jsonName]}))
	diags = appendError(diags, err)

	return true, diags
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

	return resourceAgentRead(ctx, d, meta)
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
