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
