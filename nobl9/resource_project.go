package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"description":  schemaDescription(),
			"label":        schemaLabels(),
		},
		CreateContext: resourceProjectApply,
		UpdateContext: resourceProjectApply,
		DeleteContext: resourceProjectDelete,
		ReadContext:   resourceProjectRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Project configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#project)",
	}
}

func marshalProject(d *schema.ResourceData) (*v1alpha.Project, diag.Diagnostics) {
	labelsMarshalled, diags := getMarshalledLabels(d)
	if diags.HasError() {
		return nil, diags
	}

	return &v1alpha.Project{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindProject,
		Metadata: v1alpha.ProjectMetadata{
			Name:        d.Get("name").(string),
			DisplayName: d.Get("display_name").(string),
			Labels:      labelsMarshalled,
		},
		Spec: v1alpha.ProjectSpec{
			Description: d.Get("description").(string),
		},
	}, diags
}

func unmarshalProject(d *schema.ResourceData, objects []v1alpha.Project) diag.Diagnostics {
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

	if labelsRaw := metadata.Labels; len(labelsRaw) > 0 {
		err = d.Set("label", unmarshalLabels(labelsRaw))
		diags = appendError(diags, err)
	}

	spec := object.Spec
	err = d.Set("description", spec.Description)
	diags = appendError(diags, err)

	return diags
}

func resourceProjectApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project, diags := marshalProject(d)
	if diags.HasError() {
		return diags
	}
	resultProject := manifest.SetDefaultProject([]manifest.Object{project}, config.Project)
	err := client.ApplyObjects(ctx, resultProject, false)
	if err != nil {
		return diag.Errorf("could not add project: %s", err.Error())
	}
	d.SetId(project.Metadata.Name)
	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	objects, err := client.GetObjects(ctx, d.Id(), manifest.KindProject, nil, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalProject(d, manifest.FilterByKind[v1alpha.Project](objects))
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	err := client.DeleteObjectsByName(ctx, d.Id(), manifest.KindProject, false, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
