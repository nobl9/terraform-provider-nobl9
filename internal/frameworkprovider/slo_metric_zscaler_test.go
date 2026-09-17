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
	for _, metric := range []string{"score", "pft"} {
		for _, withLocation := range []bool{false, true} {
			name := metric + "/all-locations"
			if withLocation {
				name = metric + "/filtered-location"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				src := &v1alphaSLO.ZscalerMetric{Type: "application", AppID: 9007199254740993, Metric: metric}
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
		Type:  types.StringValue("application"),
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
		Type: types.StringUnknown(), DeviceID: types.Int64Unknown(), ProbeID: types.Int64Unknown(),
		LegSrc: types.StringUnknown(), LegDst: types.StringUnknown(),
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

func TestZscalerProbeMetricRoundTrip(t *testing.T) {
	deviceID, probeID := int64(67890), int64(24680)
	src, dst := "client", "egress"
	for _, query := range []v1alphaSLO.ZscalerMetric{
		{Type: "web-probe", Metric: "ttfb", AppID: 12345, DeviceID: &deviceID, ProbeID: &probeID},
		{Type: "cloudpath", Metric: "loss", AppID: 12345, DeviceID: &deviceID, ProbeID: &probeID},
		{Type: "cloudpath", Metric: "latency", AppID: 12345,
			DeviceID: &deviceID, ProbeID: &probeID, LegSrc: &src, LegDst: &dst},
	} {
		t.Run(query.Type+"/"+query.Metric, func(t *testing.T) {
			model := zscalerToModel(&query)
			assert.True(t, model.LocationID.IsNull())
			assert.Equal(t, query.LegSrc == nil, model.LegSrc.IsNull())
			assert.Equal(t, query.LegDst == nil, model.LegDst.IsNull())
			assert.Equal(t, &query, modelToZscaler(model))
		})
	}
}
