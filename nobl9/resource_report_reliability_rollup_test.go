package nobl9

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
)

func TestReliabilityRollupScoreTypeSchema(t *testing.T) {
	scoreType := reportReliabilityRollup{}.GetSchema()["reliability_score_type"]

	require.NotNil(t, scoreType)
	assert.Equal(t, schema.TypeString, scoreType.Type)
	assert.True(t, scoreType.Optional)
	assert.Equal(t, v1alphaReport.ReliabilityScoreTypeSLOTimeWindow.String(), scoreType.Default)

	path := cty.GetAttrPath("reliability_score_type")
	assert.Empty(t, scoreType.ValidateDiagFunc(v1alphaReport.ReliabilityScoreTypeSLOTimeWindow.String(), path))
	assert.Empty(t, scoreType.ValidateDiagFunc(v1alphaReport.ReliabilityScoreTypeReportTimeFrame.String(), path))
	assert.NotEmpty(t, scoreType.ValidateDiagFunc("unsupported", path))
}

func TestReliabilityRollupScoreTypeRoundTrip(t *testing.T) {
	tests := map[string]struct {
		configuredScoreType v1alphaReport.ReliabilityScoreType
		returnedScoreType   v1alphaReport.ReliabilityScoreType
		expectedScoreType   v1alphaReport.ReliabilityScoreType
	}{
		"omitted with empty legacy API response": {
			expectedScoreType: v1alphaReport.ReliabilityScoreTypeSLOTimeWindow,
		},
		"SLO time window": {
			configuredScoreType: v1alphaReport.ReliabilityScoreTypeSLOTimeWindow,
			returnedScoreType:   v1alphaReport.ReliabilityScoreTypeSLOTimeWindow,
			expectedScoreType:   v1alphaReport.ReliabilityScoreTypeSLOTimeWindow,
		},
		"report time frame": {
			configuredScoreType: v1alphaReport.ReliabilityScoreTypeReportTimeFrame,
			returnedScoreType:   v1alphaReport.ReliabilityScoreTypeReportTimeFrame,
			expectedScoreType:   v1alphaReport.ReliabilityScoreTypeReportTimeFrame,
		},
	}

	provider := reportReliabilityRollup{}
	resourceSchema := resourceReportFactory(provider).Schema
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			raw := map[string]interface{}{
				"time_frame": []interface{}{
					map[string]interface{}{
						"time_zone": "Europe/Warsaw",
						"rolling": []interface{}{
							map[string]interface{}{
								"unit":  "Week",
								"count": 1,
							},
						},
					},
				},
			}
			if test.configuredScoreType != "" {
				raw["reliability_score_type"] = test.configuredScoreType.String()
			}
			data := schema.TestResourceDataRaw(t, resourceSchema, raw)

			spec := provider.MarshalSpec(v1alphaReport.Spec{}, data)
			require.NotNil(t, spec.ReliabilityRollup)
			assert.Equal(t, test.expectedScoreType, spec.ReliabilityRollup.ReliabilityScoreType)
			spec.ReliabilityRollup.ReliabilityScoreType = test.returnedScoreType

			state := schema.TestResourceDataRaw(t, resourceSchema, nil)
			diags := provider.UnmarshalSpec(state, spec)
			require.Empty(t, diags)
			assert.Equal(t, test.expectedScoreType.String(), state.Get("reliability_score_type"))
		})
	}
}
