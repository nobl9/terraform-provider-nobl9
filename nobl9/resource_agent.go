package nobl9

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"labels":       schemaLabels(),
			"project":      schemaProject(),
			"description":  schemaDescription(),

			"source_of": {
				Type:        schema.TypeList, // TODO should provider verify the options or the API do that? is it possible to check Sets with SetFunc
				Required:    true,
				Description: "Source of either Metrics or Services",
				MinItems:    1,
				MaxItems:    2,
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "",
				},
			},

			"prometheus": {
				// TODO I don't like that is has to be a Set... maybe there is a better way?
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "",
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
			// TODO add status as computed field, do not support it in Apply, support it in Read
		},
		CreateContext: resourceAgentApply,
		UpdateContext: resourceAgentApply,
		DeleteContext: resourceAgentDelete,
		ReadContext:   resourceAgentRead,
		//Importer:  TODO impl me; discuss how project should be selected
		Description: "* [HTTP API](https://api-docs.app.nobl9.com/)",
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
			APIVersion: apiVersion,
			Kind:       "Agent",
			// TODO metadataHolder marshaler can be reused
			MetadataHolder: n9api.MetadataHolder{
				Metadata: n9api.Metadata{
					Name:        d.Get("name").(string),
					DisplayName: d.Get("display_name").(string),
					Project:     d.Get("project").(string),
					// TODO Metadata should also support labels - SDK is outdated
				},
			},
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
	prom := d.Get("prometheus").(*schema.Set).List()
	if len(prom) == 0 {
		return nil
	}

	if len(prom) > 1 {
		// TODO diag error
		return nil
	}

	url := prom[0].(map[string]interface{})["url"].(string)
	return &n9api.PrometheusConfig{
		URL:              &url,
		ServiceDiscovery: nil,
	}
}

func unmarshalAgentPrometheus(d *schema.ResourceData, objects []n9api.AnyJSONObj) error {
	// TODO how to mark that object was removed?

	if len(objects) != 1 {
		return fmt.Errorf("expected one object for id=%s but got %d", d.Id(), len(objects))
	}
	object := objects[0]

	// TODO metadata unmarshal can be reused
	metadata := object["metadata"].(map[string]interface{})
	err := d.Set("name", metadata["name"])
	if err != nil {
		return err
	}
	err = d.Set("display_name", metadata["displayName"])
	if err != nil {
		return err
	}
	err = d.Set("labels", metadata["labels"]) // TODO labels are not supported yet on Agent - check it on SLO
	if err != nil {
		return err
	}
	err = d.Set("project", metadata["project"])
	if err != nil {
		return err
	}
	err = d.Set("description", metadata["description"])
	if err != nil {
		return err
	}

	specMap := object["spec"].(map[string]interface{})

	err = d.Set("source_of", specMap["sourceOf"])
	if err != nil {
		return err
	}

	oneElementSet := func(i interface{}) int { return 0 }
	err = d.Set("prometheus", schema.NewSet(oneElementSet, []interface{}{specMap["prometheus"]}))
	if err != nil {
		return err
	}

	return nil
}

func resourceAgentApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*n9api.Client)
	var diags diag.Diagnostics
	// TODO []diags should be probably returned from marshal method to give the user all errors

	service := marshalAgent(d)
	var p n9api.Payload
	p.AddObject(service)

	err := c.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add service: %s", err.Error())
	}

	// TODO technically it is correct but it might break ForceNew set on name field - check it
	d.SetId(service.Metadata.Name)

	return diags
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*n9api.Client)
	// TODO if project is set in the provider, what should we do?
	//  we will read from the project from provider but we should use project from the resource
	//  Maybe project from provider should be only used as a default, when not set in resources and additionally in imports?

	objects, err := c.GetObject(n9api.ObjectAgent, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := unmarshalAgentPrometheus(d, objects); err != nil {
		// TODO []diags should be probably returned from unmarshal method to give the user all errors
		return diag.FromErr(err)
	}

	return nil
}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*n9api.Client)

	err := c.DeleteObjectsByName(n9api.ObjectAgent, d.Id())
	if err != nil {
		diag.FromErr(err)
	}

	return nil
}
