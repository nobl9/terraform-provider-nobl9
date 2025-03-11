package frameworkprovider

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringValue returns [types.String] from a string.
// If the string is empty, it returns [types.StringNull].
func stringValue(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

// sortListBasedOnReferenceList sorts the provided list based on another list as a reference for sorting order.
// Each element of the provided list is matched by equalsFunc to its counterpart
// in the reference list and appended under the same index in the sorted list.
// If an element is not found in the reference list,
// it is appended to the end of the sorted list.
func sortListBasedOnReferenceList[S ~[]E, E any](target, reference S, equalsFunc func(E, E) bool) S {
	if target == nil {
		return nil
	}
	sortedPointers := make([]*E, len(reference))
	extraElements := make(S, 0)
	for _, el := range target {
		matched := false
		for i, refEl := range reference {
			if equalsFunc(el, refEl) {
				sortedPointers[i] = &el
				matched = true
				break
			}
		}
		if !matched {
			extraElements = append(extraElements, el)
		}
	}
	// Remove potential empty elements.
	sortedPointers = slices.DeleteFunc(sortedPointers, func(v *E) bool { return v == nil })
	// Convert pointers back to values.
	sorted := make(S, 0, len(sortedPointers))
	for _, v := range sortedPointers {
		sorted = append(sorted, *v)
	}
	// Add removed elements to the end of the list.
	sorted = append(sorted, extraElements...)
	return sorted
}

func addInvalidSDKClientTypeDiag(diags *diag.Diagnostics, data any) {
	diags.AddError(
		"Unexpected Configure Type",
		fmt.Sprintf(
			"Expected *sdkClient, got: %T. Please report this issue to the provider developers.",
			data,
		),
	)
}
