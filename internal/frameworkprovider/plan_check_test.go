package frameworkprovider

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
)

var _ plancheck.PlanCheck = expectChangesInPlanChecker{}

// expectChangesInResourcePlan creates a [plancheck.PlanCheck]
// that compares expected [planDiff] with actual plan for a single resource.
func expectChangesInResourcePlan(expected planDiff) expectChangesInPlanChecker {
	return expectChangesInPlanChecker{resourcesPlan: map[string]planDiff{"": expected}}
}

// expectChangesInResourcesPlan creates a [plancheck.PlanCheck]
// that compares expected [planDiff] with actual plan for multiple resources.
func expectChangesInResourcesPlan(expected map[string]planDiff) expectChangesInPlanChecker {
	return expectChangesInPlanChecker{resourcesPlan: expected}
}

type expectChangesInPlanChecker struct {
	resourcesPlan map[string]planDiff
}

// CheckPlan implements the [plancheck.PlanCheck] interface.
func (e expectChangesInPlanChecker) CheckPlan(
	_ context.Context,
	req plancheck.CheckPlanRequest,
	resp *plancheck.CheckPlanResponse,
) {
	if len(req.Plan.ResourceChanges) != len(e.resourcesPlan) {
		resp.Error = fmt.Errorf("expected exactly %d resources to change, but got %d",
			len(e.resourcesPlan), len(req.Plan.ResourceChanges))
		return
	}

	for _, resource := range req.Plan.ResourceChanges {
		var expected planDiff
		switch len(e.resourcesPlan) {
		case 1:
			expected = slices.Collect(maps.Values(e.resourcesPlan))[0]
		default:
			var ok bool
			expected, ok = e.resourcesPlan[resource.Address]
			if !ok {
				resp.Error = fmt.Errorf("unexpected resource '%s' in the plan, expected one of %v",
					resource.Address, strings.Join(slices.Collect(maps.Keys(e.resourcesPlan)), ", "))
				return
			}
		}

		change := resource.Change
		before, _ := change.Before.(map[string]any)
		after, _ := change.After.(map[string]any)
		diff := calculatePlanDiff(before, after)

		t := &testingRecorder{}
		assert.ElementsMatchf(
			t,
			expected.Added,
			diff.Added,
			"unexpected added attributes in the %s plan",
			resource.Address,
		)
		if t.err != nil {
			resp.Error = t.err
			return
		}
		assert.ElementsMatch(
			t,
			expected.Removed,
			diff.Removed,
			"unexpected removed attributes in the %s plan",
			resource.Address,
		)
		if t.err != nil {
			resp.Error = t.err
			return
		}
		assert.ElementsMatch(
			t,
			expected.Modified,
			diff.Modified,
			"unexpected modified attributes in the %s plan",
			resource.Address,
		)
		if t.err != nil {
			resp.Error = t.err
			return
		}
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
	return diff
}

// testingRecorder is a mock implementation of the [assert.TestingT] that captures errors.
type testingRecorder struct{ err error }

func (m *testingRecorder) Errorf(format string, args ...interface{}) {
	m.err = fmt.Errorf(format, args...)
}
