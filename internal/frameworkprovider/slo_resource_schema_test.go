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

func TestSLOMetricSpecModelsMatchSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		blocks  map[string]schema.Block
		model   any
		decoded any
		rawOnly bool
	}{
		{"count metrics", sloResourceCountMetricSpecBlocks(), CountMetricSpecModel{}, &CountMetricSpecModel{}, false},
		{"raw metric", sloResourceRawMetricSpecBlocks(), MetricSpecModel{}, &MetricSpecModel{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			objectType := schema.NestedBlockObject{Blocks: test.blocks}.Type().(types.ObjectType)

			_, hasThousandEyes := objectType.AttrTypes["thousandeyes"]
			_, hasZscaler := objectType.AttrTypes["zscaler"]
			require.Equal(t, test.rawOnly, hasThousandEyes)
			require.Equal(t, test.rawOnly, hasZscaler)

			// act
			value, diags := types.ObjectValueFrom(ctx, objectType.AttrTypes, test.model)
			require.False(t, diags.HasError(), "%v", diags)
			decodeDiags := value.As(ctx, test.decoded, basetypes.ObjectAsOptions{})

			// assert
			require.False(t, decodeDiags.HasError(), "%v", decodeDiags)
		})
	}
}

// ClickHouse parameters live in a map attribute; no SDK example sets them.
func TestSLOResourceClickHouseParametersRoundTrip(t *testing.T) {
	t.Parallel()
	spec := &v1alphaSLO.MetricSpec{
		ClickHouse: &v1alphaSLO.ClickHouseMetric{
			Query:      "SELECT n9date, n9value FROM t WHERE ts >= {n9date_from:DateTime} AND ts < {n9date_to:DateTime}",
			Parameters: map[string]string{"service": "api-server"},
		},
	}

	t.Run("count metrics", func(t *testing.T) {
		t.Parallel()
		decoded := &CountMetricSpecModel{}
		roundTripMetricSpecModel(t, sloResourceCountMetricSpecBlocks(), countMetricSpecToModel(spec), decoded)
		require.Equal(t, spec, decoded.ToManifest())
	})
	t.Run("raw metric", func(t *testing.T) {
		t.Parallel()
		decoded := &MetricSpecModel{}
		roundTripMetricSpecModel(t, sloResourceRawMetricSpecBlocks(), metricSpecToModel(spec), decoded)
		require.Equal(t, spec, decoded.ToManifest())
	})
}

func roundTripMetricSpecModel(t *testing.T, blocks map[string]schema.Block, model, decoded any) {
	t.Helper()
	ctx := context.Background()
	objectType := schema.NestedBlockObject{Blocks: blocks}.Type().(types.ObjectType)
	value, diags := types.ObjectValueFrom(ctx, objectType.AttrTypes, model)
	require.False(t, diags.HasError(), "%v", diags)
	diags = value.As(ctx, decoded, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), "%v", diags)
}
