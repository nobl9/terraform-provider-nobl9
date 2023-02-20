resource "nobl9_direct_bigquery" "test-bigquery" {
  name                = "test-bigquery"
  project             = "terraform"
  description         = "desc"
  source_of           = ["Metrics", "Services"]
  service_account_key = "secret"
}