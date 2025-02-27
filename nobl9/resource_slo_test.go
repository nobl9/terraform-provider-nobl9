package nobl9

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestAcc_Nobl9SLO(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-amazonprometheus", testAmazonPrometheusSLO},
		{"test-appdynamics", testAppdynamicsSLO},
		{"test-azure-monitor-metrics", testAzureMonitorMetricsSLO},
		{"test-azure-monitor-logs", testAzureMonitorLogsSLO},
		{"test-bigquery", testBigQuerySLO},
		{"test-cloudwatch-with-json", testCloudWatchWithJSON},
		{"test-cloudwatch-with-sql", testCloudWatchWithSQL},
		{"test-cloudwatch-with-stat", testCloudWatchWithStat},
		{"test-cloudwatch-with-stat-and-cross-account", testCloudWatchWithStatAndCrossAccount},
		{"test-cloudwatch-with-bad-over-total", testCloudWatchWithBadOverTotal},
		{"test-composite-occurrences-deprecated", testCompositeSLOOccurrencesDeprecated},
		{"test-composite-time-slices-deprecated", testCompositeSLOTimeSlicesDeprecated},
		{"test-composite-occurrences", testCompositeSLOOccurrences},
		{"test-composite-time-slices", testCompositeSLOTimeSlices},
		{"test-composite-with-value", testCompositeSLOValueZeroBackwardCompatibility},
		{"test-datadog", testDatadogSLO},
		{"test-dynatrace", testDynatraceSLO},
		{"test-google-cloud-monitoring", testGoogleCloudMonitoringPromQLSLO},
		{"test-grafanaloki", testGrafanaLokiSLO},
		{"test-graphite", testGraphiteSLO},
		{"test-influxdb", testInfluxDBSLO},
		{"test-instana-infra", testInstanaInfrastructureSLO},
		{"test-instana-app", testInstanaApplicationSLO},
		{"test-lightstep", testLightstepSLO},
		{"test-logic-monitor", testLogicMonitorDeviceMetricsSLO},
		{"test-logic-monitor-website", testLogicMonitorWebsiteMetricsSLO},
		{"test-multiple-ap", testMultipleAlertPolicies},
		{"test-newrelic", testNewRelicSLO},
		{"test-opentsdb", testOpenTSDBSLO},
		{"test-pingdom", testPingdomSLO},
		{"test-prom-full", testPrometheusSLOFull},
		{"test-prom-with-ap", testPrometheusSLOWithAlertPolicy},
		{"test-prom-with-attachments-deprecated", testPrometheusWithAttachmentsDeprecated},
		{"test-prom-with-attachment", testPrometheusWithAttachment},
		{"test-prom-with-countmetrics", testPrometheusSLOWithCountMetrics},
		{"test-prom-with-multiple-objectives", testPrometheusSLOWithMultipleObjectives},
		{"test-prom-with-raw-metric-in-objective", testPrometheusSLOWithRawMetricInObjective},
		{"test-prom-with-time-slices", testPrometheusSLOWithTimeSlices},
		{"test-prometheus", testPrometheusSLO},
		{"test-redshift", testRedshiftSLO},
		{"test-splunk", testSplunkSLO},
		{"test-splunk-observability", testSplunkObservabilitySLO},
		{"test-splunk-single-query", testSplunkSingleQuerySLO},
		{"test-sumologic", testSumoLogicSLO},
		{"test-thousandeyes", testThousandeyesSLO},
		{"test-anomaly-config-same-project", testAnomalyConfigNoDataSameProject},
		{"test-anomaly-config-different-project", testAnomalyConfigNoDataDifferentProject},
		{"test-max-one-primary-objective", testMaxOnePrimaryObjective},
		{"test-no-primary-objective", testNoPrimaryObjective},
		{"test-metadata-annotations", testMetadataAnnotations},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_slo", manifest.KindSLO),
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

func TestAcc_Nobl9SLOErrors(t *testing.T) {
	cases := []struct {
		name         string
		configFunc   func(string) string
		errorMessage string
	}{
		{"test-prom-with-conflict-attachments",
			testPrometheusWithAttachmentsConflict,
			`attachments": conflicts with attachment`,
		},
		{"test-metric-spec-required",
			testMetricSpecRequired,
			`one of \[goodTotal, total\] properties must be set, none was provided`,
		},
		{"test-more-than-one-primary-objective",
			testMoreThanOnePrimaryObjective,
			`there can be max 1 primary objective`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_slo.", manifest.KindSLO),
				Steps: []resource.TestStep{
					{
						Config:      tc.configFunc(tc.name),
						ExpectError: regexp.MustCompile(tc.errorMessage),
					},
				},
			})
		})
	}
}

func testAmazonPrometheusSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testAmazonPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        amazon_prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testAppdynamicsSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testAppDynamicsAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        appdynamics {
          application_name = "my_app"
          metric_path = "End User Experience|App|End User Response Time 95th percentile (ms)"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

// nolint: lll
func testAzureMonitorMetricsSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testAzureMonitorAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        azure_monitor {
          data_type = "metrics"
          resource_id = "/subscriptions/9c26f90e-24bb-4d20-a648-c6e3e1cde26a/resourceGroups/azure-monitor-test-sources/providers/microsoft.insights/components/n9-web-app"
          metric_namespace = ""
          metric_name = "requests/duration"
          aggregation = "Avg"
          dimensions {
		  	name = "request/resultCode"
			value = "200"
         }
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

// nolint: lll
func testAzureMonitorLogsSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testAzureMonitorAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        azure_monitor {
          data_type = "logs"
          workspace {
			subscription_id = "9c26f90e-24bb-4d20-a648-c6e3e1cde26a"
			resource_group = "azure-monitor-test-sources"
			workspace_id = "e5da9ba8-cb8f-437e-aec0-61d21aab2bcd"
          }
          kql_query = "AppRequests | where AppRoleName == \"n9-web-app\" | summarize n9_value = avg(DurationMs) by bin(TimeGenerated, 15s) | project n9_time = TimeGenerated, n9_value"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testBigQuerySLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testBigQueryAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        bigquery {
          project_id = "project"
          location = "EU"
          query = <<-EOT
			SELECT response_time AS n9value, created AS n9date
			FROM 'project.metrics.http_response'
			WHERE date_col BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to)
			EOT
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCloudWatchWithJSON(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testCloudWatchAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        cloudwatch {
		region = "eu-central-1"
		json = jsonencode(
		[
			{
				"Id": "e1",
				"Expression": "m1 / m2",
				"Period": 60
			},
			{
				"Id": "m1",
				"MetricStat": {
					"Metric": {
						"Namespace": "AWS/ApplicationELB",
						"MetricName": "HTTPCode_Target_2XX_Count",
						"Dimensions": [
							{
								"Name": "name1",
								"Value": "name2"
							}
						]
					},
					"Period": 60,
					"Stat": "SampleCount"
				},
				"ReturnData": false
			},
			{
				"Id": "m2",
				"MetricStat": {
					"Metric": {
						"Namespace": "AWS/ApplicationELB",
						"MetricName": "RequestCount",
						"Dimensions": [
							{
								"Name": "name2",
								"Value": "value2"
							}
						]
					},
					"Period": 60,
					"Stat": "SampleCount"
				},
				"ReturnData": false
			}
		])
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCloudWatchWithSQL(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testCloudWatchAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        cloudwatch {
		  region = "eu-central-1"
		  sql = "SELECT AVG(CPUUtilization)FROM \"AWS/EC2\""
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCloudWatchWithStat(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testCloudWatchAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        cloudwatch {
		region = "eu-central-1"
		namespace = "namespace"
		metric_name = "metric_name"
		stat        = "Sum"
		dimensions {
		  name  = "name1"
			value = "value1"
		}
		dimensions {
			name  = "name2"
			value = "value3"
		}
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCloudWatchWithBadOverTotal(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testCloudWatchAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    count_metrics {
      incremental = true
      bad {
        cloudwatch {
          region = "eu-central-1"
          namespace = "namespace"
          metric_name = "metric_name"
          stat        = "Sum"
          dimensions {
            name  = "name1"
            value = "value1"
          }
          dimensions {
            name  = "name2"
            value = "value3"
          }
        }
      }
      total {
        cloudwatch {
          region = "eu-central-1"
          namespace = "namespace"
          metric_name = "metric_name"
          stat        = "Sum"
          dimensions {
            name  = "name1"
            value = "value1"
          }
          dimensions {
            name  = "name2"
            value = "value3"
          }
        }
      }
    }
  }


  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCloudWatchWithStatAndCrossAccount(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testCloudWatchDirectBeta(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        cloudwatch {
		region = "eu-central-1"
		account_id = "123456789012"
		namespace = "namespace"
		metric_name = "metric_name"
		stat        = "Sum"
		dimensions {
		  name  = "name1"
			value = "value1"
		}
		dimensions {
			name  = "name2"
			value = "value3"
		}
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_direct_cloudwatch.:agentName.name
    project = ":project"
    kind    = "Direct"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCompositeSLOOccurrencesDeprecated(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  composite {
    burn_rate_condition {
      op    = "gt"
      value = 1
    }
    target = 0.5
  }

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  objective {
    display_name = "obj2"
    name         = "tf-objective-2"
    target       = 0.8
    value        = 1.5
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCompositeSLOTimeSlicesDeprecated(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  budgeting_method = "Timeslices"

  composite {
    target = 0.5
  }

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 15
    time_slice_target = 0.7
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  objective {
    display_name      = "obj2"
    name              = "tf-objective-2"
    target            = 0.5
	value             = 10
	time_slice_target = 0.5
	op                = "lt"
	raw_metric {
	  query {
	    prometheus {
	      promql = "1.0"
	    }
	  }
	}
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCompositeSLOOccurrences(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	var sloDependencyName = name + "-dependency-slo-1"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) +
			testCompositeDependencySLO(serviceName, agentName, sloDependencyName) + `
resource "nobl9_slo" ":name" {
 name         = ":name"
 display_name = ":name"
 project      = ":project"
 service      = nobl9_service.:serviceName.name


 depends_on = [nobl9_slo.:sloDependencyName]

 label {
  key = "team"
  values = ["green","sapphire"]
 }

 label {
  key = "env"
  values = ["dev", "staging", "prod"]
 }

 budgeting_method = "Occurrences"

 objective {
   display_name = "obj1"
   name         = "tf-objective-1"
   target       = 0.7
   composite {
     max_delay = "45m"
     components {
       objectives {
         composite_objective {
           project      = ":project"
           slo          = ":sloDependencyName"
           objective    = "objective-1"
           weight       = 0.8
           when_delayed = "CountAsGood"
         }
		 composite_objective {
           project      = ":project"
           slo          = ":sloDependencyName"
           objective    = "objective-2"
           weight       = 1.0
           when_delayed = "CountAsBad"
         }
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
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)
	config = strings.ReplaceAll(config, ":sloDependencyName", sloDependencyName)

	return config
}

func testCompositeSLOTimeSlices(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	var sloDependencyName = name + "-dependency-slo-2"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) +
			testCompositeDependencySLO(serviceName, agentName, sloDependencyName) + `
resource "nobl9_slo" ":name" {
 name         = ":name"
 display_name = ":name"
 project      = ":project"
 service      = nobl9_service.:serviceName.name

 depends_on = [nobl9_slo.:sloDependencyName]

 budgeting_method = "Timeslices"

 objective {
   display_name = "obj1"
   name         = "tf-objective-1"
   target       = 0.7
   time_slice_target = 0.7
   composite {
     max_delay = "45m"
     components {
       objectives {
         composite_objective {
           project      = ":project"
           slo          = ":sloDependencyName"
           objective    = "objective-1"
           weight       = 0.8
           when_delayed = "CountAsGood"
         }
		 composite_objective {
           project      = ":project"
           slo          = ":sloDependencyName"
           objective    = "objective-2"
           weight       = 1.0
           when_delayed = "CountAsBad"
         }
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
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)
	config = strings.ReplaceAll(config, ":sloDependencyName", sloDependencyName)

	return config
}

func testCompositeSLOValueZeroBackwardCompatibility(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	var sloDependencyName = name + "-dependency-slo-1"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) +
			testCompositeDependencySLO(serviceName, agentName, sloDependencyName) + `
resource "nobl9_slo" ":name" {
 name         = ":name"
 display_name = ":name"
 project      = ":project"
 service      = nobl9_service.:serviceName.name


 depends_on = [nobl9_slo.:sloDependencyName]

 label {
  key = "team"
  values = ["green","sapphire"]
 }

 label {
  key = "env"
  values = ["dev", "staging", "prod"]
 }

 budgeting_method = "Occurrences"

 objective {
   display_name = "obj1"
   name         = "tf-objective-1"
   target       = 0.7
   value        = 0
   composite {
     max_delay = "45m"
     components {
       objectives {
         composite_objective {
           project      = ":project"
           slo          = ":sloDependencyName"
           objective    = "objective-1"
           weight       = 0.8
           when_delayed = "CountAsGood"
         }
		 composite_objective {
           project      = ":project"
           slo          = ":sloDependencyName"
           objective    = "objective-2"
           weight       = 1.0
           when_delayed = "CountAsBad"
         }
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
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)
	config = strings.ReplaceAll(config, ":sloDependencyName", sloDependencyName)

	return config
}

func testDatadogSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testDatadogAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        datadog {
          query = "avg:system.cpu.user{cluster_name:main}"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testDynatraceSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testDynatraceAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        dynatrace {
          metric_selector = <<-EOT
			builtin:synthetic.http.duration.geo:filter(
			and(in("dt.entity.http_check",entitySelector("type(http_check),entityName(~"API Sample~")")),
				in("dt.entity.synthetic_location",entitySelector("type(synthetic_location),entityName(~"N. California~")")))
			):splitBy("dt.entity.http_check","dt.entity.synthetic_location"):avg:auto:sort(value(avg,descending)):limit(20)
			EOT
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testGoogleCloudMonitoringPromQLSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testGoogleCloudMonitoringAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        gcm {
          project_id = "project1"
		  promql = "sum(rate(http_requests_total{job=\"api-server\"}[5m]))"
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
      kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testGrafanaLokiSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testGrafanaLokiAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        grafana_loki {
          logql = <<-EOT
			sum(
				sum_over_time(
					{topic="topic", consumergroup="group", cluster="main"} |= "kafka_consumergroup_lag" |
					logfmt |
					line_format "{{.kafka_consumergroup_lag}}" |
					unwrap kafka_consumergroup_lag [1m]
			)
			)
			EOT
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
      kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testGraphiteSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testGraphiteAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        graphite {
          metric_path = "TODO"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
      kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testInfluxDBSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testInfluxDBAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
	  	influxdb {
		  query = <<-EOT
			from(bucket: "integrations")
			|> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))
			|> aggregateWindow(every: 15s, fn: mean, createEmpty: false)
			|> filter(fn: (r) => r["_measurement"] == "internal_write")
			|> filter(fn: (r) => r["_field"] == "write_time_ns")'
		    EOT
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
      kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testInstanaInfrastructureSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testInstanaAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        instana {
          metric_type = "infrastructure"
		  infrastructure {
		    metric_retrieval_method = "query"
		    metric_id               = "outstanding_requests"
		    plugin_id               = "zooKeeper"
		    query                   = "entity.selfType:zookeeper"
		  }
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
      kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testInstanaApplicationSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testInstanaAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        instana {
          metric_type = "application"
		  application {
		    metric_id         = "latency"
		    aggregation       = "p99"
		    group_by {
		  	  tag                  = "endpointname"
			  tag_entity           = "DESTINATION"
			  tag_second_level_key = ""
		    }
		    include_internal  = false
		    include_synthetic = false
		    api_query = <<-EOT
			{
				"type": "EXPRESSION",
				"logicalOperator": "AND",
				"elements": [
					{
						"type": "TAG_FILTER",
						"name": "service.name",
						"operator": "EQUALS",
						"entity": "DESTINATION",
						"value": "master"
					},
					{
						"type": "TAG_FILTER",
						"name": "call.type",
						"operator": "EQUALS",
						"entity": "NOT_APPLICABLE",
						"value": "HTTP"
					}
				]
			}
			EOT
		  }
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
      kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testLightstepSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testLightstepAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        lightstep {
          stream_id = "id"
          type_of_data = "latency"
          percentile = 95
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testLogicMonitorDeviceMetricsSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testLogicMonitorAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        logic_monitor {
          query_type = "device_metrics"
          device_data_source_instance_id = "775430648"
          graph_id = "11354"
          line = "AVERAGE"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testLogicMonitorWebsiteMetricsSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testLogicMonitorAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        logic_monitor {
          query_type = "website_metrics"
          website_id = "1"
          checkpoint_id = "775430648"
          graph_name = "responseTime"
          line = "AVERAGE"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testMultipleAlertPolicies(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) +
			testAlertPolicyWithoutAnyAlertMethod(name+"-fast") +
			testAlertPolicyWithoutAnyAlertMethod(name+"-slow") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }

  alert_policies = [
    nobl9_alert_policy.:name-slow.name,
    nobl9_alert_policy.:name-fast.name
    ]
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testNewRelicSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testNewrelicAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        newrelic {
          nrql = "SELECT average(duration * 1000) FROM Transaction TIMESERIES"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"

  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testOpenTSDBSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testOpenTSDBAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        opentsdb {
          query = "m=none:{{.N9RESOLUTION}}-avg-zero:cpu{cpu.usage=core.1}"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPingdomSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPingdomAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
	  	pingdom {
		  check_id   = "100000"
		  check_type = "uptime"
		  status     = "up"
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOFull(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  objective {
    display_name = "obj2"
    name         = "tf-objective-2"
    target       = 0.5
    value        = 10
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    calendar {
      start_time = "2020-03-09 00:00:00"
      time_zone = "Europe/Warsaw"
    }
    count      = 7
    unit       = "Day"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithAlertPolicy(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) +
			testAlertPolicyWithoutAnyAlertMethod(name+"-ap") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }

  alert_policies = [ nobl9_alert_policy.:name-ap.name ]
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusWithAttachmentsDeprecated(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"

  }

  attachments {
    display_name = "test"
    url          = "https://google.com"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusWithAttachment(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"

  }

  attachment {
    display_name = "test"
    url          = "https://google.com"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusWithAttachmentsConflict(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"

  }

  attachment {
    display_name = "test1"
    url          = "https://google.com"
  }

  attachment {
    display_name = "test1"
    url          = "https://google.com"
  }

  attachments {
    display_name = "test2"
    url          = "https://google.com"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithCountMetrics(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    count_metrics {
      incremental = true
      good {
            prometheus {
                promql = "1.0"
            }
      }
      total {
            prometheus {
                promql = "1.0"
            }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithMultipleObjectives(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  objective {
    display_name = "obj2"
    name         = "tf-objective-2"
    target       = 0.5
    value        = 10
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithRawMetricInObjective(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  budgeting_method = "Timeslices"

  objective {
    display_name      = "obj2"
    name              = "tf-objective-2"
    target            = 0.5
    value             = 10
    time_slice_target = 0.5
    op                = "lt"
    raw_metric {
      query{
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithTimeSlices(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  budgeting_method = "Timeslices"

  objective {
    display_name      = "obj2"
    name              = "tf-objective-2"
    target            = 0.5
    value             = 10
    time_slice_target = 0.5
    op                = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        prometheus {
          promql = "1.0"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testRedshiftSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testRedshiftAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
	    redshift {
		  region        = "eu-central-1"
		  cluster_id    = "redshift"
		  database_name = "dev"
		  query         = <<-EOT
			SELECT value as n9value, timestamp as n9date
			FROM sinusoid
			WHERE timestamp BETWEEN :n9date_from AND :n9date_to
		  EOT
	    }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testSplunkSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testSplunkAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        splunk {
          query = <<-EOT
			search index=events source=udp:5072 sourcetype=syslog status<400 |
			bucket _time span=1m |
			stats avg(response_time) as n9value by _time | rename _time as n9time | fields n9time n9value"
            EOT
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testSplunkObservabilitySLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testSplunkObservabilityAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        splunk_observability {
          program = "TODO"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testSplunkSingleQuerySLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testSplunkAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    primary      = false
    target       = 0.7
    time_slice_target = 0
    value        = 1
    count_metrics {
      incremental = true
      good_total {
		splunk {
		query = "| mstats avg(\"spl.intr.resource_usage.IOWait.data.avg_cpu_pct\") as n9good WHERE index=\"_metrics\"` +
			` span=15s | join type=left _time [ | mstats avg(\"spl.intr.resource_usage.IOWait.data.max_cpus_pct\") ` +
			`as n9total WHERE index=\"_metrics\" span=15s] | rename _time as n9time | fields n9time n9good n9total"
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testSumoLogicSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testSumoLogicAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
	    sumologic {
   			type         = "metrics"
            query        = "kube_node_status_condition | min"
            rollup       = "Min"
            quantization = "15s"
		}
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testThousandeyesSLO(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testThousandEyesAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        thousandeyes {
          test_id = 11
          test_type = "web-dom-load"
        }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testAnomalyConfigNoDataSameProject(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	var alertMethodName = name + "-tf-alertmethod"

	config := testService(serviceName) +
		testThousandEyesAgent(agentName) +
		mockAlertMethod(alertMethodName, testProject) + `

		resource "nobl9_slo" ":name" {
			name         = ":name"
			display_name = ":name"
			project      = ":project"
			service      = nobl9_service.:serviceName.name

			budgeting_method = "Occurrences"

			objective {
				display_name = "obj1"
				name         = "tf-objective-1"
				target       = 0.7
				value        = 1
				op           = "lt"
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			time_window {
				count      = 10
				is_rolling = true
				unit       = "Minute"
			}

			indicator {
				name    = nobl9_agent.:agentName.name
				project = ":project"
				kind    = "Agent"
			}

			anomaly_config {
				no_data {
					alert_method {
						name = ":alertMethodName"
						project = ":project"
					}
				}
			}
		}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)
	config = strings.ReplaceAll(config, ":alertMethodName", alertMethodName)

	return config
}

func testAnomalyConfigNoDataDifferentProject(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	var alertMethodName = name + "-tf-alertmethod"
	var alertMethodProject = name + "-tf-alertmethod-project"

	config := testService(serviceName) +
		testThousandEyesAgent(agentName) +
		mockAlertMethod(alertMethodName, alertMethodProject) + `

		resource "nobl9_slo" ":name" {
			name         = ":name"
			display_name = ":name"
			project      = ":project"
			service      = nobl9_service.:serviceName.name

			budgeting_method = "Occurrences"

			objective {
				display_name = "obj1"
				name         = "tf-objective-1"
				target       = 0.7
				value        = 1
				op           = "lt"
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			time_window {
				count      = 10
				is_rolling = true
				unit       = "Minute"
			}

			indicator {
				name    = nobl9_agent.:agentName.name
				project = ":project"
				kind    = "Agent"
			}

			anomaly_config {
				no_data {
					alert_method {
						name = ":alertMethodName"
						project = ":alertMethodProject"
					}
				}
			}
		}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)
	config = strings.ReplaceAll(config, ":alertMethodName", alertMethodName)
	config = strings.ReplaceAll(config, ":alertMethodProject", alertMethodProject)

	return config
}

func testMetricSpecRequired(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  label {
   key = "team"
   values = ["green","sapphire"]
  }

  label {
   key = "env"
   values = ["dev", "staging", "prod"]
  }

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    name         = "tf-objective-1"
    target       = 0.7
    value        = 1
    count_metrics {
      incremental = true
      good {
            prometheus {
                promql = "1.0"
            }
      }
    }
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:agentName.name
    project = ":project"
    kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testMaxOnePrimaryObjective(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"

	config := testService(serviceName) +
		testThousandEyesAgent(agentName) + `

		resource "nobl9_slo" ":name" {
			name         = ":name"
			display_name = ":name"
			project      = ":project"
			service      = nobl9_service.:serviceName.name

			budgeting_method = "Occurrences"

			objective {
				display_name = "obj1"
				name         = "tf-objective-1"
				target       = 0.7
				value        = 1
				op           = "lt"
                primary      = true
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			objective {
				display_name = "obj2"
				name         = "tf-objective-2"
				target       = 0.6
				value        = 1.1
				op           = "lt"
                primary      = false
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			time_window {
				count      = 10
				is_rolling = true
				unit       = "Minute"
			}

			indicator {
				name    = nobl9_agent.:agentName.name
				project = ":project"
				kind    = "Agent"
			}
		}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testNoPrimaryObjective(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"

	config := testService(serviceName) +
		testThousandEyesAgent(agentName) + `

		resource "nobl9_slo" ":name" {
			name         = ":name"
			display_name = ":name"
			project      = ":project"
			service      = nobl9_service.:serviceName.name

			budgeting_method = "Occurrences"

			objective {
				display_name = "obj1"
				name         = "tf-objective-1"
				target       = 0.7
				value        = 1
				op           = "lt"
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			objective {
				display_name = "obj2"
				name         = "tf-objective-2"
				target       = 0.6
				value        = 1.1
				op           = "lt"
                primary      = false
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			time_window {
				count      = 10
				is_rolling = true
				unit       = "Minute"
			}

			indicator {
				name    = nobl9_agent.:agentName.name
				project = ":project"
				kind    = "Agent"
			}
		}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testMoreThanOnePrimaryObjective(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"

	config := testService(serviceName) +
		testThousandEyesAgent(agentName) + `

		resource "nobl9_slo" ":name" {
			name         = ":name"
			display_name = ":name"
			project      = ":project"
			service      = nobl9_service.:serviceName.name

			budgeting_method = "Occurrences"

			objective {
				display_name = "obj1"
				name         = "tf-objective-1"
				target       = 0.7
				value        = 1
				op           = "lt"
                primary      = true
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			objective {
				display_name = "obj2"
				name         = "tf-objective-2"
				target       = 0.6
				value        = 1.1
				op           = "lt"
                primary      = true
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}

			time_window {
				count      = 10
				is_rolling = true
				unit       = "Minute"
			}

			indicator {
				name    = nobl9_agent.:agentName.name
				project = ":project"
				kind    = "Agent"
			}
		}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testCompositeDependencySLO(serviceName, agentName, name string) string {
	return fmt.Sprintf(`
resource "nobl9_slo" "%s" {
  name             = "%s"
  service          = nobl9_service.%s.name
  budgeting_method = "Occurrences"
  project          = "%s"

  time_window {
    unit       = "Day"
    count      = 14
    is_rolling = true
  }

  objective {
    target       = 0.999
    value        = 5
    display_name = "Good"
    name         = "objective-1"
    op           = "lte"

    raw_metric {
      query {
        prometheus {
          promql = "server_requestMsec{host=\"*\",instance=\"143.146.168.125:9913\",job=\"nginx\"}"
        }
      }
    }
  }
  objective {
    target       = 0.80
    value        = 2
    display_name = "Moderate"
    name         = "objective-2"
    op           = "lte"

    raw_metric {
      query {
        prometheus {
          promql = "server_requestMsec{host=\"*\",instance=\"143.146.168.125:9913\",job=\"nginx\"}"
        }
      }
    }
  }

  indicator {
    name    = nobl9_agent.%s.name
    kind    = "Agent"
    project = "%s"
  }
}
`, name, name, serviceName, testProject, agentName, testProject)
}

func testMetadataAnnotations(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"

	config := testService(serviceName) +
		testThousandEyesAgent(agentName) + `
		resource "nobl9_slo" ":name" {
			name         = ":name"
			display_name = ":name"
			project      = ":project"
			service      = nobl9_service.:serviceName.name
			annotations = {
				env  = "development"
				name = "example annotation"
			}
			budgeting_method = "Occurrences"
			objective {
				display_name = "obj1"
				name         = "tf-objective-1"
				target       = 0.7
				value        = 1
				op           = "lt"
				raw_metric {
					query {
						thousandeyes {
							test_id = 11
							test_type = "web-dom-load"
						}
					}
				}
			}
			time_window {
				count      = 10
				is_rolling = true
				unit       = "Minute"
			}
			indicator {
				name    = nobl9_agent.:agentName.name
				project = ":project"
				kind    = "Agent"
			}
		}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":serviceName", serviceName)
	config = strings.ReplaceAll(config, ":agentName", agentName)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testAlertPolicyWithoutAnyAlertMethod(name string) string {
	return fmt.Sprintf(`
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
}
`, name, name, testProject)
}

func testService(name string) string {
	return fmt.Sprintf(`
resource "nobl9_service" "%s" {
  name              = "%s"
  display_name = "%s"
  project             = "%s"
  description       = "Test of service"

  label {
   key = "env"
   values = ["green","sapphire"]
  }

  label {
   key = "dev"
   values = ["dev", "staging", "prod"]
  }

  annotations = {
   env = "development"
   name = "example annotation"
  }
}
`, name, name, name, testProject)
}
