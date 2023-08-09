package nobl9

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	n9api "github.com/nobl9/nobl9-go"
	v1alpha "github.com/nobl9/nobl9-go"
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
	if name == "" {
		id, _ := uuid.NewUUID() // NewUUID returns always nil error
		name = id.String()
	}
	// FIXME: delete ObjectInternal field after SDK update - for now it's hardcoded organization.
	return &n9api.RoleBinding{
		ObjectInternal: v1alpha.ObjectInternal{
			Organization: "nobl9-dev",
		},
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
	_, isProjectRole := d.GetOk("project_ref")
	roleBinding := findRoleBindingByType(isProjectRole, objects)
	if roleBinding == nil {
		d.SetId("")
		return nil
	}

	var diags diag.Diagnostics
	metadata := roleBinding["metadata"].(map[string]interface{})
	err := d.Set("name", metadata["name"])
	diags = appendError(diags, err)

	spec := roleBinding["spec"].(map[string]interface{})
	err = d.Set("user", spec["user"])
	diags = appendError(diags, err)
	err = d.Set("role_ref", spec["roleRef"])
	diags = appendError(diags, err)
	err = d.Set("project_ref", spec["projectRef"])
	diags = appendError(diags, err)

	return diags
}

func findRoleBindingByType(projectRole bool, objects []n9api.AnyJSONObj) n9api.AnyJSONObj {
	for _, object := range objects {
		if projectRole && containsProjectRef(object) {
			return object
		} else if !projectRole && !containsProjectRef(object) {
			return object
		}
	}
	return nil
}

func containsProjectRef(obj n9api.AnyJSONObj) bool {
	spec := obj["spec"].(map[string]interface{})
	return spec["projectRef"] != nil
}

func resourceRoleBindingApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	ap := marshalRoleBinding(d)

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		err := clientApplyObject(ctx, client, ap)
		if err != nil {
			// FIXME: Uncomment after sdk fix.
			//if errors.Is(err, sdk.ErrConcurrencyIssue) {
			//	return resource.RetryableError(err)
			//}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.Errorf("could not add project: %s", err.Error())
	}

	d.SetId(ap.Metadata.Name)

	return resourceRoleBindingRead(ctx, d, meta)
}

func resourceRoleBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	objects, err := client.GetObjects(ctx, project, 11, nil, d.Id()) // FIXME: Can it be just '11' here?
	if err != nil {
		return diag.FromErr(err)
	}

	return unmarshalRoleBinding(d, objects)
}

func resourceRoleBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getNewClient(config)
	if ds != nil {
		return ds
	}

	project := d.Get("project").(string)
	if project == "" {
		project = wildcardProject
	}
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		err := client.DeleteObjectsByName(ctx, project, 11, false, d.Id()) // FIXME: Can it be just '11' here?
		if err != nil {
			// FIXME: Uncomment after sdk fix.
			//if errors.Is(err, sdk.ErrConcurrencyIssue) {
			//	return resource.RetryableError(err)
			//}
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
