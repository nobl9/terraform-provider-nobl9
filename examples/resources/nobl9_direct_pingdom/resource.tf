resource "nobl9_direct_pingdom" "test-pingdom" {
  name                   = "test-pingdom"
  project                = "terraform"
  description            = "desc"
  source_of              = ["Metrics", "Services"]
  api_token              = "secret"
  log_collection_enabled = true
}
