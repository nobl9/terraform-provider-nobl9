package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const cloudWatchFailoverConfigKey = "failover"

func getCloudWatchFailoverSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Description: "Region failover of the data source. Nobl9 reads the queries from the failover region " +
			"only after a failover event is sent to Nobl9, and goes back to the primary region after a recovery event.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"primary_region": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      "AWS region that Nobl9 reads from outside a failover.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				},
				"failover_region": {
					Type:     schema.TypeString,
					Required: true,
					Description: "AWS region that Nobl9 reads from during a failover. " +
						"It must differ from `primary_region`.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
				},
			},
		},
	}
}

func marshalCloudWatchFailover(raw interface{}) *v1alpha.CloudWatchFailoverConfig {
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 || list[0] == nil {
		return nil
	}
	data := list[0].(map[string]interface{})
	return &v1alpha.CloudWatchFailoverConfig{
		PrimaryRegion:  data["primary_region"].(string),
		FailoverRegion: data["failover_region"].(string),
	}
}

func unmarshalCloudWatchFailover(config *v1alpha.CloudWatchFailoverConfig) []interface{} {
	if config == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"primary_region":  config.PrimaryRegion,
		"failover_region": config.FailoverRegion,
	}}
}
