package nobl9

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
)

func TestZscalerDirectSpec(t *testing.T) {
	t.Parallel()
	spec := zscalerDirectSpec{}
	resourceSchema := resourceDirectFactory(spec).Schema
	for _, field := range []string{"client_id", "client_secret"} {
		assert.True(t, resourceSchema[field].Sensitive)
		assert.True(t, resourceSchema[field].Optional)
		assert.True(t, resourceSchema[field].Computed)
	}
	assert.Equal(t, v1alpha.ReleaseChannelBeta.String(), resourceSchema[releaseChannel].Default)
	data := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"name": "zscaler", "project": "default", "vanity_domain": "example",
		"client_id": "example-client-id", "client_secret": "example-client-secret",
	})
	actual := spec.MarshalSpec(data)
	assert.Equal(t, &v1alphaDirect.ZscalerConfig{
		VanityDomain: "example", ClientID: "example-client-id", ClientSecret: "example-client-secret",
	}, actual.Zscaler)
}

func TestZscalerDirectSpecPreservesCredentialsOnRead(t *testing.T) {
	t.Parallel()
	for _, remoteSecret := range []string{"[hidden]", ""} {
		t.Run(remoteSecret, func(t *testing.T) {
			t.Parallel()
			spec := zscalerDirectSpec{}
			data := schema.TestResourceDataRaw(t, resourceDirectFactory(spec).Schema, map[string]interface{}{
				"vanity_domain": "old", "client_id": "stored-client-id", "client_secret": "stored-client-secret",
			})
			diags := spec.UnmarshalSpec(data, v1alphaDirect.Spec{
				Description: "updated",
				Zscaler: &v1alphaDirect.ZscalerConfig{
					VanityDomain: "example", ClientID: remoteSecret, ClientSecret: remoteSecret,
				},
			})
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, "example", data.Get("vanity_domain"))
			assert.Equal(t, "updated", data.Get("description"))
			assert.Equal(t, "stored-client-id", data.Get("client_id"))
			assert.Equal(t, "stored-client-secret", data.Get("client_secret"))
		})
	}
}

func TestZscalerDirectSpecMissingConfig(t *testing.T) {
	t.Parallel()
	assert.True(t, (zscalerDirectSpec{}).UnmarshalSpec(nil, v1alphaDirect.Spec{}).HasError())
}

func TestUnmarshalDirectPreservesReplayAndSkipsUnsupportedLogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resource directSpecResource
		spec     v1alphaDirect.Spec
		wantLogs bool
	}{
		{"sumologic", sumologicDirectSpec{}, v1alphaDirect.Spec{
			SumoLogic: &v1alphaDirect.SumoLogicConfig{URL: "https://example.com"},
		}, true},
		{"thousandeyes", thousandeyesDirectSpec{}, v1alphaDirect.Spec{
			ThousandEyes: &v1alphaDirect.ThousandEyesConfig{},
		}, true},
		{"elasticsearch", elasticsearchDirectSpec{}, v1alphaDirect.Spec{
			Elasticsearch: &v1alphaDirect.ElasticsearchConfig{URL: "https://example.com"},
		}, false},
		{"splunk_observability", splunkObservabilityDirectSpec{}, v1alphaDirect.Spec{
			SplunkObservability: &v1alphaDirect.SplunkObservabilityConfig{Realm: "us0"},
		}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			resourceSchema := resourceDirectFactory(test.resource).Schema
			data := schema.TestResourceDataRaw(t, resourceSchema, nil)
			value := 1
			logsEnabled := true
			spec := test.spec
			spec.HistoricalDataRetrieval = &v1alpha.HistoricalDataRetrieval{
				DefaultDuration: v1alpha.HistoricalRetrievalDuration{Value: &value, Unit: v1alpha.HRDDay},
				MaxDuration:     v1alpha.HistoricalRetrievalDuration{Value: &value, Unit: v1alpha.HRDDay},
			}
			spec.LogCollectionEnabled = &logsEnabled
			direct := v1alphaDirect.Direct{
				Spec:   spec,
				Status: &v1alphaDirect.Status{DirectType: test.name},
			}

			// act
			diags := (directResource{test.resource}).unmarshalDirect(data, direct)

			// assert
			require.False(t, diags.HasError(), "%v", diags)
			_, hasLogs := resourceSchema[logCollectionConfigKey]
			assert.Contains(t, resourceSchema, historicalDataRetrievalConfigKey)
			assert.Equal(t, test.wantLogs, hasLogs)
			assert.Equal(t, 1, data.Get("historical_data_retrieval.0.default_duration.0.value"))
			replayed := marshalHistoricalDataRetrieval(data)
			require.NotNil(t, replayed)
			require.NotNil(t, replayed.MaxDuration.Value)
			assert.Equal(t, 1, *replayed.MaxDuration.Value)
			if test.wantLogs {
				assert.Equal(t, true, data.Get(logCollectionConfigKey))
			}
		})
	}
}

func TestSetReportsErrors(t *testing.T) {
	if os.Getenv("TF_ACC") != "" {
		t.Skip("Terraform SDK panics on invalid state writes in acceptance mode")
	}
	t.Parallel()
	data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, nil)
	var diags diag.Diagnostics

	// act
	set(data, "missing", true, &diags)

	// assert
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Invalid address to set")
}

