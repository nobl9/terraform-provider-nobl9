package frameworkprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_unorderedValuesEqual(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		a        []attr.Value
		b        []attr.Value
		expected bool
	}{
		"same order": {
			a:        stringValues("a", "b"),
			b:        stringValues("a", "b"),
			expected: true,
		},
		"different order": {
			a:        stringValues("a", "b"),
			b:        stringValues("b", "a"),
			expected: true,
		},
		"different value": {
			a: stringValues("a", "b"),
			b: stringValues("a", "c"),
		},
		"different length": {
			a: stringValues("a"),
			b: stringValues("a", "b"),
		},
		"duplicate values": {
			a:        stringValues("a", "a"),
			b:        stringValues("a", "a"),
			expected: true,
		},
		"different duplicate values": {
			a: stringValues("a", "a"),
			b: stringValues("a", "b"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, unorderedValuesEqual(test.a, test.b))
		})
	}
}

func stringValues(values ...string) []attr.Value {
	result := make([]attr.Value, 0, len(values))
	for _, value := range values {
		result = append(result, types.StringValue(value))
	}
	return result
}
