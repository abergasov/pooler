package entities_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"pooler/internal/entities"
)

func TestParseAddressRequiresPrefixedTwentyByteHex(t *testing.T) {
	// given
	valid := "0x0000000000000000000000000000000000000001"

	// when
	address, err := entities.ParseAddress(valid)

	// then
	require.NoError(t, err)
	require.Equal(t, common.HexToAddress(valid), address)

	for _, value := range []string{"0000000000000000000000000000000000000001", "0x000000000000000000000000000000000000000g", "0x0001"} {
		_, err := entities.ParseAddress(value)
		require.ErrorContains(t, err, "invalid address")
	}
}
