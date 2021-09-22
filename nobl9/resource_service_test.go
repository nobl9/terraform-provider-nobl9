package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Service(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-service", testService},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      DestroyFunc("nobl9_service", n9api.ObjectService),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_service." + tc.name),
					},
				},
			})
		})
	}
}

func testService(name string) string {
	return fmt.Sprintf(`
resource "nobl9_service" "%s" {
  name      = "%s"
  project   = "%s"
  service_spec {
	description = "Test of service"
	}
}
`, name, name, testProject)
}
