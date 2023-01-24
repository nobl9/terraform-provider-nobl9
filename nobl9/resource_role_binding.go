package nobl9

import (
	"context"
	"fmt"

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
				Description: "Automatically generated, unique name of the resource, must conform to the naming convention from [DNS RFC1123](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).",
			},
			"display_name": schemaDisplayName(),
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Okta User ID that can be retrieved from the Nobl9 UI (**Settings** > **Users**).",
			},
			"role_ref": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name; the role that you want the user to assume.",
			},
			"project_ref": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Project name, the project in which we want the user to assume the specified role. When `project_ref` is empty, `role_ref` must contain an Organization Role.",
			},
		},
		CreateContext: resourceRoleBindingApply,
		UpdateContext: resourceRoleBindingApply,
		DeleteContext: resourceRoleBindingDelete,
		ReadContext:   resourceRoleBindingRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "[Role Binding configuration | Nobl9 Documentation](https://docs.nobl9.com/yaml-guide#rolebinding)",
	}
}

func marshalRoleBinding(d *schema.ResourceData) *n9api.RoleBinding {
	name := d.Get("name").(string)
	project := d.Get("project_ref").(string)
	if name == "" {
		id, _ := uuid.NewUUID() // NewUUID returns always nil error
		name = id.String()
	}
	return &n9api.RoleBinding{
		APIVersion: n9api.APIVersion,
		Kind:       n9api.KindRoleBinding,
		Metadata: n9api.RoleBindingMetadata{
			Name: createName(name, project),
		},
		Spec: n9api.RoleBindingSpec{
			User:       d.Get("user").(string),
			RoleRef:    d.Get("role_ref").(string),
			ProjectRef: project,
		},
	}
}

func createName(name, project string) string {
	if project != "" {
		return fmt.Sprintf("%s-%s", name, project)
	}
	return name
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
