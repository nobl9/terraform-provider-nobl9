package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// sloObjectiveValuePlanModifier should not be used with [float64planmodifier.UseStateForUnknown].
// The latter will have no effect.
type sloObjectiveValuePlanModifier struct{}

func (s sloObjectiveValuePlanModifier) Description(ctx context.Context) string {
	return s.MarkdownDescription(ctx)
}

func (s sloObjectiveValuePlanModifier) MarkdownDescription(context.Context) string {
	return "Verifies that the `value` attribute in `objective` is not set when defining composite SLOs."
}

func (s sloObjectiveValuePlanModifier) PlanModifyFloat64(
	ctx context.Context,
	req planmodifier.Float64Request,
	resp *planmodifier.Float64Response,
) {
	var objectives []ObjectiveModel
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("objective"), &objectives)...)
	if resp.Diagnostics.HasError() {
		return
	}
	nullOrUnknown := isNullOrUnknown(req.PlanValue)
	compositeObjectives := hasCompositeObjectives(objectives)
	switch {
	case !nullOrUnknown && compositeObjectives:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"objective value cannot be set when defining composite SLOs",
			"",
		)
	case nullOrUnknown && !compositeObjectives:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"objective value must be set for ratio and threshold objectives",
			"",
		)
	}
}

type sloProjectPlanModifier struct{}

func (s sloProjectPlanModifier) Description(ctx context.Context) string {
	return s.MarkdownDescription(ctx)
}

func (s sloProjectPlanModifier) MarkdownDescription(context.Context) string {
	return "Modifies the SLO plan when the `project` attribute is changed. " +
		"This modifier ensures that no other attributes are changed along with the project change, " +
		"and provides warnings about the implications of moving an SLO between projects."
}

func (s sloProjectPlanModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	if isNullOrUnknown(req.StateValue) || req.StateValue == req.PlanValue {
		return
	}
	resp.Diagnostics.Append(s.modifyPlanForProjectChange(ctx, req.Plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s sloProjectPlanModifier) modifyPlanForProjectChange(ctx context.Context, plan tfsdk.Plan) diag.Diagnostics {
	diags := make(diag.Diagnostics, 0)

	var alertPolicies []string
	alertPoliciesPath := path.Root("alert_policies")
	diags.Append(plan.GetAttribute(ctx, alertPoliciesPath, &alertPolicies)...)
	if diags.HasError() {
		return diags
	}
	if len(alertPolicies) > 0 {
		diags.AddAttributeError(alertPoliciesPath,
			"Cannot move SLO between Projects with attached Alert Policies.",
			"You must first remove Alert Policies attached to this SLO before attempting to change it's Project.",
		)
		return diags
	}

	diags.AddAttributeWarning(path.Root("project"),
		"Changing the Project results in a dedicated operation which has several side effects (see details).",
		`Moving an SLO between Projects:
  - Creates a new Project and/or Service if the specified target objects do not yet exist.
    It is best practice to define these new objects in the Terraform configuration, before moving the SLO.
  - Updates SLO’s project in the composite SLO definition and Budget Adjustment filters.
    These definitions, which reference any objectives from the moved SLO need to be updated manually.
  - Updates its link — the former link won't work anymore.
  - Removes it from reports filtered by its previous path.
`)
	return diags
}
