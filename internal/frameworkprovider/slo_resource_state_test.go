package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_preserveNullIndicatorProject(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source          *SLOResourceModel
		target          *SLOResourceModel
		expectedProject types.String
	}{
		"omitted project": {
			source: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringNull(),
				}},
			},
			target: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("default"),
				}},
			},
			expectedProject: types.StringNull(),
		},
		"explicit project": {
			source: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("metrics"),
				}},
			},
			target: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("metrics"),
				}},
			},
			expectedProject: types.StringValue("metrics"),
		},
		"missing source indicator": {
			source: &SLOResourceModel{},
			target: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("default"),
				}},
			},
			expectedProject: types.StringValue("default"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			preserveNullIndicatorProject(test.source, test.target)

			assert.Equal(t, test.expectedProject, test.target.Indicator[0].Project)
		})
	}
}
