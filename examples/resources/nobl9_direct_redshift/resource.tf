resource "nobl9_direct_redshift" "test-redshift" {
  name                   = "test-redshift"
  project                = "terraform"
  description            = "desc"
  source_of              = ["Metrics", "Services"]
  secret_arn             = "aws:arn"
  role_arn               = "secret"
  log_collection_enabled = true
}