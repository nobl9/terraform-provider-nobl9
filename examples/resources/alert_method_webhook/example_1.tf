resource "nobl9_alert_method_webhook" "this" {
  name         = "my-webhook-alert"
  display_name = "Webhook Alert"
  project      = "Test Project"
  url          = "https://webhook.com/12345"

  template_fields = [
    "alert_policy_name",
    "alert_policy_description",
    "alert_policy_conditions_text",
    "project_name",
    "service_name",
    "slo_name",
    "organization",
    "objective_name",
    "severity",
    "timestamp",
    "slo_details_link",
    "alert_policy_conditions[]",
    "iso_timestamp",
    "slo_labels_text",
    "service_labels_text",
    "alert_policy_labels_text",
  ]
}
