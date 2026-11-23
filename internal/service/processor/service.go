// Package processor executes distributor reads and owner actions.
package processor

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"pooler/internal/bindings"
	"pooler/internal/config"
	internal "pooler/internal/entities"
	"pooler/internal/utils"
)

const (
	rpcProbeTimeout          = 5 * time.Second
	rpcReadTimeout           = 5 * time.Second
	confirmationTimeout      = 2 * time.Minute
	batchConfirmationTimeout = 2 * time.Minute

	batchStatusOK      = "OK"
	batchStatusSkipped = "SKIPPED"
	batchStatusFailed  = "FAILED"
)

// Processor executes distributor operations using one runtime configuration.
type Processor struct {
	configuration *config.Config
}

// New creates a processor backed by configuration.
func New(configuration *config.Config) *Processor {
	return &Processor{configuration: configuration}
}

// ShowResult is the current public on-chain configuration and USDT balance.
type ShowResult struct {
	Owner      common.Address
	FeeWallet  common.Address
	FeeBps     uint16
	Token      common.Address
	Recipients []internal.Recipient
	Balance    *big.Int
}

// TransactionResult records a submitted transaction and whether its receipt succeeded.
type TransactionResult struct {
	TransactionHash string
	Confirmed       bool
}

// DistributeAllResult contains one terminal result for every selected registry contract.
type DistributeAllResult struct {
	Results          []DistributeAllContractResult
	NothingToProcess bool
}

// DistributeAllContractResult records the terminal state for one registry contract.
type DistributeAllContractResult struct {
	Contract        internal.Contract
	Label           string
	Status          string
	TransactionHash string
	Failure         string
}

// Terminal returns the operator-facing result for one contract.
func (result *DistributeAllContractResult) Terminal() string {
	switch result.Status {
	case batchStatusOK:
		return fmt.Sprintf("%s: %s: %s", result.Label, result.Status, result.TransactionHash)
	case batchStatusSkipped:
		return fmt.Sprintf("%s: %s: empty balance", result.Label, result.Status)
	default:
		if result.TransactionHash != "" {
			return fmt.Sprintf("%s: distribution err: %s (transaction %s)", result.Label, result.Failure, result.TransactionHash)
		}
		return fmt.Sprintf("%s: distribution err: %s", result.Label, result.Failure)
	}
}

// Show reads the contract's public configuration and configured USDT balance.
func (processor *Processor) Show(ctx context.Context, network internal.Network, contract common.Address) (ShowResult, error) {
	if contract == (common.Address{}) {
		return ShowResult{}, errors.New("contract address is required")
	}
	client, err := processor.client(ctx, network)
	if err != nil {
		return ShowResult{}, err
	}
	defer client.Close()

	readContext, cancel := context.WithTimeout(ctx, rpcReadTimeout)
	defer cancel()
	return show(readContext, contract, client)
}

// ReplaceRecipients verifies the signer is the current owner, then replaces only the recipient list.
func (processor *Processor) ReplaceRecipients(ctx context.Context, network internal.Network, contract common.Address, input internal.DistributionInput) (TransactionResult, error) {
	return processor.replaceRecipients(ctx, network, contract, input, nil)
}

// ReplaceRecipientsWithSubmission replaces recipients and reports a broadcast transaction immediately.
func (processor *Processor) ReplaceRecipientsWithSubmission(ctx context.Context, network internal.Network, contract common.Address, input internal.DistributionInput, submitted func(TransactionResult)) (TransactionResult, error) {
	return processor.replaceRecipients(ctx, network, contract, input, submitted)
}

func (processor *Processor) replaceRecipients(ctx context.Context, network internal.Network, contract common.Address, input internal.DistributionInput, submitted func(TransactionResult)) (TransactionResult, error) {
	if err := input.Validate(); err != nil {
		return TransactionResult{}, fmt.Errorf("validate replacement recipients: %w", err)
	}
	if contract == (common.Address{}) {
		return TransactionResult{}, errors.New("contract address is required")
	}
	client, err := processor.client(ctx, network)
	if err != nil {
		return TransactionResult{}, err
	}
	defer client.Close()

	recipients := make([]bindings.USDTRewardDistributorRecipient, len(input.Recipients))
	for index, recipient := range input.Recipients {
		recipients[index] = bindings.USDTRewardDistributorRecipient{Account: recipient.Address, ShareBps: recipient.ShareBps}
	}
	return processor.mutate(ctx, network, contract, client, "replaceRecipients", submitted, func(distributor *bindings.USDTRewardDistributor, options *bind.TransactOpts) (*types.Transaction, error) {
		return distributor.ReplaceRecipients(options, recipients)
	})
}

// Distribute verifies the signer is the current owner, then submits one full-balance distribution.
func (processor *Processor) Distribute(ctx context.Context, network internal.Network, contract common.Address) (TransactionResult, error) {
	return processor.distributeWithSubmission(ctx, network, contract, nil)
}

// DistributeWithSubmission distributes and reports a broadcast transaction immediately.
func (processor *Processor) DistributeWithSubmission(ctx context.Context, network internal.Network, contract common.Address, submitted func(TransactionResult)) (TransactionResult, error) {
	return processor.distributeWithSubmission(ctx, network, contract, submitted)
}

