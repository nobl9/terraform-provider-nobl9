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
  project   = ":testProject"
  service = "%s-service"
  
  indicator {
	name = "ind1"
	raw_metrics {
		promql = "test-query"
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
