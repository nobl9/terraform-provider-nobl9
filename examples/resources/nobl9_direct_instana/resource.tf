resource "nobl9_direct_instana" "test-instana" {
  name        = "test-instana"
  project     = "terraform"
  description = "desc"
  source_of   = ["Metrics", "Services"]
  url         = "https://web.net"
  api_token   = "secret"
}