func (processor *Processor) distributeWithSubmission(ctx context.Context, network internal.Network, contract common.Address, submitted func(TransactionResult)) (TransactionResult, error) {
	if contract == (common.Address{}) {
		return TransactionResult{}, errors.New("contract address is required")
	}
	client, err := processor.client(ctx, network)
	if err != nil {
		return TransactionResult{}, err
	}
	defer client.Close()

	return processor.distribute(ctx, network, contract, client, submitted)
}

// DistributeAll distributes funded contracts registered for network in registry order.
func (processor *Processor) DistributeAll(ctx context.Context, network internal.Network, contracts []internal.Contract) (DistributeAllResult, error) {
	return processor.distributeAll(ctx, network, contracts, nil)
}

// DistributeAllWithSubmission distributes registered contracts and reports each broadcast immediately.
func (processor *Processor) DistributeAllWithSubmission(ctx context.Context, network internal.Network, contracts []internal.Contract, submitted func(TransactionResult)) (DistributeAllResult, error) {
	return processor.distributeAll(ctx, network, contracts, submitted)
}

func (processor *Processor) distributeAll(ctx context.Context, network internal.Network, contracts []internal.Contract, submitted func(TransactionResult)) (DistributeAllResult, error) {
	contracts = contractsForNetwork(network, contracts)
	if len(contracts) == 0 {
		return DistributeAllResult{NothingToProcess: true}, nil
	}
	client, err := processor.client(ctx, network)
	if err != nil {
		return DistributeAllResult{}, err
	}
	defer client.Close()
	if err := processor.validateMutation(network, client); err != nil {
		return DistributeAllResult{}, err
	}

	labels := distributeAllLabels(contracts)
	result := DistributeAllResult{Results: make([]DistributeAllContractResult, 0, len(contracts))}
	failed := false
	for index, contract := range contracts {
		entry := DistributeAllContractResult{Contract: contract, Label: labels[index]}
		address, err := internal.ParseAddress(contract.Address)
		if err != nil || address == (common.Address{}) {
			entry.Status = batchStatusFailed
			entry.Failure = fmt.Sprintf("invalid contract address %q", contract.Address)
			result.Results = append(result.Results, entry)
			failed = true
			continue
		}

		contractContext, cancel := context.WithTimeout(ctx, batchConfirmationTimeout)
		shown, err := show(contractContext, address, client)
		if err != nil {
			cancel()
			entry.Status = batchStatusFailed
			entry.Failure = fmt.Errorf("read balance: %w", err).Error()
			result.Results = append(result.Results, entry)
			failed = true
			continue
		}
		if shown.Balance.Sign() == 0 {
			cancel()
			entry.Status = batchStatusSkipped
			result.Results = append(result.Results, entry)
			continue
		}

		transaction, err := processor.distribute(contractContext, network, address, client, submitted)
		cancel()
		entry.TransactionHash = transaction.TransactionHash
		if err != nil {
			entry.Status = batchStatusFailed
			entry.Failure = err.Error()
			result.Results = append(result.Results, entry)
			failed = true
			continue
		}
		if !transaction.Confirmed {
			entry.Status = batchStatusFailed
			entry.Failure = "transaction was not confirmed"
			result.Results = append(result.Results, entry)
			failed = true
			continue
		}

		entry.Status = batchStatusOK
		result.Results = append(result.Results, entry)
	}

	if failed {
		return result, errors.New("one or more distributions failed")
	}
	return result, nil
}

type actionBackend interface {
	bind.ContractBackend
	bind.DeployBackend
}

func (processor *Processor) client(ctx context.Context, network internal.Network) (*ethclient.Client, error) {
	if processor == nil || processor.configuration == nil {
		return nil, errors.New("configuration is required")
	}
	return utils.SelectRPC(ctx, network, processor.configuration.RPC[network.RPCKey], rpcProbeTimeout)
}

func show(ctx context.Context, contractAddress common.Address, backend bind.ContractCaller) (ShowResult, error) {
	distributor, err := bindings.NewUSDTRewardDistributorCaller(contractAddress, backend)
	if err != nil {
		return ShowResult{}, errors.New("read contract owner")
	}
	options := &bind.CallOpts{Context: ctx}
	owner, err := distributor.Owner(options)
	if err != nil {
		return ShowResult{}, errors.New("read contract owner")
	}
	feeWallet, err := distributor.FeeWallet(options)
	if err != nil {
		return ShowResult{}, errors.New("read contract fee wallet")
	}
	feeBps, err := distributor.FeeBps(options)
	if err != nil {
		return ShowResult{}, errors.New("read contract fee rate")
	}
	token, err := distributor.Usdt(options)
	if err != nil {
		return ShowResult{}, errors.New("read contract token")
	}
	recipients, err := distributor.GetRecipients(options)
	if err != nil {
		return ShowResult{}, errors.New("read contract recipients")
	}

	tokenContract, err := bindings.NewIERC20BalanceCaller(token, backend)
	if err != nil {
		return ShowResult{}, errors.New("read contract token balance")
	}
	balance, err := tokenContract.BalanceOf(options, contractAddress)
	if err != nil {
		return ShowResult{}, errors.New("read contract token balance")
	}
	converted := make([]internal.Recipient, len(recipients))
	for index, recipient := range recipients {
		converted[index] = internal.Recipient{Address: recipient.Account, ShareBps: recipient.ShareBps}
	}

	return ShowResult{
		Owner:      owner,
		FeeWallet:  feeWallet,
		FeeBps:     feeBps,
		Token:      token,
		Recipients: converted,
		Balance:    new(big.Int).Set(balance),
	}, nil
}

