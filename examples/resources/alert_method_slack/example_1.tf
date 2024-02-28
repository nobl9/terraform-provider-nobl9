resource "nobl9_alert_method_slack" "this" {
  name         = "my-slack-alert"
  display_name = "My Slack Alert"
  project      = "Test Project"
  description = "slack"
  url         = "https://slack.com"
}
