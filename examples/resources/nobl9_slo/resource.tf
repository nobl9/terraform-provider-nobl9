resource "nobl9_project" "this" {
  display_name = "Test N9 Terraform"
  name         = "test-n9-terraform"
  description  = "An example N9 Terraform project"
}

resource "nobl9_service" "this" {
  name         = "my-front-page"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Front Page"
  description  = "Front page service"
}

resource "nobl9_slo" "this" {
  name             = "${nobl9_project.this.name}-latency"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  label {
    key    = "env"
    values = ["dev", "prod"]
  }

  label {
    key    = "team"
    values = ["red"]
  }

  attachment {
    utl          = "https://www.nobl9.com/"
    display_name = "SLO provider"
  }

  attachment {
    utl          = "https://duckduckgo.com/"
    display_name = "Nice search engine"
  }

  alert_policies = [
    "foo-front-page-latency"
  ]

  time_window {
    unit       = "Day"
    count      = 30
    is_rolling = true
  }

  objective {
    name         = "tf-objective-1"
    target       = 0.99
    display_name = "OK"
    value        = 2000
    op           = "gte"
    
    raw_metric {
      query {
        prometheus {
          promql = <<EOT
          latency_west_c7{code="ALL",instance="localhost:3000",job="prometheus",service="globacount"}
          EOT
        }
      }
    }
  }

  indicator {
    name = "test-terraform-prom-agent"
  }
}

resource "nobl9_slo" "this" {
  name             = "${nobl9_project.this.name}-ratio"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  time_window {
    unit       = "Day"
    count      = 30
    is_rolling = true
  }

  objective {
    name         = "tf-objective-1"
    target       = 0.99
    display_name = "OK"
    value        = 1

    count_metrics {
      incremental = true
      good {
        prometheus {
          promql = "1.0"
        }
      }
      total {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  indicator {
    name = "test-terraform-prom-agent"
  }

  anomaly_config {
    no_data {
      alert_method {
        name = "foo-method-method"
        project = "default"
      }

      alert_method {
        name = "bar-alert-method"
        project = "default"
      }
    }
  }
}