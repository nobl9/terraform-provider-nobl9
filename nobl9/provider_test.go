package nobl9

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

//nolint:gochecknoglobals
var (
	testProvider *schema.Provider
	testProject  string
)

//nolint:gochecknoinits
func init() {
	testProject = os.Getenv("NOBL9_PROJECT")
	if testProject == "" {
		testProject = "default"
	}
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
		client, ds := getClient(config)
		if ds.HasError() {
			return fmt.Errorf("unable create client when deleting objects")
		}

		ctx := context.Background()
		for _, rs := range s.RootModule().Resources {
			if rs.Type != rsType {
				continue
			}

			if _, err := client.Objects().V1().Get(
				ctx,
				kind,
				http.Header{sdk.HeaderProject: []string{testProject}},
				url.Values{v1Objects.QueryKeyName: []string{rs.Primary.ID}},
			); err != nil {
				return err
			}
		}

		return nil
	}
}
