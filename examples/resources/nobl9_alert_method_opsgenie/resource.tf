resource "nobl9_alert_method_webhook" "this" {
  name         = "my-opsgenie-alert"
  display_name = "My Opsgenie Alert"
  project      = "Test Project"
  description = "My Opsgenie Alert"
  url         = "https://discord.com"
  auth		  = "GenieKey 12345"
}
