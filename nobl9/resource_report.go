package nobl9

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

type reportResource struct {
	reportProvider
}

type reportProvider interface {
	GetSchema() map[string]*schema.Schema
	GetDescription() string
	MarshalSpec(spec v1alphaReport.Spec, resource resourceInterface) v1alphaReport.Spec
	UnmarshalSpec(d *schema.ResourceData, spec v1alphaReport.Spec) diag.Diagnostics
}

func resourceReportFactory(provider reportProvider) *schema.Resource {
	i := reportResource{reportProvider: provider}
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"shared":       schemaShared(),
			"filters":      schemaFilters(),
		},
		CustomizeDiff: i.resourceReportValidate,
		CreateContext: i.resourceReportApply,
		UpdateContext: i.resourceReportApply,
		DeleteContext: resourceReportDelete,
		ReadContext:   i.resourceReportRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: provider.GetDescription(),
	}

	for k, v := range provider.GetSchema() {
		resource.Schema[k] = v
	}

	return resource
}

func schemaShared() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Is report shared for all users with access to included projects.",
	}
}

func schemaFilters() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MinItems:    1,
		MaxItems:    1,
		Description: "Filters are used to select scope for Report.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"projects": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Projects to pull data for report from.",
					Elem: &schema.Schema{
						Type:        schema.TypeString,
						Description: "Project name, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
					},
				},
				"service": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Services to pull data for report from.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
							},
							"project": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
							},
						},
					},
				},
				"slo": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "SLOs to pull data for report from.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
							},
							"project": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the Nobl9 project the resource sits in, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
							},
						},
					},
				},
				"label": schemaReportLabels(),
			},
		},
	}
}

func (r reportResource) marshalReport(ri resourceInterface) *v1alphaReport.Report {
	spec := v1alphaReport.Spec{
		Shared:  ri.Get("shared").(bool),
		Filters: marshalReportFilters(ri.Get("filters")),
	}
	report := v1alphaReport.New(
		v1alphaReport.Metadata{
			Name:        ri.Get("name").(string),
			DisplayName: ri.Get("display_name").(string),
		},
		r.MarshalSpec(spec, ri),
	)
	return &report
}

func (r reportResource) unmarshalReport(d *schema.ResourceData, report v1alphaReport.Report) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = appendError(diags, d.Set("name", report.Metadata.Name))
	diags = appendError(diags, d.Set("display_name", report.Metadata.DisplayName))
	diags = appendError(diags, d.Set("shared", report.Spec.Shared))
	diags = appendError(diags, unmarshalReportFilters(d, report.Spec.Filters))

	errs := r.UnmarshalSpec(d, report.Spec)
	diags = append(diags, errs...)
	return diags
}

func marshalReportFilters(filtersRaw interface{}) *v1alphaReport.Filters {
	if len(filtersRaw.([]interface{})) == 0 {
		return nil
	}
	filters := filtersRaw.([]interface{})[0].(map[string]interface{})

	projectList := filters["projects"].([]interface{})
	projects := make([]string, 0, len(projectList))
	for _, filter := range projectList {
		projects = append(projects, filter.(string))
	}

	serviceList := filters["service"].([]interface{})
	services := make([]v1alphaReport.Service, 0, len(serviceList))
	for _, filter := range serviceList {
		f := filter.(map[string]interface{})
		service := v1alphaReport.Service{
			Name:    f["name"].(string),
			Project: f["project"].(string),
		}
		services = append(services, service)
	}

	sloList := filters["slo"].([]interface{})
	slos := make([]v1alphaReport.SLO, 0, len(sloList))
	for _, filter := range sloList {
		f := filter.(map[string]interface{})
		slo := v1alphaReport.SLO{
			Name:    f["name"].(string),
			Project: f["project"].(string),
		}
		slos = append(slos, slo)
	}

	return &v1alphaReport.Filters{
		Projects: projects,
		Services: services,
		SLOs:     slos,
		Labels:   marshalReportLabels(filters["label"].([]interface{})),
	}
}

func unmarshalReportFilters(d *schema.ResourceData, filters *v1alphaReport.Filters) error {
	services := make([]map[string]interface{}, 0, len(filters.Services))
	for _, service := range filters.Services {
		serviceMap := map[string]interface{}{
			"name":    service.Name,
			"project": service.Project,
		}
		services = append(services, serviceMap)
	}

	slos := make([]map[string]interface{}, 0, len(filters.SLOs))
	for _, slo := range filters.SLOs {
		sloMap := map[string]interface{}{
			"name":    slo.Name,
			"project": slo.Project,
		}
		slos = append(slos, sloMap)
	}

	f := map[string]interface{}{
		"projects": filters.Projects,
		"service":  services,
		"slo":      slos,
	}

	if len(filters.Labels) > 0 {
		f["label"] = unmarshalReportLabels(filters.Labels)
	}

	return d.Set("filters", []interface{}{f})
}

