resource "nobl9_direct_splunk_observability" "test-splunkobservability" {
  name         = "test-splunkobservability"
  project      = "terraform"
  description  = "desc"
  source_of    = ["Metrics", "Services"]
  realm        = "eu"
  access_token = "secret"
}