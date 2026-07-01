package nobl9

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
)

func TestReportFiltersProjectScopeSchema(t *testing.T) {
	projectScope := reportProjectScopeSchema(t)

	assert.Equal(t, schema.TypeString, projectScope.Type)
	assert.True(t, projectScope.Optional)
	assert.True(t, projectScope.Computed)
	assert.True(t, projectScope.DiffSuppressOnRefresh)
	assert.True(t, projectScope.DiffSuppressFunc("filters.0.project_scope", "", "selected", nil))
	assert.True(t, projectScope.DiffSuppressFunc("filters.0.project_scope", "selected", "", nil))
	assert.False(t, projectScope.DiffSuppressFunc("filters.0.project_scope", "", "all", nil))
	assert.False(t, projectScope.DiffSuppressFunc("filters.0.project_scope", "selected", "all", nil))

	path := cty.GetAttrPath("filters").IndexInt(0).GetAttr("project_scope")
	assert.Empty(t, projectScope.ValidateDiagFunc("selected", path))
	assert.Empty(t, projectScope.ValidateDiagFunc("all", path))
	assert.NotEmpty(t, projectScope.ValidateDiagFunc("unsupported", path))
}

func TestMarshalReportFiltersProjectScope(t *testing.T) {
	filters := marshalReportFilters([]interface{}{
		map[string]interface{}{
			"project_scope": "all",
			"projects":      []interface{}{},
			"service":       []interface{}{},
			"slo":           []interface{}{},
			"label":         []interface{}{},
		},
	})

	require.NotNil(t, filters)
	assert.Equal(t, v1alphaReport.ProjectScopeAll, filters.ProjectScope)
}

func TestMarshalReportFiltersProjectScopeAllWithLabels(t *testing.T) {
	filters := marshalReportFilters([]interface{}{
		map[string]interface{}{
			"project_scope": "all",
			"projects":      []interface{}{},
			"service":       []interface{}{},
			"slo":           []interface{}{},
			"label": []interface{}{
				map[string]interface{}{
					"key":    "team",
					"values": []interface{}{"platform"},
				},
			},
		},
	})

	require.NotNil(t, filters)
	assert.Equal(t, v1alphaReport.ProjectScopeAll, filters.ProjectScope)
	assert.Equal(t, v1alpha.Labels{"team": []string{"platform"}}, filters.Labels)

	redLTE := 0.8
	greenGT := 0.95
	report := v1alphaReport.New(
		v1alphaReport.Metadata{
			Name:        "test-report",
			DisplayName: "Test Report",
		},
		v1alphaReport.Spec{
			Shared:  true,
			Filters: filters,
			SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
				TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
					Snapshot: v1alphaReport.SnapshotTimeFrame{
						Point: v1alphaReport.SnapshotPointLatest,
					},
					TimeZone: "Europe/Warsaw",
				},
				RowGroupBy: v1alphaReport.RowGroupByProject,
				Columns: []v1alphaReport.ColumnSpec{
					{
						DisplayName: "Column 1",
						Labels:      v1alpha.Labels{"team": []string{"platform"}},
					},
				},
				Thresholds: v1alphaReport.Thresholds{
					RedLessThanOrEqual: &redLTE,
					GreenGreaterThan:   &greenGT,
					ShowNoData:         true,
				},
			},
		},
	)

	assert.Empty(t, manifest.Validate([]manifest.Object{report}))
}

func TestUnmarshalReportFiltersProjectScope(t *testing.T) {
	reportResource := resourceReportFactory(reportSystemHealthReview{})
	data := schema.TestResourceDataRaw(t, reportResource.Schema, nil)
	filters := &v1alphaReport.Filters{ProjectScope: v1alphaReport.ProjectScopeAll}

	require.NoError(t, unmarshalReportFilters(data, filters))

	rawFilters := data.Get("filters").([]interface{})
	require.Len(t, rawFilters, 1)
	filtersMap := rawFilters[0].(map[string]interface{})
	assert.Equal(t, "all", filtersMap["project_scope"])
}

func reportProjectScopeSchema(t *testing.T) *schema.Schema {
	t.Helper()

	filtersSchema := schemaFilters()
	filtersResource, ok := filtersSchema.Elem.(*schema.Resource)
	require.True(t, ok)
	projectScope, ok := filtersResource.Schema["project_scope"]
	require.True(t, ok)
	return projectScope
}
