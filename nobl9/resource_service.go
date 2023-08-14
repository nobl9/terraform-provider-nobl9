package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),
			"label":        schemaLabels(),
			"status": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Status of created service.",
				Elem: &schema.Schema{
					Type: schema.TypeFloat,
				},
			},
		},
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

func marshalService(d *schema.ResourceData) (*v1alpha.Service, diag.Diagnostics) {
	metadata, diags := marshalMetadata(d)
	if diags.HasError() {
		return nil, diags
	}
	return &v1alpha.Service{
		// FIXME: delete ObjectInternal field after SDK update - for now it's hardcoded organization.
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindService,
		Metadata:   metadata,
		Spec: v1alpha.ServiceSpec{
			Description: d.Get("description").(string),
		},
	}, diags
}

func unmarshalService(d *schema.ResourceData, objects []v1alpha.Service) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	err := d.Set("name", object.Metadata.Name)
	diags = appendError(diags, err)
	err = d.Set("display_name", object.Metadata.DisplayName)
	diags = appendError(diags, err)

	err = d.Set("project", object.Metadata.Project)
	diags = appendError(diags, err)

	labelsRaw := object.Metadata.Labels
	// FIXME: is this condition correct and necessary?
	if labelsRaw != nil {
		err = d.Set("label", unmarshalLabels(labelsRaw))
		diags = appendError(diags, err)
	}

	status := object.Status
	err = d.Set("status", status)
	diags = appendError(diags, err)

	spec := object.Spec

	err = d.Set("description", spec.Description)
	diags = appendError(diags, err)

	return diags
}

func resourceServiceApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	service, diags := marshalService(d)
	if diags.HasError() {
		return diags
	}

	err := clientApplyObject(ctx, client, service)
	if err != nil {
		return diag.Errorf("could not add service: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	objects, err := client.GetObjects(ctx, project, manifest.KindService, nil, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalService(d, manifest.FilterByKind[v1alpha.Service](objects))
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project").(string)
	err := client.DeleteObjectsByName(ctx, project, manifest.KindService, false, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
