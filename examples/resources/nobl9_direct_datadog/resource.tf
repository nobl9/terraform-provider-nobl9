resource "nobl9_direct_datadog" "test-datadog" {
  name                   = "test-datadog"
  project                = "terraform"
  description            = "desc"
  site                   = "datadoghq.eu"
  api_key                = "secret"
  application_key        = "secret"
  log_collection_enabled = true
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 0
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
    triggered_by_slo_creation {
      unit  = "Day"
      value = 10
    }
    triggered_by_slo_edit {
      unit  = "Day"
      value = 10
    }
  }
}