func marshalReportLabels(labelList []interface{}) v1alphaReport.Labels {
	labels, _ := marshalLabels(labelList)
	reportLabels := make(map[v1alphaReport.LabelKey][]v1alphaReport.LabelValue, len(labels))
	for key, values := range labels {
		reportLabels[key] = append(reportLabels[key], values...)
	}
	return reportLabels
}

func unmarshalReportLabels(labelsRaw v1alphaReport.Labels) interface{} {
	resultLabels := make([]map[string]interface{}, 0)

	for labelKey, labelValuesRaw := range labelsRaw {
		var labelValuesStr []string
		labelValuesStr = append(labelValuesStr, labelValuesRaw...)
		labelKeyWithValues := make(map[string]interface{})
		labelKeyWithValues["key"] = labelKey
		labelKeyWithValues["values"] = labelValuesStr

		resultLabels = append(resultLabels, labelKeyWithValues)
	}

	return resultLabels
}

func (r reportResource) resourceReportValidate(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	report := r.marshalReport(d)
	errs := manifest.Validate([]manifest.Object{report})
	if errs != nil {
		return formatErrorsAsSingleError(errs)
	}
	return nil
}

func (r reportResource) resourceReportApply(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	report := r.marshalReport(d)
	err := client.Objects().V1().Apply(ctx, []manifest.Object{report})
	if err != nil {
		return diag.Errorf("could not add report: %s", err.Error())
	}
	d.SetId(report.Metadata.Name)
	return r.resourceReportRead(ctx, d, meta)
}

func (r reportResource) resourceReportRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	reports, err := client.Objects().V1().GetReports(ctx, v1Objects.GetReportsRequest{
		Names: []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return handleResourceReadResult(d, reports, r.unmarshalReport)
}

func resourceReportDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	err := client.Objects().V1().DeleteByName(ctx, manifest.KindReport, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

type reportSystemHealthReview struct{}

func (r reportSystemHealthReview) GetDescription() string {
	return "[System Health Review Report | Nobl9 Documentation](https://docs.nobl9.com/reports/system-health-review/)"
}

func (r reportSystemHealthReview) GetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"time_frame": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"time_zone": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Timezone name in IANA Time Zone Database.",
					},
					"snapshot": {
						Type:     schema.TypeSet,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"point": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The method of reporting time frame [past/latest]",
									ValidateDiagFunc: validation.ToDiagFunc(
										validation.StringInSlice([]string{
											v1alphaReport.SnapshotPointPast.String(),
											v1alphaReport.SnapshotPointLatest.String(),
										}, false),
									),
								},
								"date_time": {
									Type:             schema.TypeString,
									Optional:         true,
									ValidateDiagFunc: validateDateTime,
									Description:      "Date and time of the past snapshot in RFC3339 format.",
								},
								"rrule": {
									Type:             schema.TypeString,
									Optional:         true,
									ValidateDiagFunc: validateRrule,
									Description: "The recurrence rule for the report past snapshot. " +
										"The expected value is a string in RRULE format. " +
										"Example: `FREQ=MONTHLY;BYMONTHDAY=1`",
								},
							},
						},
					},
				},
			},
		},
		"row_group_by": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Grouping methods of report table rows [project/service]",
			ValidateDiagFunc: validation.ToDiagFunc(
				validation.StringInSlice([]string{
					v1alphaReport.RowGroupByProject.String(),
					v1alphaReport.RowGroupByService.String(),
				}, false),
			),
		},
		"column": {
			Type:        schema.TypeList,
			MinItems:    1,
			Required:    true,
			Description: "Columns to display in the report table.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"display_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Column display name.",
					},
					"label": schemaColumnLabels(),
				},
			},
		},
		"thresholds": {
			Type:     schema.TypeSet,
			MinItems: 1,
			MaxItems: 1,
			Required: true,
			Description: "Thresholds for Green, Yellow and Red statuses (e.g. healthy, at risk, exhausted budget). " +
				"Yellow is calculated as the difference between Red and Green thresholds. " +
				"If Red and Green are the same, Yellow is not used on the report.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"green_gt": {
						Type:        schema.TypeFloat,
						Required:    true,
						Description: "Min value for the Green status (e.g. healthy).",
					},
					"red_lte": {
						Type:        schema.TypeFloat,
						Required:    true,
						Description: "Max value for the Red status (e.g. exhausted budget).",
					},
					"show_no_data": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "ShowNoData customizes the report to either show or hide rows with no data.",
					},
				},
			},
		},
	}
}

func schemaReportLabels() *schema.Schema {
	s := schemaLabels()
	s.DiffSuppressFunc = nil
	return s
}

func schemaColumnLabels() *schema.Schema {
	s := schemaReportLabels()
	s.Optional = false
	s.Required = true
	s.MinItems = 1
	return s
}

