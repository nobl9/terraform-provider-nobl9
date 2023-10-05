resource "nobl9_direct_cloudwatch" "test-cloudwatch" {
  name                   = "test-cloudwatch"
  project                = "terraform"
  description            = "desc"
  source_of              = ["Metrics", "Services"]
  role_arn               = "secret"
  log_collection_enabled = true
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 0
    }
    max_duration {
      unit  = "Day"
      value = 15
    }
  }
}