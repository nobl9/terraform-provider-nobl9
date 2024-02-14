Terraform Nobl9 Provider
=========================

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x

Example
----------------------

```terraform
terraform {
  required_providers {
    nobl9 = {
      source = "nobl9/nobl9"
      version = "0.23.0-beta"
    }
  }
}

provider "nobl9" {
  client_id = "<CLIENT_ID>"
  client_secret = "<CLIENT_SECRET>"
}

resource "nobl9_project" "test" {
  name = "test"
}

resource "nobl9_service" "test" {
  name    = "test"
  project = "test"
}
```

Documentation
-------------------

Documentation is generated using the
[tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) tool.
In order to generate or update the docs run the following command:

```sh
go generate
```
