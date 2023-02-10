package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Direct(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-appd", testAppDynamicsDirect},
		{"test-bigquery", testBigQueryDirect},
		{"test-cloudwatch", testCloudWatchDirect},
		{"test-ddog", testDatadogDirect},
		{"test-dynatrace", testDynatraceDirect},
		{"test-gcm", testGoogleCloudMonitoringDirect},
		{"test-influxdb", testInfluxDBDirect},
		{"test-instana", testInstanaDirect},
		{"test-lightstep", testLightstepDirect},
		{"test-newrelic", testNewrelicDirect},
		{"test-pingdom", testPingdomDirect},
		{"test-redshift", testRedshiftDirect},
		{"test-splunk", testSplunkDirect},
		{"test-splunkobs", testSplunkObservabilityDirect},
		{"test-sumologic", testSumoLogicDirect},
		{"test-thousandeyes", testThousandEyesDirect},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_direct", n9api.ObjectDirect),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check:  CheckObjectCreated("nobl9_direct." + tc.name),
					},
				},
			})
		})
	}
}

func testAppDynamicsDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "appdynamics"
  appdynamics_config {
	url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testBigQueryDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
 name      = "%s"
 project   = "%s"
 source_of = ["Metrics", "Services"]
 direct_type = "bigquery"
}
`, name, name, testProject)
}

func testCloudWatchDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "cloudwatch"
}
`, name, name, testProject)
}

func testDatadogDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "datadog"
  datadog_config {
    site = "eu"
	api_key = "secret"
	application_key = "secret"
  }
}
`, name, name, testProject)
}

func testDynatraceDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "dynatrace"
  dynatrace_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testElasticsearchDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "elasticsearch"
  elasticsearch_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testGoogleCloudMonitoringDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "gcm"
}
`, name, name, testProject)
}

func testGrafanaLokiDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "grafana_loki"
  grafana_loki_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testGraphiteDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "graphite"
  graphite_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testInfluxDBDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "influxdb"
  influxdb_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testInstanaDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "instana"
  instana_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testLightstepDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "lightstep"
  lightstep_config {
    organization = "acme"
	project		 = "project1"
  }
}
`, name, name, testProject)
}

func testNewrelicDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "newrelic"
  newrelic_config {
    account_id = 1234
  }
}
`, name, name, testProject)
}

func testOpenTSDBDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "opentsdb"
  opentsdb_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testPingdomDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "pingdom"
}
`, name, name, testProject)
}

func testPrometheusDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "prometheus"
  prometheus_config {
	url = "http://web.net"
	}
}
`, name, name, testProject)
}

func testRedshiftDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "redshift"
}
`, name, name, testProject)
}

func testSplunkDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "splunk"
  splunk_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testSplunkObservabilityDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "splunk_observability"
  splunk_observability_config {
    realm = "eu"
  }
}
`, name, name, testProject)
}

func testSumoLogicDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics"]
  direct_type = "sumologic"
  sumologic_config {
    url = "http://web.net"
  }
}
`, name, name, testProject)
}

func testThousandEyesDirect(name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct" "%s" {
  name      = "%s"
  project   = "%s"
  source_of = ["Metrics", "Services"]
  direct_type = "thousandeyes"
}
`, name, name, testProject)
}
