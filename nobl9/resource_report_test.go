package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9Reports(t *testing.T) {
	cases := []struct {
		name         string
		reportSuffix string
		configFunc   func(string) string
	}{
		{"system-health-review-by-project-latest-snapshot", "system_health_review", testSHRLatestSnapshotByProject},
		{"system-health-review-by-service-latest-snapshot", "system_health_review", testSHRLatestSnapshot},
		{"system-health-review-by-service-past-snapshot-without-rrule", "system_health_review", testSHRPastSnapshot},
		{
			"system-health-review-by-service-past-snapshot-with-rrule",
			"system_health_review",
			testSHRPastSnapshotWithRrule,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_report_"+tc.reportSuffix, manifest.KindReport),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated(fmt.Sprintf("nobl9_report_%s.%s", tc.reportSuffix, tc.name)),
					},
				},
			})
		})
	}
}

func testSHRLatestSnapshotByProject(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "project"
  filters {
    projects = ["%s"]
  }
  time_frame {
    time_zone = "Europe/Warsaw"
    snapshot {
      point = "latest"
    }
  }
  column {
    display_name = "Column 1"
    label {
      key = "key1"
      values = ["value1"]
    }
  }
  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}
`, name, name, testProject)
}

func testSHRLatestSnapshot(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "service"
  filters {
    projects = ["%s"]
  }
  time_frame {
    time_zone = "Europe/Warsaw"
    snapshot {
      point = "latest"
    }
  }
  column {
    display_name = "Column 1"
    label {
      key = "key1"
      values = ["value1"]
    }
  }
  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}
`, name, name, testProject)
}

func testSHRPastSnapshot(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "service"
  filters {
    projects = ["%s"]
  }
  time_frame {
    time_zone = "Europe/Warsaw"
    snapshot {
      point = "past"
      date_time = "2024-09-05T09:58:37Z"
    }
  }
  column {
    display_name = "Column 1"
    label {
      key = "key1"
      values = ["value1"]
    }
  }
  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}
`, name, name, testProject)
}

func testSHRPastSnapshotWithRrule(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "service"
  filters {
    projects = ["%s"]
  }
  time_frame {
    time_zone = "Europe/Warsaw"
    snapshot {
      point = "past"
      date_time = "2024-09-05T09:58:37Z"
      rrule = "FREQ=DAILY;INTERVAL=1"
    }
  }
  column {
    display_name = "Column 1"
    label {
      key = "key1"
      values = ["value1"]
    }
  }
  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}
`, name, name, testProject)
}
