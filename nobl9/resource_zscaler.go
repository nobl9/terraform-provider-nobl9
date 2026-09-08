package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
)

const (
	zscalerDirectType     = "zscaler"
	zscalerAgentType      = "zscaler"
	zscalerAgentConfigKey = "zscaler_config"
)

type zscalerDirectSpec struct{}

func (zscalerDirectSpec) GetDescription() string {
	return "Collects aggregate Zscaler Digital Experience (ZDX) application metrics through OneAPI."
}

func (zscalerDirectSpec) GetSchema() map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"vanity_domain": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "OneAPI tenant name before .zslogin.net, without a URL scheme or domain suffix.",
		},
		"client_id": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Sensitive:        true,
			Description:      "OneAPI client ID. Required when creating the data source. The client must have access to ZDX.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		},
		"client_secret": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Sensitive:        true,
			Description:      "OneAPI client secret. Required when creating the data source. Legacy ZDX API keys are not supported.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		},
		releaseChannel: {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     v1alpha.ReleaseChannelBeta.String(),
			Description: "Release channel of the data source. Only beta is supported.",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
				[]string{v1alpha.ReleaseChannelBeta.String()}, false,
			)),
		},
	}
	setHistoricalDataRetrievalSchema(result)
	setLogCollectionSchema(result)
	return result
}

func (zscalerDirectSpec) MarshalSpec(r resourceInterface) v1alphaDirect.Spec {
	return v1alphaDirect.Spec{Zscaler: &v1alphaDirect.ZscalerConfig{
		VanityDomain: r.Get("vanity_domain").(string),
		ClientID:     r.Get("client_id").(string),
		ClientSecret: r.Get("client_secret").(string),
	}}
}

func (zscalerDirectSpec) UnmarshalSpec(d *schema.ResourceData, spec v1alphaDirect.Spec) diag.Diagnostics {
	if spec.Zscaler == nil {
		return diag.Errorf("Zscaler configuration is missing from the API response")
	}
	var diags diag.Diagnostics
	set(d, "vanity_domain", spec.Zscaler.VanityDomain, &diags)
	set(d, "description", spec.Description, &diags)
	return diags
}

func schemaAgentZscaler() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		zscalerAgentConfigKey: {
			Type:     schema.TypeSet,
			Optional: true,
			MinItems: 1,
			MaxItems: 1,
			Description: "ZDX application metrics through OneAPI. Set release_channel to beta. " +
				"Provide OneAPI credentials to the Agent using ZSCALER_CLIENT_ID and ZSCALER_CLIENT_SECRET environment variables.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"vanity_domain": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "OneAPI tenant name before .zslogin.net, without a URL scheme or domain suffix.",
				},
			}},
		},
	}
}

func marshalAgentZscaler(d resourceInterface, diags diag.Diagnostics) *v1alphaAgent.ZscalerConfig {
	data := getAgentResourceData(d, zscalerAgentType, zscalerAgentConfigKey, diags)
	if data == nil {
		return nil
	}
	return &v1alphaAgent.ZscalerConfig{VanityDomain: data["vanity_domain"].(string)}
}
