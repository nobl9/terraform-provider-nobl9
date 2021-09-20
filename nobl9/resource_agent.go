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

			"prometheus": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-prometheus)",
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

			"datadog": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-datadog)",
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

			"newrelic": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-new-relic)",
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

			"appdynamics": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-appdynamics)",
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

			"splunk": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-splunk)",
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

			"lightstep": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-lightstep)",
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

			"splunk_observability": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-splunk-observability)",
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

			"dynatrace": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-dynatrace)",
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

			"thousandeyes": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-thousandeyes)",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Description: "Agent configuration is not required.",
				},
			},

			"graphite": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-graphite)",
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

			"bigquery": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-bigquery)",
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Description: "Agent configuration is not required.",
				},
			},

			"opentsdb": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[Configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent-using-opentsdb)",
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
		Description: "[Agent configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#agent)",
	}
}

func marshalAgent(d *schema.ResourceData) *n9api.Agent {
	sourceOf := d.Get("source_of").([]interface{})
	sourceOfStr := make([]string, len(sourceOf))
	for i, s := range sourceOf {
		sourceOfStr[i] = s.(string)
	}

	return &n9api.Agent{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           "Agent",
			MetadataHolder: marshalMetadata(d),
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
	}
}

func marshalAgentPrometheus(d *schema.ResourceData) *n9api.PrometheusConfig {
	p := d.Get("prometheus").(*schema.Set).List()
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
	p := d.Get("datadog").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	ddog := p[0].(map[string]interface{})

	return &n9api.DatadogAgentConfig{
		Site: ddog["site"].(string),
	}
}

func marshalAgentNewRelic(d *schema.ResourceData) *n9api.NewRelicAgentConfig {
	p := d.Get("newrelic").(*schema.Set).List()
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
	p := d.Get("appdynamics").(*schema.Set).List()
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
	p := d.Get("splunk").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunk := p[0].(map[string]interface{})

	return &n9api.SplunkAgentConfig{
		URL: splunk["url"].(string),
	}
}

func marshalAgentLightstep(d *schema.ResourceData) *n9api.LightstepAgentConfig {
	p := d.Get("lightstep").(*schema.Set).List()
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
	p := d.Get("splunk_observability").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	splunk := p[0].(map[string]interface{})

	// TODO SplunkObs now supports `realm` not `url`
	return &n9api.SplunkObservabilityAgentConfig{
		Realm: splunk["realm"].(string),
	}
}

func marshalDynatrace(d *schema.ResourceData) *n9api.DynatraceAgentConfig {
	p := d.Get("dynatrace").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	dynatrace := p[0].(map[string]interface{})

	return &n9api.DynatraceAgentConfig{
		URL: dynatrace["url"].(string),
	}
}

func marshalAgentThousandEyes(d *schema.ResourceData) *n9api.ThousandEyesAgentConfig {
	p := d.Get("thousandeyes").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}

	return &n9api.ThousandEyesAgentConfig{}
}

func marshalAgentGraphite(d *schema.ResourceData) *n9api.GraphiteAgentConfig {
	p := d.Get("graphite").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	graphite := p[0].(map[string]interface{})

	return &n9api.GraphiteAgentConfig{
		URL: graphite["url"].(string),
	}
}

func marshalAgentOpenTSDB(d *schema.ResourceData) *n9api.OpenTSDBAgentConfig {
	p := d.Get("opentsdb").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}
	graphite := p[0].(map[string]interface{})

	return &n9api.OpenTSDBAgentConfig{
		URL: graphite["url"].(string),
	}
}

func marshalAgentBigQuery(d *schema.ResourceData) *n9api.BigQueryAgentConfig {
	p := d.Get("bigquery").(*schema.Set).List()
	if len(p) == 0 {
		return nil
	}

	return &n9api.BigQueryAgentConfig{}
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
	appendError(diags, err)

	supportedAgents := []struct {
		hclName  string
		jsonName string
	}{
		{"prometheus", "prometheus"},
		{"datadog", "datadog"},
		{"newrelic", "newrelic"},
		{"appdynamics", "appDynamics"},
		{"splunk", "splunk"},
		{"lightstep", "lightstep"},
		{"splunk_observability", "splunkObservability"},
		{"dynatrace", "dynatrace"},
		{"thousandeyes", "thousandEyes"},
		{"graphite", "graphite"},
		{"bigquery", "bigQuery"},
		{"opentsdb", "opentsdb"},
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

	err := d.Set("source_of", spec["sourceOf"])
	appendError(diags, err)
	err = d.Set(hclName, schema.NewSet(oneElementSet, []interface{}{spec[jsonName]}))
	appendError(diags, err)

	return true, diags
}

func resourceAgentApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	service := marshalAgent(d)

	var p n9api.Payload
	p.AddObject(service)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add service: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return resourceAgentRead(ctx, d, meta)
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
