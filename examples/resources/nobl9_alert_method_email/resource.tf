resource "nobl9_alert_method_email" "this" {
  name         = "my-email-alert"
  display_name = "My Email Alert"
  project      = "my-project"
  description  = "teams"
  to           = ["testUser@nobl9.com"]
  cc           = ["testUser@nobl9.com"]
  bcc          = ["testUser@nobl9.com"]
}

resource "nobl9_alert_method_email" "this" {
  name               = "my-email-alert-as-plain-text"
  display_name       = "My Email Alert as plain text"
  project            = "my-project"
  description        = "plain-text"
  to                 = ["testUser@nobl9.com"]
  send_as_plain_text = true
}
