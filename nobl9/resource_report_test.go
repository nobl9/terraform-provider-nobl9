package nobl9

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_report_"+tc.reportSuffix, manifest.KindReport),
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

func TestAcc_Nobl9ReportsErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		reportSuffix string
		configFunc   func(string) string
		errorMessage string
	}{
		{"system-health-review-wrong-grouping",
			"system_health_review",
			testSHRWrongGrouping,
			`expected row_group_by to be one of`,
		},
		{"system-health-review-empty-filters",
			"system_health_review",
			testSHREmptyFilters,
			`property is required but was empty`,
		},
		{"system-health-review-wrong-filters",
			"system_health_review",
			testSHRWrongFilters,
			`at least one of the following fields is required: projects, services, slos`,
		},
		{"system-health-review-empty-column",
			"system_health_review",
			testSHREmptyColumns,
			`At least 1 "column" blocks are required`,
		},
		{"system-health-review-columns-with-no-labels",
			"system_health_review",
			testSHRColumnsWithNoLabels,
			`Insufficient label blocks`,
		},
		{"system-health-review-empty-thresholds",
			"system_health_review",
			testSHREmptyThresholds,
			`Insufficient thresholds blocks`,
		},
		{"system-health-review-wrong-thresholds",
			"system_health_review",
			testSHRWrongThresholds,
			`must be less than or equal to`,
		},
		{"system-health-review-wrong-snapshot",
			"system_health_review",
			testSHRWrongSnapshot,
			`property is required but was empty`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_report_"+tc.reportSuffix, manifest.KindReport),
				Steps: []resource.TestStep{
					{
						Config:      tc.configFunc(tc.name),
						ExpectError: regexp.MustCompile(tc.errorMessage),
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

func testSHRWrongGrouping(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "wrong"
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

func testSHREmptyFilters(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "project"
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
`, name, name)
}

func testSHRWrongFilters(name string) string {
	return fmt.Sprintf(`
resource "nobl9_report_system_health_review" "%s" {
  name         = "%s"
  display_name = "System Health Review Report"
  shared       = true
  row_group_by = "project"
  filters {
    label {
      key = "team"
      values = ["green"]
    }
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
`, name, name)
}

func testSHREmptyColumns(name string) string {
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
  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}
`, name, name, testProject)
}

func testSHRColumnsWithNoLabels(name string) string {
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
  }
  thresholds {
    red_lte      = 0.8
    green_gt     = 0.95
    show_no_data = true
  }
}
`, name, name, testProject)
}

func testSHREmptyThresholds(name string) string {
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
}
`, name, name, testProject)
}

func testSHRWrongThresholds(name string) string {
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
    red_lte      = 0.9
    green_gt     = 0.8
    show_no_data = true
  }
}
`, name, name, testProject)
}

func testSHRWrongSnapshot(name string) string {
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
      point = "past"
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
