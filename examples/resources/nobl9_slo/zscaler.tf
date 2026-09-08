resource "nobl9_slo" "zscaler" {
  name             = "zscaler-application-score"
  project          = "default"
  service          = "example"
  budgeting_method = "Occurrences"

  indicator {
    name    = "zscaler"
    project = "default"
    kind    = "Direct"
  }

  objective {
    name   = "application-score"
    target = 0.99
    op     = "gt"
    value  = 65

    raw_metric {
      query {
        zscaler {
          app_id      = 12345
          location_id = 6789
          metric      = "score"
        }
      }
    }
  }

  time_window {
    unit       = "Day"
    count      = 30
    is_rolling = true
  }
}
