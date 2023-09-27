package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const releaseChannel = "release_channel"

func schemaReleaseChannel() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Release channel of the created datasource [stable/beta]",
	}
}

func marshalReleaseChannel(d *schema.ResourceData) string {
	return d.Get(releaseChannel).(string)
}

func unmarshalReleaseChannel(d *schema.ResourceData, rc string) (diags diag.Diagnostics) {
	if rc == "" {
		return
	}
	err := d.Set(releaseChannel, rc)
	return appendError(diags, err)
}
