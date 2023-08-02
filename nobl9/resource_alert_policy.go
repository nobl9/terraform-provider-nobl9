package nobl9

import (
	"context"
	"reflect"

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
				Description: "Alert severity. One of `Low` | `Medium` | `High`.",
			},
			//nolint:lll
			"condition": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Configuration of an [alert condition](https://docs.nobl9.com/yaml-guide/#alertpolicy).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"measurement": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "One of `timeToBurnBudget` | `timeToBurnEntireBudget` | `burnRate` | `burnedBudget`.",
						},
						"value": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Description: "For `averageBurnRate`, it indicates how fast the error budget is burning. For `burnedBudget`, it tells how much error budget is already burned.",
						},
						"value_string": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Used with `timeToBurnBudget` or `timeToBurnEntireBudget`, indicates when the budget would be exhausted. The expected value is a string in time duration string format.",
						},
						"lasts_for": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Indicates how long a given condition needs to be valid to mark the condition as true.",
							Default:     "0m",
						},
					},
				},
			},

			"alert_method": {
				Type:             schema.TypeList,
				Optional:         true,
				Description:      "",
				DiffSuppressFunc: diffSuppressAlertMethods,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "Project name the Alert Method is in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)." +
								" If not defined, Nobl9 returns a default value for this field.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the previously defined alert method.",
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
		Description: "[Alert Policy configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#alertpolicy)",
	}
}

func diffSuppressAlertMethods(_, _, _ string, d *schema.ResourceData) bool {
	// the N9 API will return the alert methods in alphabetical by name order, however users
	// can have them in any order.  So we want to flatten the list into a 2D map and do a DeepEqual
	// comparison to see if we have any actual changes
	oldValue, newValue := d.GetChange("alert_method")
	alertMethodsOld := oldValue.([]interface{})
	alertMethodsNew := newValue.([]interface{})

	oldMap := transformAlertMethodsTo2DMap(alertMethodsOld)
	newMap := transformAlertMethodsTo2DMap(alertMethodsNew)

	return reflect.DeepEqual(oldMap, newMap)
}

func transformAlertMethodsTo2DMap(alertMethods []interface{}) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, method := range alertMethods {
		s := method.(map[string]interface{})

		values := make(map[string]string)

		values["name"] = s["name"].(string)
		values["project"] = s["name"].(string)
		result[s["name"].(string)] = values
	}
	return result
}

func marshalAlertPolicy(d *schema.ResourceData) (*n9api.AlertPolicy, diag.Diagnostics) {
	metadataHolder, diags := marshalMetadata(d)
	if diags.HasError() {
		return nil, diags
	}

	return &n9api.AlertPolicy{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindAlertPolicy,
			MetadataHolder: metadataHolder,
		},
		Spec: n9api.AlertPolicySpec{
			Description:  d.Get("description").(string),
			Severity:     d.Get("severity").(string),
			Conditions:   marshalAlertConditions(d),
			AlertMethods: marshalAlertMethods(d),
		},
	}, diags
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
		} else if measurement == "timeToBurnEntireBudget" {
			op = "lte"
		}

		resultConditions[i] = n9api.AlertCondition{
			Measurement:      measurement,
			Value:            value,
			LastsForDuration: condition["lasts_for"].(string),
			Operator:         op,
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

	if ds := unmarshalGenericMetadata(object, d); ds.HasError() {
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
	client, ds := getClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	ap, diags := marshalAlertPolicy(d)
	if diags.HasError() {
		return diags
	}

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
	client, ds := getClient(config, project)
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
	client, ds := getClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectAlertPolicy, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
