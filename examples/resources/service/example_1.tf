resource "nobl9_project" "this" {
  display_name = "My Project"
  name         = "my-project"
  description  = "An example N9 Terraform project"
}

resource "nobl9_service" "this" {
  name         = "${nobl9_project.this.name}-front-page"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Front Page"
  description  = "Front page service"

  label {
    key    = "env"
    values = ["dev", "prod"]
  }

  label {
    key    = "team"
    values = ["red"]
  }
}
