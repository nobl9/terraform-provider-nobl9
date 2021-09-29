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
		{"alert-policy-with-integration", testAlertPolicyWithIntegration},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy: destroyMultiple(
					[]string{"nobl9_alert_policy", "nobl9_integration_webhook"},
					[]n9api.Object{n9api.ObjectAlertPolicy, n9api.ObjectIntegration},
				),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_alert_policy." + tc.name),
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
			DestroyFunc(rsTypes[i], objectTypes[i])
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
	return testWebhookTemplateConfig(name) +
		fmt.Sprintf(`
resource "nobl9_integration_webhook" "integration-%s" {
  name        = "%s"
  project     = "%s"
  description = "wehbook"
  url         = "http://web.net"
  template    = "SLO needs attention $slo_name"
}

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

  integration {
	project = "%s"
	name	= "%s"
  }
}
`, name, name, testProject, name, name, testProject, testProject, name)
}
