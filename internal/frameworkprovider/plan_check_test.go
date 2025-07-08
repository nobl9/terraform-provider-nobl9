package frameworkprovider

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
)

var _ plancheck.PlanCheck = expectChangesInPlanChecker{}

// expectChangesInPlan creates a [plancheck.PlanCheck] that compares expected [planDiff] with actual plan.
func expectChangesInPlan(expected planDiff) expectChangesInPlanChecker {
	return expectChangesInPlanChecker{expected: expected}
}

type expectChangesInPlanChecker struct {
	expected planDiff
}

// CheckPlan implements the [plancheck.PlanCheck] interface.
func (e expectChangesInPlanChecker) CheckPlan(
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

	t := &testingRecorder{}
	assert.ElementsMatch(t, e.expected.Added, diff.Added, "unexpected added attributes in the plan")
	if t.err != nil {
		resp.Error = t.err
		return
	}
	assert.ElementsMatch(t, e.expected.Removed, diff.Removed, "unexpected removed attributes in the plan")
	if t.err != nil {
		resp.Error = t.err
		return
	}
	assert.ElementsMatch(t, e.expected.Modified, diff.Modified, "unexpected modified attributes in the plan")
	if t.err != nil {
		resp.Error = t.err
		return
	}
}

// planDiff represents the difference between 'before' and 'after' plans,
// listing attribute and block names which changed in root level.
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

// testingRecorder is a mock implementation of the [assert.TestingT] that captures errors.
type testingRecorder struct{ err error }

func (m *testingRecorder) Errorf(format string, args ...interface{}) {
	m.err = fmt.Errorf(format, args...)
}
