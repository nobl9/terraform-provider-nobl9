resource "nobl9_direct_gcm" "test-gcm" {
  name                = "test-gcm"
  project             = "terraform"
  description         = "desc"
  source_of           = ["Metrics", "Services"]
  service_account_key = "secret"
}