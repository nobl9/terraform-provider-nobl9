resource "nobl9_direct_honeycomb" "test-honeycomb" {
  name                   = "test-honeycomb"
  project                = "terraform"
  description            = "desc"
  api_key                = "secret"
  log_collection_enabled = true
  release_channel        = "beta"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 7
    }
    max_duration {
      unit  = "Day"
      value = 7
    }
  }
  query_delay {
    unit  = "Minute"
    value = 6
  }
}
