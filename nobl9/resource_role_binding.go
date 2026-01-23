package nobl9

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
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
			"account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Account ID that can be retrieved from the Nobl9 UI (for **Settings** > **Users** as User ID or **API Keys** as from Client ID).",
				ConflictsWith: []string{"user", "group_ref"},
			},
			"user": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "Use 'account_id' instead. The 'user' field is deprecated and will be removed in a future.",
				Description:   "Okta User ID that can be retrieved from the Nobl9 UI (**Settings** > **Users**). Deprecated: use 'account_id' instead.",
				ConflictsWith: []string{"account_id", "group_ref"},
			},
			"group_ref": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Group name that can be retrieved from the Nobl9 UI (**Settings** > **Groups**) or using sloctl `get usergroups` command.",
				ConflictsWith: []string{"user", "account_id"},
			},
			"role_ref": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name; the role that you want the user or group to assume.",
			},
			"project_ref": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Project name, the project in which we want the user or group to assume the specified role. When `project_ref` is empty, `role_ref` must contain an Organization Role.",
			},
		},
		CustomizeDiff: resourceRoleBindingValidate,
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

func marshalRoleBinding(r resourceInterface) *v1alphaRoleBinding.RoleBinding {
	name := r.Get("name").(string)
	if name == "" {
		id, _ := uuid.NewUUID() // NewUUID returns always nil error
		name = id.String()
	}

	// Handle user/account_id - prefer account_id, fall back to user for backward compatibility
	// Always send AccountID to the API (never the deprecated User field)
	var accountID *string
	if accountIDValue := r.Get("account_id").(string); accountIDValue != "" {
		accountID = &accountIDValue
	} else if userValue := r.Get("user").(string); userValue != "" {
		// Backward compatibility: if user is provided, use it as accountID in the SDK
		accountID = &userValue
	}

	var groupRef *string
	if groupRefValue := r.Get("group_ref").(string); groupRefValue != "" {
		groupRef = &groupRefValue
	}

	roleBinding := v1alphaRoleBinding.New(
		v1alphaRoleBinding.Metadata{
			Name: name,
		},
		v1alphaRoleBinding.Spec{
			AccountID:  accountID,
			GroupRef:   groupRef,
			RoleRef:    r.Get("role_ref").(string),
			ProjectRef: r.Get("project_ref").(string),
		},
	)
	return &roleBinding
}

func unmarshalRoleBinding(d *schema.ResourceData, objects []v1alphaRoleBinding.RoleBinding) diag.Diagnostics {
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

	// Handle backward compatibility for user/account_id
	// If state has 'user', keep it there; otherwise use 'account_id'
	_, hasUserInState := d.GetOk("user")

	// Determine which field to populate based on what's in state
	if hasUserInState {
		// Preserve 'user' field for backward compatibility with existing states
		if spec.AccountID != nil {
			err = d.Set("user", spec.AccountID)
			diags = appendError(diags, err)
		} else if spec.User != nil {
			// Fallback for API responses that still send 'user'
			err = d.Set("user", spec.User)
			diags = appendError(diags, err)
		}
	} else {
		// Use 'account_id' for new resources or migrated configs
		if spec.AccountID != nil {
			err = d.Set("account_id", spec.AccountID)
			diags = appendError(diags, err)
		} else if spec.User != nil {
			// Fallback for API responses that still send 'user'
			err = d.Set("account_id", spec.User)
			diags = appendError(diags, err)
		}
	}

	err = d.Set("group_ref", spec.GroupRef)
	diags = appendError(diags, err)
	err = d.Set("role_ref", spec.RoleRef)
	diags = appendError(diags, err)
	err = d.Set("project_ref", spec.ProjectRef)
	diags = appendError(diags, err)

	return diags
}

func findRoleBindingByType(projectRole bool, objects []v1alphaRoleBinding.RoleBinding) *v1alphaRoleBinding.RoleBinding {
	for _, object := range objects {
		if projectRole && containsProjectRef(object) {
			return &object
		} else if !projectRole && !containsProjectRef(object) {
			return &object
		}
	}
	return nil
}

func containsProjectRef(obj v1alphaRoleBinding.RoleBinding) bool {
	return obj.Spec.ProjectRef != ""
}

func resourceRoleBindingValidate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	roleBinding := marshalRoleBinding(diff)
	errs := manifest.Validate([]manifest.Object{roleBinding})
	if errs != nil {
		return formatErrorsAsSingleError(errs)
	}
	return nil
}

func resourceRoleBindingApply(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}

	roleBinding := marshalRoleBinding(d)
	if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *retry.RetryError {
		err := client.Objects().V1().Apply(ctx, []manifest.Object{roleBinding})
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.Errorf("could not add role binding: %s", err.Error())
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
	roleBindings, err := client.Objects().V1().GetV1alphaRoleBindings(ctx, v1Objects.GetRoleBindingsRequest{
		Project: project,
		Names:   []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return unmarshalRoleBinding(d, roleBindings)
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

	if err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *retry.RetryError {
		err := client.Objects().V1().DeleteByName(ctx, manifest.KindRoleBinding, project, d.Id())
		if err != nil {
			if errors.Is(err, errConcurrencyIssue) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
