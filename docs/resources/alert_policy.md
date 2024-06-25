---
page_title: "nobl9_alert_policy Resource - terraform-provider-nobl9"
description: |-
  Alert Policy configuration | Nobl9 Documentation https://docs.nobl9.com/yaml-guide#alertpolicy
---

# nobl9_alert_policy (Resource)

An **Alert Policy** expresses a set of conditions you want to track or monitor. The conditions for an Alert Policy define what is monitored and when to activate an alert: when the performance of your service is declining, Nobl9 will send a notification to a predefined channel.

A Nobl9 AlertPolicy accepts up to 3 conditions. All the specified conditions must be satisfied to trigger an alert.

For more details, refer to the [Alert Policy configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#alertpolicy).

## Example Usage

Here's an example of Alert Policy resource configuration:

```terraform
resource "nobl9_project" "this" {
  display_name = "My Project"
  name         = "my-project"
  description  = "An example N9 Terraform project"
}

resource "nobl9_service" "this" {
  name         = "${nobl9_project.this.name}-front-page"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Front Page"
  description  = "Front page service"
}

# Alert when a high momentary burn rate risks exhausting the budget (any time window).
resource "nobl9_alert_policy" "fast_burn_20x_over_5m" {
  name         = "fast-burn-20x5min"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Fast burn (20x5min)"
  severity     = "High"
  description  = "There’s been a significant spike in burn rate over a brief period."
  cooldown     = "5m"

  condition {
    measurement     = "averageBurnRate"
    value           = 20.0
    alerting_window = "5m"
    op              = "gte"
  }
}

# Alert when the budget burns slowly, and is not recovering (time windows over a week)
resource "nobl9_alert_policy" "slow_burn_1x_over_2d_and_2x_over_15m" {
  name         = "slow-burn-1x2d-and-2x15min"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Slow burn (1x2d and 2x15min)"
  severity     = "Low"
  description  = "The budget is slowly being exhausted and not recovering."
  cooldown     = "5m"

  condition {
    measurement     = "averageBurnRate"
    value           = 1.0
    alerting_window = "48h"
    op              = "gte"
  }

  condition {
    measurement     = "averageBurnRate"
    value           = 2.0
    alerting_window = "15m"
    op              = "gte"
  }
}

# Alert when the budget burns slowly and is not recovering (time windows up to one week)
resource "nobl9_alert_policy" "slow_burn_1x_over_12h_and_2x_over_15m" {
  name         = "slow-burn-1x12h-and-2x15min"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Slow burn (1x12h and 2x15min)"
  severity     = "Low"
  description  = "The budget is slowly exhausting and not recovering."
  cooldown     = "5m"

  condition {
    measurement     = "averageBurnRate"
    value           = 1.0
    alerting_window = "12h"
  }

  condition {
    measurement     = "averageBurnRate"
    value           = 2.0
    alerting_window = "15m"
  }
}

# Alert when the error budget is nearly exhausted (any time window)
resource "nobl9_alert_policy" "budget_almost_exhausted" {
  name         = "budget-almost-exhausted"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Budget almost exhausted (20%)"
  severity     = "High"
  description  = "The error budget is nearly exhausted (20% left)."
  cooldown     = "5m"

  condition {
    measurement = "burnedBudget"
    value       = 0.8
    op          = "gte"
  }
}

# There is no budget left and entire budget would be exhausted in 8h and this condition lasts for 15m.
resource "nobl9_alert_policy" "entire_exhaustion_prediction_8h_lasts_for" {
  name         = "entire-exhaustion-prediction-8h-lasts-for"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Entire budget exhaustion in 8h"
  severity     = "Low"
  description  = "No error budget left, but entire error budget would be exhausted in 8h."
  cooldown     = "15m"

  condition {
    measurement  = "timeToBurnEntireBudget"
    value_string = "8h"
    lasts_for    = "15m"
  }

  condition {
    measurement = "burnedBudget"
    value       = 1.0
    op          = "gte"
  }
}

# There is still some budget and remaining budget will be exhausted in 3 days.
resource "nobl9_alert_policy" "remaining_exhaustion_prediction_3d_lasts_for" {
  name         = "remaining-exhaustion-prediction-3d-lasts-for"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Remaining budget exhaustion in 3d"
  severity     = "Low"
  description  = "Remaining budget will be exhausted in 3 days."
  cooldown     = "15m"

  condition {
    measurement  = "timeToBurnBudget"
    value_string = "72h"
    lasts_for    = "15m"
  }

  condition {
    measurement = "burnedBudget"
    value       = 1.0
    op          = "lt"
  }
}

# Fast exhaustion below budget
resource "nobl9_alert_policy" "fast_exhaustion_below_budget_alerting_window" {
  name         = "fast-exhaustion-below-budget-alerting-window"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Fast exhaustion below budget (3d)"
  severity     = "High"
  description  = "The error budget is exhausting significantly, and there’s no remaining budget left."
  cooldown     = "5m"

  condition {
    measurement     = "timeToBurnEntireBudget"
    value_string    = "72h"
    alerting_window = "10m"
  }

  condition {
    measurement = "burnedBudget"
    value       = 1.0
    op          = "gte"
  }
}

# There is still some budget, slow exhaustion for long window SLOs
resource "nobl9_alert_policy" "slow_exhaustion_for_long_time_window_alerting_window" {
  name         = "slow-exhaustion-for-long-time-window-alerting-window"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Slow exhaustion (20d)"
  severity     = "Low"
  description  = "The error budget is above 0% is exhausting slowly and not recovering."
  cooldown     = "5m"

  condition {
    measurement     = "timeToBurnBudget"
    value_string    = "480h"
    alerting_window = "48h"
  }

  condition {
    measurement     = "timeToBurnBudget"
    value_string    = "480h"
    alerting_window = "15m"
  }

  condition {
    measurement = "burnedBudget"
    value       = 1.0
    op          = "lt"
  }
}

# There is still some budget, slow exhaustion for short window SLOs
resource "nobl9_alert_policy" "slow_exhaustion_for_short_time_window_alerting_window" {
  name         = "slow-exhaustion-for-short-time-window-alerting-window"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Slow exhaustion (5d)"
  severity     = "Low"
  description  = "The error budget is above 0% and is exhausting slowly and not recovering."
  cooldown     = "5m"

  condition {
    measurement     = "timeToBurnBudget"
    value_string    = "120h"
    alerting_window = "12h"
  }

  condition {
    measurement     = "timeToBurnBudget"
    value_string    = "120h"
    alerting_window = "15m"
  }

  condition {
    measurement = "burnedBudget"
    value       = 1.0
    op          = "lt"
  }
}

# Fast budget drop (10% over 15min)
resource "nobl9_alert_policy" "fast_budget_drop" {
  name         = "fast-budget-drop"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Fast budget drop (10% over 15 min)"
  severity     = "High"
  description  = "The budget dropped by 10% over the last 15 minutes and is not recovering."
  cooldown     = "5m"

  condition {
    measurement     = "budgetDrop"
    value           = 0.1
    alerting_window = "15m"
  }
}

# Slow budget drop (5% over 1h)
resource "nobl9_alert_policy" "slow_budget_drop" {
  name         = "slow-budget-drop"
  project      = nobl9_project.this.name
  display_name = "${nobl9_project.this.display_name} Slow budget drop (5% over 1h)"
  severity     = "Low"
  description  = "The budget dropped by 5% over the last 1 hour and is not recovering."
  cooldown     = "5m"

  condition {
    measurement     = "budgetDrop"
    value           = 0.05
    alerting_window = "1h"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `condition` (Block List, Min: 1) Configuration of an [alert condition](https://docs.nobl9.com/yaml-guide/#alertpolicy). (see [below for nested schema](#nestedblock--condition))
- `name` (String) Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `project` (String) Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
- `severity` (String) Alert severity. One of `Low` | `Medium` | `High`.

### Optional

- `alert_method` (Block List) (see [below for nested schema](#nestedblock--alert_method))
- `annotations` (Map of String) [Metadata annotations](https://docs.nobl9.com/features/labels/#metadata-annotations) attached to the resource.
- `cooldown` (String) An interval measured from the last time stamp when all alert policy conditions were satisfied before alert is marked as resolved
- `description` (String) Optional description of the resource. Here, you can add details about who is responsible for the integration (team/owner) or the purpose of creating it.
- `display_name` (String) User-friendly display name of the resource.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--condition"></a>
### Nested Schema for `condition`

Required:

- `measurement` (String) One of `timeToBurnBudget` | `timeToBurnEntireBudget` | `burnRate` | `burnedBudget` | `budgetDrop`.

Optional:

- `alerting_window` (String) Duration over which the burn rate is evaluated.
- `lasts_for` (String) Indicates how long a given condition needs to be valid to mark the condition as true.
- `op` (String) A mathematical inequality operator. One of `lt` | `lte` | `gt` | `gte`.
- `value` (Number) For `averageBurnRate`, it indicates how fast the error budget is burning. For `burnedBudget`, it tells how much error budget is already burned. For `budgetDrop`, it tells how much budget dropped.
- `value_string` (String) Used with `timeToBurnBudget` or `timeToBurnEntireBudget`, indicates when the budget would be exhausted. The expected value is a string in time duration string format.


<a id="nestedblock--alert_method"></a>
### Nested Schema for `alert_method`

Required:

- `name` (String) The name of the previously defined alert method.

Optional:

- `project` (String) Project name the Alert Method is in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names). If not defined, Nobl9 returns a default value for this field.

## Useful Links

[Alert Policy configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#alertpolicy)
