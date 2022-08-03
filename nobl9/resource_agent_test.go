package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Agent(t *testing.T) {
	// we dont' currently support creating Agents via the API
	// so skipping this for now
	t.SkipNow()
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
		{"test-opentsdb", testOpenTSDBConfig},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestory("nobl9_agent", n9api.ObjectAgent),
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

//nolint:unused
func testPrometheusConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "prometheus"
  prometheus_config {
	url = "http://web.net"
	}
}
`, name, name, testProject)
}

//nolint:unused
func testDatadogConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "datadog"
  datadog_config {
    site = "eu"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testNewrelicConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "newrelic"
  newrelic_config {
    account_id = 1234
  }
}
`, name, name, testProject)
}

//nolint:unused
func testAppDynamicsConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "appdynamics"
  appdynamics_config {
	url = "http://web.net"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testSplunkConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "splunk"
  splunk_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testLightstepConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "lightstep"
  lightstep_config {
    organization = "acme"
	project		 = "project1"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testSplunkObservabilityConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "splunk_observability"
  splunk_observability_config {
    realm = "eu"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testDynatraceConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "dynatrace"
  dynatrace_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testThousandEyesConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "thousandeyes"
}
`, name, name, testProject)
}

//nolint:unused
func testGraphiteConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "graphite"
  graphite_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

//nolint:unused
func testBigQueryConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "bigquery"
}
`, name, name, testProject)
}

//nolint:unused
func testOpenTSDBConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "opentsdb"
  opentsdb_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}
