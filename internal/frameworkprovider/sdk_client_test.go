package frameworkprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSDKClientUsesConfigFileCredentials(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`defaultContext = "default"

[Contexts]
  [Contexts.default]
    clientId = "config-client-id"
    clientSecret = "config-client-secret"
`), 0o600))

	t.Setenv("TERRAFORM_NOBL9_CONFIG_FILE_PATH", configPath)
	t.Setenv("TERRAFORM_NOBL9_NO_CONFIG_FILE", "false")
	t.Setenv("TERRAFORM_NOBL9_CLIENT_ID", "")
	t.Setenv("TERRAFORM_NOBL9_CLIENT_SECRET", "")

	client, diags := newSDKClient(ProviderModel{
		NoConfigFile: envConfigurableBool{BoolValue: basetypes.NewBoolValue(false)},
	})

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

	t.Setenv("TERRAFORM_NOBL9_CONFIG_FILE_PATH", configPath)
	t.Setenv("TERRAFORM_NOBL9_NO_CONFIG_FILE", "false")
	t.Setenv("TERRAFORM_NOBL9_CLIENT_ID", "")
	t.Setenv("TERRAFORM_NOBL9_CLIENT_SECRET", "")
	t.Setenv("TERRAFORM_NOBL9_ACCESS_TOKEN", "")

	client, diags := newSDKClient(ProviderModel{
		NoConfigFile: envConfigurableBool{BoolValue: basetypes.NewBoolValue(false)},
	})

	require.True(t, diags.HasError())
	require.Nil(t, client)
	assert.Equal(t, "missing Nobl9 client secret", diags[0].Summary())
}
