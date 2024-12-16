package frameworkprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortListBasedOnReferenceList(t *testing.T) {
	type testObject struct {
		Key string
	}
	testObjectsEqualityFunc := func(t1, t2 testObject) bool { return t1.Key == t2.Key }

	testCases := map[string]struct {
		target   []testObject
		ref      []testObject
		expected []testObject
	}{
		"no change": {
			target:   []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			ref:      []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			expected: []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
		},
		"sort": {
			target:   []testObject{{Key: "b"}, {Key: "c"}, {Key: "a"}},
			ref:      []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			expected: []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
		},
		"extra element": {
			target:   []testObject{{Key: "b"}, {Key: "d"}, {Key: "c"}, {Key: "a"}},
			ref:      []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			expected: []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}, {Key: "d"}},
		},
		"missing element": {
			target:   []testObject{{Key: "c"}, {Key: "a"}},
			ref:      []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			expected: []testObject{{Key: "a"}, {Key: "c"}},
		},
		"no ref elements": {
			target:   []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			ref:      []testObject{},
			expected: []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
		},
		"no target elements": {
			target:   []testObject{},
			ref:      []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			expected: []testObject{},
		},
		"nil target": {
			target:   nil,
			ref:      []testObject{{Key: "a"}, {Key: "b"}, {Key: "c"}},
			expected: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := sortListBasedOnReferenceList(tc.target, tc.ref, testObjectsEqualityFunc)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
