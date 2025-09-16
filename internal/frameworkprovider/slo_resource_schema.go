// nolint: lll
package frameworkprovider

import (
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

var sloResourceSchema = func() schema.Schema {
	description := "[SLO configuration | Nobl9 documentation](https://docs.nobl9.com/yaml-guide#slo)"
	return schema.Schema{
		MarkdownDescription: description,
		Description:         description,
		Attributes: map[string]schema.Attribute{
			"name":         metadataNameAttr(),
			"display_name": metadataDisplayNameAttr(),
			"project": func() schema.Attribute {
				attr := metadataProjectAttr()
				attr.PlanModifiers = []planmodifier.String{
					sloProjectPlanModifier{},
				}
				return attr
			}(),
			"description": specDescriptionAttr(),
			"annotations": metadataAnnotationsAttr(),
			"service": schema.StringAttribute{
				Required:    true,
				Description: "Name of the service.",
			},
			"budgeting_method": schema.StringAttribute{
				Required:    true,
				Description: "Method which will be use to calculate budget.",
			},
			"tier": schema.StringAttribute{
				Optional:    true,
				Description: "Internal field, do not use.",
			},
			"alert_policies": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Alert Policies attached to SLO.",
			},
			"retrieve_historical_data_from": schema.StringAttribute{
				Optional: true,
				Description: "If set, the retrieval of historical data for a newly created SLO will be triggered, " +
					"starting from the specified date. Needs to be RFC3339 format.",
				Validators: []validator.String{newDateTimeValidator(time.RFC3339)},
				WriteOnly:  true,
			},
		},
		Blocks: map[string]schema.Block{
			"label":          metadataLabelsBlock(),
			"indicator":      sloResourceIndicatorBlock(),
			"objective":      sloResourceObjectiveBlock(),
			"composite":      sloResourceCompositeV1Block(),
			"time_window":    sloResourceTimeWindowBlock(),
			"attachment":     sloResourceAttachmentBlock(),
			"anomaly_config": anomalyConfigBlock(),
		},
	}
}()

func sloResourceIndicatorBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for the metric source (Agent/Direct).",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "Name of the metric source.",
				},
				"project": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the metric source project.",
				},
				"kind": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("Agent"),
					Description: "Kind of the metric source. One of {Agent, Direct}.",
				},
			},
		},
	}
}

