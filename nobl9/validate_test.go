package nobl9

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestValidateDateTime(t *testing.T) {
	cases := []struct {
		name     string
		in       interface{}
		expected diag.Diagnostics
	}{
		{
			name:     "valid datetime",
			in:       "2021-09-26T07:00:00Z",
			expected: nil,
		},
		{
			name: "invalid datetime",
			in:   "2021-09-26T07:00:00",
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid datetime format",
				Detail:        "Invalid datetime format: 2021-09-26T07:00:00",
				AttributePath: nil,
			}},
		},
		{
			name: "empty datetime",
			in:   "",
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid datetime format",
				Detail:        "Invalid datetime format: ",
				AttributePath: nil,
			}},
		},
		{
			name: "nil datetime",
			in:   nil,
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid type",
				Detail:        "Expected string value got: <nil>",
				AttributePath: nil,
			}},
		},
		{
			name: "invalid type",
			in:   123,
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid type",
				Detail:        "Expected string value got: int",
				AttributePath: nil,
			}},
		},
		{
			name: "invalid timezone",
			in:   "2021-09-26T07:00:00+01:00Z",
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid datetime format",
				Detail:        "Invalid datetime format: 2021-09-26T07:00:00+01:00Z",
				AttributePath: nil,
			}},
		},
		{
			name: "invalid date",
			in:   "2021-09-31T07:00:00Z",
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid datetime format",
				Detail:        "Invalid datetime format: 2021-09-31T07:00:00Z",
				AttributePath: nil,
			}},
		},
		{
			name: "invalid time",
			in:   "2021-09-26T24:00:00Z",
			expected: diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid datetime format",
				Detail:        "Invalid datetime format: 2021-09-26T24:00:00Z",
				AttributePath: nil,
			}},
		},
		{
			name: "invalid seconds",
			in:   "2021-09-26T07:00:60Z",
			expected: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Invalid datetime format",
				Detail:   "Invalid datetime format: 2021-09-26T07:00:60Z",
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := validateDateTime(tc.in, nil)
			if len(tc.expected) != len(actual) {
				t.Errorf("expected %d diagnostics, got %d", len(tc.expected), len(actual))
			}
			for i, d := range actual {
				if d.Severity != tc.expected[i].Severity {
					t.Errorf("expected severity %d, got %d", d.Severity, tc.expected[i].Severity)
				}
				if d.Summary != tc.expected[i].Summary {
					t.Errorf("expected summary %s, got %s", d.Summary, tc.expected[i].Summary)
				}
				if d.Detail != tc.expected[i].Detail {
					t.Errorf("expected detail %s, got %s", d.Detail, tc.expected[i].Detail)
				}
			}
		})
	}
}
