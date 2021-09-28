package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Service(t *testing.T) {
	name := "test-service"
	config := fmt.Sprintf(`
resource "nobl9_service" "%s" {
  name         = "%s"
  display_name = "%s"
  description  = "%s"
  project      = "%s"
  //label {
  //  key   = "env"
  //  value = "prod"
  //}
  //label {
  //  key   = "team"
  //  value = "green"
  //}
  //label {
  //  key   = "team"
  //  value = "orange"
  //}
}
`, name, name, name, name, testProject)
	// TODO uncomment labels when PC-3250 is fixed

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: ProviderFactory(),
		CheckDestroy:      DestroyFunc("nobl9_service", n9api.ObjectService),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  CheckObjectCreated("nobl9_service." + name),
			},
		},
	})
}
