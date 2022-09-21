resource "nobl9_alert_method_email" "this" {
  name         = "foo-alert"
  display_name = "Foo Alert"
  project      = "Foo Project"
  description = "teams"
  to		  = [ "testUser@nobl9.com" ]
  cc		  = [ "testUser@nobl9.com" ]
  bcc		  = [ "testUser@nobl9.com" ]
  subject     = "Test email please ignore"
  body        = "This is just a test email"
}

