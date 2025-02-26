resource "nobl9_direct_appdynamics" "test-appdynamics" {
  name          = "test-appdynamics"
  project       = "terraform"
  description   = "desc"
  url           = "https://web.net"
  account_name  = "account name"
  client_secret = "secret"
  client_name   = "client name"
  log_collection_enabled = true
  release_channel = "stable"
  historical_data_retrieval {
      default_duration  {
        unit  = "Day"
        value = 0
      }
      max_duration {
        unit  = "Day"
        value = 30
      }
    }
}
