package frameworkprovider

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/terraform-provider-nobl9/nobl9"
)

func TestMain(m *testing.M) {
	e2etestutils.SetToolName("Terraform")

	code := m.Run()
	if _, ok := os.LookupEnv(resource.EnvTfAcc); ok {
		e2etestutils.Cleanup()
	}
	os.Exit(code)
}

// testAccNewMux returns a new provider server which can multiplex
// between the SDK and framework provider implementations.
func testAccNewMux(ctx context.Context) (tfprotov6.ProviderServer, error) {
	mux, err := tf6muxserver.NewMuxServer(
		ctx,
		func() tfprotov6.ProviderServer {
			srv, _ := tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer {
				return schema.NewGRPCProviderServer(nobl9.Provider())
			})
			return srv
		},
		providerserver.NewProtocol6(New()),
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

var testSDKClient = struct {
	client *sdk.Client
	once   sync.Once
}{}

// testAccSetup is a helper function that is called before running acceptance tests.
// It is used to setup [sdk.Client] which is used to interact with the Nobl9 API.
func testAccSetup(t *testing.T) {
	t.Helper()
	checkIfAcceptanceTestIsSet(t)

	// Check ENVs everytime to fail all tests using the SDK client
	for _, key := range []string{
		"NOBL9_CLIENT_ID",
		"NOBL9_CLIENT_SECRET",
	} {
		_, ok := os.LookupEnv(key)
		require.True(t, ok, "required environment variable %q is not set", key)
	}

	// Initialize the SDK client.
	testSDKClient.once.Do(func() {
		providerModel := ProviderModel{}
		diags := providerModel.setDefaultsFromEnv()
		if diags.HasError() {
			t.Fatalf("failed to set required environment variables: %v", diags.Errors())
		}
		client, diags := newSDKClient(providerModel)
		if diags.HasError() {
			t.Fatalf("failed initialize Nobl9 SDK client: %v", diags.Errors())
		}
		testSDKClient.client = client.client

		e2etestutils.SetClient(testSDKClient.client)

		org, err := testSDKClient.client.GetOrganization(context.Background())
		require.NoError(t, err)
		fmt.Printf("Running Terraform acceptance tests\nOrganization: %s\nAuth Server: %s\nClient ID: %s\n\n",
			org,
			testSDKClient.client.Config.OktaOrgURL.JoinPath(testSDKClient.client.Config.OktaAuthServer),
			testSDKClient.client.Config.ClientID)
	})
}

// assertResourceWasApplied is a test check function that asserts if the resource was applied
// and that it matches the expected [manifest.Object] shape.
func assertResourceWasApplied(t *testing.T, ctx context.Context, expected manifest.Object) resource.TestCheckFunc {
	failureErr := errors.New("failed to assert if resource was applied")
	return func(s *terraform.State) error {
		objects, err := getObjectsFromTheNobl9API(t, ctx, expected)
		if err != nil {
			return errors.Wrap(failureErr, err.Error())
		}
		if !assert.Len(t, objects, 1) {
			return errors.Wrap(failureErr, "API returned unexpected number of objects")
		}
		switch v := objects[0].(type) {
		case v1alphaProject.Project:
			v.Spec.CreatedAt = ""
			v.Spec.CreatedBy = ""
			objects[0] = v
		case v1alphaSLO.SLO:
			assert.NotEmpty(t, v.Status)
			assert.NotEmpty(t, v.Spec.CreatedAt)
			assert.NotEmpty(t, v.Spec.CreatedBy)
			v.Status = nil
			v.Spec.CreatedAt = ""
			v.Spec.CreatedBy = ""
			assert.NotEmpty(t, v.Spec.TimeWindows[0].Period)
			v.Spec.TimeWindows[0].Period = nil
			objects[0] = v
		}
		if !assert.Equal(t, expected, objects[0]) {
			return errors.Wrap(failureErr, "objects are not equal")
		}
		return nil
	}
}

// assertResourceWasDeleted is a test check function that asserts if the resource was deleted from the Nobl9 platform.
func assertResourceWasDeleted(t *testing.T, ctx context.Context, expected manifest.Object) resource.TestCheckFunc {
	t.Helper()

	failureErr := errors.New("failed to assert if resource was deleted")
	return func(s *terraform.State) error {
		objects, err := getObjectsFromTheNobl9API(t, ctx, expected)
		if err != nil {
			return errors.Wrap(failureErr, err.Error())
		}
		if !assert.Len(t, objects, 0) {
			return errors.Wrap(failureErr, "expected no objects to be returned by the API")
		}
		return nil
	}
}

func getObjectsFromTheNobl9API(t *testing.T, ctx context.Context, object manifest.Object) ([]manifest.Object, error) {
	t.Helper()

	headers := http.Header{}
	if projectScoped, ok := object.(manifest.ProjectScopedObject); ok {
		headers.Set(sdk.HeaderProject, projectScoped.GetProject())
	}
	params := url.Values{v1Objects.QueryKeyName: []string{object.GetName()}}
	objects, err := testSDKClient.client.Objects().V1().Get(ctx, object.GetKind(), headers, params)
	if !assert.NoError(t, err) {
		return nil, err
	}
	return objects, nil
}

// addTestLabels adds origin label to the provided [Labels],
// so it's easier to locate the leftovers from these tests.
// It also adds unique test identifier label to the provided labels
// so that we can reliably retrieve objects created within a given test.
func addTestLabels(t *testing.T, labels Labels) Labels {
	t.Helper()
	v1alphaLabels := e2etestutils.AnnotateLabels(t, nil)
	if labels == nil {
		labels = make(Labels, 0, len(v1alphaLabels))
	}
	for _, k := range slices.Sorted(maps.Keys(v1alphaLabels)) {
		i := slices.IndexFunc(labels, func(l LabelBlockModel) bool { return l.Key == k })
		if i >= 0 {
			labels[i].Values = v1alphaLabels[k]
		} else {
			labels = append(labels, LabelBlockModel{
				Key:    k,
				Values: v1alphaLabels[k],
			})
		}
	}
	slices.SortFunc(labels, func(a, b LabelBlockModel) int { return cmp.Compare(a.Key, b.Key) })
	return labels
}

var commonAnnotations = v1alpha.MetadataAnnotations{"origin": "sdk-e2e-test"}

// checkIfAcceptanceTestIsSet checks if the acceptance test environment variable is set.
func checkIfAcceptanceTestIsSet(t *testing.T) {
	if _, ok := os.LookupEnv(resource.EnvTfAcc); !ok {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
	}
}

// assertHCL is a helper function that checks if the provided HCL configuration is valid.
func assertHCL(t *testing.T, config string) {
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(config), "test.hcl")
	if diags != nil && diags.HasErrors() {
		t.Fatalf("failed to parse test.hcl: %v\nfile contents:\n%s", diags, config)
	}
}

func readExpectedConfig(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(fmt.Sprintf("test_data/expected/%s", filename))
	if err != nil {
		t.Fatalf("failed to read expected config %q: %v", filename, err)
	}
	return string(data)
}
