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
resource "nobl9_alert_policy" "fast_exhaustion_bewlo_budget_alerting_window" {
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