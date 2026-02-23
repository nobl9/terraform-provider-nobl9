package frameworkprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func upgradeSLOStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var rawStateData map[string]json.RawMessage
	if err := json.Unmarshal(req.RawState.JSON, &rawStateData); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Unmarshal Prior State",
			fmt.Sprintf("Failed to unmarshal SLO state during upgrade from version 0: %s", err),
		)
		return
	}

	if _, ok := rawStateData["attachments"]; ok {
		delete(rawStateData, "attachments")
		removedFieldsWarning(0, []string{"attachments"}, resp)
	}

	delete(rawStateData, "retrieve_historical_data_from")

	upgradedJSON, err := json.Marshal(rawStateData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Marshal Upgraded State",
			fmt.Sprintf("Failed to marshal SLO state during upgrade from version 0: %s", err),
		)
		return
	}

	upgradedRawState := tfprotov6.RawState{JSON: upgradedJSON}
	schemaType := sloResourceSchema.Type().TerraformType(ctx)
	rawValue, err := upgradedRawState.UnmarshalWithOpts(
		schemaType,
		tfprotov6.UnmarshalOpts{ValueFromJSONOpts: tftypes.ValueFromJSONOpts{IgnoreUndefinedAttributes: true}},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Unmarshal Upgraded State",
			fmt.Sprintf("Failed to unmarshal upgraded SLO state into schema type: %s", err),
		)
		return
	}
	resp.State = tfsdk.State{
		Schema: sloResourceSchema,
		Raw:    rawValue,
	}
}

func removedFieldsWarning(version int, fields []string, resp *resource.UpgradeStateResponse) {
	resp.Diagnostics.AddWarning(
		"SLO State Upgrade: Deprecated Fields Removed",
		fmt.Sprintf(
			"The following deprecated fields were removed from the SLO state "+
				"during migration from schema version %d: %s. "+
				"These fields are no longer supported and any previously stored values have been dropped.",
			version, strings.Join(fields, ", "),
		),
	)
}
