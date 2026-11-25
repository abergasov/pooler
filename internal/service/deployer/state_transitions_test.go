package deployer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"pooler/internal/config"
	domain "pooler/internal/entities"
	"pooler/internal/operator"
	"pooler/internal/service/deployer"
)

func TestDeployRegistersConfirmedContract(t *testing.T) {
	// given
	t.Chdir(t.TempDir())
	copyDeploymentInputs(t)
	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	address := crypto.CreateAddress(crypto.PubkeyToAddress(key.PublicKey), 0)
	var receipts atomic.Int32
	var codeReads atomic.Int32
	server := httptest.NewServer(deploymentRPC(t, new(atomic.Int32), &deploymentRPCOptions{
		receipt:         "0x1",
		code:            "0x6000",
		contractAddress: address,
		receipts:        &receipts,
		codeReads:       &codeReads,
	}))
	defer server.Close()
	previousTransport := http.DefaultTransport
	http.DefaultTransport = explorerRoundTripper(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host != "api.etherscan.io" {
			return previousTransport.RoundTrip(request)
		}
		if request.Method == http.MethodPost {
			return explorerResponse(`{"status":"1","result":"guid-1"}`), nil
		}
		return explorerResponse(`{"status":"1","result":"Pass - Verified"}`), nil
	})
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)
	input := deploymentInput()

	// when
	result, _, err := deployer.New(deploymentConfiguration(server.URL)).Deploy(context.Background(), network, &input)
	registered, loadErr := operator.LoadContracts("contracts.json")

	// then
	require.NoError(t, err)
	require.Equal(t, int32(1), receipts.Load())
	require.Equal(t, int32(1), codeReads.Load())
	require.Equal(t, address.Hex(), result.Address)
	require.NoError(t, loadErr)
	require.Equal(t, []domain.Contract{{
		Name:        input.Name,
		Description: input.Description,
		Network:     network.Name,
		Address:     strings.ToLower(address.Hex()),
	}}, registered)
}

func TestDeployReportsPostSendStatesWithoutRetry(t *testing.T) {
	for _, tc := range []struct {
		name    string
		receipt string
		code    string
		want    string
	}{
		{name: "reverted receipt", receipt: "0x0", code: "0x6000", want: "deployment transaction"},
		{name: "empty code", receipt: "0x1", code: "0x", want: "deployed bytecode is empty"},
		{name: "successful send with missing receipt", receipt: "missing", code: "0x6000", want: "confirmation is uncertain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given
			t.Chdir(t.TempDir())
			copyDeploymentInputs(t)
			var sends atomic.Int32
			server := httptest.NewServer(deploymentRPC(t, &sends, &deploymentRPCOptions{receipt: tc.receipt, code: tc.code, contractAddress: common.Address{1}}))
			defer server.Close()
			network, err := domain.ParseNetwork("arbitrum")
			require.NoError(t, err)
			input := deploymentInput()
			configuration := &config.Config{RPC: map[string][]string{"arb": {server.URL}}}
			configuration.Signing.PrivateKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			configuration.Explorer.APIKey = "key"

			// when
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			result, _, err := deployer.New(configuration).Deploy(ctx, network, &input)

			// then
			require.ErrorContains(t, err, tc.want)
			require.NotEmpty(t, result.TransactionHash)
			require.Equal(t, int32(1), sends.Load())
			registry, readErr := os.ReadFile("contracts.json")
			require.NoError(t, readErr)
			require.Equal(t, "[]\n", string(registry))
		})
	}
}

func TestDeployRetainsHashWhenBroadcastOutcomeIsUncertain(t *testing.T) {
	// given
	t.Chdir(t.TempDir())
	copyDeploymentInputs(t)
	var sends atomic.Int32
	server := httptest.NewServer(deploymentRPC(t, &sends, &deploymentRPCOptions{code: "0x6000", sendError: true}))
	defer server.Close()
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)
	configuration := deploymentConfiguration(server.URL)

	// when
	input := deploymentInput()
	result, _, err := deployer.New(configuration).Deploy(context.Background(), network, &input)

	// then
	require.ErrorContains(t, err, "broadcast outcome is uncertain")
	require.NotEmpty(t, result.TransactionHash)
	require.Equal(t, int32(1), sends.Load())
	registry, readErr := os.ReadFile("contracts.json")
	require.NoError(t, readErr)
	require.Equal(t, "[]\n", string(registry))
}

func TestDeployReportsRegistrySaveFailureAfterConfirmation(t *testing.T) {
	// given
	t.Chdir(t.TempDir())
	copyDeploymentInputs(t)
	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	address := crypto.CreateAddress(crypto.PubkeyToAddress(key.PublicKey), 0)
	var sends atomic.Int32
	server := httptest.NewServer(deploymentRPC(t, &sends, &deploymentRPCOptions{
		receipt:         "0x1",
		code:            "0x6000",
		contractAddress: address,
		afterSend: func() {
			require.NoError(t, os.Remove("contracts.json"))
			require.NoError(t, os.Mkdir("contracts.json", 0o700))
		},
	}))
	defer server.Close()
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)
	configuration := deploymentConfiguration(server.URL)

	// when
	input := deploymentInput()
	result, _, err := deployer.New(configuration).Deploy(context.Background(), network, &input)

	// then
	require.ErrorContains(t, err, "cannot save contract registry")
	require.NotEmpty(t, result.TransactionHash)
	require.Equal(t, address.Hex(), result.Address)
	require.Contains(t, err.Error(), result.TransactionHash)
	require.Contains(t, err.Error(), result.Address)
	require.Equal(t, int32(1), sends.Load())
	info, statErr := os.Stat("contracts.json")
	require.NoError(t, statErr)
	require.True(t, info.IsDir())
}

