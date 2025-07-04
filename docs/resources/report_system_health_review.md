---
page_title: "nobl9_report_system_health_review Resource - terraform-provider-nobl9"
description: |-
  System Health Review Report | Nobl9 Documentation https://docs.nobl9.com/reports/system-health-review/
---

# nobl9_report_system_health_review (Resource)

The System Health Review report facilitates recurring reliability check-ins by grouping your Nobl9 SLOs by projects or services and labels of your choice through the remaining error budget metric in a table-form report.

## Example Usage

Here's an example of Error Budget Status Report resource configuration:

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `column` (Block List, Min: 1) Columns to display in the report table. (see [below for nested schema](#nestedblock--column))
- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `row_group_by` (String) Grouping methods of report table rows [project/service]
- `thresholds` (Block Set, Min: 1, Max: 1) Thresholds for Green, Yellow and Red statuses (e.g. healthy, at risk, exhausted budget). Yellow is calculated as the difference between Red and Green thresholds. If Red and Green are the same, Yellow is not used on the report. (see [below for nested schema](#nestedblock--thresholds))
- `time_frame` (Block Set, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--time_frame))

### Optional

- `display_name` (String) User-friendly display name of the resource.
- `filters` (Block List, Max: 1) Filters are used to select scope for Report. (see [below for nested schema](#nestedblock--filters))
- `shared` (Boolean) Is report shared for all users with access to included projects.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--column"></a>
### Nested Schema for `column`

Required:

- `display_name` (String) Column display name.
- `label` (Block List, Min: 1) [Labels](https://docs.nobl9.com/features/labels/) containing a single key and a list of values. (see [below for nested schema](#nestedblock--column--label))

<a id="nestedblock--column--label"></a>
### Nested Schema for `column.label`

Required:

- `key` (String) A key for the label, unique within the associated resource.
- `values` (List of String) A list of unique values for a single key.



<a id="nestedblock--thresholds"></a>
### Nested Schema for `thresholds`

Required:

- `green_gt` (Number) Min value for the Green status (e.g. healthy).
- `red_lte` (Number) Max value for the Red status (e.g. exhausted budget).

Optional:

- `show_no_data` (Boolean) ShowNoData customizes the report to either show or hide rows with no data.


<a id="nestedblock--time_frame"></a>
### Nested Schema for `time_frame`

Required:

- `snapshot` (Block Set, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--time_frame--snapshot))
- `time_zone` (String) Timezone name in IANA Time Zone Database.

<a id="nestedblock--time_frame--snapshot"></a>
### Nested Schema for `time_frame.snapshot`

Required:

- `point` (String) The method of reporting time frame [past/latest]

Optional:

- `date_time` (String) Date and time of the past snapshot in RFC3339 format.
- `rrule` (String) The recurrence rule for the report past snapshot. The expected value is a string in RRULE format. Example: `FREQ=MONTHLY;BYMONTHDAY=1`



<a id="nestedblock--filters"></a>
### Nested Schema for `filters`

Optional:

- `label` (Block List) [Labels](https://docs.nobl9.com/features/labels/) containing a single key and a list of values. (see [below for nested schema](#nestedblock--filters--label))
- `projects` (List of String) Projects to pull data for report from.
- `service` (Block List) Services to pull data for report from. (see [below for nested schema](#nestedblock--filters--service))
- `slo` (Block List) SLOs to pull data for report from. (see [below for nested schema](#nestedblock--filters--slo))

<a id="nestedblock--filters--label"></a>
### Nested Schema for `filters.label`

Required:

- `key` (String) A key for the label, unique within the associated resource.
- `values` (List of String) A list of unique values for a single key.


<a id="nestedblock--filters--service"></a>
### Nested Schema for `filters.service`

Required:

- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).


<a id="nestedblock--filters--slo"></a>
### Nested Schema for `filters.slo`

Required:

- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).

## Useful Links

[Reports in Nobl9 | Nobl9 Documentation](https://docs.nobl9.com/reports/)

[Reports YAML Configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#report)