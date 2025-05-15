terraform {
  required_providers {
    nobl9 = {
      source = "nobl9.com/nobl9/nobl9"
      version = "0.37.2"
    }
  }
}

provider "nobl9" {
  client_id = "0oafvkkv8ibxZc1a14x7"
  client_secret = "UALHsGCg3eT6TsWujr5UEZi-OEL2dnLXP8VdcJMuqvWDEOapufG9_6VIlmm55xlt"
  okta_org_url     = "https://accounts.nobl9.dev"
  okta_auth_server = "ausdh506kj9JJVw3g4x6"
}

resource "nobl9_project" "this" {
  display_name = "Adam Test Terraform"
  name         = "adam-test-terraform"
  description  = "An example N9 Terraform project"
}

resource "nobl9_service" "this" {
  name         = "my-front-page"
  project      = nobl9_project.this.name
  display_name = "svc-${nobl9_project.this.display_name}"
  description  = "Front page service"
}

resource "nobl9_agent" "this" {
  name            = "${nobl9_project.this.name}-prom-agent"
  project         = nobl9_project.this.name
  agent_type      = "prometheus"
  release_channel = "stable"
  prometheus_config {
    url = "http://web.net"
  }
}

resource "nobl9_slo" "slo_1" {
  depends_on = [nobl9_agent.this]
  name             = "slo-1-${nobl9_project.this.name}"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  anomaly_config {
    no_data {
      alert_method {
        name = nobl9_alert_method_email.e1.name
        project = nobl9_project.this.name
      }
      alert_method {
        name = nobl9_alert_method_email.e2.name
        project = nobl9_project.this.name
      }
    }
  }

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
    name = "${nobl9_project.this.name}-prom-agent"
    project = nobl9_project.this.name
  }
}

resource "nobl9_slo" "slo_2" {
  depends_on = [nobl9_agent.this]
  name             = "slo-2-${nobl9_project.this.name}"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  anomaly_config {
    no_data {
      alert_method {
        name = nobl9_alert_method_email.e1.name
        project = nobl9_project.this.name
      }
      alert_method {
        name = nobl9_alert_method_email.e2.name
        project = nobl9_project.this.name
      }
      alert_after = "30m"
    }
  }

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
    name = "${nobl9_project.this.name}-prom-agent"
    project = nobl9_project.this.name
  }
}
/*
resource "nobl9_slo" "slo_3" {
  depends_on = [nobl9_agent.this]
  name             = "slo-3-${nobl9_project.this.name}"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  anomaly_config {
    no_data {
      alert_method {
        name = nobl9_alert_method_email.e1.name
        project = nobl9_project.this.name
      }
      alert_method {
        name = nobl9_alert_method_email.e2.name
        project = nobl9_project.this.name
      }
      alert_after = "2m" // Incorrect value (< 5m)
    }
  }

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
    name = "${nobl9_project.this.name}-prom-agent"
    project = nobl9_project.this.name
  }
}

resource "nobl9_slo" "slo_4" {
  depends_on = [nobl9_agent.this]
  name             = "slo-4-${nobl9_project.this.name}"
  service          = nobl9_service.this.name
  budgeting_method = "Occurrences"
  project          = nobl9_project.this.name

  anomaly_config {
    no_data {
      alert_method {
        name = nobl9_alert_method_email.e1.name
        project = nobl9_project.this.name
      }
      alert_method {
        name = nobl9_alert_method_email.e2.name
        project = nobl9_project.this.name
      }
      alert_after = "10m30s" // Incorrect value, not full minutes
    }
  }

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
    name = "${nobl9_project.this.name}-prom-agent"
    project = nobl9_project.this.name
  }
}
*/

resource "nobl9_alert_method_email" "e1" {
  name         = "my-email-alert-1"
  display_name = "My Email Alert 1"
  project      = nobl9_project.this.name
  description = "email"
  to		  = [ "testUser@nobl9.com" ]
}

resource "nobl9_alert_method_email" "e2" {
  name         = "my-email-alert-2"
  display_name = "My Email Alert 2"
  project      = nobl9_project.this.name
  description = "email"
  to		  = [ "test@nobl9.com" ]
}
