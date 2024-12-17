package nobl9

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Nobl9DataSource(t *testing.T) {
	cases := []struct {
		name       string
		configFunc func(name string) string
	}{
		{"test-external-id-data-source", testExternalIDDataSource},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: tc.configFunc(tc.name),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrSet(
								fmt.Sprintf("data.nobl9_aws_iam_role_external_id.%s", tc.name),
								"external_id",
							),
							resource.TestCheckResourceAttrSet(
								fmt.Sprintf("data.nobl9_aws_iam_role_external_id.%s", tc.name),
								"account_id",
							),
						),
					},
				},
			})
		})
	}
}

func testExternalIDDataSource(name string) string {
	return fmt.Sprintf(`data "nobl9_aws_iam_role_external_id" "%s" {
		name = "test"
	}`, name)
}