func copyDeploymentInputs(t *testing.T) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	var err error
	artifact, err := os.ReadFile(filepath.Join(root, "out", "USDTRewardDistributor.sol", "USDTRewardDistributor.json"))
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join(root, "contracts", "USDTRewardDistributor.sol"))
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll("out/USDTRewardDistributor.sol", 0o755))
	require.NoError(t, os.MkdirAll("contracts", 0o755))
	require.NoError(t, os.WriteFile("out/USDTRewardDistributor.sol/USDTRewardDistributor.json", artifact, 0o600))
	require.NoError(t, os.WriteFile("contracts/USDTRewardDistributor.sol", source, 0o600))
	require.NoError(t, os.WriteFile("contracts.json", []byte("[]\n"), 0o600))
}

type deploymentRPCOptions struct {
	receipt         string
	code            string
	contractAddress common.Address
	sendError       bool
	afterSend       func()
	receipts        *atomic.Int32
	codeReads       *atomic.Int32
}

type explorerRoundTripper func(*http.Request) (*http.Response, error)

func (fn explorerRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func explorerResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func deploymentRPC(t *testing.T, sends *atomic.Int32, options *deploymentRPCOptions) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var call struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&call))
		result := any("0x0")
		var rpcError string
		switch call.Method {
		case "eth_chainId":
			result = "0xa4b1"
		case "eth_getBlockByNumber":
			result = map[string]any{"baseFeePerGas": "0x1", "difficulty": "0x1", "extraData": "0x", "gasLimit": "0x989680", "gasUsed": "0x1", "hash": "0x" + fmt.Sprintf("%064x", 1), "logsBloom": "0x" + fmt.Sprintf("%0512d", 0), "miner": common.Address{}.Hex(), "mixHash": "0x" + fmt.Sprintf("%064x", 1), "nonce": "0x0000000000000000", "number": "0x1", "parentHash": "0x" + fmt.Sprintf("%064x", 0), "receiptsRoot": "0x" + fmt.Sprintf("%064x", 1), "sha3Uncles": "0x" + fmt.Sprintf("%064x", 1), "size": "0x1", "stateRoot": "0x" + fmt.Sprintf("%064x", 1), "timestamp": "0x1", "totalDifficulty": "0x1", "transactions": []string{}, "transactionsRoot": "0x" + fmt.Sprintf("%064x", 1), "uncles": []string{}}
		case "eth_gasPrice", "eth_maxPriorityFeePerGas":
			result = "0x1"
		case "eth_feeHistory":
			result = map[string]any{"baseFeePerGas": []string{"0x1", "0x1"}, "gasUsedRatio": []float64{0}, "oldestBlock": "0x1", "reward": [][]string{{"0x1"}}}
		case "eth_getTransactionCount":
			result = "0x0"
		case "eth_estimateGas":
			result = "0x100000"
		case "eth_sendRawTransaction":
			sends.Add(1)
			if options.afterSend != nil {
				options.afterSend()
			}
			if options.sendError {
				rpcError = "send outcome is uncertain"
			} else {
				result = "0x" + fmt.Sprintf("%064x", 1)
			}
		case "eth_getTransactionReceipt":
			if options.receipts != nil {
				options.receipts.Add(1)
			}
			if options.receipt == "missing" {
				result = nil
			} else {
				result = map[string]any{"status": options.receipt, "contractAddress": options.contractAddress.Hex(), "blockHash": "0x" + fmt.Sprintf("%064x", 1), "blockNumber": "0x1", "transactionHash": "0x" + fmt.Sprintf("%064x", 1), "transactionIndex": "0x0", "from": common.Address{}.Hex(), "to": nil, "cumulativeGasUsed": "0x1", "gasUsed": "0x1", "logs": []any{}, "logsBloom": "0x" + fmt.Sprintf("%0512d", 0)}
			}
		case "eth_getCode":
			if options.codeReads != nil {
				options.codeReads.Add(1)
			}
			result = options.code
		}
		payload := map[string]any{"jsonrpc": "2.0", "id": call.ID}
		if rpcError != "" {
			payload["error"] = map[string]any{"code": -32000, "message": rpcError}
		} else {
			payload["result"] = result
		}
		require.NoError(t, json.NewEncoder(response).Encode(payload))
	})
}

func deploymentConfiguration(rpcURL string) *config.Config {
	configuration := &config.Config{RPC: map[string][]string{"arb": {rpcURL}}}
	configuration.Signing.PrivateKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	configuration.Explorer.APIKey = "key"
	return configuration
}

func deploymentInput() domain.DeploymentInput {
	return domain.DeploymentInput{Name: "test", Description: "test", FeeBps: 100, Token: common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9"), Owner: common.Address{1}, FeeWallet: common.Address{2}, Recipients: []domain.Recipient{{Address: common.Address{3}, ShareBps: 10_000}}}
}
