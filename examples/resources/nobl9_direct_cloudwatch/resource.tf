resource "nobl9_direct_cloudwatch" "test-cloudwatch" {
  name                   = "test-cloudwatch"
  project                = "terraform"
  description            = "desc"
  role_arn               = "secret"
  log_collection_enabled = true
  failover {
    primary_region  = "us-west-2"
    failover_region = "us-east-2"
  }
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