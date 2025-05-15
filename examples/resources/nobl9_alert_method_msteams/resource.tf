resource "nobl9_alert_method_msteams" "this" {
  name         = "ms-teams-alert"
  display_name = "MS Teams Alert"
  project      = "Test Project"
  description  = "My MS Teams alerts"
  url          = "https://teams.com"
}

