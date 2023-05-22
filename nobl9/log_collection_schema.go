package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const logCollectionConfigKey = "log_collection_enabled"

func getLogCollectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "[Logs documentation](https://docs.nobl9.com/Features/SLO_troubleshooting/event-logs)",
	}
}

func setLogCollectionSchema(s map[string]*schema.Schema) {
	s[logCollectionConfigKey] = getLogCollectionSchema()
}

func marshalLogCollectionEnabled(d *schema.ResourceData) *bool {
	lData := d.Get(logCollectionConfigKey)
	value, ok := lData.(bool)
	if !ok {
		return nil
	}
	return &value
}

func unmarshalLogCollectionEnabled(d *schema.ResourceData, l *bool) (diags diag.Diagnostics) {
	if l == nil {
		return
	}
	set(d, logCollectionConfigKey, l, &diags)
	return
}
