terraform {
  required_providers {
    nobl9 = {
      source = "nobl9.com/nobl9/nobl9"
    }
  }
}

provider "nobl9" {
  organization = "nobl9-dev"
  project = "default"
  client_id = "0oacfsfocmdkzF7bO4x7"
  client_secret = "EPmTwXZZqOD6QpOi69Vd6wDfdhWhVbAYcg7WOq9X"
  ingest_url = "http://localhost/api"
  okta_org_url = "https://accounts.nobl9.dev"
  okta_auth_server = "ausdh5avfxFaHRKHN4x6"
}

resource "nobl9_project" "meta" {
  name = "test-terraform-langa"
}

resource "nobl9_direct_bigquery" "test-bigquery" {
  name                = "test-bigquery"
  project             = "test-terraform-langa"
  description         = "desc"
  source_of           = ["Metrics", "Services"]
  service_account_key = "secret"
  log_collection_enabled = true
}