package nobl9

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//nolint:tparallel
func TestMarshalAnomalyConfig(t *testing.T) {
	testCases := []struct {
		name                 string
		alertMethods         []interface{}
		expectedLength       int
		expectedAlertMethods []struct{ name, project string }
	}{
		{
			name:           "Empty alert method list",
			alertMethods:   []interface{}{},
			expectedLength: 0,
		},
		{
			name: "Single alert method",
			alertMethods: []interface{}{
				map[string]interface{}{
					"name":    "the-net-is-vast",
					"project": "and-infinite",
				},
			},
			expectedLength: 1,
			expectedAlertMethods: []struct{ name, project string }{
				{
					name:    "the-net-is-vast",
					project: "and-infinite",
				},
			},
		},
		{
			name: "Multiple alert methods",
			alertMethods: []interface{}{
				map[string]interface{}{
					"name":    "the-net",
					"project": "is-vast",
				},
				map[string]interface{}{
					"name":    "and",
					"project": "infinite",
				},
			},
			expectedLength: 2,
			expectedAlertMethods: []struct{ name, project string }{
				{
					name:    "the-net",
					project: "is-vast",
				},
				{
					name:    "and",
					project: "infinite",
				},
			},
		},
		{
			name: "Multiple alert methods in the same project",
			alertMethods: []interface{}{
				map[string]interface{}{
					"name":    "the",
					"project": "ghost",
				},
				map[string]interface{}{
					"name":    "net",
					"project": "ghost",
				},
				map[string]interface{}{
					"name":    "is",
					"project": "ghost",
				},
				map[string]interface{}{
					"name":    "vast",
					"project": "ghost",
				},
				map[string]interface{}{
					"name":    "and",
					"project": "ghost",
				},
				map[string]interface{}{
					"name":    "infinite",
					"project": "ghost",
				},
			},
			expectedLength: 6,
			expectedAlertMethods: []struct{ name, project string }{
				{
					name:    "the",
					project: "ghost",
				},
				{
					name:    "net",
					project: "ghost",
				},
				{
					name:    "is",
					project: "ghost",
				},
				{
					name:    "vast",
					project: "ghost",
				},
				{
					name:    "and",
					project: "ghost",
				},
				{
					name:    "infinite",
					project: "ghost",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			anomalyConfig := schema.NewSet(oneElementSet, []interface{}{
				map[string]interface{}{
					"no_data": schema.NewSet(oneElementSet, []interface{}{
						map[string]interface{}{
							"alert_method": tc.alertMethods,
						},
					}),
				},
			})

			result := marshalAnomalyConfig(anomalyConfig)
			assert.NotNil(t, result)
			assert.NotNil(t, result.NoData)
			assert.Equal(t, tc.expectedLength, len(result.NoData.AlertMethods))

			for i, expected := range tc.expectedAlertMethods {
				assert.Equal(t, expected.name, result.NoData.AlertMethods[i].Name)
				assert.Equal(t, expected.project, result.NoData.AlertMethods[i].Project)
			}
		})
	}
}

//nolint:tparallel
func TestMarshalAnomalyConfigAlertMethods(t *testing.T) {
	testCases := []struct {
		name          string
		alertMethods  []interface{}
		expectedAlert []v1alpha.AnomalyConfigAlertMethod
	}{
		{
			name:          "Empty alert methods slice",
			alertMethods:  []interface{}{},
			expectedAlert: []v1alpha.AnomalyConfigAlertMethod{},
		},
		{
			name: "Alert methods slice with nil values",
			alertMethods: []interface{}{
				nil,
				map[string]interface{}{
					"name":    "the-net-is-vast",
					"project": "and-infinite",
				},
				nil,
			},
			expectedAlert: []v1alpha.AnomalyConfigAlertMethod{
				{
					Name:    "the-net-is-vast",
					Project: "and-infinite",
				},
			},
		},
		{
			name: "Alert methods with valid data",
			alertMethods: []interface{}{
				map[string]interface{}{
					"name":    "the-net",
					"project": "is-vast",
				},
				map[string]interface{}{
					"name":    "and",
					"project": "infinite",
				},
			},
			expectedAlert: []v1alpha.AnomalyConfigAlertMethod{
				{
					Name:    "the-net",
					Project: "is-vast",
				},
				{
					Name:    "and",
					Project: "infinite",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			alertMethodsTF := tc.alertMethods
			result := marshalAnomalyConfigAlertMethods(alertMethodsTF)
			assert.Equal(t, tc.expectedAlert, result)
		})
	}
}
