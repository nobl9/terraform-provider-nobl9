package frameworkprovider

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nobl9": providerserver.NewProtocol6WithError(New("test")),
}

var testSDKClient = struct {
	client *sdk.Client
	once   sync.Once
}{}

// testAccPreCheck is a helper function that is called before running acceptance tests.
// It is used to setup [sdk.Client] which is used to interact with the Nobl9 API.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	// Initialize the SDK client.
	testSDKClient.once.Do(func() {
		for _, key := range []string{
			"NOBL9_CLIENT_ID",
			"NOBL9_CLIENT_SECRET",
		} {
			_, ok := os.LookupEnv(key)
			require.True(t, ok, "required environment variable %q is not set", key)
		}

		providerModel := ProviderModel{}
		diags := providerModel.setDefaultsFromEnv()
		if diags.HasError() {
			t.Fatalf("failed to set required environment variables: %v", diags.Errors())
		}
		client, diags := newSDKClient(providerModel, "test")
		if diags.HasError() {
			t.Fatalf("failed initialize Nobl9 SDK client: %v", diags.Errors())
		}
		testSDKClient.client = client.client
	})
}

// appendTestLabels appends a set of labels which help identify the applied objects in Nobl9 platform.
func appendTestLabels(labels Labels) Labels {
	return append(labels, LabelBlockModel{
		Key:    "origin",
		Values: []string{"terraform-acc-test"},
	})
}

// assertResourceWasApplied is a test check function that asserts if the resource was applied
// and that it matches the expected [manifest.Object] shape.
func assertResourceWasApplied(t *testing.T, expected manifest.Object) resource.TestCheckFunc {
	failureErr := errors.New("failed to assert if resource was applied")
	return func(s *terraform.State) error {
		headers := http.Header{}
		if projectScoped, ok := expected.(manifest.ProjectScopedObject); ok {
			headers.Set(sdk.HeaderProject, projectScoped.GetProject())
		}
		params := url.Values{v1Objects.QueryKeyName: []string{expected.GetName()}}
		objects, err := testSDKClient.client.Objects().V1().Get(context.Background(), expected.GetKind(), headers, params)
		if !assert.NoError(t, err) {
			return errors.Wrap(failureErr, err.Error())
		}
		if !assert.Len(t, objects, 1) {
			return errors.Wrap(failureErr, "API returned unexpected number of objects")
		}
		if !assert.Equal(t, objects[0], expected) {
			return errors.Wrap(failureErr, "objects are not equal")
		}
		return nil
	}
}
