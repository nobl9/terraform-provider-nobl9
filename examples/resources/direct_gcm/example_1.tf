resource "nobl9_direct_gcm" "test-gcm" {
  name                   = "test-gcm"
  project                = "terraform"
  description            = "desc"
  service_account_key    = "secret"
  log_collection_enabled = true
}