func sloResourceObjectiveBlock() schema.ListNestedBlock {
	description := "[Objectives documentation](https://docs.nobl9.com/yaml-guide#objective)"
	return schema.ListNestedBlock{
		Description:         description,
		MarkdownDescription: description,
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"display_name": schema.StringAttribute{
					Optional:    true,
					Description: "Name to be displayed.",
				},
				"op": schema.StringAttribute{
					Optional:    true,
					Description: "For threshold metrics, the logical operator applied to the threshold.",
					Validators: []validator.String{
						stringvalidator.AlsoRequires(
							path.MatchRoot("objective").AtAnyListIndex().AtName("raw_metric"),
						),
					},
				},
				"target": schema.Float64Attribute{
					Required:    true,
					Description: "The numeric target for your objective.",
				},
				"time_slice_target": schema.Float64Attribute{
					Optional:    true,
					Description: "Designated value for slice.",
				},
				"value": schema.Float64Attribute{
					Optional: true,
					Description: "Required for threshold and ratio metrics. Optional for existing composite SLOs. For threshold" +
						" metrics, the threshold value. For ratio metrics, this must be a unique value per objective (for" +
						" legacy reasons). For new composite SLOs, it must be omitted. If, for composite SLO, it was set" +
						" previously to a non-zero value, then it must remain unchanged.",
					PlanModifiers: []planmodifier.Float64{
						sloObjectiveValuePlanModifier{},
					},
				},
				"name": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Objective's name. This field is computed if not provided.",
				},
				"primary": schema.BoolAttribute{
					Optional:    true,
					Description: "Is objective marked as primary.",
				},
			},
			Blocks: map[string]schema.Block{
				"count_metrics": schema.ListNestedBlock{
					Description: "Compares two time series, calculating the ratio of either good or bad values to the" +
						" total number of values. Fill either the 'good' or 'bad' series, but not both.",
					Validators: []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"incremental": schema.BoolAttribute{
								Required:    true,
								Description: "Should the metrics be incrementing or not.",
							},
						},
						Blocks: map[string]schema.Block{
							"good": schema.ListNestedBlock{
								Description: "Configuration for good time series metrics.",
								Validators:  []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{
									Blocks: sloResourceMetricSpecBlocks(),
								},
							},
							"bad": schema.ListNestedBlock{
								Description: "Configuration for bad time series metrics.",
								Validators:  []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{
									Blocks: sloResourceMetricSpecBlocks(),
								},
							},
							"total": schema.ListNestedBlock{
								Description: "Configuration for metric source.",
								Validators:  []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{
									Blocks: sloResourceMetricSpecBlocks(),
								},
							},
							"good_total": schema.ListNestedBlock{
								Description: "Configuration for single query series metrics.",
								Validators:  []validator.List{listvalidator.SizeAtMost(1)},
								NestedObject: schema.NestedBlockObject{
									Blocks: sloResourceMetricSpecBlocks(),
								},
							},
						},
					},
				},
				"raw_metric": schema.ListNestedBlock{
					Description: "Raw data is used to compare objective values.",
					Validators:  []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"query": schema.ListNestedBlock{
								Description: "Configuration for metric source.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeBetween(1, 1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: sloResourceMetricSpecBlocks(),
								},
							},
						},
					},
				},
				"composite": sloResourceCompositeV2ObjectiveBlock(),
			},
		},
	}
}

func sloResourceTimeWindowBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Time window configuration for the SLO.",
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeBetween(1, 1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"count": schema.Int64Attribute{
					Required:    true,
					Description: "Count of the time unit.",
				},
				"is_rolling": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Is the window moving or not.",
				},
				"unit": schema.StringAttribute{
					Required:    true,
					Description: "Unit of time.",
				},
			},
			Blocks: map[string]schema.Block{
				"calendar": schema.ListNestedBlock{
					Description: "Calendar configuration for the time window.",
					Validators:  []validator.List{listvalidator.SizeAtMost(1)},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"start_time": schema.StringAttribute{
								Required:    true,
								Description: "Date of the start.",
							},
							"time_zone": schema.StringAttribute{
								Required:    true,
								Description: "Timezone name in IANA Time Zone Database.",
							},
						},
					},
				},
			},
		},
	}
}

func sloResourceAttachmentBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "URL attachments for the SLO.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"display_name": schema.StringAttribute{
					Optional:    true,
					Description: "Name displayed for the attachment. Max. length: 63 characters.",
					Validators:  []validator.String{stringvalidator.LengthAtMost(63)},
				},
				"url": schema.StringAttribute{
					Required:    true,
					Description: "URL to the attachment.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(20),
		},
	}
}

func anomalyConfigBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for anomaly detection.",
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"no_data": schema.ListNestedBlock{
					Description: "No data alerts configuration.",
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeBetween(1, 1),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"alert_method": schema.ListNestedBlock{
								Description: "Alert methods attached to Anomaly Config.",
								Validators:  []validator.List{listvalidator.SizeBetween(1, 5)},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required:    true,
											Description: "The name of the previously defined alert method.",
											Validators:  []validator.String{stringvalidator.LengthAtMost(63)},
										},
										"project": schema.StringAttribute{
											Required: true,
											Description: "Project name the Alert Method is in, " +
												" must conform to the naming convention from [DNS RFC1123]" +
												"(https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names)." +
												" If not defined, Nobl9 returns a default value for this field.",
										},
									},
								},
							},
						},
						Attributes: map[string]schema.Attribute{
							"alert_after": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Default:  stringdefault.StaticString("15m"),
								Description: "Specifies the duration to wait after receiving no data before triggering an alert. " +
									"The value must be a valid Go duration string, such as \"1h\" for one hour. " +
									"If not specified, the system defaults to \"15m\" (15 minutes).",
							},
						},
					},
				},
			},
		},
	}
}

