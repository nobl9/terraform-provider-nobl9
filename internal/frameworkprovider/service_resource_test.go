package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccExampleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: newServiceResource(t, serviceResourceTemplateModel{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nobl9_service.test", "name", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "nobl9_service.this",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: newServiceResource(t, serviceResourceTemplateModel{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestRenderServiceResourceTemplate(t *testing.T) {
	t.Parallel()

	actual := newServiceResource(t, serviceResourceTemplateModel{
		ResourceName: "this",
		ServiceResourceModel: ServiceResourceModel{
			Name:        "service",
			DisplayName: types.StringValue("Service"),
			Project:     "default",
			Description: types.StringValue("Example service"),
			Annotations: map[string]string{"key": "value"},
			Labels: Labels{
				{Key: "team", Values: []string{"green"}},
				{Key: "env", Values: []string{"prod", "dev"}},
			},
		},
	})

	expected := `resource "nobl9_service" "this" {
  name = "service"
  display_name = "Service"
  project = "default"
  annotations = {
    key = "value",
  }
  label {
    key = "team"
    values = [
      "green",
    ]
  }
  label {
    key = "env"
    values = [
      "prod",
      "dev",
    ]
  }
  description = "Example service"
}
`

	assert.Equal(t, expected, actual)
}

type serviceResourceTemplateModel struct {
	ResourceName string
	ServiceResourceModel
}

func newServiceResource(t *testing.T, model serviceResourceTemplateModel) string {
	return executeTemplate(t, "service_resource.hcl.tmpl", model)
}
