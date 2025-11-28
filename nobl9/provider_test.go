package nobl9

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"

	"github.com/nobl9/terraform-provider-nobl9/internal/frameworkprovider"
)

var testProject string

//nolint:gochecknoinits
func init() {
	testProject = os.Getenv("NOBL9_PROJECT")
	if testProject == "" {
		testProject = "default"
	}
}

// testAccNewMux returns a new provider server which can multiplex
// between the SDK and framework provider implementations.
func testAccNewMux(ctx context.Context) (tfprotov6.ProviderServer, error) {
	mux, err := tf6muxserver.NewMuxServer(
		ctx,
		func() tfprotov6.ProviderServer {
			srv, _ := tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer {
				return schema.NewGRPCProviderServer(Provider())
			})
			return srv
		},
		providerserver.NewProtocol6(frameworkprovider.New()),
	)
	if err != nil {
		return nil, err
	}
	return mux.ProviderServer(), nil
}

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var (
	testAccProviderServer struct {
		srv  tfprotov6.ProviderServer
		err  error
		once sync.Once
	}
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"nobl9": func() (tfprotov6.ProviderServer, error) {
			testAccProviderServer.once.Do(func() {
				testAccProviderServer.srv, testAccProviderServer.err = testAccNewMux(context.Background())
			})
			return testAccProviderServer.srv, testAccProviderServer.err
		},
	}
)

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

func CheckDestroy(rsType string, kind manifest.Kind) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		// When CheckDestroy is called, the client is already created.
		// There's no need to pass any config to this function at that point.
		client, ds := getClient(ProviderConfig{})
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
