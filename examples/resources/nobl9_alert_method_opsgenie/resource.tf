resource "nobl9_alert_method_webhook" "this" {
  name         = "foo-alert"
  display_name = "Foo Alert"
  project      = "Foo Project"
  description = "opsgenie"
  url         = "https://discord.com"
  auth		  = "GenieKey 12345"
}

