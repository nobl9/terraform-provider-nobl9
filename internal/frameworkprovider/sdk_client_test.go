package frameworkprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSDKClientUsesConfigFileCredentialsByDefault(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`defaultContext = "default"

[Contexts]
  [Contexts.default]
    clientId = "config-client-id"
    clientSecret = "config-client-secret"
`), 0o600))

	setSDKConfigFileTestEnv(t, configPath)
	unsetProviderConfigEnv(t)

	provider := ProviderModel{
		ClientID:     envConfigurableString{StringValue: basetypes.NewStringValue("")},
		ClientSecret: envConfigurableString{StringValue: basetypes.NewStringValue("")},
	}
	defaultDiags := provider.setDefaultsFromEnv()
	require.False(t, defaultDiags.HasError(), defaultDiags.Errors())

	client, diags := newSDKClient(provider)

	require.False(t, diags.HasError(), diags.Errors())
	require.NotNil(t, client)
	assert.Equal(t, "config-client-id", client.client.Config.ClientID)
	assert.Equal(t, "config-client-secret", client.client.Config.ClientSecret)
}

func TestNewSDKClientRejectsPartialConfigFileCredentials(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`defaultContext = "default"

[Contexts]
  [Contexts.default]
    clientId = "config-client-id"
`), 0o600))

	setSDKConfigFileTestEnv(t, configPath)

	client, diags := newSDKClient(ProviderModel{
		NoConfigFile: envConfigurableBool{BoolValue: basetypes.NewBoolValue(false)},
	})

	require.True(t, diags.HasError())
	require.Nil(t, client)
	assert.Equal(t, "missing Nobl9 client secret", diags[0].Summary())
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

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	value, exists := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if exists {
			require.NoError(t, os.Setenv(key, value))
			return
		}
		require.NoError(t, os.Unsetenv(key))
	})
}
