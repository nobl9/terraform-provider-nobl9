resource "nobl9_direct_bigquery" "test-bigquery" {
  name                   = "test-bigquery"
  project                = "terraform"
  description            = "desc"
  service_account_key    = "secret"
  log_collection_enabled = true
}