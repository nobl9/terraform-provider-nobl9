package nobl9

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
)

// Every site the schema description lists must pass the validation that runs at plan time
// (CustomizeDiff). datadoghq.com is a control that every SDK version accepts.
func TestDatadogSitesValidate(t *testing.T) {
	sites := []string{"datadoghq.com", "us2.ddog-gov.com", "ap2.datadoghq.com", "uk1.datadoghq.com"}

	for _, site := range sites {
		t.Run("agent/"+site, func(t *testing.T) {
			t.Parallel()
			data := schema.TestResourceDataRaw(t, agentSchema(), map[string]interface{}{
				"name":                "datadog",
				"project":             "default",
				agentTypeKey:          datadogAgentType,
				datadogAgentConfigKey: []interface{}{map[string]interface{}{"site": site}},
			})
			agent, diags := marshalAgent(data)
			require.False(t, diags.HasError(), "%v", diags)

			// act
			errs := manifest.Validate([]manifest.Object{agent})

			// assert
			assert.Empty(t, errs)
		})

		t.Run("direct/"+site, func(t *testing.T) {
			t.Parallel()
			resource := directResource{directSpecResource: datadogDirectSpec{}}
			data := schema.TestResourceDataRaw(t, resourceDirectFactory(datadogDirectSpec{}).Schema, map[string]interface{}{
				"name":            "datadog",
				"project":         "default",
				"site":            site,
				"api_key":         "api-key",
				"application_key": "application-key",
			})
			direct, diags := resource.marshalDirect(stubEmptyRawConfig{data})
			require.False(t, diags.HasError(), "%v", diags)

			// act
			errs := manifest.Validate([]manifest.Object{direct})

			// assert
			assert.Empty(t, errs)
		})
	}
}

// stubEmptyRawConfig stands in for a plan's raw config, which TestResourceDataRaw leaves null.
type stubEmptyRawConfig struct{ *schema.ResourceData }

func (stubEmptyRawConfig) GetRawConfig() cty.Value { return cty.EmptyObjectVal }
