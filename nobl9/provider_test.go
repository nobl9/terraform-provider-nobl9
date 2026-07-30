package nobl9

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			defaultToConfigFileForAcceptanceTests()
			testAccProviderServer.once.Do(func() {
				testAccProviderServer.srv, testAccProviderServer.err = testAccNewMux(context.Background())
			})
			return testAccProviderServer.srv, testAccProviderServer.err
		},
	}
)

func defaultToConfigFileForAcceptanceTests() {
	if _, ok := os.LookupEnv("NOBL9_NO_CONFIG_FILE"); !ok {
		_ = os.Setenv("NOBL9_NO_CONFIG_FILE", "false")
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderResolvesCredentials(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`defaultContext = "default"

[Contexts]
  [Contexts.default]
    clientId = "config-client-id"
    clientSecret = "config-value"
`), 0o600))

	var emptyValue string
	configValue := "config-value"
	envValue := "env-value"
	providerValue := "provider-value"
	tests := map[string]struct {
		providerConfig       map[string]any
		envClientID          string
		envClientSecret      string
		expectedClientID     string
		expectedClientSecret string
		expectedNoConfigFile bool
	}{
		"config file fallback for empty credentials": {
			providerConfig: map[string]any{
				"client_id":     "",
				"client_secret": emptyValue,
			},
			expectedClientID:     "config-client-id",
			expectedClientSecret: configValue,
		},
		"environment variables before config file": {
			envClientID:          "env-client-id",
			envClientSecret:      envValue,
			expectedClientID:     "env-client-id",
			expectedClientSecret: envValue,
		},
		"provider configuration before environment and config file": {
			providerConfig: map[string]any{
				"client_id":     "provider-client-id",
				"client_secret": providerValue,
			},
			envClientID:          "env-client-id",
			envClientSecret:      envValue,
			expectedClientID:     "provider-client-id",
			expectedClientSecret: providerValue,
		},
		"config file disabled": {
			providerConfig: map[string]any{
				"client_id":      "provider-client-id",
				"client_secret":  providerValue,
				"no_config_file": true,
			},
			expectedClientID:     "provider-client-id",
			expectedClientSecret: providerValue,
			expectedNoConfigFile: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resetProviderClient(t)
			setSDKConfigFileTestEnv(t, configPath)
			unsetProviderConfigEnv(t)
			if tt.envClientID != "" {
				t.Setenv("NOBL9_CLIENT_ID", tt.envClientID)
				t.Setenv("NOBL9_CLIENT_SECRET", tt.envClientSecret)
			}

			data := schema.TestResourceDataRaw(t, Provider().Schema, tt.providerConfig)
			client, diags := getClient(getProviderConfig(data))

			require.False(t, diags.HasError(), diags)
			require.NotNil(t, client)
			assert.Equal(t, tt.expectedClientID, client.Config.ClientID)
			assert.Equal(t, tt.expectedClientSecret, client.Config.ClientSecret)
			if tt.expectedNoConfigFile {
				assert.Nil(t, client.Config.GetFileConfig())
			}
		})
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

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	value, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset environment variable %s: %v", key, err)
	}
	t.Cleanup(func() {
		if exists {
			if err := os.Setenv(key, value); err != nil {
				t.Errorf("restore environment variable %s: %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("unset environment variable %s during cleanup: %v", key, err)
		}
	})
}

func resetProviderClient(t *testing.T) {
	t.Helper()

	sharedClient = nil
	once = sync.Once{}
	t.Cleanup(func() {
		sharedClient = nil
		once = sync.Once{}
	})
}

func setSDKConfigFileTestEnv(t *testing.T, configPath string) {
	t.Helper()

	t.Setenv("TERRAFORM_NOBL9_CONFIG_FILE_PATH", configPath)
	t.Setenv("TERRAFORM_NOBL9_NO_CONFIG_FILE", "false")
	t.Setenv("TERRAFORM_NOBL9_DEFAULT_CONTEXT", "")
	t.Setenv("TERRAFORM_NOBL9_CLIENT_ID", "")
	t.Setenv("TERRAFORM_NOBL9_CLIENT_SECRET", "")
	t.Setenv("TERRAFORM_NOBL9_ACCESS_TOKEN", "")
	t.Setenv("TERRAFORM_NOBL9_DISABLE_OKTA", "false")
	t.Setenv("TERRAFORM_NOBL9_FILES_PROMPT_ENABLED", "")
	t.Setenv("TERRAFORM_NOBL9_FILES_PROMPT_THRESHOLD", "")
	t.Setenv("TERRAFORM_NOBL9_PROJECT", "")
	t.Setenv("TERRAFORM_NOBL9_URL", "")
	t.Setenv("TERRAFORM_NOBL9_OKTA_ORG_URL", "")
	t.Setenv("TERRAFORM_NOBL9_OKTA_AUTH_SERVER", "")
	t.Setenv("TERRAFORM_NOBL9_ORGANIZATION", "")
	t.Setenv("TERRAFORM_NOBL9_TIMEOUT", "")
	t.Setenv("TERRAFORM_NOBL9_CA_CERT_FILE", "")
}

func unsetProviderConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"NOBL9_CLIENT_ID",
		"NOBL9_CLIENT_SECRET",
		"NOBL9_OKTA_URL",
		"NOBL9_OKTA_AUTH",
		"NOBL9_PROJECT",
		"NOBL9_URL",
		"NOBL9_ORG",
		"NOBL9_NO_CONFIG_FILE",
	} {
		unsetEnv(t, key)
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
