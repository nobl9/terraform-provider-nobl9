resource "nobl9_direct_appdynamics" "test-appdynamics" {
  name          = "test-appdynamics"
  project       = "terraform"
  description   = "desc"
  source_of     = ["Metrics", "Services"]
  url           = "https://web.net"
  account_name  = "account name"
  client_secret = "secret"
  client_name   = "client name"
  log_collection_enabled = true
  release_channel = "stable"
}