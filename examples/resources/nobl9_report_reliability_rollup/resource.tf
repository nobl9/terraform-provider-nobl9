resource "nobl9_report_reliability_rollup" "all_projects" {
  name         = "my-rrr-all-projects"
  display_name = "My Reliability Roll-up - All Projects"

  filters {
    project_scope = "all"
  }

  time_frame {
    time_zone = "Europe/Warsaw"

    rolling {
      unit  = "Week"
      count = 4
    }
  }
}

resource "nobl9_report_reliability_rollup" "calendar_last_month" {
  name         = "my-rrr-last-month"
  display_name = "My Reliability Roll-up - Last Month"

  filters {
    service {
      name    = "checkout-api"
      project = "customer-facing"
    }

    service {
      name    = "worker-service"
      project = "platform"
    }
  }

  time_frame {
    time_zone = "Europe/Warsaw"

    calendar {
      from = "2026-04-01"
      to   = "2026-04-30"
    }
  }
}

resource "nobl9_report_reliability_rollup" "custom_hierarchy" {
  name         = "my-rrr-custom-hierarchy"
  display_name = "My Reliability Roll-up - Custom Hierarchy"

  time_frame {
    time_zone = "Europe/Warsaw"

    rolling {
      unit  = "Week"
      count = 4
    }
  }

  custom_hierarchy = jsonencode([
    {
      displayName = "Customer-facing services"
      slos = [
        {
          name    = "checkout-availability"
          project = "customer-facing"
        },
        {
          name    = "api-latency"
          project = "customer-facing"
        },
      ]
    },
    {
      displayName = "Background processing"
      children = [
        {
          displayName = "Workers"
          slos = [
            {
              name    = "worker-throughput"
              project = "platform"
            },
          ]
        },
      ]
    },
  ])
}
