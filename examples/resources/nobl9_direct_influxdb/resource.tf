resource "nobl9_direct_influxdb" "test-influxdb" {
  name            = "test-influxdb"
  project         = "terraform"
  description     = "desc"
  source_of       = ["Metrics", "Services"]
  url             = "https://web.net"
  api_token       = "secret"
  organization_id = "secret"
}
