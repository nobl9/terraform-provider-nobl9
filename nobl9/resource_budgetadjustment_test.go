package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9BudgetAdjustments(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"single-event", testBudgetAdjustmentSingleEvent},
		{"recurring-event", testBudgetAdjustmentRecurringEvent},
		{"recurring-event-multiple-slos", testBudgetAdjustmentRecurringEventMultipleSlo},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_budget_adjustment", manifest.KindBudgetAdjustment),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated(fmt.Sprintf("nobl9_budget_adjustment.%s", tc.name)),
					},
				},
			})
		})
	}
}

func testBudgetAdjustmentSingleEvent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_budget_adjustment" "%s" {
  name              = "%s"
  first_event_start = "2022-01-01T00:00:00Z"
  duration          = "1h"
  filters {
    slos {
      slo {
        name    = "cloudwatch-ratio-slo"
        project = "cloudwatch"
      }
    }
  }
}
`, name, name)
}

func testBudgetAdjustmentRecurringEvent(name string) string {
	return fmt.Sprintf(`resource "nobl9_budget_adjustment" "%s" {
  name              = "%s"
  first_event_start = "2022-01-01T00:00:00Z"
  duration          = "1h"
  rrule             = "FREQ=MONTHLY;BYMONTHDAY=1"
  filters {
    slos {
      slo {
        name    = "cloudwatch-ratio-slo"
        project = "cloudwatch"
      }
      slo {
        name    = "cloudwatch-ratio-slo2"
        project = "cloudwatch"
      }
    }
  }
}`, name, name)
}

func testBudgetAdjustmentRecurringEventMultipleSlo(name string) string {
	return fmt.Sprintf(`
resource "nobl9_budget_adjustment" "%s" {
  name              = "%s"
  display_name      = "Recurring budget adjustment for the first day of the month."
  first_event_start = "2022-01-01T00:00:00Z"
  description       = "Recurring budget adjustment for the first day of the month."
  duration          = "1h"
  rrule             = "FREQ=MONTHLY;BYMONTHDAY=1"
  filters {
    slos {
      slo {
        name    = "ratio-slo"
        project = "default"
      }
      slo {
        name    = "ratio-slo-timeslices"
        project = "default"
      }
    }
  }
}
`, name, name)
}
