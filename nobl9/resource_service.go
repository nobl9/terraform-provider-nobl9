package nobl9

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),
			"label":        schemaLabels(),
			"annotations":  schemaAnnotations(),
			"status": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Status of created service.",
				Elem: &schema.Schema{
					Type: schema.TypeFloat,
				},
			},
		},
		CustomizeDiff: resourceServicePlanChecks,
		CreateContext: resourceServiceApply,
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceApply,
		DeleteContext: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Service configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#service)",
	}
}

type Data interface {
	Get(key string) any
}

func marshalService(d Data) (*v1alphaService.Service, diag.Diagnostics) {
	var displayName string
	if dn := d.Get("display_name"); dn != nil {
		displayName = dn.(string)
	}

	labelsMarshaled, diags := getMarshaledLabels(d)
	if diags.HasError() {
		return nil, diags
	}

	annotationsMarshaled := getMarshaledAnnotations(d)

	service := v1alphaService.New(
		v1alphaService.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: displayName,
			Project:     d.Get("project").(string),
			Labels:      labelsMarshaled,
			Annotations: annotationsMarshaled,
		},
		v1alphaService.Spec{
			Description: d.Get("description").(string),
		})
	return &service, diags
}

func unmarshalService(d *schema.ResourceData, objects []v1alphaService.Service) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	metadata := object.Metadata
	err := d.Set("name", metadata.Name)
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata.DisplayName)
	diags = appendError(diags, err)
	err = d.Set("project", metadata.Project)
	diags = appendError(diags, err)

	if labelsRaw := metadata.Labels; len(labelsRaw) > 0 {
		err = d.Set("label", unmarshalLabels(labelsRaw))
		diags = appendError(diags, err)
	}

	if len(metadata.Annotations) > 0 {
		err = d.Set("annotations", metadata.Annotations)
		diags = appendError(diags, err)
	}

	status := map[string]int{"sloCount": object.Status.SloCount}
	err = d.Set("status", status)
	diags = appendError(diags, err)

	spec := object.Spec
	err = d.Set("description", spec.Description)
	diags = appendError(diags, err)

	return diags
}

func resourceServicePlanChecks(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// TODO PC-13378: Would this work correctly?
	//	if diff.HasChange("name") {
	//		return fmt.Errorf("name cannot be changed")
	//	}
	//	if diff.HasChange("project") {
	//		return fmt.Errorf("project cannot be changed")
	//	}

	// Check if project exists.
	if diff.HasChange("project") {
		//if err := validateProjectExistence(diff); err != nil {
		//	return err
		//}
	}

	// Check if apply with dry run works.
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return fmt.Errorf("could not get client: %v", ds)
	}
	service, diags := marshalService(diff)
	if diags.HasError() {
		return fmt.Errorf("could not marshal service: %v", diags)
	}
	resultService := manifest.SetDefaultProject([]manifest.Object{service}, config.Project)
	dryRun := true
	err := client.Objects().V1().Apply(ctx, resultService, dryRun)
	if err != nil {
		return diag.Errorf("could not add service: %s", err.Error())
	}
	return nil
}

func resourceServiceApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	service, diags := marshalService(d)
	if diags.HasError() {
		return diags
	}
	resultService := manifest.SetDefaultProject([]manifest.Object{service}, config.Project)
	err := client.Objects().V1().Apply(ctx, resultService)
	if err != nil {
		return diag.Errorf("could not add service: %s", err.Error())
	}
	d.SetId(service.Metadata.Name)
	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	services, err := client.Objects().V1().GetV1alphaServices(ctx, v1Objects.GetServicesRequest{
		Project: project,
		Names:   []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalService(d, services)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	if project == "" {
		project = config.Project
	}
	err := client.Objects().V1().DeleteByName(ctx, manifest.KindService, project, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
