resource "nobl9_alert_method_webhook" "this" {
  name         = "foo-alert"
  display_name = "Foo Alert"
  project      = "Foo Project"
  description = "jira"
  url		  = "https://jira.com"
  username    = "nobleUser"
  apitoken    = "very sercret"
  project_key = "PC"
}

