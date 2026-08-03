package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResources_unknownValuesDuringPlanning(t *testing.T) {
	t.Parallel()
	testAccSetup(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "name" {}

resource "nobl9_project" "test" {
  name = terraform_data.name.id
}

resource "nobl9_service" "test" {
  name    = terraform_data.name.id
  project = "default"
}

resource "nobl9_slo" "test" {
  name             = terraform_data.name.id
  project          = "default"
  service          = "service"
  budgeting_method = "Occurrences"

  indicator {
    name    = "indicator"
    project = "default"
    kind    = "Agent"
  }

  objective {
    name   = "objective"
    op     = "lt"
    target = 0.7
    value  = 1

    raw_metric {
      query {
        appdynamics {
          application_name = "application"
          metric_path      = "metric"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }
}
`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
