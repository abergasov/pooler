package deployer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"pooler/internal/config"
	domain "pooler/internal/entities"
)

const verificationAddress = "0x0000000000000000000000000000000000000005"

func TestVerifySubmitsCanonicalChainAndPollsForSuccess(t *testing.T) {
	var submits atomic.Int32
	var polls atomic.Int32
	client := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		switch request.Method {
		case http.MethodPost:
			require.NoError(t, request.ParseForm())
			require.Equal(t, "42161", request.Form.Get("chainid"))
			require.Equal(t, "explorer-key", request.Form.Get("apikey"))
			submits.Add(1)
			return explorerHTTPResponse(`{"status":"1","result":"guid-1"}`), nil
		case http.MethodGet:
			require.Equal(t, "42161", request.URL.Query().Get("chainid"))
			require.Equal(t, "guid-1", request.URL.Query().Get("guid"))
			polls.Add(1)
			return explorerHTTPResponse(`{"status":"1","result":"Pass - Verified"}`), nil
		default:
			return explorerHTTPResponse(`{"status":"0","result":"unexpected method"}`), nil
		}
	})}
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)

	result, err := New(&config.Config{Explorer: struct {
		APIKey string `yaml:"apiKey"`
	}{APIKey: "explorer-key"}}).verify(context.Background(), &verificationRequest{
		network: network, address: verificationAddress, metadata: verificationMetadataPtr(t),
		httpClient: client, explorerURL: "https://explorer.example/v2/api", pollInterval: time.Millisecond, timeout: time.Second,
	})

	require.NoError(t, err)
	require.Equal(t, "guid-1", result.SubmissionID)
	require.Equal(t, int32(1), submits.Load())
	require.Equal(t, int32(1), polls.Load())
}

func testVerificationMetadata() VerificationMetadata {
	return VerificationMetadata{
		Network: "arbitrum", Address: verificationAddress, ContractName: "contracts/Pool.sol:Pool", CompilerVersion: "0.8.28+commit.7893614a",
		CompilerInput:        json.RawMessage(`{"language":"Solidity","sources":{"contracts/Pool.sol":{"content":"contract Pool {}"}},"settings":{"optimizer":{"enabled":false}}}`),
		ConstructorArguments: "0000000000000000000000000000000000000000000000000000000000000000",
	}
}

func verificationMetadataPtr(t *testing.T) *VerificationMetadata {
	t.Helper()
	metadata := testVerificationMetadata()
	return &metadata
}

type roundTripper func(*http.Request) (*http.Response, error)

func (fn roundTripper) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

func explorerHTTPResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}
