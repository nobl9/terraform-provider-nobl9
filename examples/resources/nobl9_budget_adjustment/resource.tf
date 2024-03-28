resource "nobl9_budget_adjustment" "single-budget-adjustment-event" {
  name              = "single-budget-adjustment-event"
  display_name      = "Single Budget Adjustment Event"
  first_event_start = "2022-01-01T00:00:00Z"
  duration          = "1h"
  description       = "Single budget adjustment event"
  filters {
    slos {
      slo {
        name    = "my-slo"
        project = "default"
      }
    }
  }
}

resource "nobl9_budget_adjustment" "recurring-budget-adjustment-event" {
  name              = "recurring-budget-adjustment-event"
  display_name      = "Recurring Budget Adjustment Event"
  first_event_start = "2022-01-01T16:00:00Z"
  duration          = "1h"
  rrule             = "FREQ=WEEKLY"
  description       = "Recurring budget adjustment event"
  filters {
    slos {
      slo {
        name    = "my-slo"
        project = "default"
      }
    }
  }
}