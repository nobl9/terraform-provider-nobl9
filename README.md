Terraform `nobl9` Provider
=========================



Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x

Example
----------------------
```sh
terraform {
  required_providers {
    nobl9 = {
      source = "nobl9/nobl9"
      version = "0.17.0"
    }
  }
}

provider "nobl9" {
  organization = "test"
  project = "test"
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

Documentation is generated using the [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) tool.
In order to generate or update the docs run the following command:

```
go generate
```