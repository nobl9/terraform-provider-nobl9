resource "nobl9_direct_logic_monitor" "logic_monitor" {
  name                   = "logic-monitor"
  project                = "logic-monitor"
  description            = "desc"
  account                = "account_name"
  account_id             = "secret"
  access_key             = "secret"
  log_collection_enabled = true
  release_channel        = "beta"
  query_delay {
    unit  = "Minute"
    value = 6
  }
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