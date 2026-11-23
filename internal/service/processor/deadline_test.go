package processor_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"pooler/internal/config"
	internal "pooler/internal/entities"
	"pooler/internal/service/processor"
)

func TestShowCancelsBlockedCallWhenReadContextExpires(t *testing.T) {
	// given
	canceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var rpcRequest struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&rpcRequest))
		if rpcRequest.Method == "eth_chainId" {
			_, _ = response.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xa4b1"}`))
			return
		}
		<-request.Context().Done()
		close(canceled)
	}))
	defer server.Close()

	configuration := &config.Config{RPC: map[string][]string{"arb": {server.URL}}}
	network, err := internal.ParseNetwork("arbitrum")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// when
	_, err = processor.New(configuration).Show(ctx, network, common.HexToAddress("0x0000000000000000000000000000000000000001"))

	// then
	require.Error(t, err)
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("blocked RPC call was not canceled")
	}
}
