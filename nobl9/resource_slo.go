package nobl9

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceSlo() *schema.Resource {
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
						"attachments": {
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
									"metricSpec": schemaMetricSpec(),
									},
								},
							},
						},
						"objectives": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: " ([Objectives documentation] https://nobl9.github.io/techdocs_YAML_Guide/#objectives)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"countMetrics": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Alert Policies attached to SLO",
										Elem: &schema.Schema{
											"good": schemaMetricSpec(),
											"incemental": {
												Type:        schema.TypeBool,
							        			Required:    true,
												Description: "Should the metrics be incrementing or not",
											},
											"metricSpec": schemaMetricSpec(),
										},
									},
									"displayname": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name to be displayed",
									},
									"op": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Type of logical operation",
									},
									"target": {
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Desiganted value",
									},
									"timeSliceTarget": {
										Type:        schema.TypeFloat,
										Optional:    true,
										Description: "Designated value for slice",
									},
									"value": {
										Type:        schema.TypeFloat,
										Optional:    true,
										Description: "Value",
									},
								},
							},
						},
						"service": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the service",
						},
						"timeWindows": {
							Type:        schema.TypeString,
							Required:    true,
							Description: " ",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"calendar": {
										Type:        schema.TypeSet,
										Optional:    true,
										Description: "Alert Policies attached to SLO",
										Elem: &schema.Schema{
											"startTime": {
												Type:        schema.TypeString,
							        			Required:    true,
												Description: "Date of the start",
											},
											"timeZone": {
												Type:        schema.TypeString,
							        			Required:    true,
												Description: "Timezone name in IANA Time Zone Database",
											},
										},
									},
									"count": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Count of the time unit",
									},
									"isRolling": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Is the window moving or not",
									},
									"period": {
										Type:        schema.TypeFloat,
										Optional:    true,
										Description: "Specific time frame",
										Elem: &schema.Schema{
											"begin": {
												Type:        schema.TypeString,
							        			Optional:    true,
												Description: "Beginning of the period",
											},
											"end": {
												Type:        schema.TypeString,
							        			Optional:    true,
												Description: "End of the period",
											},
										},
									},
									"unit": {
										Type:        schema.TypeFloat,
										Required:    true,
										Description: "Unit of time",
									},
								},
							},
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

func marshalSlo() {}

func unmarshalSlo() {}

func resourceSloApply() {}

func resourceSloRead() {}

func resourceSloDelete() {}