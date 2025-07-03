resource "nobl9_service" "this" {
  name = "service"
  display_name = "Service"
  project = "default"
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
  description = "Example service"
}
