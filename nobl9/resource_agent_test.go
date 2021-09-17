package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Agent(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-prometheus", testPrometheusConfig},
		{"test-ddog", testDatadogConfig},
		{"test-newrelic", testNewrelicConfig},
		{"test-appd", testAppDynamicsConfig},
		{"test-splunk", testSplunkConfig},
		{"test-lightstep", testLightstepConfig},
		{"test-splunkobs", testSplunkObservabilityConfig},
		{"test-dynatrace", testDynatraceConfig},
		{"test-thousandeyes", testThousandEyesConfig},
		{"test-graphite", testGraphiteConfig},
		{"test-bigquery", testBigQueryConfig},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      DestroyFunc("nobl9_agent", n9api.ObjectAgent),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_agent." + tc.name),
					},
				},
			})
		})
	}
}

func testPrometheusConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  prometheus {
	url = "http://web.net"
	}
}
`, name, name, testProject)
}

func testDatadogConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  datadog {
    site = "eu"
  }
}
`, name, name, testProject)
}

func testNewrelicConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  newrelic {
    account_id = 1234
  }
}
`, name, name, testProject)
}

func testAppDynamicsConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  appdynamics {
	url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testSplunkConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  splunk {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testLightstepConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  lightstep {
    organization = "acme"
	project		 = "project1"
  }
}
`, name, name, testProject)
}

func testSplunkObservabilityConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  splunk_observability {
    realm = "eu"
  }
}
`, name, name, testProject)
}

func testDynatraceConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  dynatrace {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testThousandEyesConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  thousandeyes {}
}
`, name, name, testProject)
}

func testGraphiteConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  graphite {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testBigQueryConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  bigquery {}
}
`, name, name, testProject)
}
