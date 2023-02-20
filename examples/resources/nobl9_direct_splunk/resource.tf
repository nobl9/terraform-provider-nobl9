resource "nobl9_direct_splunk" "test-splunk" {
  name         = "test-splunk"
  project      = "terraform"
  description  = "desc"
  source_of    = ["Metrics", "Services"]
  url          = "https://web.net"
  access_token = "secret"
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