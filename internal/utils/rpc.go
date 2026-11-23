// Package utils contains shared infrastructure helpers.
package utils

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	internal "pooler/internal/entities"
)

// SelectRPC dials endpoints in order and returns the first client on network's chain.
func SelectRPC(ctx context.Context, network internal.Network, endpoints []string, timeout time.Duration) (*ethclient.Client, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no RPC endpoints configured for %s", network.Name)
	}
	if timeout <= 0 {
		return nil, errors.New("RPC probe timeout must be positive")
	}
	expected := new(big.Int).SetUint64(network.ChainID)
	for _, endpoint := range endpoints {
		probe, cancel := context.WithTimeout(ctx, timeout)
		client, err := ethclient.DialContext(probe, endpoint)
		if err == nil {
			var chainID *big.Int
			chainID, err = client.ChainID(probe)
			if err == nil && chainID.Cmp(expected) == 0 {
				cancel()
				return client, nil
			}
			client.Close()
		}
		cancel()
	}
	return nil, fmt.Errorf("no usable RPC endpoint for %s", network.Name)
}
