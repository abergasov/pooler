package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIDeployLoadsConfigurationBeforeOperation(t *testing.T) {
	// given
	input := writeInput(t, "input files/deployment.json", deploymentInput)
	command := exec.Command("go", "run", ".", "deploy", "--network", "ethereum", "--input", input, "--contract", "0x123")

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err, "go run succeeded without its required local configuration")
	text := string(output)
	require.Contains(t, text, "configuration file not found")
	require.NotContains(t, text, "deploy preflight failed")
}

func TestCLIUsageDoesNotRequireConfiguration(t *testing.T) {
	// given
	command := exec.Command("go", "run", ".")

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err, "go run succeeded without a command")
	require.Contains(t, string(output), "usage: pooler")
	require.NotContains(t, string(output), "configuration file not found")
}

func TestCLIDistributeAllLoadsConfigurationBeforeRegistry(t *testing.T) {
	// given
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "pooler")
	build := exec.Command("go", "build", "-o", binary, ".")
	require.NoError(t, build.Run())
	command := exec.Command(binary, "distribute-all", "--network", "ethereum")
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err, "distribute-all succeeded without configuration")
	text := string(output)
	require.Contains(t, text, "configuration file not found")
	require.NotContains(t, text, "contracts.json")
	require.NotContains(t, text, "nothing to process")
}

func TestMakeTargetsForwardOverriddenConfiguration(t *testing.T) {
	// given
	root := filepath.Clean(filepath.Join("..", ".."))
	contract := "0x0000000000000000000000000000000000000005"

	for _, target := range []string{"deploy", "set-distribution", "show", "distribute", "distribute-all", "verify"} {
		t.Run(target, func(t *testing.T) {
			input := writeInput(t, "input files/operation.json", inputForTarget(target))
			command := exec.Command("make", "-n", target, "NETWORK=ethereum", "INPUT="+input, "CONTRACT="+contract)
			command.Dir = root

			// when
			output, err := command.CombinedOutput()
			// then
			require.NoError(t, err, "make dry run failed: %s", output)
			text := string(output)
			for _, want := range []string{
				"pooler " + target,
				`--network "ethereum"`,
				`--input "` + input + `"`,
				`--contract "` + contract + `"`,
			} {
				require.Contains(t, text, want)
			}
		})
	}
}

func TestMakeDefaults(t *testing.T) {
	// given
	command := exec.Command("make", "deploy")
	command.Dir = filepath.Clean(filepath.Join("..", ".."))

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err, "make deploy succeeded without its required local configuration")
	text := string(output)
	require.Contains(t, text, "configuration file not found")
	require.NotContains(t, text, "deploy preflight failed")
}

const deploymentInput = `{
  "name": "Reward pool",
  "description": "Quarterly rewards",
  "owner": "0x0000000000000000000000000000000000000001",
  "feeWallet": "0x0000000000000000000000000000000000000002",
  "feeBps": 0,
  "token": "0x0000000000000000000000000000000000000003",
  "recipients": [{"address": "0x0000000000000000000000000000000000000004", "shareBps": 10000}]
}`

const distributionInput = `{
  "recipients": [{"address": "0x0000000000000000000000000000000000000004", "shareBps": 10000}]
}`

func inputForTarget(target string) string {
	if target == "deploy" {
		return deploymentInput
	}
	return distributionInput
}

func writeInput(t *testing.T, name, input string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(input), 0o600))
	return path
}
