package reflectiontuils_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/nobl9/terraform-provider-nobl9/internal/frameworkprovider"
	"github.com/nobl9/terraform-provider-nobl9/internal/reflectiontuils"
)

// Test structs for GetAttributeTypes function
type TestModelWithTags struct {
	StringField  types.String  `tfsdk:"string_field"`
	IntField     types.Int64   `tfsdk:"int_field"`
	BoolField    types.Bool    `tfsdk:"bool_field"`
	FloatField   types.Float64 `tfsdk:"float_field"`
	ObjectField  types.Object  `tfsdk:"object_field"`
	ListField    types.List    `tfsdk:"list_field"`
	SetField     types.Set     `tfsdk:"set_field"`
	MapField     types.Map     `tfsdk:"map_field"`
	DynamicField types.Dynamic `tfsdk:"dynamic_field"`
}

type TestModelWithoutTags struct {
	StringField types.String
	IntField    types.Int64
	BoolField   types.Bool
}

type TestModelWithMixedTags struct {
	WithTag    types.String `tfsdk:"with_tag"`
	WithoutTag types.String
	SkipTag    types.String `tfsdk:"-"`
	EmptyTag   types.String `tfsdk:""`
}

type TestModelWithNonTypeFields struct {
	StringField   types.String `tfsdk:"string_field"`
	RegularString string       `tfsdk:"regular_string"`
	RegularInt    int          `tfsdk:"regular_int"`
}

type TestModelWithUnexportedFields struct {
	ExportedField   types.String `tfsdk:"exported_field"`
	unexportedField types.String `tfsdk:"unexported_field"`
}

type TestModelEmpty struct{}

func TestGetAttributeTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]attr.Type
	}{
		{
			name:  "struct with all terraform types",
			input: TestModelWithTags{},
			expected: map[string]attr.Type{
				"string_field":  types.StringType,
				"int_field":     types.Int64Type,
				"bool_field":    types.BoolType,
				"float_field":   types.Float64Type,
				"object_field":  types.ObjectType{AttrTypes: map[string]attr.Type{}},
				"list_field":    types.ListType{},
				"set_field":     types.SetType{},
				"map_field":     types.MapType{},
				"dynamic_field": types.DynamicType,
			},
		},
		{
			name:     "struct without tfsdk tags",
			input:    TestModelWithoutTags{},
			expected: map[string]attr.Type{},
		},
		{
			name:  "struct with mixed tags",
			input: TestModelWithMixedTags{},
			expected: map[string]attr.Type{
				"with_tag": types.StringType,
			},
		},
		{
			name:  "struct with non-Type fields",
			input: TestModelWithNonTypeFields{},
			expected: map[string]attr.Type{
				"string_field": types.StringType,
				// regular_string and regular_int should be skipped as they don't have Type() methods
			},
		},
		{
			name:  "struct with unexported fields",
			input: TestModelWithUnexportedFields{},
			expected: map[string]attr.Type{
				"exported_field": types.StringType,
				// unexported_field should be skipped
			},
		},
		{
			name:     "empty struct",
			input:    TestModelEmpty{},
			expected: map[string]attr.Type{},
		},
		{
			name:  "pointer to struct",
			input: &TestModelWithTags{},
			expected: map[string]attr.Type{
				"string_field":  types.StringType,
				"int_field":     types.Int64Type,
				"bool_field":    types.BoolType,
				"float_field":   types.Float64Type,
				"object_field":  types.ObjectType{AttrTypes: map[string]attr.Type{}},
				"list_field":    types.ListType{},
				"set_field":     types.SetType{},
				"map_field":     types.MapType{},
				"dynamic_field": types.DynamicType,
			},
		},
		{
			name: "nil pointer to struct",
			input: func() *TestModelWithTags {
				var ptr *TestModelWithTags
				return ptr
			}(),
			expected: map[string]attr.Type{
				"string_field":  types.StringType,
				"int_field":     types.Int64Type,
				"bool_field":    types.BoolType,
				"float_field":   types.Float64Type,
				"object_field":  types.ObjectType{AttrTypes: map[string]attr.Type{}},
				"list_field":    types.ListType{},
				"set_field":     types.SetType{},
				"map_field":     types.MapType{},
				"dynamic_field": types.DynamicType,
			},
		},
		{
			name:     "non-struct input (string)",
			input:    "not a struct",
			expected: map[string]attr.Type{},
		},
		{
			name:     "non-struct input (int)",
			input:    42,
			expected: map[string]attr.Type{},
		},
		{
			name:     "non-struct input (slice)",
			input:    []string{"test"},
			expected: map[string]attr.Type{},
		},
		{
			name:     "non-struct input (map)",
			input:    map[string]string{"key": "value"},
			expected: map[string]attr.Type{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reflectiontuils.GetAttributeTypes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAttributeTypes_RealWorldExample(t *testing.T) {
	// Test with actual PeriodModel from the SLO code
	result := reflectiontuils.GetAttributeTypes(frameworkprovider.PeriodModel{})

	expected := map[string]attr.Type{
		"begin": types.StringType,
		"end":   types.StringType,
	}

	assert.Equal(t, expected, result)
}

func TestGetAttributeTypes_WithInitializedValues(t *testing.T) {
	// Test that the function works with initialized values
	model := TestModelWithTags{
		StringField: types.StringValue("test"),
		IntField:    types.Int64Value(42),
		BoolField:   types.BoolValue(true),
	}

	result := reflectiontuils.GetAttributeTypes(model)

	expected := map[string]attr.Type{
		"string_field":  types.StringType,
		"int_field":     types.Int64Type,
		"bool_field":    types.BoolType,
		"float_field":   types.Float64Type,
		"object_field":  types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"list_field":    types.ListType{},
		"set_field":     types.SetType{},
		"map_field":     types.MapType{},
		"dynamic_field": types.DynamicType,
	}

	assert.Equal(t, expected, result)
}

func BenchmarkGetAttributeTypes(b *testing.B) {
	model := TestModelWithTags{}
	for b.Loop() {
		reflectiontuils.GetAttributeTypes(model)
	}
}
