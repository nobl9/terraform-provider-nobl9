resource "nobl9_alert_method_jira" "this" {
  name         = "my-jira-alert"
  display_name = "My Jira Alert"
  project      = "My Jira Project"
  description = "My jira alert"
  url		  = "https://jira.com"
  username    = "nobl9User"
  apitoken    = "secret_api_token"
  project_key = "PC"
}
