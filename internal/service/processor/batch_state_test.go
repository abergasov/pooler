package processor_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"pooler/internal/config"
	internal "pooler/internal/entities"
	"pooler/internal/service/processor"
)

func TestDistributeAllReportsTerminalStatesWithoutRetry(t *testing.T) {
	// given
	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	fixture := newBatchRPC(t, crypto.PubkeyToAddress(key.PublicKey))
	server := httptest.NewServer(fixture)
	defer server.Close()
	network, err := internal.ParseNetwork("arbitrum")
	require.NoError(t, err)
	contracts := []internal.Contract{
		{Name: "success", Network: network.Name, Address: batchAddress(1).Hex()},
		{Name: "uncertain", Network: network.Name, Address: batchAddress(2).Hex()},
		{Name: "empty", Network: network.Name, Address: batchAddress(3).Hex()},
		{Name: "success", Network: network.Name, Address: batchAddress(4).Hex()},
	}
	configuration := &config.Config{RPC: map[string][]string{"arb": {server.URL}}}
	configuration.Signing.PrivateKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	var submitted []string

	// when
	result, err := processor.New(configuration).DistributeAllWithSubmission(context.Background(), network, contracts, func(transaction processor.TransactionResult) {
		submitted = append(submitted, transaction.TransactionHash)
	})

	// then
	require.ErrorContains(t, err, "one or more distributions failed")
	require.False(t, result.NothingToProcess)
	require.Equal(t, contracts, contractsFromResults(result.Results))
	require.Equal(t, []string{"OK", "FAILED", "SKIPPED", "OK"}, statuses(result.Results))
	require.Equal(t, "success ("+contracts[0].Address+")", result.Results[0].Label)
	require.Equal(t, "success ("+contracts[3].Address+")", result.Results[3].Label)
	require.NotEmpty(t, result.Results[0].TransactionHash)
	require.NotEmpty(t, result.Results[1].TransactionHash)
	require.Empty(t, result.Results[2].TransactionHash)
	require.NotEmpty(t, result.Results[3].TransactionHash)
	require.Equal(t, []string{result.Results[0].TransactionHash, result.Results[3].TransactionHash}, submitted)
	require.Equal(t, 3, fixture.sendCount())
	require.Empty(t, fixture.unexpectedCalls())
}

func TestShowReadsRecipientsAfterAnExternalChange(t *testing.T) {
	// given
	fixture := newBatchRPC(t, batchAddress(1))
	fixture.setRecipients([]batchRecipient{{account: batchAddress(10), shareBps: 10_000}})
	server := httptest.NewServer(fixture)
	defer server.Close()
	network, err := internal.ParseNetwork("arbitrum")
	require.NoError(t, err)
	service := processor.New(&config.Config{RPC: map[string][]string{"arb": {server.URL}}})
	contract := batchAddress(1)

	initial, err := service.Show(context.Background(), network, contract)
	require.NoError(t, err)

	// when
	fixture.setRecipients([]batchRecipient{{account: batchAddress(11), shareBps: 4_000}, {account: batchAddress(12), shareBps: 6_000}})
	updated, err := service.Show(context.Background(), network, contract)

	// then
	require.NoError(t, err)
	require.Equal(t, []internal.Recipient{{Address: batchAddress(10), ShareBps: 10_000}}, initial.Recipients)
	require.Equal(t, []internal.Recipient{{Address: batchAddress(11), ShareBps: 4_000}, {Address: batchAddress(12), ShareBps: 6_000}}, updated.Recipients)
	require.Empty(t, fixture.unexpectedCalls())
}

