package operator_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	internal "pooler/internal/entities"
	"pooler/internal/utils"
)

func TestParseNetworkCanonicalizesSupportedNames(t *testing.T) {
	tests := map[string]struct {
		input   string
		name    string
		rpcKey  string
		chainID uint64
	}{
		"ethereum": {input: "ethereum", name: "ethereum", rpcKey: "eth", chainID: 1},
		"arbitrum": {input: " ARBITRUM ", name: "arbitrum", rpcKey: "arb", chainID: 42161},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			network, err := internal.ParseNetwork(test.input)

			// then
			require.NoError(t, err)
			require.Equal(t, test.name, network.Name)
			require.Equal(t, test.rpcKey, network.RPCKey)
			require.Equal(t, test.chainID, network.ChainID)
		})
	}

	// when
	_, err := internal.ParseNetwork("sepolia")

	// then
	require.ErrorContains(t, err, "unsupported network")
}

func TestSelectRPCSkipsUnreachableAndWrongChainThenStops(t *testing.T) {
	// given
	unreachable := httptest.NewServer(http.NotFoundHandler())
	unreachableURL := unreachable.URL
	unreachable.Close()

	var calls []string
	wrongChain := chainIDServer(t, "wrong", "0xa4b1", &calls)
	defer wrongChain.Close()
	working := chainIDServer(t, "working", "0x1", &calls)
	defer working.Close()
	afterSuccess := chainIDServer(t, "after-success", "0x1", &calls)
	defer afterSuccess.Close()
	network, err := internal.ParseNetwork("ethereum")
	require.NoError(t, err)

	// when
	client, err := utils.SelectRPC(
		context.Background(),
		network,
		[]string{unreachableURL, wrongChain.URL, working.URL, afterSuccess.URL},
		time.Second,
	)

	// then
	require.NoError(t, err)
	client.Close()
	require.Equal(t, []string{"wrong", "working"}, calls)
}

func TestSelectRPCRejectsEmptyAndUnusableLists(t *testing.T) {
	// given
	network, err := internal.ParseNetwork("arbitrum")
	require.NoError(t, err)
	wrongChain := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`))
	}))
	defer wrongChain.Close()

	// when
	_, emptyErr := utils.SelectRPC(context.Background(), network, nil, time.Second)
	_, unusableErr := utils.SelectRPC(context.Background(), network, []string{wrongChain.URL}, time.Second)

	// then
	require.ErrorContains(t, emptyErr, "no RPC endpoints configured")
	require.ErrorContains(t, unusableErr, "no usable RPC endpoint")
	require.NotContains(t, unusableErr.Error(), wrongChain.URL)
}

func chainIDServer(t *testing.T, name, chainID string, calls *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		*calls = append(*calls, name)
		require.Equal(t, http.MethodPost, request.Method)
		writer.Header().Set("Content-Type", "application/json")
		_, err := writer.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + chainID + `"}`))
		require.NoError(t, err)
	}))
}
