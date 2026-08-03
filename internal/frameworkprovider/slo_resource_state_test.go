package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_preserveNullIndicatorProject(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source          *SLOResourceModel
		target          *SLOResourceModel
		expectedProject types.String
	}{
		"omitted project": {
			source: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringNull(),
				}},
			},
			target: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("default"),
				}},
			},
			expectedProject: types.StringNull(),
		},
		"explicit project": {
			source: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("metrics"),
				}},
			},
			target: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("metrics"),
				}},
			},
			expectedProject: types.StringValue("metrics"),
		},
		"missing source indicator": {
			source: &SLOResourceModel{},
			target: &SLOResourceModel{
				Indicator: []IndicatorModel{{
					Project: types.StringValue("default"),
				}},
			},
			expectedProject: types.StringValue("default"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			preserveNullIndicatorProject(test.source, test.target)

			assert.Equal(t, test.expectedProject, test.target.Indicator[0].Project)
		})
	}
}

func Test_preserveNullLogicMonitorIDs(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source   *LogicMonitorModel
		target   *LogicMonitorModel
		expected types.Int64
	}{
		"omitted with API zero": {
			source:   new(logicMonitorModel(types.Int64Null())),
			target:   new(logicMonitorModel(types.Int64Value(0))),
			expected: types.Int64Null(),
		},
		"omitted with API nonzero": {
			source:   new(logicMonitorModel(types.Int64Null())),
			target:   new(logicMonitorModel(types.Int64Value(123))),
			expected: types.Int64Value(123),
		},
		"explicit zero": {
			source:   new(logicMonitorModel(types.Int64Value(0))),
			target:   new(logicMonitorModel(types.Int64Value(0))),
			expected: types.Int64Value(0),
		},
		"explicit nonzero": {
			source:   new(logicMonitorModel(types.Int64Value(123))),
			target:   new(logicMonitorModel(types.Int64Value(123))),
			expected: types.Int64Value(123),
		},
		"explicit nonzero with API zero": {
			source:   new(logicMonitorModel(types.Int64Value(123))),
			target:   new(logicMonitorModel(types.Int64Value(0))),
			expected: types.Int64Value(0),
		},
		"API-only zero": {
			target:   new(logicMonitorModel(types.Int64Value(0))),
			expected: types.Int64Value(0),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			preserveNullLogicMonitorIDs(test.source, test.target)

			assert.Equal(t, test.expected, test.target.DeviceDataSourceInstanceID)
			assert.Equal(t, test.expected, test.target.GraphID)
		})
	}
}

func Test_preserveNullLogicMonitorMetricIDs_matchesObjectivesByName(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source    *SLOResourceModel
		target    *SLOResourceModel
		sortLists bool
		expected  []types.Int64
	}{
		"removed objective": {
			source: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("removed"), types.Int64Null()),
					newLogicMonitorRawMetricObjective(types.StringValue("retained"), types.Int64Value(0)),
				},
			},
			target: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("retained"), types.Int64Value(0)),
				},
			},
			expected: []types.Int64{types.Int64Value(0)},
		},
		"reordered objectives": {
			source: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("first"), types.Int64Null()),
					newLogicMonitorRawMetricObjective(types.StringValue("second"), types.Int64Value(0)),
				},
			},
			target: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("second"), types.Int64Value(0)),
					newLogicMonitorRawMetricObjective(types.StringValue("first"), types.Int64Value(0)),
				},
			},
			expected: []types.Int64{types.Int64Value(0), types.Int64Null()},
		},
		"computed name with equal cardinality": {
			source: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringUnknown(), types.Int64Null()),
				},
			},
			target: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("generated"), types.Int64Value(0)),
				},
			},
			expected: []types.Int64{types.Int64Null()},
		},
		"computed name with unequal cardinality": {
			source: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringUnknown(), types.Int64Null()),
					newLogicMonitorRawMetricObjective(types.StringValue("other"), types.Int64Null()),
				},
			},
			target: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("generated"), types.Int64Value(0)),
				},
			},
			expected: []types.Int64{types.Int64Value(0)},
		},
		"computed and explicit names": {
			source: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringUnknown(), types.Int64Null()),
					newLogicMonitorRawMetricObjective(types.StringValue("explicit"), types.Int64Value(0)),
				},
			},
			target: &SLOResourceModel{
				Objectives: []ObjectiveModel{
					newLogicMonitorRawMetricObjective(types.StringValue("generated"), types.Int64Value(0)),
					newLogicMonitorRawMetricObjective(types.StringValue("explicit"), types.Int64Value(0)),
				},
			},
			sortLists: true,
			expected:  []types.Int64{types.Int64Value(0), types.Int64Null()},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if test.sortLists {
				(&SLOResource{}).sortLists(test.source, test.target)
			}
			preserveNullLogicMonitorMetricIDs(test.source, test.target)

			for i, expected := range test.expected {
				assertLogicMonitorRawMetricIDs(t, test.target.Objectives[i], expected)
			}
		})
	}
}

