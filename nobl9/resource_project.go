package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name":         schemaName(),
			"display_name": schemaDisplayName(),
			"description":  schemaDescription(),
			"label":        schemaLabels(),
			"annotations":  schemaAnnotations(),
		},
		CustomizeDiff: resourceProjectValidate,
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

func marshalProject(d resourceInterface) (*v1alphaProject.Project, diag.Diagnostics) {
	labelsMarshaled, diags := getMarshaledLabels(d)
	if diags.HasError() {
		return nil, diags
	}

	annotationsMarshaled := getMarshaledAnnotations(d)

	project := v1alphaProject.New(
		v1alphaProject.Metadata{
			Name:        d.Get("name").(string),
			DisplayName: d.Get("display_name").(string),
			Labels:      labelsMarshaled,
			Annotations: annotationsMarshaled,
		},
		v1alphaProject.Spec{
			Description: d.Get("description").(string),
		},
	)
	return &project, diags
}

func unmarshalProject(d *schema.ResourceData, project v1alphaProject.Project) diag.Diagnostics {
	var diags diag.Diagnostics

	metadata := project.Metadata
	err := d.Set("name", metadata.Name)
	diags = appendError(diags, err)
	err = d.Set("display_name", metadata.DisplayName)
	diags = appendError(diags, err)

	if labelsRaw := metadata.Labels; len(labelsRaw) > 0 {
		err = d.Set("label", unmarshalLabels(labelsRaw))
		diags = appendError(diags, err)
	}

	if len(metadata.Annotations) > 0 {
		err = d.Set("annotations", metadata.Annotations)
		diags = appendError(diags, err)
	}

	spec := project.Spec
	err = d.Set("description", spec.Description)
	diags = appendError(diags, err)

	return diags
}

func resourceProjectValidate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	project, diags := marshalProject(diff)
	if diags.HasError() {
		return diagsToSingleError(diags)
	}
	errs := manifest.Validate([]manifest.Object{project})
	if errs != nil {
		return formatErrorsAsSingleError(errs)
	}
	return nil
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
	err := client.Objects().V1().Apply(ctx, resultProject)
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
	projects, err := client.Objects().V1().GetV1alphaProjects(ctx, v1Objects.GetProjectsRequest{
		Names: []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return handleResourceReadResult(d, projects, unmarshalProject)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	err := client.Objects().V1().DeleteByName(ctx, manifest.KindProject, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
