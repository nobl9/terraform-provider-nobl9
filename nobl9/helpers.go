package nobl9

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Data interface {
	Get(key string) any
	GetOk(key string) (any, bool)
}

func exactlyOneStringEmpty(str1, str2 string) bool {
	return (str1 == "" && str2 != "") || (str1 != "" && str2 == "")
}

// oneElementSet implements schema.SchemaSetFunc and created only one element set.
// Never use it for sets with more elements as new elements will override the old ones.
func oneElementSet(_ interface{}) int {
	return 0
}

func set(d *schema.ResourceData, key string, value interface{}, diags *diag.Diagnostics) {
	appendError(*diags, d.Set(key, value))
}

func appendError(d diag.Diagnostics, err error) diag.Diagnostics {
	if err != nil {
		return append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Error(),
		})
	}

	return d
}

func diagsToSingleError(diags diag.Diagnostics) error {
	if len(diags) == 0 {
		return nil
	}

	var errsStrings []string
	for _, d := range diags {
		errsStrings = append(errsStrings, fmt.Sprintf("%s: %s", d.Summary, d.Detail))
	}
	combinedErrs := strings.Join(errsStrings, "; ")
	return fmt.Errorf("validation failed: %s", combinedErrs)
}

func formatErrorsAsSingleError(errs []error) error {
	var errsStrings []string
	for _, err := range errs {
		errsStrings = append(errsStrings, err.Error())
	}
	combinedErrs := strings.Join(errsStrings, "; ")
	return fmt.Errorf("validation failed: %s", strings.Trim(combinedErrs, "[]"))
}

func toStringSlice(in []interface{}) []string {
	ret := make([]string, len(in))
	for i, v := range in {
		ret[i] = v.(string)
	}
	return ret
}
