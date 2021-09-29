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
		{"test-prom-with-ap", testPrometheusSLOWithAlerPolicy},
		{"test-prom-with-countmetrics", testPrometheusSLOWithCountMetrics},
		{"test-prom-with-multiple-objectives", testPrometheusSLOWithMultipleObjectives},
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
	config := testService(name+"-service") +
		testPrometheusConfig(name+"-agent") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = "terraform"
  service      = nobl9_service.:name-service.name

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:name-agent.name
    project = ":project"
	kind    = "Agent"
    raw_metric {
      prometheus_metric {
        promql = "1.0"
      }
    }
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithAlerPolicy(name string) string {
	config := testService(name+"-service") +
		testPrometheusConfig(name+"-agent") +
		testAlertPolicyWithoutIntegration(name+"-ap") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = "terraform"
  service      = nobl9_service.:name-service.name

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:name-agent.name
    project = ":project"
	kind    = "Agent"
    raw_metric {
      prometheus_metric {
        promql = "1.0"
      }
    }
  }

  alert_policies = [ nobl9_alert_policy.:name-ap.name ]
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}

func testPrometheusSLOWithCountMetrics(name string) string {
	config := testService(name+"-service") +
		testPrometheusConfig(name+"-agent") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = "terraform"
  service      = nobl9_service.:name-service.name

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
	count_metrics {
	  incremental = true
	  good {
		prometheus_metric {
		  promql = "1.0"
		}
	  }
	  total {
		prometheus_metric {
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
    name    = nobl9_agent.:name-agent.name
    project = ":project"
	kind    = "Agent"
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}
func testPrometheusSLOWithMultipleObjectives(name string) string {
	config := testService(name+"-service") +
		testPrometheusConfig(name+"-agent") + `
resource "nobl9_slo" ":name" {
  name         = ":name"
  display_name = ":name"
  project      = "terraform"
  service      = nobl9_service.:name-service.name

  budgeting_method = "Occurrences"

  objective {
    display_name = "obj1"
    target       = 0.7
    value        = 1
    op           = "lt"
  }

  objective {
    display_name = "obj2"
    target       = 0.5
    value        = 10
    op           = "lt"
  }

  time_window {
    count      = 10
    is_rolling = true
    unit       = "Minute"
  }

  indicator {
    name    = nobl9_agent.:name-agent.name
    project = ":project"
	kind    = "Agent"
    raw_metric {
      prometheus_metric {
        promql = "1.0"
      }
    }
  }
}
`
	config = strings.ReplaceAll(config, ":name", name)
	config = strings.ReplaceAll(config, ":project", testProject)

	return config
}
