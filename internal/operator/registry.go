// Package operator contains local state used by the pooler command.
package operator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/flock"

	domain "pooler/internal/entities"
)

// LoadContracts reads a contract registry. A missing registry is an empty one.
func LoadContracts(path string) ([]domain.Contract, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read contract registry: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()

	var contracts []domain.Contract
	if err := decoder.Decode(&contracts); err != nil {
		return nil, fmt.Errorf("decode contract registry: %w", err)
	}
	if err := requireOnlyWhitespace(decoder); err != nil {
		return nil, err
	}
	if contracts == nil {
		return nil, errors.New("decode contract registry: expected a JSON array")
	}

	return validateContracts(contracts)
}

// RegisterContract adds contract unless its canonical network and address are known.
func RegisterContract(contracts []domain.Contract, contract domain.Contract) ([]domain.Contract, error) {
	validated, err := validateContracts(append(append([]domain.Contract(nil), contracts...), contract))
	if err != nil {
		return nil, err
	}
	return validated, nil
}

// RegisterContractAtPath locks, reloads, registers, and saves one contract registry entry.
func RegisterContractAtPath(path string, contract domain.Contract) (err error) {
	lock := flock.New(path + ".lock")
	if err = lock.Lock(); err != nil {
		return fmt.Errorf("lock contract registry: %w", err)
	}
	defer func() {
		if unlockErr := lock.Unlock(); unlockErr != nil && err == nil {
			err = fmt.Errorf("unlock contract registry: %w", unlockErr)
		}
	}()

	contracts, err := LoadContracts(path)
	if err != nil {
		return err
	}
	contracts, err = RegisterContract(contracts, contract)
	if err != nil {
		return err
	}
	return SaveContracts(path, contracts)
}

// SaveContracts validates and atomically replaces the contract registry at path.
func SaveContracts(path string, contracts []domain.Contract) error {
	validated, err := validateContracts(contracts)
	if err != nil {
		return err
	}

	contents, err := json.MarshalIndent(validated, "", "  ")
	if err != nil {
		return fmt.Errorf("encode contract registry: %w", err)
	}
	contents = append(contents, '\n')

	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".contracts-*.json")
	if err != nil {
		return fmt.Errorf("create temporary contract registry: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary contract registry: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary contract registry permissions: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary contract registry: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary contract registry: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace contract registry: %w", err)
	}

	return nil
}

func requireOnlyWhitespace(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("decode contract registry: multiple JSON values")
		}
		return fmt.Errorf("decode contract registry: %w", err)
	}
	return nil
}

func validateContracts(contracts []domain.Contract) ([]domain.Contract, error) {
	validated := make([]domain.Contract, len(contracts))
	identities := make(map[string]struct{}, len(contracts))
	for index, contract := range contracts {
		if err := domain.ValidateLabel("contract name", contract.Name); err != nil {
			return nil, fmt.Errorf("validate contract %d: %w", index, err)
		}
		if err := domain.ValidateLabel("contract description", contract.Description); err != nil {
			return nil, fmt.Errorf("validate contract %d: %w", index, err)
		}
		network, err := normalizeNetwork(contract.Network)
		if err != nil {
			return nil, fmt.Errorf("validate contract %d: %w", index, err)
		}
		address, err := normalizeAddress(contract.Address)
		if err != nil {
			return nil, fmt.Errorf("validate contract %d: %w", index, err)
		}

		contract.Network = network
		contract.Address = address
		identity := network + ":" + address
		if _, exists := identities[identity]; exists {
			return nil, fmt.Errorf("validate contract %d: duplicate network and address", index)
		}
		identities[identity] = struct{}{}
		validated[index] = contract
	}
	return validated, nil
}

func normalizeNetwork(network string) (string, error) {
	parsed, err := domain.ParseNetwork(network)
	if err != nil {
		return "", err
	}
	return parsed.Name, nil
}

func normalizeAddress(address string) (string, error) {
	parsed, err := domain.ParseAddress(address)
	if err != nil || parsed == (common.Address{}) {
		return "", fmt.Errorf("invalid contract address %q", address)
	}
	return strings.ToLower(parsed.Hex()), nil
}
