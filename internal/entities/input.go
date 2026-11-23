package entities

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
)

const (
	BasisPoints   = 10_000
	MaxRecipients = 100
)

// Contract identifies a deployment in the local registry.
type Contract struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Network     string `json:"network"`
	Address     string `json:"address"`
}

// Recipient is one validated participant allocation.
type Recipient struct {
	Address  common.Address
	ShareBps uint16
}

// DistributionInput is the recipient-list operation input.
type DistributionInput struct{ Recipients []Recipient }

// DeploymentInput is the operation input required to deploy a distributor.
type DeploymentInput struct {
	Name        string
	Description string
	Owner       common.Address
	FeeWallet   common.Address
	FeeBps      uint16
	Token       common.Address
	Recipients  []Recipient
}

// ParseAddress strictly parses a hexadecimal EVM address.
func ParseAddress(value string) (common.Address, error) {
	if !strings.HasPrefix(value, "0x") || !common.IsHexAddress(value) {
		return common.Address{}, fmt.Errorf("invalid address %q", value)
	}
	return common.HexToAddress(value), nil
}

// Validate checks deployment-only constraints before a deployment can be submitted.
func (input *DeploymentInput) Validate() error {
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("deployment name is required")
	}
	if strings.TrimSpace(input.Description) == "" {
		return errors.New("deployment description is required")
	}
	if err := ValidateLabel("deployment name", input.Name); err != nil {
		return err
	}
	if err := ValidateLabel("deployment description", input.Description); err != nil {
		return err
	}
	if input.Owner == (common.Address{}) || input.FeeWallet == (common.Address{}) || input.Token == (common.Address{}) {
		return errors.New("deployment addresses must not be the zero address")
	}
	if input.FeeBps >= BasisPoints {
		return fmt.Errorf("feeBps must be less than %d", BasisPoints)
	}
	return (&DistributionInput{Recipients: input.Recipients}).Validate()
}

// ValidateLabel rejects control characters from an operator-facing label.
func ValidateLabel(field, value string) error {
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return fmt.Errorf("%s contains control characters", field)
	}
	return nil
}

// Validate checks recipient-list constraints shared by deployment and replacement.
func (input DistributionInput) Validate() error {
	if len(input.Recipients) == 0 || len(input.Recipients) > MaxRecipients {
		return fmt.Errorf("recipients must contain between 1 and %d entries", MaxRecipients)
	}
	seen := make(map[common.Address]struct{}, len(input.Recipients))
	var total uint32
	for index := range input.Recipients {
		recipient := &input.Recipients[index]
		if recipient.Address == (common.Address{}) {
			return errors.New("recipient address must not be the zero address")
		}
		if recipient.ShareBps == 0 {
			return errors.New("recipient shareBps must be positive")
		}
		if _, ok := seen[recipient.Address]; ok {
			return fmt.Errorf("duplicate recipient %s", recipient.Address)
		}
		seen[recipient.Address] = struct{}{}
		total += uint32(recipient.ShareBps)
	}
	if total != BasisPoints {
		return fmt.Errorf("recipient shares must total %d basis points", BasisPoints)
	}
	return nil
}
