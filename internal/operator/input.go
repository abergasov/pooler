// Package operator contains CLI parsing and local contract selection helpers.
package operator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	internal "pooler/internal/entities"
)

// Command identifies one supported pooler command.
type Command string

const (
	CommandDeploy          Command = "deploy"
	CommandSetDistribution Command = "set-distribution"
	CommandShow            Command = "show"
	CommandDistribute      Command = "distribute"
	CommandDistributeAll   Command = "distribute-all"
	CommandVerify          Command = "verify"
)

// ParseCommand validates a command name at the CLI boundary.
func ParseCommand(value string) (Command, error) {
	command := Command(value)
	switch command {
	case CommandDeploy, CommandSetDistribution, CommandShow, CommandDistribute, CommandDistributeAll, CommandVerify:
		return command, nil
	default:
		return "", fmt.Errorf("unknown command %q", value)
	}
}

// LoadDeploymentInput opens, parses, and validates a deployment input file.
func LoadDeploymentInput(path string) (internal.DeploymentInput, error) {
	return loadInput(path, ParseDeploymentInput)
}

// LoadDistributionInput opens, parses, and validates a recipient-list input file.
func LoadDistributionInput(path string) (internal.DistributionInput, error) {
	return loadInput(path, ParseDistributionInput)
}

// ParseDeploymentInput parses and validates deployment-specific fields from JSON.
func ParseDeploymentInput(reader io.Reader) (internal.DeploymentInput, error) {
	fields, err := parseInput(reader)
	if err != nil {
		return internal.DeploymentInput{}, err
	}

	name, err := requiredString(fields, "name")
	if err != nil {
		return internal.DeploymentInput{}, err
	}
	description, err := requiredString(fields, "description")
	if err != nil {
		return internal.DeploymentInput{}, err
	}
	owner, err := requiredAddress(fields, "owner")
	if err != nil {
		return internal.DeploymentInput{}, err
	}
	feeWallet, err := requiredAddress(fields, "feeWallet")
	if err != nil {
		return internal.DeploymentInput{}, err
	}
	feeBps, err := requiredUint16(fields, "feeBps")
	if err != nil {
		return internal.DeploymentInput{}, err
	}
	token, err := requiredAddress(fields, "token")
	if err != nil {
		return internal.DeploymentInput{}, err
	}
	recipients, err := parseRecipients(fields)
	if err != nil {
		return internal.DeploymentInput{}, err
	}

	input := internal.DeploymentInput{
		Name:        name,
		Description: description,
		Owner:       owner,
		FeeWallet:   feeWallet,
		FeeBps:      feeBps,
		Token:       token,
		Recipients:  recipients,
	}
	if err := input.Validate(); err != nil {
		return internal.DeploymentInput{}, err
	}
	return input, nil
}

// ParseDistributionInput parses and validates only the recipient list from JSON.
func ParseDistributionInput(reader io.Reader) (internal.DistributionInput, error) {
	fields, err := parseInput(reader)
	if err != nil {
		return internal.DistributionInput{}, err
	}
	recipients, err := parseRecipients(fields)
	if err != nil {
		return internal.DistributionInput{}, err
	}
	input := internal.DistributionInput{Recipients: recipients}
	if err := input.Validate(); err != nil {
		return internal.DistributionInput{}, err
	}
	return input, nil
}

func loadInput[T any](path string, parse func(io.Reader) (T, error)) (T, error) {
	var zero T
	if path == "" {
		return zero, errors.New("input file is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return zero, fmt.Errorf("open input file: %w", err)
	}
	defer file.Close()

	return parse(file)
}

func parseInput(reader io.Reader) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&fields); err != nil {
		return nil, fmt.Errorf("parse operation input: %w", err)
	}
	if fields == nil {
		return nil, errors.New("operation input must be an object")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("operation input must contain one JSON value")
		}
		return nil, fmt.Errorf("parse operation input: %w", err)
	}
	return fields, nil
}

func requiredString(fields map[string]json.RawMessage, name string) (string, error) {
	value, ok := fields[name]
	if !ok {
		return "", fmt.Errorf("%s is required", name)
	}
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return "", fmt.Errorf("%s must be a string: %w", name, err)
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return text, nil
}

func requiredAddress(fields map[string]json.RawMessage, name string) (common.Address, error) {
	value, err := requiredString(fields, name)
	if err != nil {
		return common.Address{}, err
	}
	address, err := internal.ParseAddress(value)
	if err != nil {
		return common.Address{}, fmt.Errorf("%s: %w", name, err)
	}
	return address, nil
}

func requiredUint16(fields map[string]json.RawMessage, name string) (uint16, error) {
	value, ok := fields[name]
	if !ok {
		return 0, fmt.Errorf("%s is required", name)
	}
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return 0, fmt.Errorf("%s must be an unsigned integer", name)
	}
	var number uint16
	if err := json.Unmarshal(value, &number); err != nil {
		return 0, fmt.Errorf("%s must be an unsigned integer: %w", name, err)
	}
	return number, nil
}

func parseRecipients(fields map[string]json.RawMessage) ([]internal.Recipient, error) {
	value, ok := fields["recipients"]
	if !ok {
		return nil, errors.New("recipients is required")
	}
	var rawRecipients []struct {
		Address  string `json:"address"`
		ShareBps uint16 `json:"shareBps"`
	}
	if err := json.Unmarshal(value, &rawRecipients); err != nil {
		return nil, fmt.Errorf("recipients must be an array: %w", err)
	}

	recipients := make([]internal.Recipient, len(rawRecipients))
	for index, recipient := range rawRecipients {
		address, err := internal.ParseAddress(recipient.Address)
		if err != nil {
			return nil, fmt.Errorf("recipients[%d].address: %w", index, err)
		}
		recipients[index] = internal.Recipient{Address: address, ShareBps: recipient.ShareBps}
	}
	return recipients, nil
}
