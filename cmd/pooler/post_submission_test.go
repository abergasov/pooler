package main_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestCLIDeployReportsConfirmedRegistryFailureAfterSubmission(t *testing.T) {
	// given
	workdir := t.TempDir()
	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	address := crypto.CreateAddress(crypto.PubkeyToAddress(key.PublicKey), 0)
	fixture := newPostSubmissionRPC(t, workdir, address)
	server := httptest.NewServer(fixture)
	defer server.Close()
	input := writePostSubmissionDeploymentFiles(t, workdir, server.URL)
	binary := filepath.Join(workdir, "pooler")
	build := exec.Command("go", "build", "-o", binary, ".")
	require.NoError(t, build.Run())
	command := exec.Command(binary, "deploy", "--network", "arbitrum", "--input", input)
	command.Dir = workdir

	// when
	output, err := command.CombinedOutput()

	// then
	require.Error(t, err)
	text := string(output)
	require.Contains(t, text, "deployment submitted: "+fixture.transactionHash())
	require.Contains(t, text, "failed after submission")
	require.Contains(t, text, address.Hex())
	require.Contains(t, text, fixture.transactionHash())
	require.NotContains(t, text, "preflight failed")
	require.Equal(t, int32(1), fixture.sends.Load())
}

func writePostSubmissionDeploymentFiles(t *testing.T, workdir, rpcURL string) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", ".."))
	artifact, err := os.ReadFile(filepath.Join(root, "out", "USDTRewardDistributor.sol", "USDTRewardDistributor.json"))
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join(root, "contracts", "USDTRewardDistributor.sol"))
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(workdir, "out", "USDTRewardDistributor.sol"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(workdir, "contracts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workdir, "out", "USDTRewardDistributor.sol", "USDTRewardDistributor.json"), artifact, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workdir, "contracts", "USDTRewardDistributor.sol"), source, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(workdir, "config.yml"), []byte("rpc:\n  eth:\n    - "+rpcURL+"\n  arb:\n    - "+rpcURL+"\nsigning:\n  privateKey: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\nexplorer:\n  apiKey: key\n"), 0o600))
	input := filepath.Join(workdir, "deployment.json")
	require.NoError(t, os.WriteFile(input, []byte(`{
  "name": "Reward pool",
  "description": "Quarterly rewards",
  "owner": "0x0000000000000000000000000000000000000001",
  "feeWallet": "0x0000000000000000000000000000000000000002",
  "feeBps": 0,
  "token": "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9",
  "recipients": [{"address": "0x0000000000000000000000000000000000000004", "shareBps": 10000}]
}`), 0o600))
	return input
}

type postSubmissionRPC struct {
	t               *testing.T
	workdir         string
	contractAddress common.Address
	sends           atomic.Int32
	mu              sync.Mutex
	hash            string
}

func newPostSubmissionRPC(t *testing.T, workdir string, contractAddress common.Address) *postSubmissionRPC {
	t.Helper()
	return &postSubmissionRPC{t: t, workdir: workdir, contractAddress: contractAddress}
}

func (rpc *postSubmissionRPC) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	var call struct {
		ID     json.RawMessage   `json:"id"`
		Method string            `json:"method"`
		Params []json.RawMessage `json:"params"`
	}
	require.NoError(rpc.t, json.NewDecoder(request.Body).Decode(&call))
	result, rpcErr := rpc.result(call.Method, call.Params)
	payload := map[string]any{"jsonrpc": "2.0", "id": call.ID}
	if rpcErr != "" {
		payload["error"] = map[string]any{"code": -32000, "message": rpcErr}
	} else {
		payload["result"] = result
	}
	require.NoError(rpc.t, json.NewEncoder(response).Encode(payload))
}

func (rpc *postSubmissionRPC) result(method string, params []json.RawMessage) (result any, rpcErr string) {
	switch method {
	case "eth_chainId":
		return "0xa4b1", ""
	case "eth_getBlockByNumber":
		return postSubmissionHeader(), ""
	case "eth_gasPrice", "eth_maxPriorityFeePerGas":
		return "0x1", ""
	case "eth_feeHistory":
		return map[string]any{"baseFeePerGas": []string{"0x1", "0x1"}, "gasUsedRatio": []float64{0}, "oldestBlock": "0x1", "reward": [][]string{{"0x1"}}}, ""
	case "eth_getTransactionCount":
		return "0x0", ""
	case "eth_estimateGas":
		return "0x100000", ""
	case "eth_getCode":
		return "0x6000", ""
	case "eth_sendRawTransaction":
		if len(params) != 1 {
			return nil, "eth_sendRawTransaction has invalid parameters"
		}
		var encoded string
		if err := json.Unmarshal(params[0], &encoded); err != nil {
			return nil, "decode submitted transaction: " + err.Error()
		}
		var transaction types.Transaction
		if err := transaction.UnmarshalBinary(common.FromHex(encoded)); err != nil {
			return nil, "decode submitted transaction: " + err.Error()
		}
		hash := transaction.Hash().Hex()
		rpc.mu.Lock()
		rpc.hash = hash
		rpc.mu.Unlock()
		rpc.sends.Add(1)
		require.NoError(rpc.t, os.Mkdir(filepath.Join(rpc.workdir, "contracts.json"), 0o700))
		return hash, ""
	case "eth_getTransactionReceipt":
		return rpc.receipt(), ""
	default:
		return nil, "unexpected RPC method " + method
	}
}

func postSubmissionHeader() map[string]any {
	zeroHash := "0x" + strings.Repeat("0", 64)
	oneHash := "0x" + strings.Repeat("0", 63) + "1"
	return map[string]any{
		"baseFeePerGas": "0x1", "difficulty": "0x1", "extraData": "0x", "gasLimit": "0x989680", "gasUsed": "0x1",
		"hash": oneHash, "logsBloom": "0x" + strings.Repeat("0", 512), "miner": common.Address{}.Hex(), "mixHash": oneHash,
		"nonce": "0x0000000000000000", "number": "0x1", "parentHash": zeroHash, "receiptsRoot": oneHash,
		"sha3Uncles": oneHash, "size": "0x1", "stateRoot": oneHash, "timestamp": "0x1", "totalDifficulty": "0x1",
		"transactions": []string{}, "transactionsRoot": oneHash, "uncles": []string{},
	}
}

func (rpc *postSubmissionRPC) receipt() map[string]any {
	return map[string]any{
		"status": "0x1", "blockHash": "0x" + strings.Repeat("0", 63) + "1", "blockNumber": "0x1",
		"transactionHash": rpc.transactionHash(), "transactionIndex": "0x0", "from": common.Address{}.Hex(), "to": nil,
		"contractAddress": rpc.contractAddress.Hex(), "cumulativeGasUsed": "0x1", "gasUsed": "0x1", "logs": []any{},
		"logsBloom": "0x" + strings.Repeat("0", 512),
	}
}

func (rpc *postSubmissionRPC) transactionHash() string {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	return rpc.hash
}
