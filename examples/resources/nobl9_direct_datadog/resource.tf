resource "nobl9_direct_datadog" "test-datadog" {
  name            = "test-datadog"
  project         = "terraform"
  description     = "desc"
  source_of       = ["Metrics", "Services"]
  site            = "eu"
  api_key         = "secret"
  application_key = "secret"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 0
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
  }
}