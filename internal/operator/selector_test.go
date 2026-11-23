package operator_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/term"

	internal "pooler/internal/entities"
	"pooler/internal/operator"
)

const (
	ethereumContract = "0x1111111111111111111111111111111111111111"
	arbitrumContract = "0x2222222222222222222222222222222222222222"
)

type contractSelector struct {
	selected   internal.Contract
	err        error
	candidates []internal.Contract
}

func (selector *contractSelector) SelectContract(candidates []internal.Contract) (internal.Contract, error) {
	selector.candidates = append([]internal.Contract(nil), candidates...)
	return selector.selected, selector.err
}

func TestResolveContractUsesExplicitAddressWithoutRegistration(t *testing.T) {
	// given
	network, err := internal.ParseNetwork("ethereum")
	require.NoError(t, err)

	// when
	address, err := operator.ResolveContract(network, "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", nil, nil)

	// then
	require.NoError(t, err)
	require.Equal(t, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", address)
}

func TestResolveContractFiltersContractsByNetworkBeforeSelecting(t *testing.T) {
	// given
	network, err := internal.ParseNetwork("arbitrum")
	require.NoError(t, err)
	selected := internal.Contract{Name: "Arbitrum pool", Description: "eligible", Network: "arbitrum", Address: arbitrumContract}
	selector := &contractSelector{selected: selected}

	// when
	address, err := operator.ResolveContract(network, "", []internal.Contract{
		{Name: "Ethereum pool", Description: "ineligible", Network: "ethereum", Address: ethereumContract},
		selected,
	}, selector)

	// then
	require.NoError(t, err)
	require.Equal(t, arbitrumContract, address)
	require.Equal(t, []internal.Contract{selected}, selector.candidates)
}

func TestResolveContractRejectsMissingOrNonInteractiveSelection(t *testing.T) {
	// given
	network, err := internal.ParseNetwork("ethereum")
	require.NoError(t, err)
	contracts := []internal.Contract{{Network: "arbitrum", Address: arbitrumContract}}

	// when
	_, emptyErr := operator.ResolveContract(network, "", contracts, nil)
	_, terminalErr := operator.ResolveContract(network, "", []internal.Contract{{Network: "ethereum", Address: ethereumContract}}, nil)

	// then
	require.ErrorContains(t, emptyErr, "no registered contracts")
	require.ErrorContains(t, emptyErr, "CONTRACT=<address>")
	require.ErrorContains(t, terminalErr, "interactive terminal")
	require.ErrorContains(t, terminalErr, "CONTRACT=<address>")
}

func TestResolveContractPropagatesCancellationWithoutSelectingAContract(t *testing.T) {
	// given
	network, err := internal.ParseNetwork("ethereum")
	require.NoError(t, err)
	selector := &contractSelector{err: operator.ErrContractSelectionCancelled}

	// when
	address, err := operator.ResolveContract(network, "", []internal.Contract{{Network: "ethereum", Address: ethereumContract}}, selector)

	// then
	require.Empty(t, address)
	require.ErrorIs(t, err, operator.ErrContractSelectionCancelled)
}

func TestTerminalContractSelectorRejectsNoninteractiveInput(t *testing.T) {
	// given
	input, err := os.CreateTemp(t.TempDir(), "input")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, input.Close()) })
	var output bytes.Buffer
	selector := operator.TerminalContractSelector{Input: input, Output: &output}

	// when
	_, err = selector.SelectContract([]internal.Contract{{Network: "ethereum", Address: ethereumContract}})

	// then
	require.ErrorContains(t, err, "interactive terminal")
	require.ErrorContains(t, err, "CONTRACT=<address>")
}

func TestTerminalContractSelectorPTYSelectsWithArrowsAndEnter(t *testing.T) {
	// given
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("requires a terminal")
	}
	selector := operator.TerminalContractSelector{Input: os.Stdin, Output: os.Stdout}
	candidates := []internal.Contract{
		{Name: "First pool", Description: "first description", Network: "ethereum", Address: ethereumContract},
		{Name: "Second pool", Description: "second description", Network: "ethereum", Address: arbitrumContract},
	}

	// when
	selected, err := selector.SelectContract(candidates)

	// then
	require.NoError(t, err)
	require.Equal(t, ethereumContract, selected.Address)
}

func TestTerminalContractSelectorPTYCancelsWithQ(t *testing.T) {
	// given
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("requires a terminal")
	}
	selector := operator.TerminalContractSelector{Input: os.Stdin, Output: os.Stdout}

	// when
	_, err := selector.SelectContract([]internal.Contract{{Name: "Only pool", Description: "only description", Network: "ethereum", Address: ethereumContract}})

	// then
	require.ErrorIs(t, err, operator.ErrContractSelectionCancelled)
}

func TestRequiresContract(t *testing.T) {
	// when, then
	require.False(t, operator.RequiresContract(operator.CommandDeploy))
	require.True(t, operator.RequiresContract(operator.CommandSetDistribution))
	require.True(t, operator.RequiresContract(operator.CommandShow))
	require.True(t, operator.RequiresContract(operator.CommandDistribute))
	require.False(t, operator.RequiresContract(operator.CommandDistributeAll))
	require.True(t, operator.RequiresContract(operator.CommandVerify))
}

func TestResolveContractRejectsASelectionOutsideTheCandidates(t *testing.T) {
	// given
	network, err := internal.ParseNetwork("ethereum")
	require.NoError(t, err)
	selector := &contractSelector{selected: internal.Contract{Network: "arbitrum", Address: arbitrumContract}}

	// when
	_, err = operator.ResolveContract(network, "", []internal.Contract{{Network: "ethereum", Address: ethereumContract}}, selector)

	// then
	require.Error(t, err)
	require.False(t, errors.Is(err, operator.ErrContractSelectionCancelled))
}
