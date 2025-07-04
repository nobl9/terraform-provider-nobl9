resource "nobl9_project" "this" {
  name = "project"
  display_name = "Project"
  annotations = {
    key = "value",
  }
  label {
    key = "team"
    values = [
      "green",
      "orange",
    ]
  }
  label {
    key = "env"
    values = [
      "prod",
    ]
  }
  label {
    key = "empty"
    values = [
      "",
    ]
  }
  description = "Example project"
}
