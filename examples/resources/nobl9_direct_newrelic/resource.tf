resource "nobl9_direct_newrelic" "test-newrelic" {
  name               = "test-newrelic"
  project            = "terraform"
  description        = "desc"
  source_of          = ["Metrics", "Services"]
  account_id         = "1234"
  insights_query_key = "secret"
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