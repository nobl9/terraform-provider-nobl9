package frameworkprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
