package nobl9

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// validateDataTime validates the datetime format in RFC3339
func validateDateTime(v interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if _, ok := v.(string); !ok {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid type",
			Detail:        fmt.Sprintf("Expected string value got: %T", v),
			AttributePath: path,
		})
		return diags
	}

	_, err := time.Parse(time.RFC3339, v.(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid datetime format",
			Detail:        fmt.Sprintf("Invalid datetime format: %s", v),
			AttributePath: path,
		})
	}
	return diags
}

func validateMaxLength(fieldName string, maxLength int) func(interface{}, cty.Path) diag.Diagnostics {
	return func(v any, _ cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		if len(v.(string)) > 63 {
			diagnostic := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("%s is too long", fieldName),
				Detail:   fmt.Sprintf("%s cannot be longer than %d characters", fieldName, maxLength),
			}
			diags = append(diags, diagnostic)
		}
		return diags
	}
}
