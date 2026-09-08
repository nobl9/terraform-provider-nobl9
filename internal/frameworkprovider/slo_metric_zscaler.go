package frameworkprovider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// ZscalerModel represents an aggregate ZDX application metric query.
type ZscalerModel struct {
	AppID      types.Int64  `tfsdk:"app_id"`
	LocationID types.Int64  `tfsdk:"location_id"`
	Metric     types.String `tfsdk:"metric"`
}

func zscalerToModel(src *v1alphaSLO.ZscalerMetric) *ZscalerModel {
	if src == nil {
		return nil
	}
	return &ZscalerModel{
		AppID:      types.Int64Value(src.AppID),
		LocationID: types.Int64PointerValue(src.LocationID),
		Metric:     types.StringValue(src.Metric),
	}
}

func modelToZscaler(model *ZscalerModel) *v1alphaSLO.ZscalerMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.ZscalerMetric{
		AppID:      model.AppID.ValueInt64(),
		LocationID: model.LocationID.ValueInt64Pointer(),
		Metric:     model.Metric.ValueString(),
	}
}

func sloMetricZscalerBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Optional ZDX application query. At most one block is allowed. Supports raw metrics only. " +
			"Values are aggregated across users, optionally filtered by location.",
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"app_id": schema.Int64Attribute{
				Required:    true,
				Description: "Positive integer application ID from ZDX.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"location_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Positive integer ZDX location ID. Omit to aggregate across all locations.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"metric": schema.StringAttribute{
				Required:    true,
				Description: "ZDX metric: score (0-100), pft or dns (milliseconds), or availability (percent).",
				Validators:  []validator.String{stringvalidator.OneOf("score", "pft", "dns", "availability")},
			},
		}},
	}
}