func (processor *Processor) distribute(ctx context.Context, network internal.Network, contract common.Address, backend actionBackend, submitted func(TransactionResult)) (TransactionResult, error) {
	return processor.mutate(ctx, network, contract, backend, "distribute", submitted, func(distributor *bindings.USDTRewardDistributor, options *bind.TransactOpts) (*types.Transaction, error) {
		return distributor.Distribute(options)
	})
}

func (processor *Processor) mutate(ctx context.Context, network internal.Network, contractAddress common.Address, backend actionBackend, method string, submitted func(TransactionResult), transact func(*bindings.USDTRewardDistributor, *bind.TransactOpts) (*types.Transaction, error)) (TransactionResult, error) {
	if err := processor.validateMutation(network, backend); err != nil {
		return TransactionResult{}, err
	}
	key, err := signingKey(processor.configuration)
	if err != nil {
		return TransactionResult{}, err
	}
	signer := crypto.PubkeyToAddress(key.PublicKey)
	distributor, err := bindings.NewUSDTRewardDistributor(contractAddress, backend)
	if err != nil {
		return TransactionResult{}, errors.New("read contract owner before transaction")
	}
	readContext, cancelRead := context.WithTimeout(ctx, rpcReadTimeout)
	defer cancelRead()
	owner, err := distributor.Owner(&bind.CallOpts{Context: readContext})
	if err != nil {
		return TransactionResult{}, errors.New("read contract owner before transaction")
	}
	if owner != signer {
		return TransactionResult{}, fmt.Errorf("signer %s is not the current contract owner", signer.Hex())
	}

	auth, err := bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(network.ChainID))
	if err != nil {
		return TransactionResult{}, errors.New("create transaction signer")
	}
	preparationContext, cancelPreparation := context.WithTimeout(ctx, rpcReadTimeout)
	defer cancelPreparation()
	auth.Context = preparationContext
	auth.NoSend = true
	transaction, err := transact(distributor, auth)
	if err != nil {
		return TransactionResult{}, fmt.Errorf("submit %s transaction", method)
	}

	result := TransactionResult{TransactionHash: transaction.Hash().Hex()}
	broadcastContext, cancelBroadcast := context.WithTimeout(ctx, rpcReadTimeout)
	defer cancelBroadcast()
	if err := backend.SendTransaction(broadcastContext, transaction); err != nil {
		return result, fmt.Errorf("%s transaction %s submitted; confirmation is uncertain", method, result.TransactionHash)
	}
	if submitted != nil {
		submitted(result)
	}
	confirmationContext, cancel := context.WithTimeout(ctx, confirmationTimeout)
	defer cancel()
	receipt, err := bind.WaitMined(confirmationContext, backend, transaction)
	if err != nil {
		return result, fmt.Errorf("%s transaction %s submitted; confirmation is uncertain", method, result.TransactionHash)
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return result, fmt.Errorf("%s transaction %s failed", method, result.TransactionHash)
	}

	result.Confirmed = true
	return result, nil
}

func (processor *Processor) validateMutation(network internal.Network, backend actionBackend) error {
	if backend == nil {
		return errors.New("contract backend is required")
	}
	if network.ChainID == 0 {
		return errors.New("network chain ID is required")
	}
	_, err := signingKey(processor.configuration)
	return err
}

func signingKey(configuration *config.Config) (*ecdsa.PrivateKey, error) {
	if configuration == nil {
		return nil, errors.New("configuration is required")
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(configuration.Signing.PrivateKey, "0x"))
	if err != nil {
		return nil, errors.New("invalid signing private key")
	}
	return key, nil
}

func contractsForNetwork(network internal.Network, contracts []internal.Contract) []internal.Contract {
	candidates := make([]internal.Contract, 0, len(contracts))
	for _, contract := range contracts {
		if contract.Network == network.Name {
			candidates = append(candidates, contract)
		}
	}
	return candidates
}

func distributeAllLabels(contracts []internal.Contract) []string {
	counts := make(map[string]int, len(contracts))
	for _, contract := range contracts {
		counts[contract.Name]++
	}

	labels := make([]string, len(contracts))
	for index, contract := range contracts {
		if contract.Name == "" {
			labels[index] = contract.Address
			continue
		}
		if counts[contract.Name] > 1 {
			labels[index] = fmt.Sprintf("%s (%s)", contract.Name, contract.Address)
			continue
		}
		labels[index] = contract.Name
	}
	return labels
}
