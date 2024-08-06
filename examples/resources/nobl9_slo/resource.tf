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

resource "nobl9_slo" "slo_1" {
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
    url          = "https://www.nobl9.com/"
    display_name = "Nobl9 Reliability Center"
  }

  attachment {
    url          = "https://duckduckgo.com/"
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
    primary      = true

    raw_metric {
      query {
        prometheus {
          promql = <<EOT
          latency_west_c7{code="ALL",instance="localhost:3000",job="prometheus",service="glob_account"}
          EOT
        }
      }
    }
  }

  indicator {
    name = "test-n9-terraform-prom-agent"
  }
}

resource "nobl9_slo" "slo_2" {
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
    primary      = false

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
    name = "test-n9-terraform-prom-agent"
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

# Composite 2.0 example.
resource "nobl9_slo" "composite_slo" {
  name             = "${nobl9_project.this.name}-composite"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  # List the names of component SLOs your composite 2.0 must include
  depends_on = [nobl9_slo.slo_1, nobl9_slo.slo_2]

  time_window {
    unit       = "Day"
    count      = 3
    is_rolling = true
  }

  objective {
    display_name = "OK"
    name         = "tf-objective-1"
    target       = 0.8
    value        = 1
    composite {
      max_delay = "45m"
      components {
        objectives {
          composite_objective {
            project      = nobl9_slo.slo_1.project
            slo          = nobl9_slo.slo_1.name
            objective    = "tf-objective-1"
            weight       = 0.8
            when_delayed = "CountAsGood"
          }
          composite_objective {
            project      = nobl9_slo.slo_2.project
            slo          = nobl9_slo.slo_2.name
            objective    = "tf-objective-1"
            weight       = 1.5
            when_delayed = "Ignore"
          }
        }
      }
    }
  }
}