func TestReplaceRecipientsReportsConfirmedRevertedAndUncertainOutcomes(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		receiptStatus uint64
		uncertainSend int
		confirmed     bool
		errorText     string
		submitted     bool
	}{
		{name: "confirmed", receiptStatus: types.ReceiptStatusSuccessful, uncertainSend: -1, confirmed: true, submitted: true},
		{name: "reverted", receiptStatus: types.ReceiptStatusFailed, uncertainSend: -1, errorText: "failed", submitted: true},
		{name: "uncertain", receiptStatus: types.ReceiptStatusSuccessful, uncertainSend: 1, errorText: "confirmation is uncertain"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// given
			key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
			require.NoError(t, err)
			fixture := newBatchRPC(t, crypto.PubkeyToAddress(key.PublicKey))
			fixture.receiptStatus = testCase.receiptStatus
			fixture.uncertainSend = testCase.uncertainSend
			server := httptest.NewServer(fixture)
			defer server.Close()
			network, err := internal.ParseNetwork("arbitrum")
			require.NoError(t, err)
			configuration := &config.Config{RPC: map[string][]string{"arb": {server.URL}}}
			configuration.Signing.PrivateKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
			var submitted []string

			// when
			result, err := processor.New(configuration).ReplaceRecipientsWithSubmission(context.Background(), network, batchAddress(1), internal.DistributionInput{Recipients: []internal.Recipient{{Address: batchAddress(10), ShareBps: 10_000}}}, func(transaction processor.TransactionResult) {
				submitted = append(submitted, transaction.TransactionHash)
			})

			// then
			require.NotEmpty(t, result.TransactionHash)
			require.Equal(t, testCase.confirmed, result.Confirmed)
			if testCase.errorText == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, testCase.errorText)
			}
			if testCase.submitted {
				require.Equal(t, []string{result.TransactionHash}, submitted)
			} else {
				require.Empty(t, submitted)
			}
			require.Equal(t, []string{"0xa6e845d1"}, fixture.submittedMethods())
			require.Empty(t, fixture.unexpectedCalls())
		})
	}
}

func contractsFromResults(results []processor.DistributeAllContractResult) []internal.Contract {
	contracts := make([]internal.Contract, len(results))
	for index := range results {
		contracts[index] = results[index].Contract
	}
	return contracts
}

func statuses(results []processor.DistributeAllContractResult) []string {
	statuses := make([]string, len(results))
	for index := range results {
		statuses[index] = results[index].Status
	}
	return statuses
}

func batchAddress(value byte) common.Address { return common.BytesToAddress([]byte{value}) }

type batchRPC struct {
	t                  *testing.T
	owner              common.Address
	token              common.Address
	recipients         []batchRecipient
	uncertainSend      int
	receiptStatus      uint64
	mu                 sync.Mutex
	sends              int
	nonces             uint64
	submittedSelectors []string
	unexpected         []string
}

func newBatchRPC(t *testing.T, owner common.Address) *batchRPC {
	t.Helper()
	return &batchRPC{t: t, owner: owner, token: batchAddress(20), uncertainSend: 2, receiptStatus: types.ReceiptStatusSuccessful}
}

