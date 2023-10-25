package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9Project(t *testing.T) {
	name := "test-project"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: ProviderFactory(),
		CheckDestroy:      CheckDestroy("nobl9_project", manifest.KindProject),
		Steps: []resource.TestStep{
			{
				Config: testProjectConfig(name),
				Check:  CheckObjectCreated("nobl9_project." + name),
			},
			{
				Config: testProjectConfigNoLabels(name),
				Check:  CheckObjectCreated("nobl9_project." + name),
			},
		},
	})
}

func TestAcc_NewNobl9ProjectReference(t *testing.T) {
	name := "test-project"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: ProviderFactory(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			CheckDestroy("nobl9_agent", manifest.KindAgent),
			CheckDestroy("nobl9_project", manifest.KindProject),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "nobl9_project" "%s" {
					  display_name = "%s"
					  name         = "%s"
					  description  = "A terraform project"
					}
					resource "nobl9_agent" "%s" {
					 name      = "%s"
					 project   = nobl9_project.%s.name
					 source_of = ["Metrics", "Services"]
					 agent_type = "bigquery"
					 release_channel = "stable"
					 query_delay {
						unit = "Second"
						value = 0
					  }
					}
				`, name, name, name, name, name, name),
				Check: resource.ComposeTestCheckFunc(
					CheckObjectCreated("nobl9_project."+name),
					CheckObjectCreated("nobl9_agent."+name),
				),
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

  label {
    key    = "team"
    values = ["green", "sapphire"]
  }

  label {
    key    = "env"
    values = ["dev", "staging", "prod"]
  }
}
`, name, name, name)
}

func testProjectConfigNoLabels(name string) string {
	return fmt.Sprintf(`
resource "nobl9_project" "%s" {
  name         = "%s"
  display_name = "%s"
  description  = "A terraform project"
}
`, name, name, name)
}
