resource "nobl9_project" "this" {
  display_name = "Foo Project"
  name         = "foo-project"
  description  = "An example terraform project"

  label {
    key    = "env"
    values = ["dev", "prod"]
  }

  label {
    key    = "team"
    values = ["red"]
  }
}

