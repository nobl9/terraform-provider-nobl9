package frameworkprovider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSLOResourceMetricSpecBlocksWithout(t *testing.T) {
	blocks := sloResourceMetricSpecBlocksWithout("clickhouse")

	require.NotContains(t, blocks, "clickhouse")
	require.Contains(t, blocks, "prometheus")
}
