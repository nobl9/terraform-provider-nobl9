package nobl9

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"project":      schemaProject(),
			"description":  schemaDescription(),
			"labels":       schemaLabels(),
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
		Description: "* [Service configuration documentation](https://docs.nobl9.com/yaml-guide#service)",
	}
}

func marshalService(d *schema.ResourceData) (*n9api.Service, diag.Diagnostics) {
	metadataHolder, diags := marshalMetadata(d)
	if diags.HasError() {
		return nil, diags
	}
	return &n9api.Service{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindService,
			MetadataHolder: metadataHolder,
		},
		Spec: n9api.ServiceSpec{
			Description: d.Get("description").(string),
		},
	}, diags
}

func unmarshalService(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	if ds := unmarshalMetadata(object, d); ds.HasError() {
		diags = append(diags, ds...)
	}

	status := object["status"].(map[string]interface{})
	err := d.Set("status", status)
	diags = appendError(diags, err)

	spec := object["spec"].(map[string]interface{})

	err = d.Set("description", spec["description"])
	diags = appendError(diags, err)

	return diags
}

func resourceServiceApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds != nil {
		return ds
	}

	service, diags := marshalService(d)
	if diags.HasError() {
		return diags
	}

	var p n9api.Payload
	p.AddObject(service)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add service: %s", err.Error())
	}

	d.SetId(service.Metadata.Name)

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	client, ds := newClient(config, project)
	if ds.HasError() {
		return ds
	}

	objects, err := client.GetObject(n9api.ObjectService, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalService(d, objects)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, d.Get("project").(string))
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectService, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
