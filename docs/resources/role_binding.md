---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "nobl9_role_binding Resource - terraform-provider-nobl9"
subcategory: ""
description: |-
  RoleBinding configuration documentation https://nobl9.github.io/techdocs_YAML_Guide/
---

# nobl9_role_binding (Resource)

[RoleBinding configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/)



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **role_ref** (String) Role name.
- **user** (String) ID of the user.

### Optional

- **display_name** (String) Display name of the resource.
- **id** (String) The ID of this resource.
- **name** (String) Automatically generated, unique name of the resource. Must match [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- **project_ref** (String) Project name. When empty, `role_ref` has to be Organization Role.

