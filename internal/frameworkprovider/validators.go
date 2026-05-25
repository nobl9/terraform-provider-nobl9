package frameworkprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

type sumoLogicQueriesTypeValidator struct{}

func (v sumoLogicQueriesTypeValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v sumoLogicQueriesTypeValidator) MarkdownDescription(_ context.Context) string {
	return "Ensure that queries block is not used with logs type"
}

func (v sumoLogicQueriesTypeValidator) ValidateList(
	ctx context.Context,
	req validator.ListRequest,
	resp *validator.ListResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}
	typePath := req.Path.ParentPath().AtName("type")
	var typeVal types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, typePath, &typeVal)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if typeVal.ValueString() == "logs" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"queries block is not supported for logs type",
			"Multi-query (ABC pattern) is only supported for metrics type. Use the 'query' attribute for logs.",
		)
	}
}

type sumoLogicQueriesConflictWithQueryValidator struct{}

func (v sumoLogicQueriesConflictWithQueryValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v sumoLogicQueriesConflictWithQueryValidator) MarkdownDescription(_ context.Context) string {
	return "Ensure that 'queries' block conflicts with deprecated 'query' attribute"
}

func (v sumoLogicQueriesConflictWithQueryValidator) ValidateList(
	ctx context.Context,
	req validator.ListRequest,
	resp *validator.ListResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}
	queryPath := req.Path.ParentPath().AtName("query")
	var queryVal types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, queryPath, &queryVal)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !queryVal.IsNull() && !queryVal.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting attributes",
			"'queries' block conflicts with the deprecated 'query' attribute. Use one or the other, not both.",
		)
	}
}

type sumoLogicQueriesUniqueRowIDValidator struct{}

func (v sumoLogicQueriesUniqueRowIDValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v sumoLogicQueriesUniqueRowIDValidator) MarkdownDescription(_ context.Context) string {
	return "Ensure that row_id values in queries block are unique"
}

func (v sumoLogicQueriesUniqueRowIDValidator) ValidateList(
	_ context.Context,
	req validator.ListRequest,
	resp *validator.ListResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || len(req.ConfigValue.Elements()) == 0 {
		return
	}
	seen := make(map[string]bool)
	for _, elem := range req.ConfigValue.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		rowIDAttr, exists := obj.Attributes()["row_id"]
		if !exists {
			continue
		}
		rowID, ok := rowIDAttr.(types.String)
		if !ok || rowID.IsNull() || rowID.IsUnknown() {
			continue
		}
		id := rowID.ValueString()
		if seen[id] {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"duplicate row_id in queries block",
				fmt.Sprintf("Each query must have a unique row_id. Found duplicate: %q", id),
			)
			return
		}
		seen[id] = true
	}
}
