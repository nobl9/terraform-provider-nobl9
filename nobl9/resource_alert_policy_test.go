package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9AlertPolicy(t *testing.T) {
	name := "test-alert-policy"
	config := fmt.Sprintf(`
resource "nobl9_alert_policy" "%s" {
  name       = "%s"
  project    = "%s"
  severity   = "Medium"

  condition {
	  measurement = "burnedBudget"
	  value 	  = 0.9
	}

  condition {
	  measurement = "averageBurnRate"
	  value 	  = 3
	  lasts_for	  = "1m"
	}

  condition {
	  measurement  = "timeToBurnBudget"
	  value_string = "1h"
	  lasts_for	   = "300s"
	}
  
  integration {
	project = "%s"
	name	= "webhook"
  }
}
`, name, name, testProject, testProject)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: ProviderFactory(),
		CheckDestroy:      DestroyFunc("nobl9_alert_policy", n9api.ObjectAlertPolicy),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  CheckObjectCreated("nobl9_alert_policy." + name),
			},
		},
	})
}
