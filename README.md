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
      source = "nobl9.com/nobl9/nobl9"
      version = "0.1.0"
    }
  }
}

provider "nobl9" {
  organization = "test"
  project = "test"
  client_id = "<CLIENT_ID>"
  client_secret = "<CLIENT_SECRET>"
}

resource "nobl9_service" "test" {
  metadata {
    name = "test"
  }
}
```