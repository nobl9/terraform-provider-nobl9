resource "nobl9_slo" "this" {
  name = "slo"
  display_name = "SLO"
  project = "default"
  description = "Example SLO"

  service = "service"
  budgeting_method = "Occurrences"

  composite {
    target = 0.95
    burn_rate_condition {
      op = "gt"
      value = 2
    }
    burn_rate_condition {
      op = "lt"
      value = 1.5
    }
  }

  time_window {
    count = 10
    is_rolling = true
    unit = "Minute"
  }
}
