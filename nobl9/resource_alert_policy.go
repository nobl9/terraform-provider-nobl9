package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceAlertPolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),

			"severity": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Alert severity. One of Low | Medium | High.",
			},

			"condition": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Configuration of an alert condition.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"measurement": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "One of timeToBurnBudget | burnRate | burnedBudget.",
						},
						"value": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Description: "For averageBurnRate it tells how fast the error budget is burning. For burnedBudget it tells how much error budget is already burned.",
						},
						"value_string": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Used with timeToBurnBudget. When the budget would be exhausted. Expected value is a string in time duration string format.",
						},
						"lasts_for": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "How long a given condition needs to be valid to mark a condition as true. Time duration string.",
							Default:     "0m",
						},
					},
				},
			},

			"alert_method": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional, if not defined project is the same as an Alert Policy.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the alertMethod defined earlier.",
						},
					},
				},
			},
		},
		CreateContext: resourceAlertPolicyApply,
		UpdateContext: resourceAlertPolicyApply,
		DeleteContext: resourceAlertPolicyDelete,
		ReadContext:   resourceAlertPolicyRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[AlertPolicy configuration documentation](https://docs.nobl9.com/yaml-guide#alertpolicy)",
	}
}

func marshalAlertPolicy(d *schema.ResourceData) *n9api.AlertPolicy {
	return &n9api.AlertPolicy{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindAlertPolicy,
			MetadataHolder: marshalMetadata(d),
		},
		Spec: n9api.AlertPolicySpec{
			Description:  d.Get("description").(string),
			Severity:     d.Get("severity").(string),
			Conditions:   marshalAlertConditions(d),
			AlertMethods: marshalAlertMethods(d),
		},
	}
}

func marshalAlertMethods(d *schema.ResourceData) []n9api.PublicAlertMethod {
	methods := d.Get("alert_method").([]interface{})
	resultConditions := make([]n9api.PublicAlertMethod, len(methods))
	for i, m := range methods {
		method := m.(map[string]interface{})

		resultConditions[i] = n9api.PublicAlertMethod{
			ObjectHeader: n9api.ObjectHeader{
				MetadataHolder: n9api.MetadataHolder{
					Metadata: n9api.Metadata{
						Project: method["project"].(string),
						Name:    method["name"].(string),
					},
				},
			},
		}
	}

	return resultConditions
}

func marshalAlertConditions(d *schema.ResourceData) []n9api.AlertCondition {
	conditions := d.Get("condition").([]interface{})
	resultConditions := make([]n9api.AlertCondition, len(conditions))
	for i, c := range conditions {
		condition := c.(map[string]interface{})
		value := condition["value"]
		if value == 0.0 {
			value = condition["value_string"]
		}

		measurement := condition["measurement"].(string)
		op := "gte"
		if measurement == "timeToBurnBudget" {
			op = "lt"
		}

		resultConditions[i] = n9api.AlertCondition{
			Measurement:      measurement,
			Value:            value,
			LastsForDuration: condition["lasts_for"].(string),
			Operation:        op,
		}
	}

	return resultConditions
}

func unmarshalAlertPolicy(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	spec := object["spec"].(map[string]interface{})
	err := d.Set("description", spec["description"])
	diags = appendError(diags, err)
	err = d.Set("severity", spec["severity"])
	diags = appendError(diags, err)

	conditions := spec["conditions"].([]interface{})
	err = d.Set("condition", unmarshalAlertPolicyConditions(conditions))
	diags = appendError(diags, err)

	alertMethods := spec["alertMethods"].([]interface{})
	err = d.Set("alert_method", unmarshalAlertMethods(alertMethods))
	diags = appendError(diags, err)

	return diags
}

func unmarshalAlertPolicyConditions(conditions []interface{}) interface{} {
	resultConditions := make([]map[string]interface{}, len(conditions))

	for i, c := range conditions {
		condition := c.(map[string]interface{})
		var value float64
		if v, ok := condition["value"].(float64); ok {
			value = v
		}
		var valueStr string
		if v, ok := condition["value"].(string); ok {
			valueStr = v
		}

		resultConditions[i] = map[string]interface{}{
			"measurement":  condition["measurement"].(string),
			"value":        value,
			"value_string": valueStr,
			"lasts_for":    condition["lastsFor"].(string),
		}
	}

	return resultConditions
}

func unmarshalAlertMethods(alertMethods []interface{}) interface{} {
	resultMethods := make([]map[string]interface{}, len(alertMethods))

	for i, m := range alertMethods {
		method := m.(map[string]interface{})
		metadata := method["metadata"].(map[string]interface{})

		resultMethods[i] = map[string]interface{}{
			"name":    metadata["name"],
			"project": metadata["project"],
		}
	}

	return resultMethods
}

func resourceAlertPolicyApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	ap := marshalAlertPolicy(d)

	var p n9api.Payload
	p.AddObject(ap)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add alertPolicy: %s", err.Error())
	}

	d.SetId(ap.Metadata.Name)

	return resourceAlertPolicyRead(ctx, d, meta)
}

func resourceAlertPolicyRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	objects, err := client.GetObject(n9api.ObjectAlertPolicy, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalAlertPolicy(d, objects)
}

func resourceAlertPolicyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectAlertPolicy, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