func sloResourceMetricSpecBlocks() map[string]schema.Block {
	return map[string]schema.Block{
		"amazon_prometheus": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/Amazon_Prometheus/#creating-slos-with-ams-prometheus)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"promql": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"appdynamics": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/appdynamics#creating-slos-with-appdynamics)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"application_name": schema.StringAttribute{
						Required:    true,
						Description: "Name of the added application",
					},
					"metric_path": schema.StringAttribute{
						Required:    true,
						Description: "Path to the metrics",
					},
				},
			},
		},
		"azure_monitor": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/azure-monitor#creating-slos-with-azure-monitor)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"data_type": schema.StringAttribute{
						Required:    true,
						Description: "Specifies source: 'metrics' or 'logs'",
						Validators: []validator.String{stringvalidator.OneOf(
							v1alphaSLO.AzureMonitorDataTypeMetrics,
							v1alphaSLO.AzureMonitorDataTypeLogs,
						)},
					},
					"resource_id": schema.StringAttribute{
						Optional:    true,
						Description: "Identifier of the Azure Cloud resource [Required for metrics]",
					},
					"metric_namespace": schema.StringAttribute{
						Optional:    true,
						Description: "Namespace of the metric [Optional for metrics]",
					},
					"metric_name": schema.StringAttribute{
						Optional:    true,
						Description: "Name of the metric [Required for metrics]",
					},
					"aggregation": schema.StringAttribute{
						Optional:    true,
						Description: "Aggregation type [Required for metrics]",
					},
					"kql_query": schema.StringAttribute{
						Optional:    true,
						Description: "Logs query in Kusto Query Language [Required for logs]",
					},
				},
				Blocks: map[string]schema.Block{
					"dimensions": schema.SetNestedBlock{
						Description: "Dimensions of the metric [Optional for metrics]",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Name",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "Value",
								},
							},
						},
					},
					"workspace": schema.ListNestedBlock{
						Description: "Log analytics workspace [Required for logs]",
						Validators:  []validator.List{listvalidator.SizeAtMost(1)},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"subscription_id": schema.StringAttribute{
									Required:    true,
									Description: "Subscription ID of the workspace",
								},
								"resource_group": schema.StringAttribute{
									Required:    true,
									Description: "Resource group of the workspace",
								},
								"workspace_id": schema.StringAttribute{
									Required:    true,
									Description: "ID of the workspace",
								},
							},
						},
					},
				},
			},
		},
		"bigquery": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/bigquery#creating-slos-with-bigquery)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"location": schema.StringAttribute{
						Required:    true,
						Description: "Location of you BigQuery",
					},
					"project_id": schema.StringAttribute{
						Required:    true,
						Description: "Project ID",
					},
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"cloudwatch": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/Amazon_CloudWatch/#creating-slos-with-cloudwatch)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Optional:    true,
						Description: "AccountID used with cross-account observability feature",
						Validators: []validator.String{stringvalidator.RegexMatches(
							regexp.MustCompile(`^\d{12}$`),
							"account_id must be 12-digit identifier",
						)},
					},
					"region": schema.StringAttribute{
						Required:    true,
						Description: "Region of the CloudWatch instance",
					},
					"namespace": schema.StringAttribute{
						Optional:    true,
						Description: "Namespace of the metric",
					},
					"metric_name": schema.StringAttribute{
						Optional:    true,
						Description: "Metric name",
					},
					"stat": schema.StringAttribute{
						Optional:    true,
						Description: "Metric data aggregations",
					},
					"sql": schema.StringAttribute{
						Optional:    true,
						Description: "SQL query",
					},
					"json": schema.StringAttribute{
						Optional:    true,
						Description: "JSON query",
					},
				},
				Blocks: map[string]schema.Block{
					"dimensions": schema.SetNestedBlock{
						Description: "Set of name/value pairs that is part of the identity of a metric",
						Validators:  []validator.Set{setvalidator.SizeBetween(1, 10)},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "Name",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "Value",
								},
							},
						},
					},
				},
			},
		},
		"datadog": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/datadog#creating-slos-with-datadog)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"dynatrace": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/dynatrace#creating-slos-with-dynatrace)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"metric_selector": schema.StringAttribute{
						Required:    true,
						Description: "Selector for the metrics",
					},
				},
			},
		},
		"elasticsearch": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/elasticsearch#creating-slos-with-elasticsearch)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"index": schema.StringAttribute{
						Required:    true,
						Description: "Index of metrics we want to query",
					},
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"gcm": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/sources/google-cloud-monitoring/#creating-slos-with-google-cloud-monitoring)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"project_id": schema.StringAttribute{
						Required:    true,
						Description: "Project ID",
					},
					"query": schema.StringAttribute{
						Optional:    true,
						Description: "Query for the metrics in MQL format [deprecated](https://cloud.google.com/stackdriver/docs/deprecations/mql)",
					},
					"promql": schema.StringAttribute{
						Optional:    true,
						Description: "Query for the metrics in PromQL format",
					},
				},
			},
		},
		"grafana_loki": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/grafana-loki#creating-slos-with-grafana-loki)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"logql": schema.StringAttribute{
						Required:    true,
						Description: "Query for the logs",
					},
				},
			},
		},
		"graphite": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/graphite#creating-slos-with-graphite)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"metric_path": schema.StringAttribute{
						Required:    true,
						Description: "Path to the metrics",
					},
				},
			},
		},
		"honeycomb": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/honeycomb#creating-slos-with-honeycomb)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"attribute": schema.StringAttribute{
						Optional:    true,
						Description: "Column name - required for all calculation types besides 'CONCURRENCY' and 'COUNT'",
					},
				},
			},
		},
		"influxdb": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/influxdb#creating-slos-with-influxdb)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"instana": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/instana#creating-slos-with-instana)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"metric_type": schema.StringAttribute{
						Required:    true,
						Description: "Instana metric type 'application' or 'infrastructure'",
						Validators:  []validator.String{stringvalidator.OneOf("application", "infrastructure")},
					},
				},
				Blocks: map[string]schema.Block{
					"infrastructure": schema.ListNestedBlock{
						Description: "Infrastructure metric type",
						Validators:  []validator.List{listvalidator.SizeAtMost(1)},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"metric_retrieval_method": schema.StringAttribute{
									Required:    true,
									Description: "Metric retrieval method 'query' or 'snapshot'",
								},
								"query": schema.StringAttribute{
									Optional:    true,
									Description: "Query for the metrics",
								},
								"snapshot_id": schema.StringAttribute{
									Optional:    true,
									Description: "Snapshot ID",
								},
								"metric_id": schema.StringAttribute{
									Required:    true,
									Description: "Metric ID",
								},
								"plugin_id": schema.StringAttribute{
									Required:    true,
									Description: "Plugin ID",
								},
							},
						},
					},
					"application": schema.ListNestedBlock{
						Description: "Application metric type",
						Validators:  []validator.List{listvalidator.SizeAtMost(1)},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"metric_id": schema.StringAttribute{
									Required:    true,
									Description: "Metric ID one of 'calls', 'erroneousCalls', 'errors', 'latency'",
								},
								"aggregation": schema.StringAttribute{
									Required:    true,
									Description: "Depends on the value specified for 'metric_id'- more info in N9 docs",
								},
								"api_query": schema.StringAttribute{
									Required:    true,
									Description: "API query user passes in a JSON format",
								},
								"include_internal": schema.BoolAttribute{
									Optional:    true,
									Description: "Include internal",
								},
								"include_synthetic": schema.BoolAttribute{
									Optional:    true,
									Description: "Include synthetic",
								},
							},
							Blocks: map[string]schema.Block{
								"group_by": schema.ListNestedBlock{
									Description: "Group by method",
									Validators:  []validator.List{listvalidator.SizeAtMost(1)},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"tag": schema.StringAttribute{
												Required:    true,
												Description: "Group by tag",
											},
											"tag_entity": schema.StringAttribute{
												Required:    true,
												Description: "Tag entity - one of 'DESTINATION', 'SOURCE', 'NOT_APPLICABLE'",
											},
											"tag_second_level_key": schema.StringAttribute{
												Optional:    true,
												Description: "Second level key for the tag",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"lightstep": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/lightstep#creating-slos-with-lightstep)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"percentile": schema.Float64Attribute{
						Optional:    true,
						Description: "Optional value to filter by percentiles",
					},
					"stream_id": schema.StringAttribute{
						Optional:    true,
						Description: "ID of the metrics stream",
					},
					"type_of_data": schema.StringAttribute{
						Required:    true,
						Description: "Type of data to filter by",
					},
					"uql": schema.StringAttribute{
						Optional:    true,
						Description: "UQL query",
					},
				},
			},
		},
		"logic_monitor": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/logic-monitor#creating-slos-with-logic-monitor)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"query_type": schema.StringAttribute{
						Required:    true,
						Description: "Query type: device_metrics or website_metrics",
						Validators: []validator.String{stringvalidator.OneOf(
							v1alphaSLO.LMQueryTypeDeviceMetrics,
							v1alphaSLO.LMQueryTypeWebsiteMetrics,
						)},
					},
					"device_data_source_instance_id": schema.Int64Attribute{
						Optional:    true,
						Description: "Device Datasource Instance ID. Used by Query type = device_metrics",
					},
					"graph_id": schema.Int64Attribute{
						Optional:    true,
						Description: "Graph ID. Used by Query type = device_metrics",
					},
					"website_id": schema.StringAttribute{
						Optional:    true,
						Description: "Website ID. Used by Query type = website_metrics",
					},
					"checkpoint_id": schema.StringAttribute{
						Optional:    true,
						Description: "Checkpoint ID. Used by Query type = website_metrics",
					},
					"graph_name": schema.StringAttribute{
						Optional:    true,
						Description: "Graph Name. Used by Query type = website_metrics",
					},
					"line": schema.StringAttribute{
						Required:    true,
						Description: "Line",
					},
				},
			},
		},
		"newrelic": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/new-relic#creating-slos-with-new-relic)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"nrql": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"opentsdb": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/opentsdb#creating-slos-with-opentsdb)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"pingdom": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/pingdom#creating-slos-with-pingdom)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"check_id": schema.StringAttribute{
						Required:    true,
						Description: "Pingdom uptime or transaction check's ID",
					},
					"check_type": schema.StringAttribute{
						Optional:    true,
						Description: "Pingdom check type - uptime or transaction",
					},
					"status": schema.StringAttribute{
						Optional:    true,
						Description: "Optional for the Uptime checks. Use it to filter the Pingdom check results by status",
					},
				},
			},
		},
		"prometheus": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/prometheus#creating-slos-with-prometheus)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"promql": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"redshift": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/Amazon_Redshift/#creating-slos-with-amazon-redshift)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						Required:    true,
						Description: "Region of the Redshift instance",
					},
					"cluster_id": schema.StringAttribute{
						Required:    true,
						Description: "Redshift custer ID",
					},
					"database_name": schema.StringAttribute{
						Required:    true,
						Description: "Database name",
					},
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"splunk": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"splunk_observability": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk-observability)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"program": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"sumologic": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/sumo-logic#creating-slos-with-sumo-logic)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:    true,
						Description: "Sumologic source - metrics or logs",
					},
					"query": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
					"rollup": schema.StringAttribute{
						Optional:    true,
						Description: "Aggregation function - avg, sum, min, max, count, none",
					},
					"quantization": schema.StringAttribute{
						Optional:    true,
						Description: "Period of data aggregation",
					},
				},
			},
		},
		"thousandeyes": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/Sources/thousandeyes#creating-slos-with-thousandeyes)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"test_id": schema.Int64Attribute{
						Required:    true,
						Description: "ID of the test",
					},
					"test_type": schema.StringAttribute{
						Optional:    true,
						Description: "Type of the test",
					},
				},
			},
		},
		"azure_prometheus": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/sources/create-slo/azure-prometheus)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"promql": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
		"coralogix": schema.ListNestedBlock{
			Description: "[Configuration documentation](https://docs.nobl9.com/sources/create-slo/coralogix)",
			Validators:  []validator.List{listvalidator.SizeAtMost(1)},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"promql": schema.StringAttribute{
						Required:    true,
						Description: "Query for the metrics",
					},
				},
			},
		},
	}
}

