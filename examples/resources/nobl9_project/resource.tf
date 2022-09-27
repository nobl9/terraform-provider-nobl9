resource "nobl9_project" "this" {
  display_name = "My Project"
  name         = "my-project"
  description  = "An example N9 Terraform project"

  label {
    key    = "env"
    values = ["dev", "prod"]
  }

  label {
    key    = "team"
    values = ["red"]
  }
}

