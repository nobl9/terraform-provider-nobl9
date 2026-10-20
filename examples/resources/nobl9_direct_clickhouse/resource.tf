resource "nobl9_direct_clickhouse" "test-clickhouse" {
  name                   = "test-clickhouse"
  project                = "terraform"
  description            = "desc"
  url                    = "https://clickhouse.example.com:8443"
  database               = "observability"
  username               = "readonly_slo"
  password               = "secret"
  log_collection_enabled = true
  release_channel        = "beta"
  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 15
    }
    max_duration {
      unit  = "Day"
      value = 30
    }
  }
  query_delay {
    unit  = "Second"
    value = 31
  }
}
