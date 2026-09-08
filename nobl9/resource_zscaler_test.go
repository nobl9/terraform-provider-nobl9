package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
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

func TestZscalerAgentRoundTrip(t *testing.T) {
	t.Parallel()
	data := schema.TestResourceDataRaw(t, agentSchema(), map[string]interface{}{
		"name": "zscaler", "project": "default", "agent_type": "zscaler", "release_channel": "beta",
		"zscaler_config": []interface{}{map[string]interface{}{"vanity_domain": "example"}},
	})
	agent, diags := marshalAgent(data)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, agent.Spec.Zscaler)
	require.Empty(t, manifest.Validate([]manifest.Object{agent}))
	assert.Equal(t, "example", agent.Spec.Zscaler.VanityDomain)
	agent.Spec.Zscaler.VanityDomain = "updated"
	agent.Status = &v1alphaAgent.Status{AgentType: "Zscaler"}
	require.Empty(t, unmarshalAgent(data, *agent))
	config := data.Get(zscalerAgentConfigKey).(*schema.Set).List()[0].(map[string]interface{})
	assert.Equal(t, map[string]interface{}{"vanity_domain": "updated"}, config)
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
      value = 14
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

func testZscalerAgent(name string) string {
	return fmt.Sprintf(`
resource "nobl9_agent" "%s" {
  name = "%s"
  project = "%s"
  agent_type = "zscaler"
  zscaler_config { vanity_domain = "example" }
  release_channel = "beta"
  historical_data_retrieval {
    default_duration {
      value = 7
      unit = "Day"
    }
    max_duration {
      value = 14
      unit = "Day"
    }
  }
  query_delay {
    value = 20
    unit = "Minute"
  }
}
`, name, name, testProject)
}
