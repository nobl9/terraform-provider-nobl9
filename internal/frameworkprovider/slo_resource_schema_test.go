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
// Read/Create/Update/Delete for existing GA SLOs of every other source. The
// metric-spec schema must stay uniform across good/bad/total/good_total;
// source capability is enforced by validation, not by schema shape.
func TestSLOResourceGoodTotalDecodesMetricSpecModel(t *testing.T) {
	ctx := context.Background()
	objType := goodTotalNestedObjectType(t)

	// A non-ClickHouse source: that is the population the regression broke.
	model := metricSpecToModel(&v1alphaSLO.MetricSpec{
		Prometheus: &v1alphaSLO.PrometheusMetric{
			PromQL: stringPtr(`sum(rate(http_requests_total[5m]))`),
		},
	})

	value, diags := types.ObjectValueFrom(ctx, objType.AttrTypes, model)
	require.Empty(t, diags,
		"good_total object type must accept the full MetricSpecModel: %v", diags)

	var decoded MetricSpecModel
	diags = value.As(ctx, &decoded, basetypes.ObjectAsOptions{})
	require.Empty(t, diags, "good_total value must decode back into MetricSpecModel: %v", diags)
	require.Equal(t, model, decoded)
}

// The good_total block must expose the exact same metric-spec blocks as
// good/bad/total so the shared MetricSpecModel decodes all four.
func TestSLOResourceGoodTotalSchemaIsUniform(t *testing.T) {
	countMetrics := countMetricsNestedBlock(t)

	goodTotal, ok := countMetrics.NestedObject.Blocks["good_total"].(schema.ListNestedBlock)
	require.True(t, ok)
	good, ok := countMetrics.NestedObject.Blocks["good"].(schema.ListNestedBlock)
	require.True(t, ok)

	require.Equal(t, good.NestedObject.Type(), goodTotal.NestedObject.Type(),
		"good_total must expose the same metric-spec object type as good")
	require.Contains(t, goodTotal.NestedObject.Blocks, "clickhouse")
	require.Contains(t, goodTotal.NestedObject.Blocks, "prometheus")
}

// Manifest round-trip through the model for a good_total counter.
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

// countMetricsNestedBlock walks the real resource schema down to the
// count_metrics block (objective -> count_metrics).
func countMetricsNestedBlock(t *testing.T) schema.ListNestedBlock {
	t.Helper()
	objective, ok := sloResourceSchema.Blocks["objective"].(schema.ListNestedBlock)
	require.True(t, ok)
	countMetrics, ok := objective.NestedObject.Blocks["count_metrics"].(schema.ListNestedBlock)
	require.True(t, ok)
	return countMetrics
}

// goodTotalNestedObjectType extracts the good_total nested-object type from the
// real resource schema.
func goodTotalNestedObjectType(t *testing.T) basetypes.ObjectType {
	t.Helper()
	countMetrics := countMetricsNestedBlock(t)
	goodTotal, ok := countMetrics.NestedObject.Blocks["good_total"].(schema.ListNestedBlock)
	require.True(t, ok)
	objType, ok := goodTotal.NestedObject.Type().(basetypes.ObjectType)
	require.True(t, ok)
	return objType
}

func stringPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }
