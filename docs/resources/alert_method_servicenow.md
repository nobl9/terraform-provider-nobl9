---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "nobl9_alert_method_servicenow Resource - terraform-provider-nobl9"
subcategory: ""
description: |-
  Integration configuration documentation https://docs.nobl9.com/Alert_Methods/servicenow
---

# nobl9_alert_method_servicenow (Resource)

[Integration configuration documentation](https://docs.nobl9.com/Alert_Methods/servicenow)



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `instance_name` (String) ServiceNow InstanceName. For details see documentation.
- `name` (String) Unique name of the resource. Must match [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the project the resource is in. Must match [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `username` (String) ServiceNow username.

### Optional

- `description` (String) Optional description of the resource.
- `display_name` (String) Display name of the resource.
- `password` (String, Sensitive) ServiceNow password.

### Read-Only

- `id` (String) The ID of this resource.


