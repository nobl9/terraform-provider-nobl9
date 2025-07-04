resource "nobl9_direct_newrelic" "test-newrelic" {
  name                   = "test-newrelic"
  project                = "terraform"
  description            = "desc"
  account_id             = "1234"
  insights_query_key     = "secret"
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