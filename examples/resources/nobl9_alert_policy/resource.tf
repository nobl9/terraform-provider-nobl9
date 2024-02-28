resource "nobl9_project" "this" {
  display_name = "My Project"
  name         = "my-project"
  description  = "An example N9 Terraform project"
}

resource "nobl9_service" "this" {
  name         = "${nobl9_project.this.name}-front-page"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Front Page"
  description  = "Front page service"
}

resource "nobl9_alert_policy" "this" {
  name         = "${nobl9_project.this.name}-front-page-latency"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Front Page Latency"
  severity     = "High"
  description  = "Alert when page latency is > 2000 and error budget would be exhausted"
  cool_down    = "5m"

  condition {
    measurement  = "timeToBurnBudget"
    value_string = "72h"
    lasts_for    = "30m"
  }

  alert_method {
    name = "my-alert-method"
  }
}

resource "nobl9_alert_policy" "this" {
  name         = "${nobl9_project.this.name}-slow-burn"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Slow Burn (1x12h and 2x15min)"
  severity     = "Low"
  description  = "The budget is slowly exhausting and not recovering."
  cool_down    = "5m"

  condition {
    measurement  = "averageBurnRate"
    value = "1"
    alerting_window    = "12h"
  }

  condition {
    measurement  = "averageBurnRate"
    value = "2"
    alerting_window    = "15m"
  }
}

resource "nobl9_alert_policy" "this" {
  name         = "${nobl9_project.this.name}-fast-burn"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Fast Burn (20x5min)"
  severity     = "High"
  description  = "There’s been a significant spike in burn rate over a brief period."
  cool_down    = "5m"

  condition {
    measurement  = "averageBurnRate"
    value = "20"
    alerting_window    = "5m"
  }
}
