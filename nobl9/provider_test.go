package nobl9

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/nobl9/nobl9-go/manifest"
)

//nolint:gochecknoglobals
var (
	testProvider *schema.Provider
	testProject  string
)

//nolint:gochecknoinits
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

func CheckStateContainData(key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[key]
		if !ok {
			return fmt.Errorf("not found: %s", key)
		}
		if len(rs.String()) == 0 {
			return fmt.Errorf("data not set")
		}
		return nil
	}
}

func CheckDestroy(rsType string, kind manifest.Kind) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		config, ok := testProvider.Meta().(ProviderConfig)
		if !ok {
			return fmt.Errorf("could not cast data to ProviderConfig")
		}
		client := getClient(config)

		ctx := context.Background()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != rsType {
				continue
			}

			if _, err := client.GetObjects(ctx, testProject, kind, nil, rs.Primary.ID); err != nil {
				return err
			}
		}

		return nil
	}
}
