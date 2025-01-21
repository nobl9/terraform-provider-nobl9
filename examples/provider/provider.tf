terraform {
  required_providers {
    nobl9 = {
      source  = "nobl9/nobl9"
      version = "0.35.0"
    }
  }
}

provider "nobl9" {
  client_id     = "<client_id>"
  client_secret = "<client_secret>"
}
