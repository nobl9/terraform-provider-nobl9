resource "nobl9_alert_method_email" "this" {
  name         = "my-email-alert"
  display_name = "My Email Alert"
  project      = "My Project"
  description = "teams"
  to		  = [ "testUser@nobl9.com" ]
  cc		  = [ "testUser@nobl9.com" ]
  bcc		  = [ "testUser@nobl9.com" ]
}
