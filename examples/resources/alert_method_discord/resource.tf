resource "nobl9_alert_method_discord" "this" {
  name         = "foo-alert"
  display_name = "Foo Alert"
  project      = "Foo Project"
  description = "discord"
  url         = "https://discord.webhook.url"
}

