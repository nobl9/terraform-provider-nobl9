package frameworkprovider

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/sdk"
	v1Objects "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/terraform-provider-nobl9/nobl9"
)

const originLabelValue = "terraform-acc-test"

var (
	objectsCounter            = atomic.Int64{}
	testStartTime             = time.Now()
	uniqueTestIdentifierLabel = struct {
		Key   string
		Value string
	}{
		Key:   "terraform-acc-test-id",
		Value: strconv.Itoa(int(testStartTime.UnixNano())),
	}
)

// testAccNewMux returns a new provider server which can multiplex
// between the SDK and framework provider implementations.
func testAccNewMux(ctx context.Context, version string) (tfprotov5.ProviderServer, error) {
	mux, err := tf5muxserver.NewMuxServer(
		ctx,
		func() tfprotov5.ProviderServer { return schema.NewGRPCProviderServer(nobl9.Provider(version)) },
		providerserver.NewProtocol5(New(version)),
	)
	if err != nil {
		return nil, err
	}
	return mux.ProviderServer(), nil
}

// testAccProtoV5ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"nobl9": func() (tfprotov5.ProviderServer, error) {
		return testAccNewMux(context.Background(), "test")
	},
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

// applyNobl9Objects is a helper function that applies the provided objects to the Nobl9 platform.
func applyNobl9Objects(t *testing.T, ctx context.Context, objects ...manifest.Object) {
	t.Helper()

	err := testSDKClient.client.Objects().V1().Apply(ctx, objects)
	assert.NoError(t, err)
}

// deleteNobl9Objects is a helper function that deletes the provided objects from the Nobl9 platform.
func deleteNobl9Objects(t *testing.T, ctx context.Context, objects ...manifest.Object) {
	t.Helper()

	err := testSDKClient.client.Objects().V1().Delete(ctx, objects)
	assert.NoError(t, err)
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

// generateName generates a unique name for the test object.
func generateName() string {
	return fmt.Sprintf("terraform-acc-%d-%d", objectsCounter.Add(1), testStartTime.UnixNano())
}

// annotateV1alphaLabels adds origin label to the provided [v1alpha.Labels],
// so it's easier to locate the leftovers from these tests.
// It also adds unique test identifier label to the provided labels
// so that we can reliably retrieve objects created within a given test.
func annotateV1alphaLabels(t *testing.T, labels v1alpha.Labels) v1alpha.Labels {
	t.Helper()
	if labels == nil {
		labels = make(v1alpha.Labels, 3)
	}
	labels["origin"] = []string{originLabelValue}
	labels[uniqueTestIdentifierLabel.Key] = []string{uniqueTestIdentifierLabel.Value}
	labels["terraform-test-name"] = []string{t.Name()}
	return labels
}

// annotateLabels adds origin label to the provided [Labels],
// so it's easier to locate the leftovers from these tests.
// It also adds unique test identifier label to the provided labels
// so that we can reliably retrieve objects created within a given test.
func annotateLabels(t *testing.T, labels Labels) Labels {
	t.Helper()
	if labels == nil {
		labels = make(Labels, 0, 3)
	}
	v1alphaLabels := annotateV1alphaLabels(t, nil)
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
	return labels
}
