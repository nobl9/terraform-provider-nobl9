package frameworkprovider

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var _ plancheck.PlanCheck = expectOnlyOneChangeInPlan{}

// expectOnlyOneChangeInPlan is a plan check that asserts there is only one change in the plan
// and that this change is on the given attribute.
type expectOnlyOneChangeInPlan struct {
	attrName string
}

// CheckPlan implements the [plancheck.PlanCheck] interface.
func (e expectOnlyOneChangeInPlan) CheckPlan(
	_ context.Context,
	req plancheck.CheckPlanRequest,
	resp *plancheck.CheckPlanResponse,
) {
	if len(req.Plan.ResourceChanges) != 1 {
		resp.Error = fmt.Errorf("expected exactly one resource change, but got %d", len(req.Plan.ResourceChanges))
		return
	}
	change := req.Plan.ResourceChanges[0].Change
	before, _ := change.Before.(map[string]any)
	after, _ := change.After.(map[string]any)
	diff := calculatePlanDiff(before, after)

	if len(diff.Removed) > 0 {
		resp.Error = fmt.Errorf("expected no removals in the plan, but got removed attrs: %v", diff.Removed)
		return
	}
	if len(diff.Added) > 0 {
		resp.Error = fmt.Errorf("expected no additions in the plan, but got added attrs: %v", diff.Added)
		return
	}
	if len(diff.Modified) != 1 {
		resp.Error = fmt.Errorf("expected exactly one modified attribute, but got %d: %v",
			len(diff.Modified), diff.Modified)
		return
	}
	if reflect.DeepEqual(before[e.attrName], after[e.attrName]) {
		resp.Error = fmt.Errorf("expected '%s' to change, but it did not: before=%v, after=%v",
			e.attrName, before[e.attrName], after[e.attrName])
	}
}

// planDiff represents the difference between before and after keys
type planDiff struct {
	Added    []string
	Removed  []string
	Modified []string
}

// calculatePlanDiff compares two plans and returns the differences between them.
func calculatePlanDiff(before, after map[string]any) planDiff {
	diff := planDiff{
		Added:    []string{},
		Removed:  []string{},
		Modified: []string{},
	}

	for key := range after {
		beforeValue, exists := before[key]
		switch {
		case !exists:
			diff.Added = append(diff.Added, key)
		case !reflect.DeepEqual(beforeValue, after[key]):
			diff.Modified = append(diff.Modified, key)
		}
	}
	for key := range before {
		if _, exists := after[key]; !exists {
			diff.Removed = append(diff.Removed, key)
		}
	}
	slices.Sort(diff.Added)
	slices.Sort(diff.Removed)
	slices.Sort(diff.Modified)
	return diff
}
