// Package entities contains Pooler's shared domain types.
package entities

import (
	"fmt"
	"strings"
)

const (
	EthereumNetwork = "ethereum"
	ArbitrumNetwork = "arbitrum"
	EthereumRPCKey  = "eth"
	ArbitrumRPCKey  = "arb"
)

// Network is a supported EVM network and its RPC configuration key.
type Network struct {
	Name    string
	RPCKey  string
	ChainID uint64
}

// ParseNetwork normalizes a CLI network name.
func ParseNetwork(value string) (Network, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case EthereumNetwork:
		return Network{Name: EthereumNetwork, RPCKey: EthereumRPCKey, ChainID: 1}, nil
	case ArbitrumNetwork:
		return Network{Name: ArbitrumNetwork, RPCKey: ArbitrumRPCKey, ChainID: 42161}, nil
	default:
		return Network{}, fmt.Errorf("unsupported network %q", value)
	}
}
