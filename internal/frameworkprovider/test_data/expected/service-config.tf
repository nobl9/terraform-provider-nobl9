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

  responsible_users = [
    {
      id = "userID1"
    },
    {
      id = "userID2"
    },
  ]
  review_cycle = {
    rrule      = "FREQ=DAILY"
    start_time = "2024-01-01T08:00:00"
    time_zone  = "Asia/Tokyo"
  }
}
