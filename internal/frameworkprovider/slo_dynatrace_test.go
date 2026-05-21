package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelToDynatrace(t *testing.T) {
	t.Run("metric selector", func(t *testing.T) {
		metric := (&DynatraceModel{
			MetricSelector: types.StringValue("builtin:host.cpu.usage"),
		}).ToManifest()

		require.NotNil(t, metric)
		assert.Equal(t, "builtin:host.cpu.usage", *metric.MetricSelector)
		assert.Nil(t, metric.DQL)
	})

	t.Run("dql", func(t *testing.T) {
		metric := (&DynatraceModel{
			DQL: []DynatraceDQLModel{{
				Query:    "timeseries value = avg(dt.host.cpu.usage)",
				Interval: types.StringValue("1m"),
			}},
		}).ToManifest()

		require.NotNil(t, metric)
		assert.Nil(t, metric.MetricSelector)
		require.NotNil(t, metric.DQL)
		assert.Equal(t, "timeseries value = avg(dt.host.cpu.usage)", metric.DQL.Query)
		assert.Equal(t, "1m", metric.DQL.Interval)
	})
}

func TestDynatraceToModel(t *testing.T) {
	t.Run("metric selector", func(t *testing.T) {
		selector := "builtin:host.cpu.usage"
		model := dynatraceToModel(&v1alphaSLO.DynatraceMetric{
			MetricSelector: &selector,
		})

		require.NotNil(t, model)
		assert.Equal(t, "builtin:host.cpu.usage", model.MetricSelector.ValueString())
		assert.Empty(t, model.DQL)
	})

	t.Run("dql", func(t *testing.T) {
		model := dynatraceToModel(&v1alphaSLO.DynatraceMetric{
			DQL: &v1alphaSLO.DynatraceDQL{
				Query:    "timeseries value = avg(dt.host.cpu.usage)",
				Interval: "1m",
			},
		})

		require.NotNil(t, model)
		assert.True(t, model.MetricSelector.IsNull())
		require.Len(t, model.DQL, 1)
		assert.Equal(t, "timeseries value = avg(dt.host.cpu.usage)", model.DQL[0].Query)
		assert.Equal(t, "1m", model.DQL[0].Interval.ValueString())
	})
}