func (rpc *batchRPC) ServeHTTP(response http.ResponseWriter, request *http.Request) {
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

func (rpc *batchRPC) result(method string, params []json.RawMessage) (result any, rpcErr string) {
	switch method {
	case "eth_chainId":
		return "0xa4b1", ""
	case "eth_call":
		return rpc.callResult(params)
	case "eth_getBlockByNumber":
		return rpc.header(), ""
	case "eth_gasPrice", "eth_maxPriorityFeePerGas":
		return "0x1", ""
	case "eth_feeHistory":
		return map[string]any{"baseFeePerGas": []string{"0x1", "0x1"}, "gasUsedRatio": []float64{0}, "oldestBlock": "0x1", "reward": [][]string{{"0x1"}}}, ""
	case "eth_getCode":
		return "0x6000", ""
	case "eth_getTransactionCount":
		rpc.mu.Lock()
		nonce := rpc.nonces
		rpc.nonces++
		rpc.mu.Unlock()
		return fmt.Sprintf("0x%x", nonce), ""
	case "eth_estimateGas":
		return "0x100000", ""
	case "eth_sendRawTransaction":
		method, err := submittedMethod(params)
		if err != nil {
			return nil, err.Error()
		}
		rpc.mu.Lock()
		rpc.sends++
		send := rpc.sends
		rpc.submittedSelectors = append(rpc.submittedSelectors, method)
		uncertainSend := rpc.uncertainSend
		rpc.mu.Unlock()
		if send == uncertainSend {
			return nil, "send outcome is uncertain"
		}
		return "0x" + fmt.Sprintf("%064x", send), ""
	case "eth_getTransactionReceipt":
		return rpc.receipt(), ""
	default:
		rpc.mu.Lock()
		rpc.unexpected = append(rpc.unexpected, method)
		rpc.mu.Unlock()
		return nil, "unexpected RPC method " + method
	}
}

func (rpc *batchRPC) callResult(params []json.RawMessage) (result any, rpcErr string) {
	if len(params) == 0 {
		return nil, "eth_call has no parameters"
	}
	var call struct {
		To    string `json:"to"`
		Input string `json:"input"`
		Data  string `json:"data"`
	}
	if err := json.Unmarshal(params[0], &call); err != nil {
		return nil, "decode eth_call: " + err.Error()
	}
	input := call.Input
	if input == "" {
		input = call.Data
	}
	if len(input) < 10 {
		return nil, "eth_call missing selector"
	}
	selector := input[:10]
	switch selector {
	case "0x8da5cb5b":
		rpc.mu.Lock()
		owner := rpc.owner
		rpc.mu.Unlock()
		return abiWord(owner), ""
	case "0xf25f4b56":
		return abiWord(common.Address{}), ""
	case "0x24a9d853":
		return abiWord(big.NewInt(0)), ""
	case "0xd78d610b":
		rpc.mu.Lock()
		recipients := append([]batchRecipient(nil), rpc.recipients...)
		rpc.mu.Unlock()
		return recipientsResponse(recipients), ""
	case "0x2f48ab7d":
		return abiWord(rpc.token), ""
	case "0x70a08231":
		contract := common.HexToAddress("0x" + input[len(input)-40:])
		if contract == batchAddress(3) {
			return abiWord(big.NewInt(0)), ""
		}
		return abiWord(big.NewInt(1)), ""
	default:
		return nil, "unexpected eth_call selector " + selector
	}
}

func (rpc *batchRPC) header() map[string]any {
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

func (rpc *batchRPC) receipt() map[string]any {
	rpc.mu.Lock()
	status := rpc.receiptStatus
	rpc.mu.Unlock()
	return map[string]any{
		"status": fmt.Sprintf("0x%x", status), "blockHash": "0x" + strings.Repeat("0", 63) + "1", "blockNumber": "0x1",
		"transactionHash": "0x" + strings.Repeat("0", 63) + "1", "transactionIndex": "0x0", "from": common.Address{}.Hex(),
		"to": batchAddress(1).Hex(), "cumulativeGasUsed": "0x1", "gasUsed": "0x1", "logs": []any{}, "logsBloom": "0x" + strings.Repeat("0", 512),
	}
}

type batchRecipient struct {
	account  common.Address
	shareBps uint16
}

func (rpc *batchRPC) setRecipients(recipients []batchRecipient) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.recipients = append([]batchRecipient(nil), recipients...)
}

func recipientsResponse(recipients []batchRecipient) string {
	response := make([]byte, 64+64*len(recipients))
	response[31] = 32
	new(big.Int).SetUint64(uint64(len(recipients))).FillBytes(response[32:64])
	for index, recipient := range recipients {
		offset := 64 + index*64
		copy(response[offset+12:offset+32], recipient.account[:])
		new(big.Int).SetUint64(uint64(recipient.shareBps)).FillBytes(response[offset+32 : offset+64])
	}
	return "0x" + hex.EncodeToString(response)
}

func submittedMethod(params []json.RawMessage) (string, error) {
	if len(params) == 0 {
		return "", errors.New("eth_sendRawTransaction has no transaction")
	}
	var raw string
	if err := json.Unmarshal(params[0], &raw); err != nil {
		return "", fmt.Errorf("decode submitted transaction: %w", err)
	}
	encoded, err := hex.DecodeString(strings.TrimPrefix(raw, "0x"))
	if err != nil {
		return "", fmt.Errorf("decode submitted transaction: %w", err)
	}
	transaction := new(types.Transaction)
	if err := transaction.UnmarshalBinary(encoded); err != nil {
		return "", fmt.Errorf("decode submitted transaction: %w", err)
	}
	data := transaction.Data()
	if len(data) < 4 {
		return "", errors.New("submitted transaction has no method selector")
	}
	return "0x" + hex.EncodeToString(data[:4]), nil
}

func abiWord(value any) string {
	word := make([]byte, 32)
	switch value := value.(type) {
	case common.Address:
		copy(word[12:], value[:])
	case *big.Int:
		value.FillBytes(word)
	}
	return "0x" + hex.EncodeToString(word)
}

func (rpc *batchRPC) sendCount() int {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	return rpc.sends
}

func (rpc *batchRPC) unexpectedCalls() []string {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	return append([]string(nil), rpc.unexpected...)
}

func (rpc *batchRPC) submittedMethods() []string {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	return append([]string(nil), rpc.submittedSelectors...)
}
