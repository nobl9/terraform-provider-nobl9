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

// ZscalerModel represents a ZDX report query.
type ZscalerModel struct {
	Type       types.String `tfsdk:"type"`
	AppID      types.Int64  `tfsdk:"app_id"`
	LocationID types.Int64  `tfsdk:"location_id"`
	Metric     types.String `tfsdk:"metric"`
	DeviceID   types.Int64  `tfsdk:"device_id"`
	ProbeID    types.Int64  `tfsdk:"probe_id"`
	LegSrc     types.String `tfsdk:"leg_src"`
	LegDst     types.String `tfsdk:"leg_dst"`
}

func zscalerToModel(src *v1alphaSLO.ZscalerMetric) *ZscalerModel {
	if src == nil {
		return nil
	}
	return &ZscalerModel{
		Type:       types.StringValue(src.Type),
		AppID:      types.Int64Value(src.AppID),
		LocationID: types.Int64PointerValue(src.LocationID),
		Metric:     types.StringValue(src.Metric),
		DeviceID:   types.Int64PointerValue(src.DeviceID),
		ProbeID:    types.Int64PointerValue(src.ProbeID),
		LegSrc:     types.StringPointerValue(src.LegSrc),
		LegDst:     types.StringPointerValue(src.LegDst),
	}
}

func modelToZscaler(model *ZscalerModel) *v1alphaSLO.ZscalerMetric {
	if model == nil {
		return nil
	}
	return &v1alphaSLO.ZscalerMetric{
		Type:       model.Type.ValueString(),
		AppID:      model.AppID.ValueInt64(),
		LocationID: model.LocationID.ValueInt64Pointer(),
		Metric:     model.Metric.ValueString(),
		DeviceID:   model.DeviceID.ValueInt64Pointer(),
		ProbeID:    model.ProbeID.ValueInt64Pointer(),
		LegSrc:     model.LegSrc.ValueStringPointer(),
		LegDst:     model.LegDst.ValueStringPointer(),
	}
}

func sloMetricZscalerBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Optional ZDX report query. At most one block is allowed. Supports raw metrics only. " +
			"Selects one returned time series without aggregating across devices or probes.",
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required: true,
				Description: "Report type: application, web-probe, or cloudpath. " +
					"Application queries require app_id and allow location_id. " +
					"Probe queries require app_id, device_id, and probe_id and do not allow location_id.",
				Validators: []validator.String{stringvalidator.OneOf("application", "web-probe", "cloudpath")},
			},
			"app_id": schema.Int64Attribute{
				Required:    true,
				Description: "Positive integer application ID from ZDX.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"location_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Positive integer ZDX location ID for application queries only. Omit to include all locations.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"metric": schema.StringAttribute{
				Required:    true,
				Description: "Metric for the selected type: application supports score and pft; web-probe supports pft, ttfb, dns, and availability; cloudpath supports latency and loss. Times are milliseconds; availability and loss are percentages; score is 0-100.",
				Validators: []validator.String{
					stringvalidator.OneOf("score", "pft", "ttfb", "dns", "availability", "latency", "loss"),
				},
			},
			"device_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Positive integer device ID. Required for web-probe and cloudpath; not allowed for application.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"probe_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Positive integer configured probe ID. Required for web-probe and cloudpath; not allowed for application.",
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"leg_src": schema.StringAttribute{
				Optional:    true,
				Description: "CloudPath source segment label, matched exactly against leg_src in the response. Supply together with leg_dst or omit both to select end/end. Not allowed for other report types.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"leg_dst": schema.StringAttribute{
				Optional:    true,
				Description: "CloudPath destination segment label, matched exactly against leg_dst in the response. Supply together with leg_src or omit both to select end/end. Not allowed for other report types.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
		}},
	}
}
