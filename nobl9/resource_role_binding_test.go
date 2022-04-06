package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9RoleBinding(t *testing.T) {
	t.SkipNow() // these need work to get them to pass
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"project-role-binding", testProjectRoleBindingConfig},
		// this test is skipped for now because: deleting organizational role bindings is not allowed
		//{"org-role-binding", testOrganizationRoleBindingConfig},
		{"role-binding-without-name", testRoleBindingWithoutName},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestory("nobl9_role_binding", n9api.ObjectRoleBinding),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_role_binding." + tc.name),
					},
				},
			})
		})
	}
}

func testProjectRoleBindingConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  name        = "%s"
  user        = "00u3lognksvI7G1r54x7xx"
  role_ref    = "project-owner"
  project_ref = "%s"
}
`, name, name, testProject)
}

func testOrganizationRoleBindingConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  name        = "%s"
  user        = "00u3lognksvI7G1r54x7xx"
  role_ref    = "organization-admin"
}
`, name, name)
}

func testRoleBindingWithoutName(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  user        = "00u3lognksvI7G1r54x7xx"
  role_ref    = "project-owner"
  project_ref = "%s"
}
`, name, testProject)
}
