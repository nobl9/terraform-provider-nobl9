package nobl9

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/teambition/rrule-go"
)

func budgetAdjustment() *schema.Resource {
	return &schema.Resource{
		Schema:        schemaBudgetAdjustment(),
		CreateContext: resourceBudgetAdjustmentApply,
		UpdateContext: resourceBudgetAdjustmentApply,
		DeleteContext: resourceBudgetAdjustmentDelete,
		ReadContext:   resourceBudgetAdjustmentRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Budget adjustment configuration documentation](https://docs.nobl9.com/yaml-guide#budget-adjustment)",
	}
}

func schemaBudgetAdjustment() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name":         schemaName(),
		"display_name": schemaDisplayName(),
		"description":  schemaDescription(),
		"first_event_start": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validateDateTime,
			Description: "The time of the first event start. " +
				"The expected value is a string with date in RFC3339 format. " +
				"Example: `2022-12-31T00:00:00Z`",
		},
		"duration": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validateDuration,
			Description: "The duration of the budget adjustment event. " +
				"The expected value is a string in time duration string format. " +
				"Duration must be defined with 1 minute precision." +
				"Example: `1h10m`",
		},
		"rrule": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validateRrule,
			Description: "The recurrence rule for the budget adjustment event. " +
				"The expected value is a string in RRULE format. " +
				"Example: `FREQ=MONTHLY;BYMONTHDAY=1`",
		},
		"filters": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"slos": {
						Type:     schema.TypeSet,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"slo": {
									Type:     schema.TypeList,
									MinItems: 1,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name":    schemaName(),
											"project": schemaProject(),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func validateDuration(v interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	_, err := time.ParseDuration(v.(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid duration format",
			Detail:        fmt.Sprintf("Invalid duration format: %s", v),
			AttributePath: path,
		})
	}
	return diags
}

func validateRrule(v interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	_, err := rrule.StrToRRule(v.(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid rrule format",
			Detail:        fmt.Sprintf("Invalid rrule format: %s", v),
			AttributePath: path,
		})
	}
	return diags
}

func resourceBudgetAdjustmentApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	budgetAdjustment := marshalBudgetAdjustment(d)

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		err := client.Objects().V1().Apply(ctx, []manifest.Object{budgetAdjustment})
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(budgetAdjustment.Metadata.Name)
	return resourceBudgetAdjustmentRead(ctx, d, meta)
}

func marshalBudgetAdjustment(d *schema.ResourceData) *budgetadjustment.BudgetAdjustment {
	firstEventStart, _ := time.Parse(time.RFC3339, d.Get("first_event_start").(string))

	adjustment := budgetadjustment.New(
		budgetadjustment.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: d.Get("display_name").(string),
		},
		budgetadjustment.Spec{
			Description:     d.Get("description").(string),
			FirstEventStart: firstEventStart,
			Duration:        d.Get("duration").(string),
			Rrule:           d.Get("rrule").(string),
			Filters:         marshalFilters(d.Get("filters")),
		})

	return &adjustment
}

func marshalFilters(filters interface{}) budgetadjustment.Filters {
	filtersSet := filters.(*schema.Set)
	if filtersSet.Len() == 0 {
		return budgetadjustment.Filters{}
	}
	slos := filtersSet.List()[0].(map[string]interface{})["slos"].(*schema.Set)
	slosList := slos.List()[0].(map[string]interface{})["slo"].([]interface{})
	sloRef := make([]budgetadjustment.SLORef, 0, len(slosList))
	for _, filter := range slosList {
		f := filter.(map[string]interface{})
		slo := budgetadjustment.SLORef{
			Name:    f["name"].(string),
			Project: f["project"].(string),
		}
		sloRef = append(sloRef, slo)
	}

	return budgetadjustment.Filters{
		SLOs: sloRef,
	}
}

func resourceBudgetAdjustmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	budgetAdjustments, err := client.Objects().V1().GetBudgetAdjustments(ctx, v1.GetBudgetAdjustmentRequest{
		Names: []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalBudgetAdjustment(d, budgetAdjustments)
}

func unmarshalBudgetAdjustment(d *schema.ResourceData, objects []budgetadjustment.BudgetAdjustment) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics
	var err error

	err = d.Set("name", object.Metadata.Name)
	diags = appendError(diags, err)

	err = d.Set("display_name", object.Metadata.DisplayName)
	diags = appendError(diags, err)

	err = d.Set("description", object.Spec.Description)
	diags = appendError(diags, err)

	err = d.Set("first_event_start", object.Spec.FirstEventStart.Format(time.RFC3339))
	diags = appendError(diags, err)

	err = d.Set("duration", object.Spec.Duration)
	diags = appendError(diags, err)

	err = d.Set("rrule", object.Spec.Rrule)
	diags = appendError(diags, err)

	err = unmarshalFilters(d, object.Spec.Filters)
	diags = appendError(diags, err)

	return diags
}

func unmarshalFilters(d *schema.ResourceData, filters budgetadjustment.Filters) error {
	slos := make([]map[string]interface{}, 0, len(filters.SLOs))
	for _, slo := range filters.SLOs {
		sloMap := map[string]interface{}{
			"name":    slo.Name,
			"project": slo.Project,
		}
		slos = append(slos, sloMap)
	}

	f := map[string]interface{}{
		"slos": schema.NewSet(oneElementSet, []interface{}{
			map[string]interface{}{
				"slo": slos,
			},
		}),
	}

	return d.Set("filters", schema.NewSet(oneElementSet, []interface{}{f}))
}

func resourceBudgetAdjustmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		err := client.Objects().V1().DeleteByName(ctx, manifest.KindBudgetAdjustment, "", d.Id())
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
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