func (r reportSystemHealthReview) MarshalSpec(spec v1alphaReport.Spec, ri resourceInterface) v1alphaReport.Spec {
	rowGroupBy, _ := v1alphaReport.ParseRowGroupBy(ri.Get("row_group_by").(string))
	spec.SystemHealthReview = &v1alphaReport.SystemHealthReviewConfig{
		TimeFrame:  marshalReportTimeFrame(ri.Get("time_frame").(*schema.Set)),
		RowGroupBy: rowGroupBy,
		Columns:    marshalReportColumns(ri.Get("column").([]interface{})),
		Thresholds: marshalThresholds(ri.Get("thresholds").(*schema.Set)),
	}
	return spec
}

func marshalReportTimeFrame(timeFrameSet *schema.Set) v1alphaReport.SystemHealthReviewTimeFrame {
	if timeFrameSet.Len() == 0 {
		return v1alphaReport.SystemHealthReviewTimeFrame{}
	}
	timeFrame := timeFrameSet.List()[0].(map[string]interface{})
	snapshotSet := timeFrame["snapshot"].(*schema.Set)
	snapshotConfig := snapshotSet.List()[0].(map[string]interface{})
	var dateTime *time.Time
	if snapshotConfig["date_time"] != nil {
		if dt, err := time.Parse(time.RFC3339, snapshotConfig["date_time"].(string)); err == nil {
			dateTime = &dt
		}
	}
	point, _ := v1alphaReport.ParseSnapshotPoint(snapshotConfig["point"].(string))

	snapshot := v1alphaReport.SnapshotTimeFrame{
		Point:    point,
		DateTime: dateTime,
	}
	if snapshotConfig["rrule"] != nil {
		snapshot.Rrule = snapshotConfig["rrule"].(string)
	}
	return v1alphaReport.SystemHealthReviewTimeFrame{
		TimeZone: timeFrame["time_zone"].(string),
		Snapshot: snapshot,
	}
}

func marshalThresholds(thresholdsSet *schema.Set) v1alphaReport.Thresholds {
	if thresholdsSet.Len() == 0 {
		return v1alphaReport.Thresholds{}
	}
	thresholds := thresholdsSet.List()[0].(map[string]interface{})
	redLte := thresholds["red_lte"].(float64)
	greenGte := thresholds["green_gt"].(float64)
	return v1alphaReport.Thresholds{
		RedLessThanOrEqual: &redLte,
		GreenGreaterThan:   &greenGte,
		ShowNoData:         thresholds["show_no_data"].(bool),
	}
}

func marshalReportColumns(columnsRaw []interface{}) []v1alphaReport.ColumnSpec {
	columns := make([]v1alphaReport.ColumnSpec, 0, len(columnsRaw))
	for _, column := range columnsRaw {
		c := column.(map[string]interface{})
		columns = append(columns, v1alphaReport.ColumnSpec{
			DisplayName: c["display_name"].(string),
			Labels:      marshalReportLabels(c["label"].([]interface{})),
		})
	}
	return columns
}

func (r reportSystemHealthReview) UnmarshalSpec(d *schema.ResourceData, spec v1alphaReport.Spec) diag.Diagnostics {
	config := spec.SystemHealthReview
	var diags diag.Diagnostics

	diags = appendError(diags, d.Set("row_group_by", config.RowGroupBy.String()))
	diags = appendError(diags, unmarshalReportTimeFrame(d, config.TimeFrame))
	diags = appendError(diags, unmarshalReportColumns(d, config.Columns))
	diags = appendError(diags, unmarshalReportThresholds(d, config.Thresholds))
	return diags
}

func unmarshalReportTimeFrame(d *schema.ResourceData, timeFrame v1alphaReport.SystemHealthReviewTimeFrame) error {
	snapshot := map[string]interface{}{
		"point": timeFrame.Snapshot.Point.String(),
		"rrule": timeFrame.Snapshot.Rrule,
	}
	if timeFrame.Snapshot.DateTime != nil {
		snapshot["date_time"] = timeFrame.Snapshot.DateTime.Format(time.RFC3339)
	}
	return d.Set("time_frame", schema.NewSet(oneElementSet, []interface{}{
		map[string]interface{}{
			"time_zone": timeFrame.TimeZone,
			"snapshot":  schema.NewSet(oneElementSet, []interface{}{snapshot}),
		},
	}))
}

func unmarshalReportColumns(d *schema.ResourceData, columns []v1alphaReport.ColumnSpec) error {
	columnMap := make([]map[string]interface{}, 0, len(columns))
	for _, column := range columns {
		columnMap = append(columnMap, map[string]interface{}{
			"display_name": column.DisplayName,
			"label":        unmarshalReportLabels(column.Labels),
		})
	}
	return d.Set("column", columnMap)
}

func unmarshalReportThresholds(d *schema.ResourceData, thresholds v1alphaReport.Thresholds) error {
	return d.Set("thresholds", schema.NewSet(oneElementSet, []interface{}{
		map[string]interface{}{
			"green_gt":     thresholds.GreenGreaterThan,
			"red_lte":      thresholds.RedLessThanOrEqual,
			"show_no_data": thresholds.ShowNoData,
		},
	}))
}
