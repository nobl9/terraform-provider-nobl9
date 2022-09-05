package nobl9

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9SLO(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		configFunc func(string) string
	}{
		{"test-appdynamics", testAppdynamicsSLO},
		{"test-bigquery", testBigQuerySLO},
		{"test-cloudwatch-with-json", testCloudWatchWithJSON},
		{"test-cloudwatch-with-sql", testCloudWatchWithSQL},
		{"test-cloudwatch-with-stat", testCloudWatchWithStat},
		{"test-composite-occurrences", testCompositeSLOOccurrences},
		{"test-composite-time-slices", testCompositeSLOTimeSlices},
		{"test-datadog", testDatadogSLO},
		{"test-dynatrace", testDynatraceSLO},
		{"test-graphite", testGraphiteSLO},
		{"test-lightstep", testLightstepSLO},
		{"test-multiple-ap", testMultipleAlertPolicies},
		{"test-newrelic", testNewRelicSLO},
		{"test-opentsdb", testOpenTSDBSLO},
		{"test-prom-full", testPrometheusSLOFULL},
		{"test-prom-with-ap", testPrometheusSLOWithAlertPolicy},
		{"test-prom-with-attachments", testPrometheusWithAttachments},
		{"test-prom-with-countmetrics", testPrometheusSLOWithCountMetrics},
		{"test-prom-with-multiple-objectives", testPrometheusSLOWithMultipleObjectives},
		{"test-prom-with-raw-metric-in-objective", testPrometheusSLOWithRawMetricInObjective},
		{"test-prom-with-time-slices", testPrometheusSLOWithTimeSlices},
		{"test-prometheus", testPrometheusSLO},
		{"test-splunk", testSplunkSLO},
		{"test-splunk-observability", testSplunkObservabilitySLO},
		{"test-thousandeyes", testThousandeyesSLO},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestory("nobl9_slo", n9api.ObjectSLO),
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        appdynamics {
          application_name = "polakpotrafi"
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

//nolint:lll
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        bigquery {
          project_id = "bdwtest-256112"
          location = "EU"
          query = "SELECT response_time AS n9value, created AS n9date FROM 'bdwtest-256112.metrics.http_response' WHERE date_col BETWEEN DATETIME(@n9date_from) AND DATETIME(@n9date_to) "
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

func testCompositeSLOOccurrences(name string) string {
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

func testCompositeSLOTimeSlices(name string) string {
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

//nolint:lll
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        dynatrace {
          metric_selector = <<EOT
builtin:synthetic.http.duration.geo:filter(and(in("dt.entity.http_check",entitySelector("type(http_check),entityName(~"API Sample~")")),in("dt.entity.synthetic_location",entitySelector("type(synthetic_location),entityName(~"N. California~")")))):splitBy("dt.entity.http_check","dt.entity.synthetic_location"):avg:auto:sort(value(avg,descending)):limit(20)
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        lightstep {
          stream_id = "DzpxcSRh"
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

func testMultipleAlertPolicies(name string) string {
	var serviceName = name + "-tf-service"
	var agentName = name + "-tf-agent"
	config :=
		testService(serviceName) +
			testPrometheusAgent(agentName) +
			testAlertPolicyWithoutIntegration(name+"-fast") +
			testAlertPolicyWithoutIntegration(name+"-slow") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
    project      = ":project"
  service      = nobl9_service.:serviceName.name

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

func testPrometheusSLOFULL(name string) string {
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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
			testAlertPolicyWithoutIntegration(name+"-ap") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = ":project"
  service      = nobl9_service.:serviceName.name

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

func testPrometheusWithAttachments(name string) string {
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

//nolint:lll
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        splunk {
          query = "search index=polakpotrafi-events source=udp:5072 sourcetype=syslog status<400 | bucket _time span=1m | stats avg(response_time) as n9value by _time | rename _time as n9time | fields n9time n9value"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
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

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
    raw_metric {
      query {
        thousandeyes {
          test_id = 11
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
