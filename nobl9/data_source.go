package nobl9

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func dataSourceAWSIAMRoleAuthExternalID() *schema.Resource {
	return &schema.Resource{
		Description: "[Cross account IAM roles](https://docs.nobl9.com/Sources/Amazon_CloudWatch/#cross-account-iam-roles-new)",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Data source name",
			},
			"external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Nobl9 AWS Account ID",
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External ID",
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
	objects, err := client.GetAWSIAMRoleAuthExternalIDs(ctx, directName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(directName)

	return unmarshallDataSourceAWSIAMRoleAuthExternalID(d, objects)
}

func unmarshallDataSourceAWSIAMRoleAuthExternalID(
	d *schema.ResourceData,
	objects *v1alpha.AWSIAMRoleAuthExternalIDs,
) diag.Diagnostics {
	var diags diag.Diagnostics

	set(d, "external_id", objects.ExternalID, &diags)
	set(d, "account_id", objects.AccountID, &diags)

	return diags
}
