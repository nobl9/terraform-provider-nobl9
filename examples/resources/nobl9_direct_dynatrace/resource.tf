resource "nobl9_direct_dynatrace" "test-dynatrace" {
  name            = "test-dynatrace"
  project         = "terraform"
  description     = "desc"
  source_of       = ["Metrics", "Services"]
  url             = "https://web.net"
  dynatrace_token = "secret"
  log_collection_enabled = true
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
}
