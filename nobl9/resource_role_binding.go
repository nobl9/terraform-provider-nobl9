package nobl9

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	n9api "github.com/nobl9/nobl9-go"
)

const wildcardProject = "*"

//nolint:lll
func resourceRoleBinding() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Automatically generated, unique name of the resource. Must match [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
			},
			"display_name": schemaDisplayName(),
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the user.",
			},
			"role_ref": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name.",
			},
			"project_ref": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Project name. When empty, `role_ref` has to be Organization Role.",
			},
		},
		CreateContext: resourceRoleBindingApply,
		UpdateContext: resourceRoleBindingApply,
		DeleteContext: resourceRoleBindingDelete,
		ReadContext:   resourceRoleBindingRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[RoleBinding configuration documentation](https://docs.nobl9.com/yaml-guide#rolebinding)",
	}
}

func marshalRoleBinding(d *schema.ResourceData) *n9api.RoleBinding {
	name := d.Get("name").(string)
	if name == "" {
		id, _ := uuid.NewUUID() // NewUUID returns always nil error
		name = id.String()
	}
	return &n9api.RoleBinding{
		APIVersion: n9api.APIVersion,
		Kind:       n9api.KindRoleBinding,
		Metadata: n9api.RoleBindingMetadata{
			Name: name,
		},
		Spec: n9api.RoleBindingSpec{
			User:       d.Get("user").(string),
			RoleRef:    d.Get("role_ref").(string),
			ProjectRef: d.Get("project_ref").(string),
		},
	}
}

func unmarshalRoleBinding(d *schema.ResourceData, objects []n9api.AnyJSONObj) diag.Diagnostics {
	if len(objects) != 1 {
		d.SetId("")
		return nil
	}
	object := objects[0]
	var diags diag.Diagnostics

	metadata := object["metadata"].(map[string]interface{})
	err := d.Set("name", metadata["name"])
	diags = appendError(diags, err)

	spec := object["spec"].(map[string]interface{})
	err = d.Set("user", spec["user"])
	diags = appendError(diags, err)
	err = d.Set("role_ref", spec["roleRef"])
	diags = appendError(diags, err)
	err = d.Set("project_ref", spec["projectRef"])
	diags = appendError(diags, err)

	return diags
}

func resourceRoleBindingApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config, wildcardProject)
	if ds != nil {
		return ds
	}

	ap := marshalRoleBinding(d)

	var p n9api.Payload
	p.AddObject(ap)

	err := client.ApplyObjects(p.GetObjects())
	if err != nil {
		return diag.Errorf("could not add project: %s", err.Error())
	}

	d.SetId(ap.Metadata.Name)

	return resourceRoleBindingRead(ctx, d, meta)
}

func resourceRoleBindingRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config, wildcardProject)
	if ds.HasError() {
		return ds
	}

	objects, err := client.GetObject(n9api.ObjectRoleBinding, "", d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalRoleBinding(d, objects)
}

func resourceRoleBindingDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	project := d.Get("project_ref").(string)
	if project == "" {
		project = wildcardProject
	}
	client, ds := getClient(config, project)
	if ds.HasError() {
		return ds
	}

	err := client.DeleteObjectsByName(n9api.ObjectRoleBinding, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
