resource "nobl9_direct_lightstep" "test-lightstep" {
  name                   = "test-lightstep"
  project                = "terraform"
  description            = "desc"
  lightstep_organization = "acme"
  lightstep_project      = "project1"
  app_token              = "secret"
  log_collection_enabled = true
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 0
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
  }
}