package nobl9

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),

			"slo_spec": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "[SLO documentation](https://nobl9.github.io/techdocs_YAML_Guide/#slo)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alertPolicies": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Alert Policies attached to SLO",
							Elem: &schema.Schema{
								Type:        schema.TypeString,
								Description: "Alert Policy",
							},
						},
						"attachments" {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"displayname": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name which is dispalyed for the attachment",
									},
									"url": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Url to the attachment",
									},
								},
							},

						},
						"budgetingMethod": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Method which will be use to calculate budget",
						},
						"createdAt": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Time of creation",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the SLO",
						},
						"indicator": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: " ",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metricSourceSpec": {
										Type:        schema.TypeSet,
										Required:    true,
										Description: "",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Name of the metric source",
												},
												"project": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Name of the metric souce project",
												},
											},
										},
									},
									"metricSpec": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Configuration for metric source",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"appDynamics": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Name of the metric source",
												},
												"project": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Name of the metric souce project",
												},
											},
										},
									},
								},
							},
						},
						"objectives": {
							Type:        schema.TypeString,
							Required:    true,
							Description: " ",

						},
						"service": {
							Type:        schema.TypeString,
							Required:    true,
							Description: " ",
						},
						"timeWindows": {
							Type:        schema.TypeString,
							Required:    true,
							Description: " ",

						},
					},
				},
			},
		},
		CreateContext: resourceSloApply,
		UpdateContext: resourceSloApply,
		DeleteContext: resourceSloDelete,
		ReadContext:   resourceSloRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[SLO configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#Slo)",
	}
}