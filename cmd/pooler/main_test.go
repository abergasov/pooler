package main_test

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

func TestCLIRejectsInvalidInputBeforeBroadcast(t *testing.T) {
	// given
	workdir := t.TempDir()
	rpc := newValidationRPC(t, "0xa4b1", common.Address{})
	server := httptest.NewServer(rpc)
	defer server.Close()
	writeValidationConfig(t, workdir, server.URL)
	input := writeInput(t, "input.json", strings.Replace(deploymentInput, "\"shareBps\": 10000", "\"shareBps\": 9999", 1))
	binary := buildPooler(t, workdir)
	command := exec.Command(binary, "deploy", "--network", "arbitrum", "--input", input)
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err)
	require.Contains(t, string(output), "validate deployment input")
	require.Zero(t, rpc.sends.Load())
}

func TestCLIRejectsWrongRPCChainBeforeBroadcast(t *testing.T) {
	// given
	workdir := t.TempDir()
	rpc := newValidationRPC(t, "0x1", common.Address{})
	server := httptest.NewServer(rpc)
	defer server.Close()
	writeValidationConfig(t, workdir, server.URL)
	input := writeInput(t, "input.json", strings.Replace(deploymentInput, "0x0000000000000000000000000000000000000003", "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9", 1))
	binary := buildPooler(t, workdir)
	command := exec.Command(binary, "deploy", "--network", "arbitrum", "--input", input)
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err)
	require.Contains(t, string(output), "no usable RPC endpoint")
	require.Zero(t, rpc.sends.Load())
}

func TestCLIRejectsWrongSignerBeforeBroadcast(t *testing.T) {
	// given
	workdir := t.TempDir()
	rpc := newValidationRPC(t, "0xa4b1", common.HexToAddress("0x0000000000000000000000000000000000000001"))
	server := httptest.NewServer(rpc)
	defer server.Close()
	writeValidationConfig(t, workdir, server.URL)
	binary := buildPooler(t, workdir)
	command := exec.Command(binary, "distribute", "--network", "arbitrum", "--contract", "0x0000000000000000000000000000000000000002")
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err)
	require.Contains(t, string(output), "is not the current contract owner")
	require.Zero(t, rpc.sends.Load())
}

func TestCLIDistributeAllDoesNotRequireInput(t *testing.T) {
	// given
	workdir := t.TempDir()
	writeValidationConfig(t, workdir, "http://127.0.0.1:1")
	binary := buildPooler(t, workdir)
	command := exec.Command(binary, "distribute-all", "--network", "arbitrum")
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.NoError(t, err, string(output))
	require.Contains(t, string(output), "nothing to process")
	require.NotContains(t, string(output), "open input file")
}

func TestCLIShowRequiresConfigurationBeforeOperation(t *testing.T) {
	// given
	workdir := t.TempDir()
	binary := buildPooler(t, workdir)
	command := exec.Command(binary, "show", "--network", "arbitrum", "--contract", "0x0000000000000000000000000000000000000002")
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err)
	require.Contains(t, string(output), "configuration file not found")
	require.NotContains(t, string(output), "show preflight failed")
}

func TestCLIShowRejectsIncompleteConfigurationBeforeRPC(t *testing.T) {
	for name, contents := range map[string]string{
		"missing signing key":  "rpc:\n  eth:\n    - http://unused\n  arb:\n    - http://unused\nexplorer:\n  apiKey: key\n",
		"missing explorer key": "rpc:\n  eth:\n    - http://unused\n  arb:\n    - http://unused\nsigning:\n  privateKey: key\n",
	} {
		t.Run(name, func(t *testing.T) {
			// given
			workdir := t.TempDir()
			rpc := newValidationRPC(t, "0xa4b1", common.Address{})
			server := httptest.NewServer(rpc)
			defer server.Close()
			contents = strings.ReplaceAll(contents, "http://unused", server.URL)
			require.NoError(t, os.WriteFile(filepath.Join(workdir, "config.yml"), []byte(contents), 0o600))
			binary := buildPooler(t, workdir)
			command := exec.Command(binary, "show", "--network", "arbitrum", "--contract", "0x0000000000000000000000000000000000000002")
			command.Dir = workdir

			// when
			output, err := command.CombinedOutput()

			// then
			require.Error(t, err)
			require.Contains(t, string(output), "is required")
			require.Zero(t, rpc.calls.Load())
		})
	}
}

