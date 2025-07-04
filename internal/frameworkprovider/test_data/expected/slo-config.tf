resource "nobl9_slo" "this" {
  name = "slo"
  display_name = "SLO"
  project = "default"
  annotations = {
    key = "value",
  }
  label {
    key = "team"
    values = [
      "green",
      "orange",
    ]
  }
  label {
    key = "env"
    values = [
      "prod",
    ]
  }
  label {
    key = "empty"
    values = [
      "",
    ]
  }
  description = "Example SLO"

  service = "service"
  budgeting_method = "Occurrences"
  alert_policies = [
    "alert-policy",
  ]

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
        appdynamics {
          application_name = "my_app"
          metric_path = "End User Experience|App|End User Response Time 95th percentile (ms)"
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
