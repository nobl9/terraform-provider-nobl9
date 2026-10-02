package frameworkprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
