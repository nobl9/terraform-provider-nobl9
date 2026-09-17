resource "nobl9_agent" "zscaler" {
  name            = "zscaler"
  project         = "default"
  agent_type      = "zscaler"
  release_channel = "beta"

  zscaler_config {
    vanity_domain = "example"
  }

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
