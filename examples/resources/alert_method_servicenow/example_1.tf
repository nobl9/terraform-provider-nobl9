resource "nobl9_alert_method_servicenow" "this" {
  name         = "my-servicenow-alert"
  display_name = "My ServiceNow Alert"
  project      = "Test Project"
  description    = "ServiceNow alert"
  username       = "nobl9User"
  password       = "secret"
  instance_name  = "my_snow_instance_name"
}
