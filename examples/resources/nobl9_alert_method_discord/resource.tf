resource "nobl9_alert_method_discord" "this" {
  name         = "my-discord-alert"
  display_name = "My Discord alert"
  project      = "My Discord alert"
  description  = "My Discord alert method"
  url          = "https://discord.webhook.url"
}

