package frameworkprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/stretchr/testify/require"
)

// Regression test: good_total was once built with a reduced block set (without
// clickhouse) while CountMetricsModel.GoodTotal kept the full MetricSpecModel.
// terraform-plugin-framework requires an exact bidirectional 1:1 match between
// object type and struct, so every non-empty good_total list failed with
// "Struct defines fields not found in object: clickhouse" — breaking
// Read/Create/Update/Delete for existing GA SLOs of every other source.
// MetricSpecModel is decoded at five sites (good, bad, total, good_total,
// raw_metric.query); reducing ANY of them reintroduces the bug, so each site
// is round-tripped against the real schema here. Source capability is
// enforced by validation, not by schema shape.
func TestSLOResourceMetricSpecSitesDecodeMetricSpecModel(t *testing.T) {
	ctx := context.Background()

	// A non-ClickHouse source: that is the population the regression broke.
	model := metricSpecToModel(&v1alphaSLO.MetricSpec{
		Prometheus: &v1alphaSLO.PrometheusMetric{
			PromQL: stringPtr(`sum(rate(http_requests_total[5m]))`),
		},
	})

	for _, site := range []string{"good", "bad", "total", "good_total", "raw_metric.query"} {
		t.Run(site, func(t *testing.T) {
			objType := metricSpecObjectType(t, site)

			value, diags := types.ObjectValueFrom(ctx, objType.AttrTypes, model)
			require.Empty(t, diags,
				"%s object type must accept the full MetricSpecModel: %v", site, diags)

			var decoded MetricSpecModel
			diags = value.As(ctx, &decoded, basetypes.ObjectAsOptions{})
			require.Empty(t, diags,
				"%s value must decode back into MetricSpecModel: %v", site, diags)
			require.Equal(t, model, decoded)
		})
	}
}

// Every metric-spec site must expose the exact same object type so the shared
// MetricSpecModel decodes all of them; clickhouse must be among the blocks.
func TestSLOResourceMetricSpecSchemaIsUniform(t *testing.T) {
	good := metricSpecObjectType(t, "good")
	for _, site := range []string{"bad", "total", "good_total", "raw_metric.query"} {
		require.Equal(t, good, metricSpecObjectType(t, site),
			"%s must expose the same metric-spec object type as good", site)
	}
	require.Contains(t, good.AttrTypes, "clickhouse")
	require.Contains(t, good.AttrTypes, "prometheus")
}

// Manifest round-trip through the model for a good_total counter. This covers
// the countMetricsToModel/ToManifest mapping only; the schema decode paths
// are exercised by the tests above.
func TestSLOResourceGoodTotalManifestRoundTrip(t *testing.T) {
	spec := &v1alphaSLO.CountMetricsSpec{
		Incremental: boolPtr(false),
		GoodTotalMetric: &v1alphaSLO.MetricSpec{
			Prometheus: &v1alphaSLO.PrometheusMetric{
				PromQL: stringPtr(`sum(rate(http_requests_total[5m]))`),
			},
		},
	}

	model := countMetricsToModel(spec)
	require.Len(t, model.GoodTotal, 1)

	roundTripped := model.ToManifest()
	require.Equal(t, spec, roundTripped)
}

// metricSpecObjectType walks the real resource schema down to the metric-spec
// nested-object type at the given decode site (objective -> count_metrics ->
// good/bad/total/good_total, or objective -> raw_metric -> query).
func metricSpecObjectType(t *testing.T, site string) basetypes.ObjectType {
	t.Helper()
	objective := nestedBlock(t, sloResourceSchema.Blocks, "objective")
	var block schema.ListNestedBlock
	if site == "raw_metric.query" {
		rawMetric := nestedBlock(t, objective.NestedObject.Blocks, "raw_metric")
		block = nestedBlock(t, rawMetric.NestedObject.Blocks, "query")
	} else {
		countMetrics := nestedBlock(t, objective.NestedObject.Blocks, "count_metrics")
		block = nestedBlock(t, countMetrics.NestedObject.Blocks, site)
	}
	objType, ok := block.NestedObject.Type().(basetypes.ObjectType)
	require.True(t, ok, "%s nested-object type must be basetypes.ObjectType", site)
	return objType
}

func nestedBlock(t *testing.T, blocks map[string]schema.Block, name string) schema.ListNestedBlock {
	t.Helper()
	block, ok := blocks[name].(schema.ListNestedBlock)
	require.True(t, ok, "block %q must be a schema.ListNestedBlock", name)
	return block
}

func stringPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }
