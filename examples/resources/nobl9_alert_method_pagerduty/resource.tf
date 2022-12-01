resource "nobl9_alert_method_webhook" "this" {
  name         = "my-pagerduty-alert"
  display_name = "My PagerDuty Alert"
  project      = "Test Project"
  description     = "My PaderDuty Alert"
  integration_key = "84dfcdf19dad8f6c82b7e22afa024065"
}

