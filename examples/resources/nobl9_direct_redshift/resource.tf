resource "nobl9_direct_redshift" "test-redshift" {
  name              = "test-redshift"
  project           = "terraform"
  description       = "desc"
  source_of         = ["Metrics", "Services"]
  secret_arn        = "aws:arn"
  access_key_id     = "secret"
  secret_access_key = "secret"
}