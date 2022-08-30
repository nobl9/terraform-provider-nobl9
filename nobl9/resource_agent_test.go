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
		{"test-amazonprometheus", testAmazonPrometheusConfig},
		{"test-appd", testAppDynamicsConfig},
		{"test-bigquery", testBigQueryConfig},
		{"test-cloudwatch", testCloudWatchConfig},
		{"test-ddog", testDatadogConfig},
		{"test-dynatrace", testDynatraceConfig},
		{"test-elasticsearch", testElasticsearchConfig},
		{"test-gcm", testGoogleCloudMonitoringConfig},
		{"test-grafanaloki", testGrafanaLokiConfig},
		{"test-graphite", testGraphiteConfig},
		{"test-influxdb", testInfluxDBConfig},
		{"test-instana", testInstanaConfig},
		{"test-lightstep", testLightstepConfig},
		{"test-newrelic", testNewrelicConfig},
		{"test-opentsdb", testOpenTSDBConfig},
		{"test-pingdom", testPingdomConfig},
		{"test-prometheus", testPrometheusConfig},
		{"test-redshift", testRedshiftConfig},
		{"test-splunk", testSplunkConfig},
		{"test-splunkobs", testSplunkObservabilityConfig},
		{"test-sumologic", testSumoLogicConfig},
		{"test-thousandeyes", testThousandEyesConfig},
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

func testAmazonPrometheusConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "amazonprometheus"
  amazonprometheus_config {
	url = "http://web.net"
	region = "eu-central-1"
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
  agent_type = "appdynamics"
  appdynamics_config {
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
 agent_type = "bigquery"
}
`, name, name, testProject)
}

func testCloudWatchConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "cloudwatch"
}
`, name, name, testProject)
}

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

func testElasticsearchConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "elasticsearch"
  elasticsearch_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testGoogleCloudMonitoringConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "gcm"
}
`, name, name, testProject)
}

func testGrafanaLokiConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "grafanaloki"
  grafanaloki_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

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

func testInfluxDBConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "influxdb"
  influxdb_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testInstanaConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "instana"
  instana_config {
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
  agent_type = "lightstep"
  lightstep_config {
    organization = "acme"
	project		 = "project1"
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
  agent_type = "newrelic"
  newrelic_config {
    account_id = 1234
  }
}
`, name, name, testProject)
}

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

func testPingdomConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "pingdom"
}
`, name, name, testProject)
}

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

func testRedshiftConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "redshift"
}
`, name, name, testProject)
}

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

func testSumoLogicConfig(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics"]
  agent_type = "sumologic"
  sumologic_config {
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
  agent_type = "thousandeyes"
}
`, name, name, testProject)
}
