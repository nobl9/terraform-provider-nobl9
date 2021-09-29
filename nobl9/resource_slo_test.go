package nobl9

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9SLO(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-prometheus", testPrometheusSLO},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      DestroyFunc("nobl9_slo", n9api.ObjectSLO),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_slo." + tc.name),
					},
				},
			})
		})
	}
}

func testPrometheusSLO(name string) string {
	config := `
resource "nobl9_service" ":name-service" {
	name = ":name-service"
	project = ":project"
}

resource "nobl9_slo" ":name" {
  name      = ":name"
  project   = ":project"
  service = "%s-service"

  budgeting_method = "budgeting_method"

  objective {
	target = 0.7
	value = 1
	op = "lt"
  }

  time_window {
	  count = 10
	  is_rolling = true
	  period {
		begin = "2021-09-29T10:18:39Z"
		end = "2021-09-29T10:28:39Z"
	  }
	  unit = "Minute"
  }

  indicator {
	name = "ind1"
	project = ":project"
    raw_metric {
	  prometheus_metric {
	    promql = "1.0"
 	  }
	}

	// time_windows {
	// 	count = 
	// 	unit = 
	// }
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}
