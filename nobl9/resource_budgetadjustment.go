package nobl9

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func budgetAdjustment() *schema.Resource {
	return &schema.Resource{
		Schema:        schemaBudgetAdjustment(),
		CustomizeDiff: resourceBudgetAdjustmentValidation,
		CreateContext: resourceBudgetAdjustmentApply,
		UpdateContext: resourceBudgetAdjustmentApply,
		DeleteContext: resourceBudgetAdjustmentDelete,
		ReadContext:   resourceBudgetAdjustmentRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Budget adjustment configuration documentation](https://docs.nobl9.com/features/budget-adjustment)",
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
			Description: "The time at which the first event is scheduled to start. " +
				"The expected value must be a string representing the date and time in RFC3339 format. " +
				"Example: `2022-12-31T00:00:00Z`",
		},
		"duration": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validateDuration,
			Description: "The duration of the budget adjustment event. " +
				"The expected value for this field is a string formatted as a time duration. " +
				"The duration must be defined with a precision of 1 minute. " +
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
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Filters are used to select SLOs for the budget adjustment event.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"slos": {
						Type:     schema.TypeSet,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"slo": {
									Type:        schema.TypeList,
									MinItems:    1,
									Required:    true,
									Description: "SLO where budget adjustment event will be applied.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
											},
											"project": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
											},
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

//nolint:unparam
func resourceBudgetAdjustmentValidation(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	adjustment := marshalBudgetAdjustment(diff)
	errs := manifest.Validate([]manifest.Object{adjustment})
	if errs != nil {
		return formatErrorsAsSingleError(errs)
	}
	return nil
}

func resourceBudgetAdjustmentApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	adjustment := marshalBudgetAdjustment(d)

	if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *retry.RetryError {
		err := client.Objects().V1().Apply(ctx, []manifest.Object{adjustment})
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(adjustment.Metadata.Name)
	return resourceBudgetAdjustmentRead(ctx, d, meta)
}

func marshalBudgetAdjustment(r resourceInterface) *budgetadjustment.BudgetAdjustment {
	firstEventStart, _ := time.Parse(time.RFC3339, r.Get("first_event_start").(string))

	adjustment := budgetadjustment.New(
		budgetadjustment.Metadata{
			Name:        r.Get("name").(string),
			DisplayName: r.Get("display_name").(string),
		},
		budgetadjustment.Spec{
			Description:     r.Get("description").(string),
			FirstEventStart: firstEventStart,
			Duration:        r.Get("duration").(string),
			Rrule:           r.Get("rrule").(string),
			Filters:         marshalFilters(r.Get("filters")),
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

	diags = appendError(diags, d.Set("name", object.Metadata.Name))
	diags = appendError(diags, d.Set("display_name", object.Metadata.DisplayName))
	diags = appendError(diags, d.Set("description", object.Spec.Description))
	diags = appendError(diags, d.Set("first_event_start", object.Spec.FirstEventStart.Format(time.RFC3339)))
	diags = appendError(diags, d.Set("duration", object.Spec.Duration))
	diags = appendError(diags, d.Set("rrule", object.Spec.Rrule))

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

	if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *retry.RetryError {
		err := client.Objects().V1().DeleteByName(ctx, manifest.KindBudgetAdjustment, "", d.Id())
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
