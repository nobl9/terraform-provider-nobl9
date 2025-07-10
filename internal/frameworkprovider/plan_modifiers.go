package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

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
	// If state already has a value or the planned value is null, we do not perform any checks.
	if !isNullOrUnknown(req.StateValue) || isNullOrUnknown(req.PlanValue) {
		return
	}
	var objectives []ObjectiveModel
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("objective"), &objectives)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if s.hasCompositeObjectives(objectives) {
		resp.Diagnostics.AddError("objective value cannot be set when defining new composite SLOs", "")
	}
}

func (s sloObjectiveValuePlanModifier) hasCompositeObjectives(objectives []ObjectiveModel) bool {
	for _, objective := range objectives {
		if len(objective.Composite) > 0 {
			return true
		}
	}
	return false
}