func TestAcc_Nobl9Direct(t *testing.T) {
	cases := []struct {
		directType string
		configFunc func(string, string) string
	}{
		{appDynamicsDirectType, testAppDynamicsDirect},
		{azureMonitorDirectType, testAzureMonitorDirect},
		{bigqueryDirectType, testBigQueryDirect},
		{cloudWatchDirectType, testCloudWatchDirect},
		{datadogDirectType, testDatadogDirect},
		{dynatraceDirectType, testDynatraceDirect},
		{gcmDirectType, testGoogleCloudMonitoringDirect},
		{honeycombDirectType, testHoneycombDirect},
		{influxdbDirectType, testInfluxDBDirect},
		{instanaDirectType, testInstanaDirect},
		{lightstepDirectType, testLightstepDirect},
		{logicMonitorDirectType, testLogicMonitorDirect},
		{newRelicDirectType, testNewrelicDirect},
		{pingdomDirectType, testPingdomDirect},
		{redshiftDirectType, testRedshiftDirect},
		{splunkDirectType, testSplunkDirect},
		{splunkObservabilityDirectType, testSplunkObservabilityDirect},
		{sumologicDirectType, testSumoLogicDirect},
		{thousandeyesDirectType, testThousandEyesDirect},
		{dash0DirectType, testDash0Direct},
		{elasticsearchDirectType, testElasticsearchDirect},
		{zscalerDirectType, testZscalerDirect},
	}

	for _, tc := range cases {
		t.Run(tc.directType, func(t *testing.T) {
			testName := strings.ReplaceAll("test-"+tc.directType, "_", "")
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             CheckDestroy("nobl9_direct_%s", manifest.KindDirect),
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
  url = "https://web.net"
  account_name = "account name"
  client_secret = "secret"
  client_name = "client name"
  log_collection_enabled = true
  release_channel = "beta"
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
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testAzureMonitorDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  tenant_id = "45e4c1ed-5b6b-4555-a693-6ab7f15f3d6e"
  client_id = "secret"
  client_secret = "secret"
  log_collection_enabled = true
  release_channel = "beta"
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
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testBigQueryDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  service_account_key = "{}"
  log_collection_enabled = true
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testCloudWatchDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  role_arn = "secret"
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
  log_collection_enabled = true
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
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
  site = "datadoghq.eu"
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
  log_collection_enabled = true
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
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
  url = "https://web.net"
  platform_url = "https://web.apps.dynatrace.com"
  dynatrace_token = "secret"
  platform_token = "platform-secret"
  log_collection_enabled = true
  historical_data_retrieval {
    default_duration  {
      unit = "Day"
      value = 0
    }
    max_duration {
      unit = "Day"
      value = 0
    }
  }
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testGoogleCloudMonitoringDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  service_account_key = "{ \"key\": \"secret\" }"
  log_collection_enabled = true
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
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testHoneycombDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
	name = "%s"
	project = "%s"
	description = "desc"
	api_key = "secret"
	log_collection_enabled = true
	release_channel = "beta"
	historical_data_retrieval {
	  default_duration  {
		unit = "Day"
		value = 7
	  }
	  max_duration {
		unit = "Day"
		value = 7
	  }
	}
	query_delay {
	  unit = "Minute"
	  value = 6
	}
}
`, directType, name, name, testProject)
}

func testInfluxDBDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  url = "https://web.net"
  api_token = "secret"
  organization_id = "secret"
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testInstanaDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  url = "https://web.net"
  api_token = "secret"
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testLightstepDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  lightstep_organization = "acme"
  lightstep_project = "project1"
  url = "https://api.lightstep.com"
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
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testLogicMonitorDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  account = "account_name"
  account_id = "secret"
  access_key = "secret"
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
  log_collection_enabled = true
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
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
  account_id = "1234"
  insights_query_key = "NRIQ-secret"
  log_collection_enabled = true
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
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
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
  api_token = "secret"
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testRedshiftDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  secret_arn = "aws:arn"
  role_arn = "secret"
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testSplunkDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  url = "https://web.net"
  access_token = "secret"
  log_collection_enabled = true
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
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
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
  realm = "eu"
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
  release_channel = "alpha"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testSumoLogicDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  url = "https://web.net"
  access_id = "secret"
  access_key = "secret"
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testThousandEyesDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  oauth_bearer_token = "secret"
  log_collection_enabled = true
  release_channel = "stable"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testDash0Direct(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  url = "https://api.eu-west-1.aws.dash0.com/api/prometheus"
  auth_token = "secret"
  step = 60
  log_collection_enabled = true
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
  release_channel = "beta"
  query_delay {
    unit = "Minute"
    value = 6
  }
}
`, directType, name, name, testProject)
}

func testElasticsearchDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  description = "desc"
  url = "https://example.aws.found.io"
  api_key = "encoded-api-key"
  release_channel = "stable"
  historical_data_retrieval {
    default_duration {
      unit = "Day"
      value = 1
    }
    max_duration {
      unit = "Day"
      value = 30
    }
  }
  query_delay {
    unit = "Minute"
    value = 1
  }
}
`, directType, name, name, testProject)
}

func testZscalerDirect(directType, name string) string {
	return fmt.Sprintf(`
resource "nobl9_direct_%s" "%s" {
  name = "%s"
  project = "%s"
  vanity_domain = "example"
  client_id = "example-client-id"
  client_secret = "example-client-secret"
  release_channel = "beta"
  log_collection_enabled = true
  historical_data_retrieval {
    default_duration {
      value = 7
      unit = "Day"
    }
    max_duration {
      value = 10
      unit = "Day"
    }
  }
  query_delay {
    value = 20
    unit = "Minute"
  }
}
`, directType, name, name, testProject)
}
