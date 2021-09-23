package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Integration(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-webhhok", testWebhookTemplateConfig},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      DestroyFunc("nobl9_integration", n9api.ObjectIntegration),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_integration." + tc.name),
					},
				},
			})
		})
	}
}

func testWebhookTemplateConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_integration" "%s" {
  name             = "%s"
  project          = "%s"
  integration_type = "webhook"
  webhook_config {
    url      = "http://web.net"
    template = "SLO needs attention $slo_name"
	}
}
`, name, name, testProject)
}
