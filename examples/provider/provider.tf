terraform {
  required_providers {
    nobl9 = {
      source  = "nobl9/nobl9"
      version = "0.24.0-beta"
    }
  }
}

provider "nobl9" {
  client_id     = "<client_id>"
  client_secret = "<client_secret>"
}
