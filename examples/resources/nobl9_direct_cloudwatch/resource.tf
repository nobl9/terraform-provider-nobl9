resource "nobl9_direct_cloudwatch" "test-cloudwatch" {
  name              = "test-cloudwatch"
  project           = "terraform"
  description       = "desc"
  source_of         = ["Metrics", "Services"]
  access_key_id     = "secret"
  secret_access_key = "secret"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 1
    }
    max_duration {
      unit  = "Day"
      value = 10
    }
  }
}