package nobl9

import (
	"context"

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
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 2,
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
							Description: "",
						},
					},
				},
			},
			// TODO support other agent types
			"status": {
				Type:     schema.TypeMap,
				Computed: true,
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
			APIVersion:     apiVersion,
			Kind:           "Agent",
			MetadataHolder: marshalMetadata(d),
		},
		Spec: n9api.AgentSpec{
			Description:         d.Get("description").(string),
			SourceOf:            sourceOfStr,
			Prometheus:          marshalAgentPrometheus(d),
			Datadog:             nil,
			NewRelic:            nil,
			AppDynamics:         nil,
			Splunk:              nil,
			Lightstep:           nil,
			SplunkObservability: nil,
			Dynatrace:           nil,
			ThousandEyes:        nil,
			Graphite:            nil,
			BigQuery:            nil,
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

func unmarshalAgent(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalMetadata(object, d); len(ds) > 0 {
		diags = append(diags, ds...)
	}

	status := object["status"].(map[string]interface{})
	err := d.Set("status", status)
	appendError(diags, err)

	if ds := unmarshalAgentPrometheus(d, object); len(ds) > 0 {
		diags = append(diags, ds...)
	}

	return diags
}

func unmarshalAgentPrometheus(d *schema.ResourceData, object n9api.AnyJSONObj) diag.Diagnostics {
	var diags diag.Diagnostics
	spec := object["spec"].(map[string]interface{})

	err := d.Set("source_of", spec["sourceOf"])
	appendError(diags, err)
	err = d.Set("prometheus", schema.NewSet(oneElementSet, []interface{}{spec["prometheus"]}))
	appendError(diags, err)

	return diags
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
	if ds != nil {
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
	if ds != nil {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectAgent, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
