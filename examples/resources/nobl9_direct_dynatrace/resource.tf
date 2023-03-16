resource "nobl9_direct_dynatrace" "test-dynatrace" {
  name            = "test-dynatrace"
  project         = "terraform"
  description     = "desc"
  source_of       = ["Metrics", "Services"]
  url             = "https://web.net"
  dynatrace_token = "secret"
}