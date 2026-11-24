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

func TestVerifyReturnsFailureAfterExplorerStatus(t *testing.T) {
	// given
	var submits atomic.Int32
	var polls atomic.Int32
	client := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		switch request.Method {
		case http.MethodPost:
			submits.Add(1)
			return explorerHTTPResponse(`{"status":"1","result":"guid-1"}`), nil
		case http.MethodGet:
			polls.Add(1)
			return explorerHTTPResponse(`{"status":"0","result":"Fail - Unable to verify"}`), nil
		default:
			return explorerHTTPResponse(`{"status":"0","result":"unexpected method"}`), nil
		}
	})}
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)

	// when
	result, err := New(&config.Config{Explorer: struct {
		APIKey string `yaml:"apiKey"`
	}{APIKey: "explorer-key"}}).verify(context.Background(), &verificationRequest{
		network: network, address: verificationAddress, metadata: verificationMetadataPtr(t),
		httpClient: client, explorerURL: "https://explorer.example/v2/api", pollInterval: time.Millisecond, timeout: time.Second,
	})

	// then
	require.ErrorContains(t, err, "explorer verification failed")
	require.Equal(t, "guid-1", result.SubmissionID)
	require.Equal(t, int32(1), submits.Load())
	require.Equal(t, int32(1), polls.Load())
}

func TestVerifyReturnsFailureWhenExplorerRemainsPending(t *testing.T) {
	// given
	var polls atomic.Int32
	client := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		if request.Method == http.MethodPost {
			return explorerHTTPResponse(`{"status":"1","result":"guid-1"}`), nil
		}
		polls.Add(1)
		return explorerHTTPResponse(`{"status":"0","result":"Pending in queue"}`), nil
	})}
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)

	// when
	result, err := New(&config.Config{Explorer: struct {
		APIKey string `yaml:"apiKey"`
	}{APIKey: "explorer-key"}}).verify(context.Background(), &verificationRequest{
		network: network, address: verificationAddress, metadata: verificationMetadataPtr(t),
		httpClient: client, explorerURL: "https://explorer.example/v2/api", pollInterval: time.Millisecond, timeout: 50 * time.Millisecond,
	})

	// then
	require.ErrorContains(t, err, "remained pending until timeout")
	require.Equal(t, "guid-1", result.SubmissionID)
	require.Positive(t, polls.Load())
}

func TestVerifyTreatsAlreadyVerifiedContractAsSuccess(t *testing.T) {
	// given
	var requests atomic.Int32
	client := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		return explorerHTTPResponse(`{"status":"0","result":"Contract source code already verified"}`), nil
	})}
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)

	// when
	result, err := New(&config.Config{Explorer: struct {
		APIKey string `yaml:"apiKey"`
	}{APIKey: "explorer-key"}}).verify(context.Background(), &verificationRequest{
		network: network, address: verificationAddress, metadata: verificationMetadataPtr(t),
		httpClient: client, explorerURL: "https://explorer.example/v2/api", pollInterval: time.Millisecond, timeout: time.Second,
	})

	// then
	require.NoError(t, err)
	require.True(t, result.AlreadyKnown)
	require.Empty(t, result.SubmissionID)
	require.Equal(t, int32(1), requests.Load())
}

func TestVerifyRejectsWrongChainMetadataBeforeExplorerSubmission(t *testing.T) {
	// given
	var requests atomic.Int32
	client := &http.Client{Transport: roundTripper(func(request *http.Request) (*http.Response, error) {
		requests.Add(1)
		return explorerHTTPResponse(`{"status":"1","result":"guid-1"}`), nil
	})}
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)
	metadata := testVerificationMetadata()
	metadata.Network = "ethereum"

	// when
	_, err = New(&config.Config{Explorer: struct {
		APIKey string `yaml:"apiKey"`
	}{APIKey: "explorer-key"}}).verify(context.Background(), &verificationRequest{
		network: network, address: verificationAddress, metadata: &metadata,
		httpClient: client, explorerURL: "https://explorer.example/v2/api", pollInterval: time.Millisecond, timeout: time.Second,
	})

	// then
	require.ErrorContains(t, err, "does not match the selected network and contract")
	require.Zero(t, requests.Load())
}

func TestVerifyWithoutSavedDeploymentDoesNotContactExplorer(t *testing.T) {
	// given
	t.Chdir(t.TempDir())
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)

	// when
	_, err = New(&config.Config{}).Verify(context.Background(), network, verificationAddress)

	// then
	require.ErrorContains(t, err, "read verification metadata")
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
