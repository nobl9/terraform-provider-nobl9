package nobl9

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_budget_adjustment", manifest.KindBudgetAdjustment),
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

func testPrometheusSLOFull(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  objective {
    display_name = "obj2"
    name         = "tf-objective-2"
    target       = 0.5
    value        = 10
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    calendar {
      start_time = "2020-03-09 00:00:00"
      time_zone = "Europe/Warsaw"
    }
    count      = 7
    unit       = "Day"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testService(name string) string {
	return fmt.Sprintf(`
resource "nobl9_service" "%s" {
  name              = "%s"
  display_name = "%s"
  project             = "%s"
  description       = "Test of service"

  label {
   key = "env"
   values = ["green","sapphire"]
  }

  label {
   key = "dev"
   values = ["dev", "staging", "prod"]
  }

  annotations = {
   env = "development"
   name = "example annotation"
  }
}
`, name, name, name, testProject)
}
