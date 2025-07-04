package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1AuthData "github.com/nobl9/nobl9-go/sdk/endpoints/authdata/v1"
)

func dataSourceAWSIAMRoleAuthExternalID() *schema.Resource {
	return &schema.Resource{
		Description: "Returns external ID and AWS account ID that can be used to create [cross-account IAM roles " +
			"in AWS](https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cross-account-iam-roles-new).",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Data source name",
			},
			"external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External ID",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Nobl9 AWS Account ID",
			},
		},
		ReadContext: dataSourceAWSIAMRoleAuthExternalIDDRead,
	}
}

func dataSourceAWSIAMRoleAuthExternalIDDRead(
	ctx context.Context,
	d *schema.ResourceData,
	meta interface{},
) diag.Diagnostics {
	config := meta.(ProviderConfig)
	client, ds := getClient(config)
	if ds != nil {
		return ds
	}
	directName := d.Get("name").(string)
	objects, err := client.AuthData().V1().GetDirectIAMRoleIDs(ctx, client.Config.Project, directName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(directName)

	return unmarshalDataSourceAWSIAMRoleAuthExternalID(d, objects)
}

func unmarshalDataSourceAWSIAMRoleAuthExternalID(
	d *schema.ResourceData,
	objects *v1AuthData.IAMRoleIDs,
) diag.Diagnostics {
	var diags diag.Diagnostics
	set(d, "external_id", objects.ExternalID, &diags)
	set(d, "account_id", objects.AccountID, &diags)

	return diags
}
