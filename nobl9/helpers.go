package nobl9

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/nobl9/nobl9-go/manifest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type resourceInterface interface {
	Get(key string) any
	GetOk(key string) (any, bool)
	GetRawConfig() cty.Value
}

// TODO: Once we introduce a more structured approach to error handling in SDK, this should be removed.
var errConcurrencyIssue = errors.New("operation failed due to concurrency issue but can be retried")

func exactlyOneStringEmpty(str1, str2 string) bool {
	return (str1 == "" && str2 != "") || (str1 != "" && str2 == "")
}

// oneElementSet implements schema.SchemaSetFunc and creates only one element set.
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

	errsStrings := make([]string, 0, len(diags))
	for _, d := range diags {
		errsStrings = append(errsStrings, fmt.Sprintf("%s: %s", d.Summary, d.Detail))
	}
	combinedErrs := strings.Join(errsStrings, "; ")
	return fmt.Errorf("validation failed: %s", combinedErrs)
}

func formatErrorsAsSingleError(errs []error) error {
	errsStrings := make([]string, 0, len(errs))
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

func equalSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func exactlyOneObjectErr[T manifest.Object](objects []T) diag.Diagnostics {
	return diag.Errorf("Expected exactly one %T but got %d.\n"+
		"This is most likely an issue in Nobl9 platform.", *new(T), len(objects))
}

type objectUnmarshalFunc[T manifest.Object] func(*schema.ResourceData, T) diag.Diagnostics

func handleResourceReadResult[T manifest.Object](
	data *schema.ResourceData,
	objects []T,
	unmarshalFunc objectUnmarshalFunc[T],
) diag.Diagnostics {
	switch len(objects) {
	case 0:
		// When we deleted the object.
		data.SetId("")
		return nil
	case 1:
		// When we applied the object.
		return unmarshalFunc(data, objects[0])
	}
	return exactlyOneObjectErr(objects)
}
