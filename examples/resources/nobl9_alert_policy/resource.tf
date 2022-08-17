resource "nobl9_project" "this" {
  display_name = "Test Terraform"
  name         = "test-terraform"
  description  = "An example terraform project"
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

  condition {
    measurement  = "timeToBurnBudget"
    value_string = "72h"
    lasts_for    = "30m"
  }

  alert_method {
    name = "my-alert-method"
  }
}

