package nobl9

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const releaseChannel = "release_channel"

func schemaReleaseChannel() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Release channel of the created data source [stable/beta]",
		ValidateDiagFunc: validation.ToDiagFunc(
			validation.StringInSlice([]string{
				v1alpha.ReleaseChannelStable.String(),
				v1alpha.ReleaseChannelBeta.String(),
				v1alpha.ReleaseChannelAlpha.String(),
			}, false),
		),
	}
}

func marshalReleaseChannel(d resourceInterface, diags diag.Diagnostics) v1alpha.ReleaseChannel {
	rc, ok := d.Get(releaseChannel).(string)
	if !ok {
		return 0
	}
	result, err := v1alpha.ParseReleaseChannel(rc)
	if err != nil {
		appendError(diags, fmt.Errorf("invalid release channel '%s'", rc))
		return 0
	}
	return result
}

func unmarshalReleaseChannel(d *schema.ResourceData, rc v1alpha.ReleaseChannel) (diags diag.Diagnostics) {
	err := d.Set(releaseChannel, rc.String())
	return appendError(diags, err)
}
