resource "nobl9_project" "this" {
  display_name = "Test Terraform"
  name         = "test-terraform"
  description  = "An example terraform project"
}

resource "nobl9_agent" "this" {
  name      =  "${nobl9_project.this.name}-prom-agent"
  project   = nobl9_project.this.name
  source_of = ["Metrics", "Services"]
  agent_type = "prometheus"
  prometheus_config {
    url = "http://web.net"
  }
}