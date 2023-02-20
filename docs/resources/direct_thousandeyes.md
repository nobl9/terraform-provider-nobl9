---
page_title: "nobl9_direct_thousandeyes Resource - terraform-provider-nobl9"
description: |-
  ThousandEyes Direct | Nobl9 Documentation https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-direct.
---

# nobl9_direct_thousandeyes (Resource)

ThousandEyes monitors the performance of both local and wide-area networks. ThousandEyes combines Internet and WAN visibility, browser synthetics, end-user monitoring, and Internet Insights to deliver a holistic view of your hybrid digital ecosystem – across cloud, SaaS, and the Internet. It's a SaaS-based tool that helps troubleshoot application delivery and maps Internet performance. Nobl9 connects with ThousandEyes to collect SLI measurements and compare them to SLO targets.

For more information, refer to [ThousandEyes Direct | Nobl9 Documentation](https://docs.nobl9.com/Sources/thousandeyes#thousandeyes-direct).

## Example Usage

```terraform
resource "nobl9_direct_thousandeyes" "test-thousandeyes" {
  name               = "test-thousandeyes"
  project            = "terraform"
  description        = "desc"
  source_of          = ["Metrics", "Services"]
  oauth_bearer_token = "secret"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `source_of` (List of String) Source of Metrics and/or Services

### Optional

- `description` (String) Optional description of the resource. Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.
- `display_name` (String) User-friendly display name of the resource.
- `oauth_bearer_token` (String, Sensitive) [required] | ThousandEyes OAuth Bearer Token.

### Read-Only

- `id` (String) The ID of this resource.
- `status` (String) Status of the created direct.

## Nobl9 Official Documentation

https://docs.nobl9.com/