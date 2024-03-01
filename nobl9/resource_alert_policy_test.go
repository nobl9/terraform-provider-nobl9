package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9AlertPolicy(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(name string) string
	}{
		{"alert-policy", testAlertPolicyWithoutAnyAlertMethod},
		{"alert-policy-with-cool-down", testAlertPolicyWithCoolDown},
		{"alert-policy-with-alert-method", testAlertPolicyWithAlertMethod},
		{"alert-policy-with-multi-alert-method", testAlertPolicyWithMultipleAlertMethods},
		{"alert-policy-with-multi-alert-method-reverse", testAlertPolicyWithMultipleAlertMethodsReverseOrder},
		{"alert-policy-with-time-to-burn-entire-budget", testAlertPolicyWithTimeToBurnEntireBudgetCondition},
		{
			"alert-policy-with-average-burn-rate-and-alerting-window",
			testAlertPolicyWithAverageBurnRateAndAlertingWindow,
		},
		{"alert-policy-with-average-burn-rate-and-lasts-for", testAlertPolicyWithAverageBurnRateAndLastsFor},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: ProviderFactory(),
				CheckDestroy: destroyMultiple(
					[]string{"nobl9_alert_policy", "nobl9_alert_method_webhook"},
					[]manifest.Kind{manifest.KindAlertPolicy, manifest.KindAlertMethod},
				),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_alert_policy." + tc.name),
					},
					// make sure that applying the same config results in a no-op plan, regardless of alert_method order
					{
						Config:             tc.configFunc(tc.name),
						PlanOnly:           true,
						ExpectNonEmptyPlan: false,
					},
				},
			})
		})
	}
}

func destroyMultiple(rsTypes []string, kinds []manifest.Kind) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if len(rsTypes) != len(kinds) {
			return fmt.Errorf("resource_types (%v) must match objectTypes (%v)", rsTypes, kinds)
		}
		for i := 0; i < len(rsTypes); i++ {
			CheckDestroy(rsTypes[i], kinds[i])
		}
		return nil
	}
}

func testAlertPolicyWithoutAnyAlertMethod(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"

  condition {
	  measurement = "burnedBudget"
	  value 	  = 0.9
	}

  condition {
	  measurement = "averageBurnRate"
	  value 	  = 3
	  lasts_for	  = "1m"
	}

  condition {
	  measurement  = "timeToBurnBudget"
	  value_string = "1h"
	  lasts_for	   = "300s"
	}
}
`, name, name, testProject)
}

func testAlertPolicyWithAlertMethod(name string) string {
	return testWebhookTemplateConfig(name+"-am") +
		fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"

  condition {
    measurement = "burnedBudget"
	value 	  = 0.9
  }

  condition {
    measurement = "averageBurnRate"
	value 	  = 3
	lasts_for	  = "1m"
  }

  condition {
	measurement  = "timeToBurnBudget"
	value_string = "1h"
	lasts_for	   = "300s"
  }

  alert_method {
	project = "%s"
	name	= nobl9_alert_method_webhook.%s-am.name
  }
}
`, name, name, testProject, testProject, name)
}

func testAlertPolicyWithMultipleAlertMethods(name string) string {
	return testWebhookTemplateConfig(name+"-am") +
		testWebhookTemplateConfig(name+"-am-two") +
		fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"

  condition {
    measurement = "burnedBudget"
    value 	  = 0.9
  }

  condition {
    measurement = "averageBurnRate"
    value 	  = 3
    lasts_for	  = "1m"
  }

  condition {
    measurement  = "timeToBurnBudget"
    value_string = "1h"
    lasts_for	   = "300s"
  }

  alert_method {
    project = "%s"
    name	= nobl9_alert_method_webhook.%s-am.name
  }

  alert_method {
    project = "%s"
    name	= nobl9_alert_method_webhook.%s-am-two.name
  }
}
`, name, name, testProject, testProject, name, testProject, name)
}

func testAlertPolicyWithMultipleAlertMethodsReverseOrder(name string) string {
	return testWebhookTemplateConfig(name+"-am") +
		testWebhookTemplateConfig(name+"-am-two") +
		fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"

  condition {
    measurement = "burnedBudget"
    value 	  = 0.9
  }

  condition {
    measurement = "averageBurnRate"
    value 	  = 3
    lasts_for	  = "1m"
  }

  condition {
    measurement  = "timeToBurnBudget"
    value_string = "1h"
    lasts_for	   = "300s"
  }

  alert_method {
    project = "%s"
    name	= nobl9_alert_method_webhook.%s-am-two.name
  }

  alert_method {
    project = "%s"
    name	= nobl9_alert_method_webhook.%s-am.name
  }
}
`, name, name, testProject, testProject, name, testProject, name)
}

func testAlertPolicyWithTimeToBurnEntireBudgetCondition(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"

  condition {
    measurement  = "timeToBurnEntireBudget"
    value_string = "1h"
    lasts_for	   = "300s"
  }

  condition {
    measurement  = "timeToBurnEntireBudget"
    value_string = "1h"
  }
}
`, name, name, testProject)
}

func testAlertPolicyWithCoolDown(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"
  cooldown  = "15m"

  condition {
	  measurement = "burnedBudget"
	  value 	  = 0.9
	}

  condition {
	  measurement = "averageBurnRate"
	  value 	  = 3
	  lasts_for	  = "1m"
	}

  condition {
	  measurement  = "timeToBurnBudget"
	  value_string = "1h"
	  lasts_for	   = "300s"
	}
}
`, name, name, testProject)
}

func testAlertPolicyWithAverageBurnRateAndAlertingWindow(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"
  cooldown  = "5m"
  condition {
	  measurement = "averageBurnRate"
	  value 	  = 1
	  alerting_window	  = "1h"
	}

  condition {
	  measurement  = "averageBurnRate"
	  value = "2"
	  alerting_window	   = "15m"
	}
}
`, name, name, testProject)
}

func testAlertPolicyWithAverageBurnRateAndLastsFor(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"
  cooldown  = "5m"
  condition {
	  measurement = "averageBurnRate"
	  value 	  = 2
	  lasts_for	  = "10m"
	}
}
`, name, name, testProject)
}
