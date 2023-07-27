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
		{"test-amazonprometheus", testAmazonPrometheusAgent},
		{"test-appd", testAppDynamicsAgent},
		{"test-bigquery", testBigQueryAgent},
		{"test-cloudwatch", testCloudWatchAgent},
		{"test-ddog", testDatadogAgent},
		{"test-dynatrace", testDynatraceAgent},
		{"test-elasticsearch", testElasticsearchAgent},
		{"test-gcm", testGoogleCloudMonitoringAgent},
		{"test-grafanaloki", testGrafanaLokiAgent},
		{"test-graphite", testGraphiteAgent},
		{"test-influxdb", testInfluxDBAgent},
		{"test-instana", testInstanaAgent},
		{"test-lightstep", testLightstepAgent},
		{"test-newrelic", testNewrelicAgent},
		{"test-opentsdb", testOpenTSDBAgent},
		{"test-pingdom", testPingdomAgent},
		{"test-prometheus", testPrometheusAgent},
		{"test-redshift", testRedshiftAgent},
		{"test-splunk", testSplunkAgent},
		{"test-splunkobs", testSplunkObservabilityAgent},
		{"test-sumologic", testSumoLogicAgent},
		{"test-thousandeyes", testThousandEyesAgent},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_agent", n9api.ObjectAgent),
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

func testAmazonPrometheusAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "amazon_prometheus"
  amazon_prometheus_config {
    url = "http://web.net"
    region = "eu-central-1"
  }
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testAppDynamicsAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "appdynamics"
  appdynamics_config {
    url = "http://web.net"
  }
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testBigQueryAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
 name      = "%s"
 project   = "%s"
 source_of = ["Metrics", "Services"]
 agent_type = "bigquery"
 release_channel = "stable"
 query_delay {
  unit = "Minute"
  value = 6
}
}
`, name, name, testProject)
}

func testCloudWatchAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "cloudwatch"
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testDatadogAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "datadog"
  release_channel = "stable"
  datadog_config {
    site = "eu"
  }
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testDynatraceAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "dynatrace"
  dynatrace_config {
    url = "http://web.net"
  }
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testElasticsearchAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "elasticsearch"
  elasticsearch_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testGoogleCloudMonitoringAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "gcm"
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testGrafanaLokiAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "grafana_loki"
  grafana_loki_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testGraphiteAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "graphite"
  graphite_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testInfluxDBAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "influxdb"
  influxdb_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testInstanaAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "instana"
  instana_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testLightstepAgent(name string) string {
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
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testNewrelicAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "newrelic"
  newrelic_config {
    account_id = 1234
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testOpenTSDBAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "opentsdb"
  opentsdb_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testPingdomAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "pingdom"
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testPrometheusAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "prometheus"
  prometheus_config {
	url = "http://web.net"
	}
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testRedshiftAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "redshift"
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testSplunkAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "splunk"
  splunk_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testSplunkObservabilityAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "splunk_observability"
  splunk_observability_config {
    realm = "eu"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testSumoLogicAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics"]
  agent_type = "sumologic"
  sumologic_config {
    url = "http://web.net"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testThousandEyesAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  agent_type = "thousandeyes"
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}
