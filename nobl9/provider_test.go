package nobl9

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	n9api "github.com/nobl9/nobl9-go"
)

var testProvider *schema.Provider
var testProject string

func init() {
	testProject = os.Getenv("NOBL9_PROJECT")
}

func ProviderFactory() map[string]func() (*schema.Provider, error) {
	testProvider = Provider()
	return map[string]func() (*schema.Provider, error){
		"nobl9": func() (*schema.Provider, error) {
			return testProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("NOBL9_URL"); err == "" {
		t.Fatal("NOBL9_URL must be set for acceptance tests")
	}
	if err := os.Getenv("NOBL9_ORG"); err == "" {
		t.Fatal("NOBL9_ORG must be set for acceptance tests")
	}
	if err := os.Getenv("NOBL9_PROJECT"); err == "" {
		t.Fatal("NOBL9_PROJECT must be set for acceptance tests")
	}
	if err := os.Getenv("NOBL9_CLIENT_ID"); err == "" {
		t.Fatal("NOBL9_CLIENT_ID must be set for acceptance tests")
	}
	if err := os.Getenv("NOBL9_CLIENT_SECRET"); err == "" {
		t.Fatal("NOBL9_CLIENT_SECRET must be set for acceptance tests")
	}
	if err := os.Getenv("NOBL9_OKTA_URL"); err == "" {
		t.Fatal("NOBL9_OKTA_URL must be set for acceptance tests")
	}
	if err := os.Getenv("NOBL9_OKTA_AUTH"); err == "" {
		t.Fatal("NOBL9_OKTA_AUTH must be set for acceptance tests")
	}
}

func CheckObjectCreated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("ID not set")
		}
		return nil
	}
}

func CheckDestory(rsType string, objectType n9api.Object) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		config := testProvider.Meta().(ProviderConfig)
		client, ds := newClient(config, testProject)
		if ds.HasError() {
			return fmt.Errorf("unable create client when deleting objects")
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != rsType {
				continue
			}
			if _, err := client.GetObject(objectType, "", rs.Primary.ID); err != nil {
				return err
			}
		}

		return nil
	}
}
