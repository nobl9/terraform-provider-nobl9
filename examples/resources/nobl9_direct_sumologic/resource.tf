resource "nobl9_direct_sumologic" "test-sumologic" {
  name        = "test-sumologic"
  project     = "terraform"
  description = "desc"
  url         = "http://web.net"
  access_id   = "secret"
  access_key  = "secret"
  log_collection_enabled = true
}
