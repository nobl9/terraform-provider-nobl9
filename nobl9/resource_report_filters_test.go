package nobl9

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
