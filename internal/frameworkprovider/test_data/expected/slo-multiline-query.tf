resource "nobl9_slo" "this" {
  name = "slo"
  display_name = "SLO"
  project = "default"
  description = "Example SLO"

  service = "service"
  budgeting_method = "Occurrences"

  indicator {
    name = "indicator"
    project = "default"
    kind = "Agent"
  }

  objective {
    display_name = "obj1"
    name = "tf-objective-1"
    op = "lt"
    target = 0.7
    value = 1
    raw_metric {
      query {
        prometheus {
          promql = "sum by (job) (\n  rate(http_request_duration_seconds_count{job=\"api\"}[5m]\n)"
        }
      }
    }
  }

  time_window {
    count = 10
    is_rolling = true
    unit = "Minute"
  }
}
