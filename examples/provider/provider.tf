terraform {
  required_providers {
    nobl9 = {
      source  = "nobl9.com/nobl9/nobl9"
      version = "0.5.0"
    }
  }
}

provider "nobl9" {
  organization  = "<your org name>"
  project       = "default"
  client_id     = "<client_id>"
  client_secret = "<client_secret>"
}
