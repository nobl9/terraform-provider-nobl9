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
        sumologic {
          type = "metrics"
          quantization = "15s"
          rollup = "Avg"
          queries {
            row_id = "A"
            query = "metric=cpu_idle"
          }
          queries {
            row_id = "B"
            query = "#A + 1"
          }
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
