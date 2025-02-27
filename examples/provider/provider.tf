terraform {
  required_providers {
    nobl9 = {
      source  = "nobl9/nobl9"
      version = "0.37.1"
    }
  }
}

provider "nobl9" {
  client_id     = "<client_id>"
  client_secret = "<client_secret>"
}
