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
				"appDynamics": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-appdynamics)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"applicationName": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the added application",
							},
							"metricPath": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Path to the metrics",
							},
						},
					},
				},
				"bigQuery": {
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
							"projectID": {
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
				"dynatrace": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-dynatrace)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"metricSelector": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Selector for the metrics",
							},
						},
					},
				},
				"elasticsearch": {
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
				"graphite": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-graphite)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"metricPath": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Path to the metrics",
							},
						},
					},
				},
				"lightstep": {
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
							"streamId": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "ID of the metrics stream",
							},
							"typeOfData": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Type of data to filter by",
							},
						},
					},
				},
				"newRelic": {
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
				"opentsdb": {
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
				"prometheus": {
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
				"splunk": {
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
					},
				},
				"splunkObservability": {
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
				"thousandEyes": {
					Type:        schema.TypeSet,
					Required:    true,
					Description: "[Configuration documentation] (https://nobl9.github.io/techdocs_YAML_Guide/#slo-using-thousandeyes)",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"testID": {
								Type:        schema.TypeInteger,
								Required:    true,
								Description: "ID of the test",
							},
						},
					},
				},
			},
		},
	}
}