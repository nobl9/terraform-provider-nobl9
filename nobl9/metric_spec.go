package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//nolint:lll
func schemaMetricSpec() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Configuration for metric source",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"appdynamics": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/appdynamics#creating-slos-with-appdynamics)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"application_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the added application",
							},
							"metric_path": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Path to the metrics",
							},
						},
					},
				},

				"bigquery": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/bigquery#creating-slos-with-bigquery)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"location": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Location of you BigQuery",
							},
							"project_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Project ID",
							},
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"datadog": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/datadog#creating-slos-with-datadog)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"dynatrace": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/dynatrace#creating-slos-with-dynatrace)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"metric_selector": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Selector for the metrics",
							},
						},
					},
				},

				"elasticsearch": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/elasticsearch#creating-slos-with-elasticsearch)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"index": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Index of metrics we want to query",
							},
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"graphite": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/graphite#creating-slos-with-graphite)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"metric_path": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Path to the metrics",
							},
						},
					},
				},

				"lightstep": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/lightstep#creating-slos-with-lightstep)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"percentile": {
								Type:        schema.TypeFloat,
								Optional:    true,
								Description: "Optional value to filter by percentiles",
							},
							"stream_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "ID of the metrics stream",
							},
							"type_of_data": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Type of data to filter by",
							},
						},
					},
				},

				"newrelic": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/new-relic#creating-slos-with-new-relic)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"nrql": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"opentsdb": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/opentsdb#creating-slos-with-opentsdb)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"prometheus": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/prometheus#creating-slos-with-prometheus)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"promql": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"splunk": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"splunk_observability": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/splunk#creating-slos-with-splunk-observability)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"program": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
						},
					},
				},

				"thousandeyes": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/thousandeyes#creating-slos-with-thousandeyes)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"test_id": {
								Type:        schema.TypeInt,
								Required:    true,
								Description: "ID of the test",
							},
						},
					},
				},

				"grafana_loki": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "[Configuration documentation](https://docs.nobl9.com/Sources/grafana-loki#creating-slos-with-grafana-loki)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"logql": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the logs",
							},
						},
					},
				},
				"cloudwatch": {
					Type:     schema.TypeSet,
					Optional: true,
					Description: "[Configuration documentation]" +
						"(https://docs.nobl9.com/Sources/Amazon_CloudWatch/#creating-slos-with-cloudwatch)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"region": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Region of the CloudWatch instance",
							},
							"namespace": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Namespace of the metric",
							},
							"metric_name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Metric name",
							},
							"stat": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Metric data aggregations",
							},
							"sql": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "SQL query",
							},
							"json": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "JSON query",
							},
							"dimensions": {
								Type:        schema.TypeSet,
								Optional:    true,
								Description: "Set of name/value pairs that is a part of the identity of a metric",
								MinItems:    1,
								MaxItems:    10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"name": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Name",
										},
										"value": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "Value",
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
