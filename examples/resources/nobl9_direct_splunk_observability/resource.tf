resource "nobl9_direct_splunk_observability" "test-splunk-observability" {
  name            = "test-splunk-observability"
  project         = "terraform"
  description     = "desc"
  realm           = "eu"
  access_token    = "secret"
  release_channel = "beta"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 15
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
  }
  query_delay {
    unit  = "Minute"
    value = 6
  }
}
