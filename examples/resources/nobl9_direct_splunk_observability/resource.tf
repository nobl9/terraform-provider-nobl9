resource "nobl9_direct_splunk_observability" "test-splunk-observability" {
  name         = "test-splunk-observability"
  project      = "terraform"
  description  = "desc"
  realm        = "eu"
  access_token = "secret"
}