func TestCLIShowReadsLiveStateWithoutInput(t *testing.T) {
	// given
	workdir := t.TempDir()
	rpc := newValidationRPC(t, "0xa4b1", common.HexToAddress("0x0000000000000000000000000000000000000001"))
	server := httptest.NewServer(rpc)
	defer server.Close()
	writeValidationConfig(t, workdir, server.URL)
	binary := buildPooler(t, workdir)
	command := exec.Command(binary, "show", "--network", "arbitrum", "--contract", "0x0000000000000000000000000000000000000002")
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.NoError(t, err, string(output))
	require.Contains(t, string(output), "owner: 0x0000000000000000000000000000000000000001")
	require.Contains(t, string(output), "balance: 123")
	require.NotContains(t, string(output), "open input file")
	require.Greater(t, rpc.calls.Load(), int32(1))
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

func buildPooler(t *testing.T, workdir string) string {
	t.Helper()
	binary := filepath.Join(workdir, "pooler")
	build := exec.Command("go", "build", "-o", binary, ".")
	require.NoError(t, build.Run())
	return binary
}

func writeValidationConfig(t *testing.T, workdir, rpcURL string) {
	t.Helper()
	contents := "rpc:\n  eth:\n    - " + rpcURL + "\n  arb:\n    - " + rpcURL + "\nsigning:\n  privateKey: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\nexplorer:\n  apiKey: key\n"
	require.NoError(t, os.WriteFile(filepath.Join(workdir, "config.yml"), []byte(contents), 0o600))
}

type validationRPC struct {
	t       *testing.T
	chainID string
	owner   common.Address
	calls   atomic.Int32
	sends   atomic.Int32
}

func newValidationRPC(t *testing.T, chainID string, owner common.Address) *validationRPC {
	t.Helper()
	return &validationRPC{t: t, chainID: chainID, owner: owner}
}

func (rpc *validationRPC) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var call struct {
		ID     json.RawMessage   `json:"id"`
		Method string            `json:"method"`
		Params []json.RawMessage `json:"params"`
	}
	require.NoError(rpc.t, json.NewDecoder(request.Body).Decode(&call))
	rpc.calls.Add(1)

	result := any(nil)
	switch call.Method {
	case "eth_chainId":
		result = rpc.chainID
	case "eth_call":
		result = rpc.callResult(call.Params)
	case "eth_sendRawTransaction":
		rpc.sends.Add(1)
		result = "0x" + strings.Repeat("0", 64)
	default:
		require.FailNowf(rpc.t, "unexpected RPC method", "%s", call.Method)
	}

	response.Header().Set("Content-Type", "application/json")
	require.NoError(rpc.t, json.NewEncoder(response).Encode(map[string]any{"jsonrpc": "2.0", "id": call.ID, "result": result}))
}

func (rpc *validationRPC) callResult(params []json.RawMessage) string {
	var call struct {
		Data  string `json:"data"`
		Input string `json:"input"`
	}
	require.NotEmpty(rpc.t, params)
	require.NoError(rpc.t, json.Unmarshal(params[0], &call))
	data := call.Data
	if data == "" {
		data = call.Input
	}
	require.GreaterOrEqual(rpc.t, len(data), 10)

	switch data[:10] {
	case selector("owner()"):
		return encodedAddress(rpc.owner)
	case selector("feeWallet()"):
		return encodedAddress(common.HexToAddress("0x0000000000000000000000000000000000000002"))
	case selector("feeBps()"):
		return encodedUint(250)
	case selector("usdt()"):
		return encodedAddress(common.HexToAddress("0x0000000000000000000000000000000000000003"))
	case selector("getRecipients()"):
		return hexutil.Encode(append(common.LeftPadBytes([]byte{32}, common.HashLength), make([]byte, common.HashLength)...))
	case selector("balanceOf(address)"):
		return encodedUint(123)
	default:
		require.FailNowf(rpc.t, "unexpected contract method", "%s", data)
		return ""
	}
}

func selector(signature string) string {
	return hexutil.Encode(crypto.Keccak256([]byte(signature))[:4])
}

func encodedAddress(address common.Address) string {
	return hexutil.Encode(common.LeftPadBytes(address.Bytes(), common.HashLength))
}

func encodedUint(value uint64) string {
	return hexutil.Encode(common.LeftPadBytes(new(big.Int).SetUint64(value).Bytes(), common.HashLength))
}
