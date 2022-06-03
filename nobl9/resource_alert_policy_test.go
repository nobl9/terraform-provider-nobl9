package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9AlertPolicy(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(name string) string
	}{
		{"alert-policy", testAlertPolicyWithoutIntegration},
		{"alert-policy-with-alert-method", testAlertPolicyWithIntegration},
		{"alert-policy-with-multi-alert-method", testAlertPolicyWithMultipleIntegration},
		// This is coming from SRE-738 where the order of the alert methods was always showing a diff
		{"alert-policy-with-multi-alert-method-reverse", testAlertPolicyWithMultipleIntegrationReverseOrder},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy: destroyMultiple(
					[]string{"nobl9_alert_policy", "nobl9_alert_method_webhook"},
					[]n9api.Object{n9api.ObjectAlertPolicy, n9api.ObjectAlertMethod},
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

func destroyMultiple(rsTypes []string, objectTypes []n9api.Object) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if len(rsTypes) != len(objectTypes) {
			return fmt.Errorf("resource_types (%v) must match objectTypes (%v)", rsTypes, objectTypes)
		}
		for i := 0; i < len(rsTypes); i++ {
			CheckDestory(rsTypes[i], objectTypes[i])
		}
		return nil
	}
}

func testAlertPolicyWithoutIntegration(name string) string {
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

func testAlertPolicyWithIntegration(name string) string {
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

func testAlertPolicyWithMultipleIntegration(name string) string {
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

func testAlertPolicyWithMultipleIntegrationReverseOrder(name string) string {
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
