package nobl9

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9AlertPolicy(t *testing.T) {
	for scenario, alertPolicyConfig := range map[string]alertPolicyConfig{
		// Test optional annotations.
		"alert policy with no annotations": {
			OverrideAnnotationsBlock: ``,
		},
		"alert policy with annotations": {
			OverrideAnnotationsBlock: `
				annotations = {
					env  = "development"
					name = "example annotation"
				}`,
		},
		// Test optional description.
		"alert policy with no description": {
			OverrideDescriptionBlock: ``,
		},
		"alert policy with custom description": {
			OverrideDescriptionBlock: `description = "test test"`,
		},
		// Test optional display name.
		"alert policy with no display name": {
			OverrideDisplayNameBlock: ``,
		},
		"alert policy with custom display name": {
			OverrideDisplayNameBlock: `display_name = "test test"`,
		},
		// Test optional cooldown.
		"alert policy with no cooldown defined": {
			OverrideCooldownBlock: ``,
		},
		"alert policy with custom cooldown": {
			OverrideCooldownBlock: `cooldown = "15m"`,
		},
		// Test multiple conditions order.
		"alert policy with multiple conditions": {
			OverrideConditionsBlock: `
				condition {
					measurement = "burnedBudget"
					value 	  	= 0.9
				}

				condition {
					measurement	= "averageBurnRate"
					value 	  	= 3
					lasts_for	= "1m"
				}

				condition {
					measurement  = "timeToBurnBudget"
					value_string = "1h"
					lasts_for	 = "300s"
				}`,
		},
		"alert policy with multiple conditions reversed": {
			OverrideConditionsBlock: `
				condition {
					measurement  = "timeToBurnBudget"
					value_string = "1h"
					lasts_for	 = "300s"
				}

				condition {
					measurement = "burnedBudget"
					value 	  	= 0.9
				}

				condition {
					measurement = "averageBurnRate"
					value 	  	= 3
					lasts_for	= "1m"
				}`,
		},
		// Test alert methods
		"alert policy with no alert method": {
			OverrideAlertMethodsBlock: ``,
		},
		"alert policy with multiple alert method": {
			AdditionalResources: fmt.Sprintf(`
			resource "nobl9_alert_method_slack" "%s" {
				name        = "%s"
				project     = "%s"
				description = "slack"
				url         = "https://hooks.slack.com/services/321/123/secret"
			}
			resource "nobl9_alert_method_slack" "%s" {
				name        = "%s"
				project     = "%s"
				description = "slack"
				url         = "https://hooks.slack.com/services/321/123/secret"
			}
			resource "nobl9_alert_method_slack" "%s" {
				name        = "%s"
				project     = "%s"
				description = "slack"
				url         = "https://hooks.slack.com/services/321/123/secret"
			}
			`,
				"am1", "am1", testProject,
				"am2", "am2", testProject,
				"am3", "am3", testProject,
			),
			OverrideAlertMethodsBlock: fmt.Sprintf(`
			alert_method {
				name = nobl9_alert_method_slack.am1.name
				project = "%s"
			}
			alert_method {
				name = nobl9_alert_method_slack.am2.name
				project = "%s"
			}
			alert_method {
				name = nobl9_alert_method_slack.am3.name
				project = "%s"
			}
			`, testProject, testProject, testProject),
		},
		"alert policy with multiple alert method reversed": {
			AdditionalResources: fmt.Sprintf(`
			resource "nobl9_alert_method_slack" "%s" {
				name        = "%s"
				project     = "%s"
				description = "slack"
				url         = "https://hooks.slack.com/services/321/123/secret"
			}
			resource "nobl9_alert_method_slack" "%s" {
				name        = "%s"
				project     = "%s"
				description = "slack"
				url         = "https://hooks.slack.com/services/321/123/secret"
			}
			resource "nobl9_alert_method_slack" "%s" {
				name        = "%s"
				project     = "%s"
				description = "slack"
				url         = "https://hooks.slack.com/services/321/123/secret"
			}
			`,
				"am1", "am1", testProject,
				"am2", "am2", testProject,
				"am3", "am3", testProject,
			),
			OverrideAlertMethodsBlock: fmt.Sprintf(`
			alert_method {
				name = nobl9_alert_method_slack.am3.name
				project = "%s"
			}
			alert_method {
				name = nobl9_alert_method_slack.am2.name
				project = "%s"
			}
			alert_method {
				name = nobl9_alert_method_slack.am1.name
				project = "%s"
			}
			`, testProject, testProject, testProject),
		},
		// Measurement: burnedBudget
		"burned budget with default operator and lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement = "burnedBudget"
					value 	  	= 1.0
				}`,
		},
		"burned budget with value 0": {
			OverrideConditionsBlock: `
				condition {
					measurement = "burnedBudget"
					value 	  	= 0.0
				}`,
		},
		"burned budget with explicit default operator": {
			OverrideConditionsBlock: `
				condition {
					measurement = "burnedBudget"
					value 	  	= 1.0
					op			= "gte"
				}`,
		},
		"burned budget with custom operator": {
			OverrideConditionsBlock: `
				condition {
					measurement = "burnedBudget"
					value 	  	= 1.0
					op			= "lt"
				}`,
		},
		"burned budget with custom lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement = "burnedBudget"
					value 	  	= 1.0
					lasts_for	= "10m"
				}`,
		},
		// Measurement: averageBurnRate
		"average burn rate with default operator and lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement = "averageBurnRate"
					value 	  	= 1.0
				}`,
		},
		"average burn rate with value 0": {
			OverrideConditionsBlock: `
				condition {
					measurement = "averageBurnRate"
					value 	  	= 0.0
				}`,
		},
		"average burn rate with explicit default operator": {
			OverrideConditionsBlock: `
				condition {
					measurement = "averageBurnRate"
					value 	  	= 1.0
					op			= "gte"
				}`,
		},
		"average burn rate with custom lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement = "averageBurnRate"
					value 	  	= 1.0
					lasts_for	= "10m"
				}`,
		},
		"average burn rate with custom alerting window": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "averageBurnRate"
					value 	  		= 1.0
					alerting_window	= "10m"
				}`,
		},
		// Measurement: budgetDrop
		"budget drop with explicit operator": {
			OverrideConditionsBlock: `
				condition {
					measurement = "budgetDrop"
					value 	  	= 1.0
					op			= "gte"
					alerting_window	= "10m"
				}`,
		},
		"budget drop with default operator": {
			OverrideConditionsBlock: `
				condition {
					measurement = "averageBurnRate"
					value 	  	= 1.0
					alerting_window	= "10m"
				}`,
		},
		// Measurement: timeToBurnBudget
		"time to burn budget with default operator and lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnBudget"
					value_string 	= "6h"
				}`,
		},
		"time to burn budget with explicit default operator": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnBudget"
					value_string 	= "6h"
					op				= "lt"
				}`,
		},
		"time to burn budget with custom lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnBudget"
					value_string 	= "6h"
					lasts_for		= "10m"
				}`,
		},
		"time to burn budget with custom alerting window": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnBudget"
					value_string 	= "6h"
					alerting_window	= "10m"
				}`,
		},
		// Measurement: timeToBurnEntireBudget
		"time to burn entire budget with default operator and lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnEntireBudget"
					value_string 	= "6h"
				}`,
		},
		"time to burn entire budget with explicit default operator": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnEntireBudget"
					value_string 	= "6h"
					op				= "lte"
				}`,
		},
		"time to burn entire budget with custom lasts for": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnEntireBudget"
					value_string 	= "6h"
					lasts_for		= "10m"
				}`,
		},
		"time to burn entire budget with custom alerting window": {
			OverrideConditionsBlock: `
				condition {
					measurement 	= "timeToBurnEntireBudget"
					value_string 	= "6h"
					alerting_window	= "10m"
				}`,
		},
	} {
		t.Run(scenario, func(t *testing.T) {
			resourceName := strings.ReplaceAll(scenario, " ", "_")
			alertPolicyName := strings.ReplaceAll(scenario, " ", "-")

			res := alertPolicyConfig.Build(resourceName, alertPolicyName, testProject)

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy: destroyMultiple(
					[]string{"nobl9_alert_policy", "nobl9_alert_method_slack"},
					[]manifest.Kind{manifest.KindAlertPolicy, manifest.KindAlertMethod},
				),
				Steps: []resource.TestStep{
					{
						Config: res,
						Check:  CheckObjectCreated("nobl9_alert_policy." + resourceName),
					},
					// make sure that applying the same config results in a no-op plan, regardless of alert_method order
					{
						Config:             res,
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

type alertPolicyConfig struct {
	OverrideDisplayNameBlock  string
	OverrideDescriptionBlock  string
	OverrideCooldownBlock     string
	OverrideConditionsBlock   string
	OverrideSeverityBlock     string
	OverrideAlertMethodsBlock string
	OverrideAnnotationsBlock  string
	AdditionalResources       string
}

func (ap alertPolicyConfig) Build(resourceName, name, project string) string {
	const defaultCondition = `
	condition {
		measurement = "burnedBudget"
		value 	  = 0.9
	}`

	b := strings.Builder{}
	b.WriteString(fmt.Sprintf(`
%s
resource "nobl9_alert_policy" "%s" {`, ap.AdditionalResources, resourceName))
	b.WriteString("\n	")
	b.WriteString(fmt.Sprintf(`name = "%s"`, name))
	b.WriteString("\n	")
	b.WriteString(fmt.Sprintf(`project = "%s"`, project))
	b.WriteString("\n	")
	if ap.OverrideAnnotationsBlock != "" {
		b.WriteString(ap.OverrideAnnotationsBlock)
		b.WriteString("\n	")
	}
	if ap.OverrideSeverityBlock == "" {
		b.WriteString(`severity = "Low"`)
	} else {
		b.WriteString(ap.OverrideSeverityBlock)
	}
	b.WriteString("\n	")
	if ap.OverrideCooldownBlock != "" {
		b.WriteString(ap.OverrideCooldownBlock)
		b.WriteString("\n	")
	}
	if ap.OverrideDisplayNameBlock != "" {
		b.WriteString(ap.OverrideDisplayNameBlock)
		b.WriteString("\n	")
	}
	if ap.OverrideDescriptionBlock != "" {
		b.WriteString(ap.OverrideDescriptionBlock)
		b.WriteString("\n	")
	}
	if ap.OverrideAlertMethodsBlock != "" {
		b.WriteString(ap.OverrideAlertMethodsBlock)
		b.WriteString("\n	")
	}
	if ap.OverrideConditionsBlock == "" {
		b.WriteString(defaultCondition)
	} else {
		b.WriteString(ap.OverrideConditionsBlock)
	}
	b.WriteString(`
}`)
	return b.String()
}
