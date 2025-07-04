---
page_title: "nobl9_alert_method_email Resource - terraform-provider-nobl9"
subcategory: "Alert Methods"
description: |-
  Email Alert Method | Nobl9 Documentation https://docs.nobl9.com/alerting/alert-methods/email-alert
---

# nobl9_alert_method_email (Resource)

The **Email Alert Method** enables sending automated and customized alert messages to up to 30 different inboxes per alert to notify Nobl9 users whenever an incident is triggered.

For more details, refer to [Email Alert Method | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/email-alert).

## Example Usage

Here's an example of Email Terraform resource configuration:

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `to` (List of String) Recipients. The maximum number of recipients is 10.

### Optional

- `bcc` (List of String) Blind carbon copy recipients. The maximum number of recipients is 10.
- `cc` (List of String) Carbon copy recipients. The maximum number of recipients is 10.
- `description` (String) Optional description of the resource. Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.
- `display_name` (String) User-friendly display name of the resource.
- `send_as_plain_text` (Boolean) Send email as plain text.

### Read-Only

- `id` (String) The ID of this resource.

## Useful links

[Email alerts configuration | Nobl9 Documentation](https://docs.nobl9.com/alerting/alert-methods/email-alert)