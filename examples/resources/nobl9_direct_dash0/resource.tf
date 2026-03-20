resource "nobl9_direct_dash0" "test-dash0" {
  name                   = "test-dash0"
  project                = "terraform"
  description            = "desc"
  url                    = "https://api.eu-west-1.aws.dash0.com/api/prometheus"
  auth_token             = "secret"
  step                   = 60
  log_collection_enabled = true
  release_channel        = "beta"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 1
    }
    max_duration {
      unit  = "Day"
      value = 10
    }
  }
  query_delay {
    unit  = "Minute"
    value = 6
  }
}
