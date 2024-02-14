// Package nobl9
package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9Service(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: ProviderFactory(),
		CheckDestroy:      CheckDestroy("nobl9_service", manifest.KindService),
		Steps: []resource.TestStep{
			{
				Config: testService("test-service"),
				Check:  CheckObjectCreated("nobl9_service.test-service"),
			},
		},
	})
}

func testService(name string) string {
	return fmt.Sprintf(`
resource "nobl9_service" "%s" {
  name              = "%s"
  display_name = "%s"
  project             = "%s"
  description       = "Test of service"

  label {
   key = "env"
   values = ["green","sapphire"]
  }

  label {
   key = "dev"
   values = ["dev", "staging", "prod"]
  }
}
`, name, name, name, testProject)
}
