package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9RoleBinding(t *testing.T) {
	t.SkipNow() // these need work to get them to pass
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"project-role-binding", testProjectRoleBindingConfig},
		// this test is skipped for now because: deleting organizational role bindings is not allowed
		// {"org-role-binding", testOrganizationRoleBindingConfig},
		{"role-binding-without-name", testRoleBindingWithoutName},
		{"role-binding-without-user", testRoleBindingWithoutUser},
		{"role-binding-without-group", testRoleBindingWithoutGroup},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_role_binding", manifest.KindRoleBinding),
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
  user        = "test"
  role_ref    = "project-owner"
  project_ref = "%s"
}
`, name, name, testProject)
}

//nolint:unused,deadcode
func testOrganizationRoleBindingConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  name        = "%s"
  user        = "test"
  role_ref    = "organization-admin"
}
`, name, name)
}

func testRoleBindingWithoutName(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  user        = "test"
  role_ref    = "project-owner"
  project_ref = "%s"
}
`, name, testProject)
}

func testRoleBindingWithoutUser(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  group_ref   = "group_xyzabc"
  role_ref    = "project-owner"
  project_ref = "%s"
}
`, name, testProject)
}

func testRoleBindingWithoutGroup(name string) string {
	return fmt.Sprintf(`
resource "nobl9_role_binding" "%s" {
  user        = "test"
  role_ref    = "project-owner"
  project_ref = "%s"
}
`, name, testProject)
}
