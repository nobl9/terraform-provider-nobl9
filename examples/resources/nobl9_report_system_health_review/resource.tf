resource "nobl9_report_system_health_review" "this" {
  name         = "my-shr-report"
  display_name = "My System Health Review Report"
  shared       = true
  row_group_by = "service"

  filters {
    projects = ["project1", "project2"]
    service {
      name    = "service1"
      project = "project1"
    }
    service {
      name    = "service2"
      project = "project2"
    }
    slo {
      name    = "my-slo"
      project = "project1"
    }
    label {
      key    = "key1"
      values = ["value1"]
    }
  }

  time_frame {
    time_zone = "Europe/Warsaw"
    snapshot {
      point     = "past"
      date_time = "2024-09-05T09:58:37Z"
      rrule     = "FREQ=DAILY;INTERVAL=1"
    }
  }

  column {
    display_name = "Column 1"
    label {
      key    = "key1"
      values = ["value1"]
    }
  }

  column {
    display_name = "Column 2"
    label {
      key    = "key2"
      values = ["value2"]
    }
  }

  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}

