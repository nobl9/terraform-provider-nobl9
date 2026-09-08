package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZscalerMetricRoundTrip(t *testing.T) {
	t.Parallel()
	for _, metric := range []string{"score", "pft", "dns", "availability"} {
		for _, withLocation := range []bool{false, true} {
			name := metric + "/all-locations"
			if withLocation {
				name = metric + "/filtered-location"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				src := &v1alphaSLO.ZscalerMetric{AppID: 9007199254740993, Metric: metric}
				if withLocation {
					locationID := int64(123)
					src.LocationID = &locationID
				}
				model := metricSpecToModel(&v1alphaSLO.MetricSpec{Zscaler: src})
				require.Len(t, model.Zscaler, 1)
				assert.Equal(t, !withLocation, model.Zscaler[0].LocationID.IsNull())
				assert.Equal(t, src, model.ToManifest().Zscaler)
			})
		}
	}
}

func TestZscalerMetricOmittedAndExplicitZeroLocation(t *testing.T) {
	t.Parallel()
	assert.Nil(t, zscalerToModel(nil))
	assert.Nil(t, modelToZscaler(nil))
	assert.Nil(t, (MetricSpecModel{}).ToManifest().Zscaler)
	model := ZscalerModel{
		AppID: types.Int64Value(123), Metric: types.StringValue("score"), LocationID: types.Int64Value(0),
	}
	actual := modelToZscaler(&model)
	require.NotNil(t, actual.LocationID)
	assert.Zero(t, *actual.LocationID)
}

func TestZscalerMetricUnknownPlanValues(t *testing.T) {
	t.Parallel()
	type query struct {
		Zscaler []ZscalerModel `tfsdk:"zscaler"`
	}
	model := query{Zscaler: []ZscalerModel{{
		AppID: types.Int64Unknown(), LocationID: types.Int64Unknown(), Metric: types.StringUnknown(),
	}}}
	plan := tfsdk.Plan{Schema: schema.Schema{
		Blocks: map[string]schema.Block{"zscaler": sloMetricZscalerBlock()},
	}}
	require.False(t, plan.Set(t.Context(), model).HasError())
	var actual query
	require.False(t, plan.Get(t.Context(), &actual).HasError())
	assert.Equal(t, model, actual)
}
