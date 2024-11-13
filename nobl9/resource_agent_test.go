package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9Agent(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-amazonprometheus", testAmazonPrometheusAgent},
		{"test-amazonprometheus-historical-data-retrieval", testAmazonPrometheusAgentHistoricalDataRetrieval},
		{"test-appdynamics", testAppDynamicsAgent},
		{"test-azuremonitor", testAzureMonitorAgent},
		{"test-bigquery", testBigQueryAgent},
		{"test-cloudwatch", testCloudWatchAgent},
		{"test-ddog", testDatadogAgent},
		{"test-dynatrace", testDynatraceAgent},
		{"test-dynatrace-without-query-delay", testDynatraceAgentWithoutQueryDelay},
		{"test-elasticsearch", testElasticsearchAgent},
		{"test-gcm", testGoogleCloudMonitoringAgent},
		{"test-grafanaloki", testGrafanaLokiAgent},
		{"test-graphite", testGraphiteAgent},
		{"test-honeycomb", testHoneycombAgent},
		{"test-influxdb", testInfluxDBAgent},
		{"test-instana", testInstanaAgent},
		{"test-lightstep", testLightstepAgent},
		{"test-logicmonitor", testLogicMonitorAgent},
		{"test-newrelic", testNewrelicAgent},
		{"test-opentsdb", testOpenTSDBAgent},
		{"test-pingdom", testPingdomAgent},
		{"test-prometheus", testPrometheusAgent},
		{"test-redshift", testRedshiftAgent},
		{"test-splunk", testSplunkAgent},
		{"test-splunk-observability", testSplunkObservabilityAgent},
		{"test-sumologic", testSumoLogicAgent},
		{"test-thousandeyes", testThousandEyesAgent},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_agent", manifest.KindAgent),
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

func testAmazonPrometheusAgentHistoricalDataRetrieval(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
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
  historical_data_retrieval {
	default_duration {
		unit = "Minute"
    	value = 10
	}
	max_duration {
		unit = "Hour"
		value = 19
	}
	triggered_by_slo_creation {
		unit = "Hour"
		value = 19
	}
	triggered_by_slo_edit {
		unit = "Hour"
		value = 19
	}
  }
}
`, name, name, testProject)
}

func testAppDynamicsAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
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

func testAzureMonitorAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  agent_type = "azure_monitor"
  azure_monitor_config {
    tenant_id = "40ad1f5f-7025-4056-9b90-9f49617423ac"
  }
  release_channel = "beta"
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
  agent_type = "cloudwatch"
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testCloudWatchDirectBeta(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_cloudwatch" "%s" {
  name      = "%s"
  project   = "%s"
  release_channel = "beta"
  role_arn = "test"
  historical_data_retrieval {
   default_duration {
	  unit  = "Day"
	  value = 0
	}
	max_duration {
	  unit  = "Day"
	  value = 15
	}
	triggered_by_slo_creation {
	  unit = "Hour"
      value = 19
	}
	triggered_by_slo_edit {
	  unit = "Hour"
	  value = 19
	}
  }
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
  agent_type = "datadog"
  release_channel = "stable"
  datadog_config {
    site = "datadoghq.eu"
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

func testDynatraceAgentWithoutQueryDelay(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  agent_type = "dynatrace"
  dynatrace_config {
    url = "http://web.net"
  }
  release_channel = "stable"
}
`, name, name, testProject)
}

func testElasticsearchAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
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
  agent_type = "gcm"
  release_channel = "beta"
  historical_data_retrieval {
	default_duration {
      unit = "Minute"
      value = 10
	}
	max_duration {
      unit = "Hour"
      value = 19
	}
	triggered_by_slo_creation {
	  unit = "Hour"
	  value = 19
	}
	triggered_by_slo_edit {
	  unit = "Hour"
	  value = 19
	}
  }
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

func testHoneycombAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
	name      = "%s"
	project   = "%s"
	agent_type = "honeycomb"
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
  agent_type = "lightstep"
  lightstep_config {
    organization = "acme"
    project		 = "project1"
    url			 = "https://api.lightstep.com"
  }
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}

func testLogicMonitorAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name      = "%s"
  project   = "%s"
  agent_type = "logic_monitor"
  logic_monitor_config {
    account = "account-name"
  }
  release_channel = "beta"
  historical_data_retrieval {
	default_duration {
      unit = "Minute"
      value = 10
	}
	max_duration {
      unit = "Hour"
      value = 19
	}
	triggered_by_slo_creation {
	  unit = "Hour"
	  value = 19
	}
	triggered_by_slo_edit {
	  unit = "Hour"
	  value = 19
	}
  }
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
  agent_type = "thousandeyes"
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, name, name, testProject)
}
