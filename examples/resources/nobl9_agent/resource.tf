resource "nobl9_project" "this" {
  display_name = "Test N9 Terraform"
  name         = "test-n9-terraform"
  description  = "An example N9 Terraform project"
}

resource "nobl9_agent" "this" {
  name            = "${nobl9_project.this.name}-prom-agent"
  project         = nobl9_project.this.name
  agent_type      = "prometheus"
  release_channel = "stable"
  prometheus_config {
    url = "http://web.net"
  }
}
