package frameworkprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type dateTimeValidator struct{}

func (v dateTimeValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v dateTimeValidator) MarkdownDescription(_ context.Context) string {
	return "Ensure that the attribute is a valid date time in RFC3339 notation"
}

func (v dateTimeValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if _, err := time.Parse(time.RFC3339, req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid datetime format",
			fmt.Sprintf("Invalid datetime format: %s", v),
		)
	}
}
