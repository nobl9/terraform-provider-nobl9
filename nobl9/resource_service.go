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
			"description":  schemaDescription(),
			"project":      schemaProject(),
			//"label":        schemaLabels(), // TODO enable when PC-3250 is done

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
		UpdateContext: resourceServiceApply,
		DeleteContext: resourceServiceDelete,
		ReadContext:   resourceServiceRead,
		Description:   "[Service configuration documentation](https://nobl9.github.io/techdocs_YAML_Guide/#service)",
	}
}

func marshalService(d *schema.ResourceData) *n9api.Service {
	return &n9api.Service{
		ObjectHeader: n9api.ObjectHeader{
			APIVersion:     n9api.APIVersion,
			Kind:           n9api.KindService,
			MetadataHolder: marshalMetadataWithLabels(d),
		},
		Spec: n9api.ServiceSpec{
			Description: d.Get("description").(string),
		},
	}
}

func unmarshalService(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]

	diags := unmarshalMetadataWithLabels(object, d)

	err := d.Set("status", object["status"])
	diags = appendError(diags, err)

	return diags
}

func resourceServiceApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	project := d.Get("project").(string)
	if project == "" {
		// project is empty when importing
		project = config.Project
	}
	client, ds := newClient(config, project)
	if ds != nil {
		return ds
	}

	ap := marshalService(d)

	var p n9api.Payload
	p.AddObject(ap)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add project: %s", err.Error())
	}

	d.SetId(ap.Metadata.Name)

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceServiceDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, "")
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectService, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
