resource "nobl9_alert_method_webhook" "this" {
  name         = "foo-alert"
  display_name = "Foo Alert"
  project      = "Foo Project"
  description     = "paderduty"
  integration_key = "84dfcdf19dad8f6c82b7e22afa024065"
}

