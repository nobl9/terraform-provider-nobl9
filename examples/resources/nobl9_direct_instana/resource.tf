resource "nobl9_direct_instana" "test-instana" {
  name        = "test-instana"
  project     = "terraform"
  description = "desc"
  url         = "https://web.net"
  api_token   = "secret"
  log_collection_enabled = true
}