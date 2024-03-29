package nobl9

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
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
			"cooldown": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "5m",
				//nolint:lll
				Description: "An interval measured from the last time stamp when all alert policy conditions were satisfied before alert is marked as resolved",
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
							DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
								// To be backward compatible with lasts for with default=0m that was set before.
								return oldValue == "0m" && newValue == ""
							},
						},
						"alerting_window": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Duration over which the burn rate is evaluated.",
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

func marshalAlertPolicy(d *schema.ResourceData) (*v1alphaAlertPolicy.AlertPolicy, diag.Diagnostics) {
	var displayName string
	if dn := d.Get("display_name"); dn != nil {
		displayName = dn.(string)
	}

	labelsMarshaled, diags := getMarshaledLabels(d)
	if diags.HasError() {
		return nil, diags
	}

	alertPolicy := v1alphaAlertPolicy.New(
		v1alphaAlertPolicy.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: displayName,
			Project:     d.Get("project").(string),
			Labels:      labelsMarshaled,
		},
		v1alphaAlertPolicy.Spec{
			Description:      d.Get("description").(string),
			Severity:         d.Get("severity").(string),
			CoolDownDuration: d.Get("cooldown").(string),
			Conditions:       marshalAlertConditions(d),
			AlertMethods:     marshalAlertMethods(d),
		})
	return &alertPolicy, diags
}

func marshalAlertMethods(d *schema.ResourceData) []v1alphaAlertPolicy.AlertMethodRef {
	methods := d.Get("alert_method").([]interface{})
	resultConditions := make([]v1alphaAlertPolicy.AlertMethodRef, len(methods))
	for i, m := range methods {
		method := m.(map[string]interface{})
		resultConditions[i] = v1alphaAlertPolicy.AlertMethodRef{
			Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
				Name:    method["name"].(string),
				Project: method["project"].(string),
			},
		}
	}
	return resultConditions
}

func marshalAlertConditions(d *schema.ResourceData) []v1alphaAlertPolicy.AlertCondition {
	conditions := d.Get("condition").([]interface{})
	resultConditions := make([]v1alphaAlertPolicy.AlertCondition, len(conditions))
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

		lastsFor := condition["lasts_for"].(string)
		alertingWindow := condition["alerting_window"].(string)

		if lastsFor == "0m" && alertingWindow != "" {
			// To be backward compatible with lasts for with default=0m that was set before, when user
			// wants to switch to use alerting_window instead of lasts_for.
			lastsFor = ""
		}

		resultConditions[i] = v1alphaAlertPolicy.AlertCondition{
			Measurement:      measurement,
			Value:            value,
			LastsForDuration: lastsFor,
			AlertingWindow:   alertingWindow,
			Operator:         op,
		}
	}

	return resultConditions
}

func unmarshalAlertPolicy(d *schema.ResourceData, objects []v1alphaAlertPolicy.AlertPolicy) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	metadata := object.Metadata
	err := d.Set("name", metadata.Name)
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata.DisplayName)
	diags = appendError(diags, err)
	err = d.Set("project", metadata.Project)
	diags = appendError(diags, err)

	if labelsRaw := metadata.Labels; len(labelsRaw) > 0 {
		err = d.Set("label", unmarshalLabels(labelsRaw))
		diags = appendError(diags, err)
	}

	spec := object.Spec
	err = d.Set("description", spec.Description)
	diags = appendError(diags, err)
	err = d.Set("severity", spec.Severity)
	diags = appendError(diags, err)
	err = d.Set("cooldown", spec.CoolDownDuration)
	diags = appendError(diags, err)

	conditions := spec.Conditions
	err = d.Set("condition", unmarshalAlertPolicyConditions(conditions))
	diags = appendError(diags, err)

	alertMethods := spec.AlertMethods
	err = d.Set("alert_method", unmarshalAlertMethods(alertMethods))
	diags = appendError(diags, err)

	return diags
}

func unmarshalAlertPolicyConditions(conditions []v1alphaAlertPolicy.AlertCondition) interface{} {
	resultConditions := make([]map[string]interface{}, len(conditions))
	for i, condition := range conditions {
		var value json.Number
		if v, ok := condition.Value.(json.Number); ok {
			value = v
		}
		var valueStr string
		if v, ok := condition.Value.(string); ok {
			valueStr = v
		}
		resultConditions[i] = map[string]interface{}{
			"measurement":     condition.Measurement,
			"value":           value,
			"value_string":    valueStr,
			"lasts_for":       condition.LastsForDuration,
			"alerting_window": condition.AlertingWindow,
		}
	}

	return resultConditions
}

func unmarshalAlertMethods(alertMethods []v1alphaAlertPolicy.AlertMethodRef) interface{} {
	resultMethods := make([]map[string]interface{}, len(alertMethods))

	for i, method := range alertMethods {
		metadata := method.Metadata

		resultMethods[i] = map[string]interface{}{
			"name":    metadata.Name,
			"project": metadata.Project,
		}
	}

	return resultMethods
}

func resourceAlertPolicyApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	ap, diags := marshalAlertPolicy(d)
	if diags.HasError() {
		return diags
	}
	resultAp := manifest.SetDefaultProject([]manifest.Object{ap}, config.Project)
	err := client.Objects().V1().Apply(ctx, resultAp)
	if err != nil {
		return diag.Errorf("could not add alertPolicy: %s", err.Error())
	}
	d.SetId(ap.Metadata.Name)
	return resourceAlertPolicyRead(ctx, d, meta)
}

func resourceAlertPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	alertPolicies, err := client.Objects().V1().GetV1alphaAlertPolicies(ctx, v1Objects.GetAlertPolicyRequest{
		Project: project,
		Names:   []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalAlertPolicy(d, alertPolicies)
}

func resourceAlertPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	err := client.Objects().V1().DeleteByName(ctx, manifest.KindAlertPolicy, project, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
