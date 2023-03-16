package nobl9

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	n9api "github.com/nobl9/nobl9-go"
)

func TestAcc_Nobl9Direct(t *testing.T) {
	cases := []struct {
		directType string
		configFunc func(string, string) string
	}{
		{appDynamicsDirectType, testAppDynamicsDirect},
		{bigqueryDirectType, testBigQueryDirect},
		{cloudWatchDirectType, testCloudWatchDirect},
		{datadogDirectType, testDatadogDirect},
		{dynatraceDirectType, testDynatraceDirect},
		{gcmDirectType, testGoogleCloudMonitoringDirect},
		{influxdbDirectType, testInfluxDBDirect},
		{instanaDirectType, testInstanaDirect},
		{lightstepDirectType, testLightstepDirect},
		{newRelicDirectType, testNewrelicDirect},
		{pingdomDirectType, testPingdomDirect},
		{redshiftDirectType, testRedshiftDirect},
		{splunkDirectType, testSplunkDirect},
		{splunkObservabilityDirectType, testSplunkObservabilityDirect},
		{sumologicDirectType, testSumoLogicDirect},
		{thousandeyesDirectType, testThousandEyesDirect},
	}

	for _, tc := range cases {
		t.Run(tc.directType, func(t *testing.T) {
			testName := strings.ReplaceAll("test-"+tc.directType, "_", "")
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: ProviderFactory(),
				CheckDestroy:      CheckDestroy("nobl9_direct_%s", n9api.ObjectDirect),
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.directType, testName),
						Check:  CheckObjectCreated("nobl9_direct_" + tc.directType + "." + testName),
					},
				},
			})
		})
	}
}

func testAppDynamicsDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  url = "https://web.net"
  account_name = "account name"
  client_secret = "secret"
  client_name = "client name"
}
`, directType, name, name, testProject)
}

func testBigQueryDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  service_account_key = "secret"  	
}
`, directType, name, name, testProject)
}

func testCloudWatchDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  access_key_id = "secret"
  secret_access_key = "secret"
  historical_data_retrieval {
    default_duration  {
      unit = "Day"
      value = 1
    }
    max_duration {
      unit = "Day"
      value = 10
    }
  }
}
`, directType, name, name, testProject)
}

func testDatadogDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  site = "eu"
  api_key = "secret"
  application_key = "secret"
  historical_data_retrieval {
    default_duration  {
      unit = "Day"
      value = 1
    }
    max_duration {
      unit = "Day"
      value = 10
    }
  }
}
`, directType, name, name, testProject)
}

func testDynatraceDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  url = "https://web.net"
  dynatrace_token = "secret"
}
`, directType, name, name, testProject)
}

func testGoogleCloudMonitoringDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  service_account_key = "secret"
}
`, directType, name, name, testProject)
}

func testInfluxDBDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  url = "https://web.net"
  api_token = "secret"
  organization_id = "secret"
}
`, directType, name, name, testProject)
}

func testInstanaDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  url = "https://web.net"
  api_token = "secret"
}
`, directType, name, name, testProject)
}

func testLightstepDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  lightstep_organization = "acme"
  lightstep_project = "project1"
  app_token = "secret"
  historical_data_retrieval {
    default_duration  {
      unit = "Day"
      value = 1
    }
    max_duration {
      unit = "Day"
      value = 10
    }
  }
}
`, directType, name, name, testProject)
}

func testNewrelicDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  account_id = "1234"
  insights_query_key = "secret"
  historical_data_retrieval {
    default_duration  {
      unit = "Day"
      value = 1
    }
    max_duration {
      unit = "Day"
      value = 10
    }
  }
}
`, directType, name, name, testProject)
}

func testPingdomDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  api_token = "secret"
}
`, directType, name, name, testProject)
}

func testRedshiftDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  secret_arn = "aws:arn"
  access_key_id = "secret"
  secret_access_key = "secret"
}
`, directType, name, name, testProject)
}

func testSplunkDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  url = "https://web.net"
  access_token = "secret"
  historical_data_retrieval {
    default_duration  {
      unit = "Day"
      value = 1
    }
    max_duration {
      unit = "Day"
      value = 10
    }
  }
}
`, directType, name, name, testProject)
}

func testSplunkObservabilityDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  realm = "eu"
  access_token = "secret"
}
`, directType, name, name, testProject)
}

func testSumoLogicDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics"]
  url = "http://web.net"
  access_id = "secret"
  access_key = "secret"
}
`, directType, name, name, testProject)
}

func testThousandEyesDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  source_of = ["Metrics", "Services"]
  oauth_bearer_token = "secret"
}
`, directType, name, name, testProject)
}
