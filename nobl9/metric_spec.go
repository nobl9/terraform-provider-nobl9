package nobl9

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func schemaMetricSpec() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Configuration for metric source",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"appdynamics_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-appdynamics)",
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
				"bigquery_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-bigquery)",
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
				"datadog_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-datadog)",
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
				"dynatrace_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-dynatrace)",
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
				"elasticsearch_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-elasticsearch)",
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
				"graphite_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-graphite)",
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
				"lightstep_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-lightstep)",
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
				"newrelic_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-newrelic)",
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
				"opentsdb_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-opentsdb)",
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
				"prometheus_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-prometheus)",
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
				"splunk_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-splunk)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query for the metrics",
							},
							"field_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name for metrics",
							},
					},
				},
				"splunk_observability_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-splunk-observability)",
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
				"thousandeyes_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-thousandeyes)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"test_id": {
								Type:        schema.TypeInteger,
								Required:    true,
								Description: "ID of the test",
							},
						},
					},
				},
				"grafana_loki_metric": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-loki",
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
			},
		},
	}
}