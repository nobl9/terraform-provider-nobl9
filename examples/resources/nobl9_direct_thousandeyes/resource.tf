resource "nobl9_direct_thousandeyes" "test-thousandeyes" {
  name               = "test-thousandeyes"
  project            = "terraform"
  description        = "desc"
  source_of          = ["Metrics", "Services"]
  oauth_bearer_token = "secret"
}