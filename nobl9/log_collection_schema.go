package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const logCollectionConfigKey = "log_collection_enabled"

func schemaLogCollection() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "[Logs documentation](https://docs.nobl9.com/Features/direct-logs)",
	}
}

func marshalLogCollectionEnabled(d *schema.ResourceData) *bool {
	lData := d.Get(logCollectionConfigKey)
	value := lData.(bool)
	return &value
}

func unmarshalLogCollectionEnabled(d *schema.ResourceData, l *bool) (diags diag.Diagnostics) {
	if l == nil {
		return
	}
	set(d, logCollectionConfigKey, l, &diags)
	return
}
