package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9AlertMethod(t *testing.T) {
	cases := []struct {
		name           string
		resourceSuffix string
		configFunc     func(string) string
	}{
		{"test-webhook", "webhook", testWebhookTemplateConfig},
		{"test-webhook-fields", "webhook", testWebhookTemplateFieldsConfig},
		{"test-webhook-headers", "webhook", testWebhookHeadersConfig},
		{"test-pagerduty", "pagerduty", testPagerDutyConfig},
		{"test-pagerduty-send-resolution", "pagerduty", testPagerDutyWithSendResolutionConfig},
		{"test-pagerduty-send-resolution-message", "pagerduty", testPagerDutyWithSendResolutionWithMessageEmptyConfig},
		{"test-slack", "slack", testSlackConfig},
		{"test-discord", "discord", testDiscordConfig},
		{"test-opsgenie", "opsgenie", testOpsgenieConfig},
		{"test-servicenow", "servicenow", testServiceNowConfig},
		{"test-servicenow-send-resolution", "servicenow", testServiceNowWithSendResolutionConfig},
		{"test-servicenow-send-resolution-message", "servicenow", testServiceNowWithSendResolutionMessageEmptyConfig},
		{"test-jira", "jira", testJiraConfig},
		{"test-teams", "msteams", testTeamsConfig},
		{"test-email", "email", testEmailConfig},
		{"test-email-plain-text", "email", testEmailAsPlainTextConfig},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				CheckDestroy: CheckDestroy(
					"nobl9_alert_method_"+tc.resourceSuffix,
					manifest.KindAlertMethod,
				),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated(fmt.Sprintf("nobl9_alert_method_%s.%s", tc.resourceSuffix, tc.name)),
					},
				},
			})
		})
	}
}

func testWebhookTemplateConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_webhook" "%s" {
  name        = "%s"
  project     = "%s"
  description = "WebHook"
  url         = "http://web.net"
  template    = "SLO needs attention $slo_name"
}
`, name, name, testProject)
}

func testWebhookTemplateFieldsConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_webhook" "%s" {
  name            = "%s"
  project         = "%s"
  description	  = "WebHook"
  url             = "http://web.net"
  template_fields = [ "slo_name", "slo_details_link" ]
}
`, name, name, testProject)
}

func testWebhookHeadersConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_webhook" "%s" {
  name            = "%s"
  project         = "%s"
  description	    = "WebHook"
  url             = "http://web.net"
  template        = "SLO needs attention $slo_name"
  headers         = {
    "X-Custom-Header" = "custom value"
  }
  sensitive_headers = {
    "Authorization" = "Bearer xyz"
  }
}
`, name, name, testProject)
}

func testPagerDutyConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_pagerduty" "%s" {
  name            = "%s"
  project         = "%s"
  description     = "PagerDuty"
  integration_key = "84dfcdf19dad8f6c82b7e22afa024065"
}
`, name, name, testProject)
}

func testPagerDutyWithSendResolutionConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_pagerduty" "%s" {
  name            = "%s"
  project         = "%s"
  description     = "PagerDuty"
  integration_key = "84dfcdf19dad8f6c82b7e22afa024065"

  send_resolution {
    message = "Alert is now resolved"
  }
}
`, name, name, testProject)
}

func testPagerDutyWithSendResolutionWithMessageEmptyConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_pagerduty" "%s" {
  name            = "%s"
  project         = "%s"
  description     = "PagerDuty"
  integration_key = "84dfcdf19dad8f6c82b7e22afa024065"

  send_resolution {
  }
}
`, name, name, testProject)
}

func testSlackConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_slack" "%s" {
  name        = "%s"
  project     = "%s"
  description = "slack"
  url         = "https://hooks.slack.com/services/321/123/secret"
}
`, name, name, testProject)
}

func testDiscordConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_discord" "%s" {
  name        = "%s"
  project     = "%s"
  description = "discord"
  url         = "https://discord.com"
}
`, name, name, testProject)
}

func testOpsgenieConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_opsgenie" "%s" {
  name        = "%s"
  project     = "%s"
  description = "opsgenie"
  url         = "https://api.opsgenie.com"
  auth		  = "GenieKey 12345"
}
`, name, name, testProject)
}

func testServiceNowConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_servicenow" "%s" {
  name           = "%s"
  project        = "%s"
  description    = "servicenow"
  username       = "nobleUser"
  password       = "very secret"
  instance_name  = "name"
}
`, name, name, testProject)
}

func testServiceNowWithSendResolutionConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_servicenow" "%s" {
  name           = "%s"
  project        = "%s"
  description    = "servicenow"
  username       = "nobleUser"
  password       = "very secret"
  instance_name  = "name"

  send_resolution {
    message = "Alert is now resolved"
  }
}
`, name, name, testProject)
}

func testServiceNowWithSendResolutionMessageEmptyConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_servicenow" "%s" {
  name           = "%s"
  project        = "%s"
  description    = "servicenow"
  username       = "nobleUser"
  password       = "very secret"
  instance_name  = "name"

  send_resolution {
  }
}
`, name, name, testProject)
}

func testJiraConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_jira" "%s" {
  name        = "%s"
  project     = "%s"
  description = "jira"
  url		  = "https://jira.com"
  username    = "nobleUser"
  apitoken    = "very secret"
  project_key = "PC"
}
`, name, name, testProject)
}

func testTeamsConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_msteams" "%s" {
  name        = "%s"
  project     = "%s"
  description = "teams"
  url		  = "https://teams.com"
}
`, name, name, testProject)
}

func testEmailConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_email" "%s" {
  name        = "%s"
  project     = "%s"
  description = "email"
  to		  = [ "testUser@nobl9.com" ]
  cc		  = [ "testUser@nobl9.com" ]
  bcc		  = [ "testUser@nobl9.com" ]
}
`, name, name, testProject)
}

func testEmailAsPlainTextConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_email" "%s" {
  name        = "%s"
  project     = "%s"
  description = "email"
  to		  = [ "testUser@nobl9.com" ]
  cc		  = [ "testUser@nobl9.com" ]
  bcc		  = [ "testUser@nobl9.com" ]
  send_as_plain_text = true
}
`, name, name, testProject)
}

func mockAlertMethod(name, project string) string {
	return fmt.Sprintf(`
resource "nobl9_alert_method_slack" "%s" {
  name        = "%s"
  project     = "%s"
  description = "slack"
  url         = "https://hooks.slack.com/services/321/123/secret"
}
`, name, name, project)
}
