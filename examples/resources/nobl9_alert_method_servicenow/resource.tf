resource "nobl9_alert_method_webhook" "this" {
  name         = "foo-alert"
  display_name = "Foo Alert"
  project      = "Foo Project"
  description    = "servicenow"
  username       = "nobleUser"
  password       = "very sercret"
  instance_name  = "name"
}

