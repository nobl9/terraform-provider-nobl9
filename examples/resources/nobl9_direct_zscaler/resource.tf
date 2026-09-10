variable "zscaler_client_id" {
  type      = string
  sensitive = true
}

variable "zscaler_client_secret" {
  type      = string
  sensitive = true
}

resource "nobl9_direct_zscaler" "this" {
  name            = "zscaler"
  project         = "default"
  vanity_domain   = "example"
  client_id       = var.zscaler_client_id
  client_secret   = var.zscaler_client_secret
  release_channel = "beta"

  query_delay {
    unit  = "Minute"
    value = 20
  }

  historical_data_retrieval {
    default_duration {
      unit  = "Day"
      value = 7
    }
    max_duration {
      unit  = "Day"
      value = 14
    }
  }
}
