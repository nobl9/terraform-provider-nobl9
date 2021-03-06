package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	n9api "github.com/nobl9/nobl9-go"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"description":  schemaDescription(),
		},
		CreateContext: resourceProjectApply,
		UpdateContext: resourceProjectApply,
		DeleteContext: resourceProjectDelete,
		ReadContext:   resourceProjectRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Project configuration documentation]()",
	}
}

func marshalProject(d *schema.ResourceData) *n9api.Project {
	return &n9api.Project{
		APIVersion: n9api.APIVersion,
		Kind:       n9api.KindProject,
		Metadata: n9api.ProjectMetadata{
			Name:        d.Get("name").(string),
			DisplayName: d.Get("display_name").(string),
		},
		Spec: n9api.ProjectSpec{
			Description: d.Get("description").(string),
		},
	}
}

func unmarshalProject(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	metadata := object["metadata"].(map[string]interface{})
	err := d.Set("name", metadata["name"])
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata["displayName"])
	diags = appendError(diags, err)

	spec := object["spec"].(map[string]interface{})
	err = d.Set("description", spec["description"])
	diags = appendError(diags, err)

	return diags
}

func resourceProjectApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, "")
	if ds != nil {
		return ds
	}

	ap := marshalProject(d)

	var p n9api.Payload
	p.AddObject(ap)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add project: %s", err.Error())
	}

	d.SetId(ap.Metadata.Name)

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, "")
	if ds.HasError() {
		return ds
	}

	objects, err := client.GetObject(n9api.ObjectProject, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalProject(d, objects)
}

func resourceProjectDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := newClient(config, "")
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectProject, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
