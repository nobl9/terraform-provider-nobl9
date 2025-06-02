resource "nobl9_direct_azure_monitor" "test-azure-monitor" {
  name                   = "test-azure-monitor"
  project                = "terraform"
  description            = "desc"
  tenant_id              = "45e4c1ed-5b6b-4555-a693-6ab7f15f3d6e"
  client_id              = "secret"
  client_secret          = "secret"
  log_collection_enabled = true
  release_channel        = "beta"
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
  query_delay {
    unit  = "Minute"
    value = 6
  }
}