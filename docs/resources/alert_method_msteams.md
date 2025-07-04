---
page_title: "nobl9_alert_method_msteams Resource - terraform-provider-nobl9"
subcategory: "Alert Methods"
description: |-
  MS Teams Alert Method | Nobl9 Documentation https://docs.nobl9.com/alerting/alert-methods/ms-teams
---

# nobl9_alert_method_msteams (Resource)

The **MS Teams Alert Method** enables sending alerts through MS Teams to notify Nobl9 users whenever an incident is triggered.

For more details, refer to [MS Teams Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/ms-teams).

## Example Usage

Here's an example of MS Teams Terraform resource configuration:

```terraform
resource "nobl9_alert_method_msteams" "this" {
  name         = "ms-teams-alert"
  display_name = "MS Teams Alert"
  project      = "Test Project"
  description  = "My MS Teams alerts"
  url          = "https://teams.com"
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
- `url` (String, Sensitive) Either [MS Teams Workflow URL](https://docs.nobl9.com/alerting/alert-methods/ms-teams/#2) or deprecated [Webhook URL](https://docs.nobl9.com/alerting/alert-methods/ms-teams/#webhook-url-).

### Read-Only

- `id` (String) The ID of this resource.

## Useful Links

[MS Teams alerts configuration | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/ms-teams)

[Retirement of Office 365 connectors within Microsoft Teams | MS Teams Documentation](https://devblogs.microsoft.com/microsoft365dev/retirement-of-office-365-connectors-within-microsoft-teams/)

[MS Teams webhooks | MS Teams Documentation](https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook)