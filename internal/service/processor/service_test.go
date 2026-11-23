package processor

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	internal "pooler/internal/entities"
)

func TestShowReadsCurrentOnChainConfiguration(t *testing.T) {
	contract := common.HexToAddress("0x0000000000000000000000000000000000001000")
	token := common.HexToAddress("0x0000000000000000000000000000000000000020")
	backend := showBackend{responses: map[string][]byte{
		showCallKey(contract, "owner()"):         showAddressResponse(common.HexToAddress("0x0000000000000000000000000000000000000010")),
		showCallKey(contract, "feeWallet()"):     showAddressResponse(common.HexToAddress("0x0000000000000000000000000000000000000030")),
		showCallKey(contract, "feeBps()"):        showUintResponse(250),
		showCallKey(contract, "usdt()"):          showAddressResponse(token),
		showCallKey(contract, "getRecipients()"): showRecipientsResponse([]showRecipient{{common.HexToAddress("0x0000000000000000000000000000000000000040"), 4_000}, {common.HexToAddress("0x0000000000000000000000000000000000000050"), 6_000}}),
		showCallKey(token, "balanceOf(address)"): showUintResponse(123_456_789),
	}}

	result, err := show(context.Background(), contract, backend)

	require.NoError(t, err)
	require.Equal(t, "0x0000000000000000000000000000000000000010", result.Owner.String())
	require.Equal(t, "0x0000000000000000000000000000000000000030", result.FeeWallet.String())
	require.Equal(t, uint16(250), result.FeeBps)
	require.Equal(t, token.Hex(), result.Token.String())
	require.Equal(t, "123456789", result.Balance.String())
	require.Equal(t, []internal.Recipient{{Address: common.HexToAddress("0x0000000000000000000000000000000000000040"), ShareBps: 4_000}, {Address: common.HexToAddress("0x0000000000000000000000000000000000000050"), ShareBps: 6_000}}, result.Recipients)
}

func TestDistributeAllSkipsEmptyNetworkWithoutConfiguration(t *testing.T) {
	result, err := New(nil).DistributeAll(context.Background(), internal.Network{Name: "arbitrum"}, []internal.Contract{{Network: "ethereum", Address: "0x0000000000000000000000000000000000000001"}})

	require.NoError(t, err)
	require.True(t, result.NothingToProcess)
	require.Empty(t, result.Results)
}

type showBackend struct{ responses map[string][]byte }

func (showBackend) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{1}, nil
}

//nolint:gocritic // bind.ContractCaller requires ethereum.CallMsg by value.
func (backend showBackend) CallContract(_ context.Context, call ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	if call.To == nil || len(call.Data) < 4 {
		return nil, errors.New("unexpected contract call")
	}
	response, ok := backend.responses[call.To.Hex()+":"+hex.EncodeToString(call.Data[:4])]
	if !ok {
		return nil, errors.New("unexpected contract call")
	}
	return response, nil
}

func showCallKey(address common.Address, signature string) string {
	return address.Hex() + ":" + hex.EncodeToString(crypto.Keccak256([]byte(signature))[:4])
}

func showAddressResponse(address common.Address) []byte {
	response := make([]byte, 32)
	copy(response[12:], address[:])
	return response
}

func showUintResponse(value uint64) []byte {
	response := make([]byte, 32)
	new(big.Int).SetUint64(value).FillBytes(response)
	return response
}

type showRecipient struct {
	account  common.Address
	shareBps uint16
}

func showRecipientsResponse(recipients []showRecipient) []byte {
	response := make([]byte, 64+64*len(recipients))
	response[31] = 32
	new(big.Int).SetUint64(uint64(len(recipients))).FillBytes(response[32:64])
	for index, recipient := range recipients {
		offset := 64 + index*64
		copy(response[offset+12:offset+32], recipient.account[:])
		new(big.Int).SetUint64(uint64(recipient.shareBps)).FillBytes(response[offset+32 : offset+64])
	}
	return response
}
