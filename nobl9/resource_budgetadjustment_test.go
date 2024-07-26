package nobl9

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9BudgetAdjustments(t *testing.T) {
	futureTime := time.Now().AddDate(0, 0, 1).UTC().Format(time.RFC3339)
	cases := []struct {
		name            string
		configFunc      func(string, string) string
		firstEventStart string
	}{
		{"single-event", testBudgetAdjustmentSingleEvent, futureTime},
		{"recurring-event", testBudgetAdjustmentRecurringEvent, futureTime},
		{"recurring-event-multiple-slos", testBudgetAdjustmentRecurringEventMultipleSlo, futureTime},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_budget_adjustment", manifest.KindBudgetAdjustment),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name, tc.firstEventStart),
						Check:  CheckObjectCreated(fmt.Sprintf("nobl9_budget_adjustment.%s", tc.name)),
					},
				},
			})
		})
	}
}

func testBudgetAdjustmentSingleEvent(name, futureTime string) string {
	const sloName = "test-slo"
	return testPrometheusSLOFull(sloName) +
		fmt.Sprintf(`
resource "nobl9_budget_adjustment" "%s" {
  name              = "%s"
  first_event_start = "%s"
  duration          = "1h"
  filters {
    slos {
      slo {
        name    = nobl9_slo.%s.name
        project = "%s"
      }
    }
  }
}
`, name, name, futureTime, sloName, testProject)
}

func testBudgetAdjustmentRecurringEvent(name, futureTime string) string {
	const sloName = "test-slo2"
	const sloName2 = "test-slo3"

	return testPrometheusSLOFull(sloName) + testPrometheusSLOFull(sloName2) +
		fmt.Sprintf(`
resource "nobl9_budget_adjustment" "%s" {
  name              = "%s"
  first_event_start = "%s"
  duration          = "1h"
  rrule             = "FREQ=MONTHLY;BYMONTHDAY=1"
  filters {
    slos {
      slo {
        name    = nobl9_slo.%s.name
        project = "%s"
      }
      slo {
        name    = nobl9_slo.%s.name
        project = "%s"
      }
    }
  }
}`, name, name, futureTime, sloName, testProject, sloName2, testProject)
}

func testBudgetAdjustmentRecurringEventMultipleSlo(name, futureTime string) string {
	const sloName = "test-slo4"
	const sloName2 = "test-slo5"

	return testPrometheusSLOFull(sloName) + testPrometheusSLOFull(sloName2) +
		fmt.Sprintf(`
resource "nobl9_budget_adjustment" "%s" {
  name              = "%s"
  display_name      = "Recurring budget adjustment for the first day of the month."
  first_event_start = "%s"
  description       = "Recurring budget adjustment for the first day of the month."
  duration          = "1h"
  rrule             = "FREQ=MONTHLY;BYMONTHDAY=1"
  filters {
    slos {
      slo {
        name    = nobl9_slo.%s.name
        project = "%s"
      }
      slo {
        name    = nobl9_slo.%s.name
        project = "%s"
      }
    }
  }
}
`, name, name, futureTime, sloName, testProject, sloName2, testProject)
}