func sloResourceCompositeV2ObjectiveBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "An assembly of objectives from different SLOs reflecting their combined performance.",
		Validators:  []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"max_delay": schema.StringAttribute{
					Required:    true,
					Description: "Maximum time for your composite SLO to wait for data from objectives.",
				},
			},
			Blocks: map[string]schema.Block{
				"components": schema.ListNestedBlock{
					Description: "Objectives to be assembled in your composite SLO.",
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"objectives": schema.ListNestedBlock{
								Description: "An additional nesting for the components of your composite SLO.",
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"composite_objective": schema.ListNestedBlock{
											Description: "Your composite SLO component.",
											NestedObject: schema.NestedBlockObject{
												Attributes: map[string]schema.Attribute{
													"project": schema.StringAttribute{
														Required:    true,
														Description: "Project name.",
													},
													"slo": schema.StringAttribute{
														Required:    true,
														Description: "SLO name.",
													},
													"objective": schema.StringAttribute{
														Required:    true,
														Description: "SLO objective name.",
													},
													"weight": schema.Float64Attribute{
														Required:    true,
														Description: "Weights determine each component's contribution to the composite SLO.",
													},
													"when_delayed": schema.StringAttribute{
														Required:    true,
														Description: "Defines how to treat missing component data on `max_delay` expiry.",
														Validators: []validator.String{
															stringvalidator.OneOf(v1alphaSLO.WhenDelayedNames()...),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func sloResourceCompositeV1Block() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description:        "(\"composite\" is deprecated, use [composites 2.0 schema](https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo#nested-schema-for-objectivecomposite) instead) [Composite SLO documentation](https://docs.nobl9.com/yaml-guide/#slo)",
		DeprecationMessage: "\"composite\" is deprecated, use \"objective.composite\" instead.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"target": schema.Float64Attribute{
					Required:    true,
					Description: "The numeric target for your objective.",
				},
			},
			Blocks: map[string]schema.Block{
				"burn_rate_condition": schema.SetNestedBlock{
					Description:        "(\"burn_rate_condition\" is part of deprecated composites 1.0, use [composites 2.0](https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo#nested-schema-for-objectivecomposite) instead) Condition when the Composite SLO's error budget is burning.",
					DeprecationMessage: "\"burn_rate_condition\" is part of deprecated composites 1.0, use composites 2.0 (https://registry.terraform.io/providers/nobl9/nobl9/latest/docs/resources/slo#nested-schema-for-objectivecomposite) instead",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"op": schema.StringAttribute{
								Required:    true,
								Description: "Type of logical operation.",
							},
							"value": schema.Float64Attribute{
								Required:    true,
								Description: "Burn rate value.",
							},
						},
					},
				},
			},
		},
	}
}
