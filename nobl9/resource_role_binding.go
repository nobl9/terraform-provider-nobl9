package nobl9

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk"
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

func marshalRoleBinding(d *schema.ResourceData) *v1alpha.RoleBinding {
	name := d.Get("name").(string)
	if name == "" {
		id, _ := uuid.NewUUID() // NewUUID returns always nil error
		name = id.String()
	}
	userRoleBindingSpec := d.Get("user").(string)
	return &v1alpha.RoleBinding{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindRoleBinding,
		Metadata: v1alpha.RoleBindingMetadata{
			Name: name,
		},
		Spec: v1alpha.RoleBindingSpec{
			User:       &userRoleBindingSpec,
			RoleRef:    d.Get("role_ref").(string),
			ProjectRef: d.Get("project_ref").(string),
		},
	}
}

func unmarshalRoleBinding(d *schema.ResourceData, objects []v1alpha.RoleBinding) diag.Diagnostics {
	_, isProjectRole := d.GetOk("project_ref")
	roleBindingP := findRoleBindingByType(isProjectRole, objects)
	if roleBindingP == nil {
		d.SetId("")
		return nil
	}
	roleBinding := *roleBindingP

	var diags diag.Diagnostics
	metadata := roleBinding.Metadata
	err := d.Set("name", metadata.Name)
	diags = appendError(diags, err)

	spec := roleBinding.Spec
	err = d.Set("user", spec.User)
	diags = appendError(diags, err)
	err = d.Set("role_ref", spec.RoleRef)
	diags = appendError(diags, err)
	err = d.Set("project_ref", spec.ProjectRef)
	diags = appendError(diags, err)

	return diags
}

func findRoleBindingByType(projectRole bool, objects []v1alpha.RoleBinding) *v1alpha.RoleBinding {
	for _, object := range objects {
		if projectRole && containsProjectRef(object) {
			return &object
		} else if !projectRole && !containsProjectRef(object) {
			return &object
		}
	}
	return nil
}

func containsProjectRef(obj v1alpha.RoleBinding) bool {
	return obj.Spec.ProjectRef != ""
}

func resourceRoleBindingApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	roleBinding := marshalRoleBinding(d)
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		err := client.ApplyObjects(ctx, []manifest.Object{roleBinding}, false)
		if err != nil {
			if errors.Is(err, sdk.ErrConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.Errorf("could not add project: %s", err.Error())
	}

	d.SetId(roleBinding.Metadata.Name)
	return resourceRoleBindingRead(ctx, d, meta)
}

func resourceRoleBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	project := d.Get("project_ref").(string)
	if project == "" {
		project = wildcardProject
	}
	objects, err := client.GetObjects(ctx, project, manifest.KindRoleBinding, nil, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalRoleBinding(d, manifest.FilterByKind[v1alpha.RoleBinding](objects))
}

func resourceRoleBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project_ref").(string)
	if project == "" {
		project = wildcardProject
	}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		err := client.DeleteObjectsByName(ctx, project, manifest.KindRoleBinding, false, d.Id())
		if err != nil {
			if errors.Is(err, sdk.ErrConcurrencyIssue) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
