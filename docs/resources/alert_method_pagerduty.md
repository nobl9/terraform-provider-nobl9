---
page_title: "nobl9_alert_method_pagerduty Resource - terraform-provider-nobl9"
subcategory: "Alert Methods"
description: |-
  PagerDuty Alert Method | Nobl9 Documentation https://docs.nobl9.com/alerting/alert-methods/pagerduty
---

# nobl9_alert_method_pagerduty (Resource)

The **PagerDuty Alert Method** enables triggering alerts through PagerDuty to notify Nobl9 users whenever an incident is triggered.

For more details, refer to [PagerDuty Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/pagerduty).

## Example Usage

Here's an example of PagerDuty Terraform resource configuration:

```terraform
resource "nobl9_alert_method_pagerduty" "this" {
  name            = "my-pagerduty-alert"
  display_name    = "My PagerDuty Alert"
  project         = "Test Project"
  description     = "My PagerDuty Alert"
  integration_key = "84dfcdf19dad8f6c82b7e22afa024065"
  send_resolution {
    message = "Alert is now resolved"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).

### Optional

- `description` (String) Optional description of the resource. Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.
- `display_name` (String) User-friendly display name of the resource.
- `integration_key` (String, Sensitive) PagerDuty Integration Key. For more details, check [Services and integrations](https://support.pagerduty.com/docs/services-and-integrations).
- `send_resolution` (Block Set, Max: 1) Sends a notification after the cooldown period is over. (see [below for nested schema](#nestedblock--send_resolution))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--send_resolution"></a>
### Nested Schema for `send_resolution`

Optional:

- `message` (String) A message that will be attached to your 'all clear' notification.

## Useful Links

[PagerDuty alerts configuration | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/pagerduty/)