resource "nobl9_direct_splunk" "test-splunk" {
  name                   = "test-splunk"
  project                = "terraform"
  description            = "desc"
  url                    = "https://web.net"
  access_token           = "secret"
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
  }
}