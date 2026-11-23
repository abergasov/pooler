package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"pooler/internal/config"
)

func TestLoadBuildsRedactedConfiguration(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(`rpc:
  eth:
    - https://ethereum.example
  arb:
    - https://arbitrum.example
signing:
  privateKey: private-key
explorer:
  apiKey: explorer-key
`), 0o600))

	// when
	configuration, err := config.Load(path)

	// then
	require.NoError(t, err)
	require.Equal(t, []string{"https://arbitrum.example"}, configuration.RPC["arb"])
	require.Equal(t, "private-key", configuration.Signing.PrivateKey)
	require.Equal(t, "explorer-key", configuration.Explorer.APIKey)
}

func TestLoadRejectsIncompleteConfiguration(t *testing.T) {
	for name, test := range map[string]struct {
		contents string
		want     string
	}{
		"missing RPC":              {contents: "signing:\n  privateKey: private-key\nexplorer:\n  apiKey: explorer-key\n", want: "rpc is required"},
		"missing Ethereum RPC":     {contents: "rpc:\n  arb:\n    - https://arbitrum.example\nsigning:\n  privateKey: private-key\nexplorer:\n  apiKey: explorer-key\n", want: "rpc.eth is required"},
		"blank RPC endpoint":       {contents: "rpc:\n  eth:\n    - ' '\n  arb:\n    - https://arbitrum.example\nsigning:\n  privateKey: private-key\nexplorer:\n  apiKey: explorer-key\n", want: "rpc.eth contains a blank endpoint"},
		"missing private key":      {contents: "rpc:\n  eth:\n    - https://ethereum.example\n  arb:\n    - https://arbitrum.example\nexplorer:\n  apiKey: explorer-key\n", want: "signing.privateKey is required"},
		"blank private key":        {contents: "rpc:\n  eth:\n    - https://ethereum.example\n  arb:\n    - https://arbitrum.example\nsigning:\n  privateKey: '  '\nexplorer:\n  apiKey: explorer-key\n", want: "signing.privateKey is required"},
		"missing explorer API key": {contents: "rpc:\n  eth:\n    - https://ethereum.example\n  arb:\n    - https://arbitrum.example\nsigning:\n  privateKey: private-key\n", want: "explorer.apiKey is required"},
		"blank explorer API key":   {contents: "rpc:\n  eth:\n    - https://ethereum.example\n  arb:\n    - https://arbitrum.example\nsigning:\n  privateKey: private-key\nexplorer:\n  apiKey: '  '\n", want: "explorer.apiKey is required"},
	} {
		t.Run(name, func(t *testing.T) {
			// given
			path := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(path, []byte(test.contents), 0o600))

			// when
			_, err := config.Load(path)

			// then
			require.EqualError(t, err, test.want)
		})
	}
}

func TestValidateDoesNotTrimSecretValues(t *testing.T) {
	// given
	configuration := &config.Config{RPC: map[string][]string{
		"eth": {"https://ethereum.example"},
		"arb": {"https://arbitrum.example"},
	}}
	configuration.Signing.PrivateKey = " private-key "
	configuration.Explorer.APIKey = " explorer-key "

	// when
	err := configuration.Validate()

	// then
	require.NoError(t, err)
	require.Equal(t, " private-key ", configuration.Signing.PrivateKey)
	require.Equal(t, " explorer-key ", configuration.Explorer.APIKey)
}

func TestLoadDoesNotExposeConfigurationValuesOnErrors(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "config.yml")
	const secret = "do-not-print-this-secret"
	require.NoError(t, os.WriteFile(path, []byte("rpc: [https://user:"+secret+"@example"), 0o600))

	// when
	_, parseErr := config.Load(path)
	_, missingErr := config.Load(filepath.Join(t.TempDir(), secret, "config.yml"))

	// then
	require.EqualError(t, parseErr, "parse configuration")
	require.EqualError(t, missingErr, "configuration file not found")
	require.NotContains(t, parseErr.Error(), secret)
	require.NotContains(t, missingErr.Error(), secret)
}
