package frameworkprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func newDateTimeValidator(layout string) dateTimeValidator {
	return dateTimeValidator{layout: layout}
}

type dateTimeValidator struct {
	layout string
}

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
	if isNullOrUnknown(req.ConfigValue) {
		return
	}
	value := req.ConfigValue.ValueString()
	if _, err := time.Parse(v.layout, value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid datetime format",
			fmt.Sprintf("Invalid datetime format: %q", value),
		)
	}
}
