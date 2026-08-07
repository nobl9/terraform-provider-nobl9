resource "nobl9_direct_elasticsearch" "example" {
  name            = "elasticsearch-direct"
  project         = "default"
  url             = "https://example.aws.found.io"
  api_key         = "encoded-api-key"
  release_channel = "beta"

  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 1
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
  }

  query_delay {
    unit  = "Minute"
    value = 1
  }
}
