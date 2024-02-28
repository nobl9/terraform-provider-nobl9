resource "nobl9_direct_thousandeyes" "test-thousandeyes" {
  name                   = "test-thousandeyes"
  project                = "terraform"
  description            = "desc"
  oauth_bearer_token     = "secret"
  log_collection_enabled = true
}
