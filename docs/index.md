---
page_title: "Nobl9 Provider"
description: |-
  The Nobl9 provider provides utilities for working with Nobl9 API.
---

# NOBL9 Provider

The Nobl9 provider provides utilities for working with Nobl9 API to create and manage resources such as:
- SLOs
- Services
- Projects
- Alert Policies
- Alert Methods
- Data Sources
- Role Bindings
Use the navigation to the left to learn more about the available resources.

This provider can be used as an alternative to [sloctl](https://docs.nobl9.com/sloctl-user-guide/).

## Example Usage

```terraform
terraform {
  required_providers {
    nobl9 = {
      source  = "nobl9/nobl9"
      version = "0.6.0"
    }
  }
}

provider "nobl9" {
  organization  = "<your org name>"
  project       = "default"
  client_id     = "<client_id>"
  client_secret = "<client_secret>"
}
```

