package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Project(t *testing.T) {
	name := "test-project"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: ProviderFactory(),
		CheckDestroy:      CheckDestory("nobl9_project", n9api.ObjectProject),
		Steps: []resource.TestStep{
			{
				Config: testProjectConfig(name),
				Check:  CheckObjectCreated("nobl9_project." + name),
			},
		},
	})
}

func testProjectConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_project" "%s" {
  name         = "%s"
  display_name = "%s"
  description  = "A terraform project"

}
`, name, name, name)
}
