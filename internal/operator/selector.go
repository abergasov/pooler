package operator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

	domain "pooler/internal/entities"
)

// ErrContractSelectionCancelled reports that the operator canceled a terminal selection.
var ErrContractSelectionCancelled = errors.New("contract selection canceled")

// ContractSelector chooses one registered contract from the supplied candidates.
type ContractSelector interface {
	SelectContract([]domain.Contract) (domain.Contract, error)
}

// TerminalContractSelector presents a minimal arrow-key menu on a terminal.
type TerminalContractSelector struct {
	Input  *os.File
	Output io.Writer
}

// RequiresContract reports whether command operates on one existing contract.
func RequiresContract(command Command) bool {
	switch command {
	case CommandSetDistribution, CommandShow, CommandDistribute, CommandVerify:
		return true
	default:
		return false
	}
}

// ResolveContract returns an explicit address or one selected from network-scoped contracts.
// Callers must return an error from this function before executing their operation, which keeps
// ErrContractSelectionCancelled and failed selection from causing a mutation.
func ResolveContract(network domain.Network, explicitAddress string, contracts []domain.Contract, selector ContractSelector) (string, error) {
	if explicitAddress != "" {
		address, err := normalizeAddress(explicitAddress)
		if err != nil {
			return "", err
		}
		return address, nil
	}

	candidates := contractsForNetwork(network, contracts)
	if len(candidates) == 0 {
		return "", fmt.Errorf("no registered contracts for network %q; set CONTRACT=<address>", network.Name)
	}
	if selector == nil {
		return "", errors.New("contract selection requires an interactive terminal; set CONTRACT=<address>")
	}

	selected, err := selector.SelectContract(candidates)
	if err != nil {
		return "", fmt.Errorf("select contract: %w", err)
	}
	for _, candidate := range candidates {
		if selected.Address == candidate.Address {
			return candidate.Address, nil
		}
	}
	return "", errors.New("selected contract is not registered for the selected network")
}

func contractsForNetwork(network domain.Network, contracts []domain.Contract) []domain.Contract {
	candidates := make([]domain.Contract, 0, len(contracts))
	for _, contract := range contracts {
		if contract.Network == network.Name {
			candidates = append(candidates, contract)
		}
	}
	return candidates
}

// SelectContract displays the candidates and returns one after the operator presses Enter.
func (selector TerminalContractSelector) SelectContract(candidates []domain.Contract) (contract domain.Contract, err error) {
	if selector.Input == nil || !term.IsTerminal(int(selector.Input.Fd())) {
		return domain.Contract{}, errors.New("contract selection requires an interactive terminal; set CONTRACT=<address>")
	}
	if selector.Output == nil {
		return domain.Contract{}, errors.New("contract selection requires terminal output")
	}
	if len(candidates) == 0 {
		return domain.Contract{}, errors.New("no contracts available for selection; set CONTRACT=<address>")
	}

	state, err := term.MakeRaw(int(selector.Input.Fd()))
	if err != nil {
		return domain.Contract{}, fmt.Errorf("prepare terminal selection: %w", err)
	}
	defer func() {
		if restoreErr := term.Restore(int(selector.Input.Fd()), state); restoreErr != nil && err == nil {
			contract = domain.Contract{}
			err = fmt.Errorf("restore terminal selection: %w", restoreErr)
		}
	}()

	selected := 0
	reader := bufio.NewReader(selector.Input)
	for {
		renderContractMenu(selector.Output, candidates, selected)
		key, err := reader.ReadByte()
		if err != nil {
			return domain.Contract{}, fmt.Errorf("read contract selection: %w", err)
		}

		switch key {
		case '\r', '\n':
			return candidates[selected], nil
		case 3, 'q':
			return domain.Contract{}, ErrContractSelectionCancelled
		case '\x1b':
			bracket, bracketErr := reader.ReadByte()
			direction, directionErr := reader.ReadByte()
			if bracketErr != nil || directionErr != nil || bracket != '[' {
				return domain.Contract{}, ErrContractSelectionCancelled
			}
			switch direction {
			case 'A':
				selected = (selected + len(candidates) - 1) % len(candidates)
			case 'B':
				selected = (selected + 1) % len(candidates)
			}
		}
	}
}

func renderContractMenu(output io.Writer, candidates []domain.Contract, selected int) {
	fmt.Fprint(output, "\033[H\033[2JSelect a contract (arrow keys, Enter; q cancels):\n")
	for index, contract := range candidates {
		marker := " "
		if index == selected {
			marker = ">"
		}
		fmt.Fprintf(output, "%s %s — %s — %s\n", marker, contract.Name, contract.Description, contract.Address)
	}
}
