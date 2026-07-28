package frameworkprovider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSLOResourceMetricSpecBlocksWithoutClickHouse(t *testing.T) {
	blocks := sloResourceMetricSpecBlocksWithoutClickHouse()

	require.NotContains(t, blocks, "clickhouse")
	require.Contains(t, blocks, "prometheus")
}
