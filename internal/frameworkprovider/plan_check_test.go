package frameworkprovider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var _ plancheck.PlanCheck = expectNoChangeInPlan{}

// ExpectNoChangeInPlan returns is a plan check that asserts there is no change in the specified attribute in the plan.
type expectNoChangeInPlan struct {
	attrName string
}

// CheckPlan implements the [plancheck.PlanCheck] interface.
func (e expectNoChangeInPlan) CheckPlan(
	_ context.Context,
	req plancheck.CheckPlanRequest,
	resp *plancheck.CheckPlanResponse,
) {
	if len(req.Plan.ResourceChanges) != 1 {
		resp.Error = fmt.Errorf("expected exactly one resource change, but got %d", len(req.Plan.ResourceChanges))
		return
	}
	change := req.Plan.ResourceChanges[0].Change
	before, _ := change.Before.(map[string]any)[e.attrName]
	after, _ := change.After.(map[string]any)[e.attrName]
	if !reflect.DeepEqual(before, after) {
		resp.Error = fmt.Errorf("expected no change in %s, but got '%v' -> '%v'", e.attrName, before, after)
	}
}
