package frameworkprovider

import "github.com/hashicorp/terraform-plugin-framework/types"

// stringValue returns [types.String] from a string.
// If the string is empty, it returns [types.StringNull].
func stringValue(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}
