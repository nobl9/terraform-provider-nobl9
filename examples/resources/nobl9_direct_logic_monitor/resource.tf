resource "nobl9_direct_logic_monitor" "logic_monitor" {
  name = "logic-monitor"
  project = "logic_monitor"
  description = "desc"
  account = "account_name"
  account_id = "secret"
  access_key = "secret"
  log_collection_enabled = true
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}