func Test_preserveNullLogicMonitorMetricIDs_countMetrics(t *testing.T) {
	t.Parallel()

	source := &SLOResourceModel{
		Objectives: []ObjectiveModel{{
			Name:         types.StringValue("objective"),
			CountMetrics: []CountMetricsModel{newLogicMonitorCountMetrics(types.Int64Null())},
		}},
	}
	target := &SLOResourceModel{
		Objectives: []ObjectiveModel{{
			Name:         types.StringValue("objective"),
			CountMetrics: []CountMetricsModel{newLogicMonitorCountMetrics(types.Int64Value(0))},
		}},
	}

	preserveNullLogicMonitorMetricIDs(source, target)

	targetCountMetrics := target.Objectives[0].CountMetrics[0]
	for name, specs := range map[string][]MetricSpecModel{
		"good":       targetCountMetrics.Good,
		"bad":        targetCountMetrics.Bad,
		"total":      targetCountMetrics.Total,
		"good_total": targetCountMetrics.GoodTotal,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logicMonitor := specs[0].LogicMonitor[0]
			assert.Equal(t, types.Int64Null(), logicMonitor.DeviceDataSourceInstanceID)
			assert.Equal(t, types.Int64Null(), logicMonitor.GraphID)
		})
	}
}

func assertLogicMonitorRawMetricIDs(t *testing.T, objective ObjectiveModel, expected types.Int64) {
	t.Helper()
	logicMonitor := objective.RawMetric[0].Query[0].LogicMonitor[0]
	assert.Equal(t, expected, logicMonitor.DeviceDataSourceInstanceID)
	assert.Equal(t, expected, logicMonitor.GraphID)
}

func newLogicMonitorRawMetricObjective(
	name types.String,
	value types.Int64,
) ObjectiveModel {
	return ObjectiveModel{
		Name: name,
		RawMetric: []RawMetricModel{{
			Query: newLogicMonitorMetricSpecs(value),
		}},
	}
}

func logicMonitorModel(value types.Int64) LogicMonitorModel {
	return LogicMonitorModel{
		DeviceDataSourceInstanceID: value,
		GraphID:                    value,
	}
}

func newLogicMonitorCountMetrics(value types.Int64) CountMetricsModel {
	return CountMetricsModel{
		Good:      newLogicMonitorMetricSpecs(value),
		Bad:       newLogicMonitorMetricSpecs(value),
		Total:     newLogicMonitorMetricSpecs(value),
		GoodTotal: newLogicMonitorMetricSpecs(value),
	}
}

func newLogicMonitorMetricSpecs(value types.Int64) []MetricSpecModel {
	return []MetricSpecModel{{
		LogicMonitor: []LogicMonitorModel{logicMonitorModel(value)},
	}}
}
