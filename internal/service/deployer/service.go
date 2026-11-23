// Package deployer deploys and verifies USDT reward distributors.
package deployer

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"pooler/internal/bindings"
	"pooler/internal/config"
	domain "pooler/internal/entities"
	"pooler/internal/operator"
	"pooler/internal/utils"
)

const (
	defaultRegistryPath            = "contracts.json"
	defaultArtifactPath            = "out/USDTRewardDistributor.sol/USDTRewardDistributor.json"
	defaultRPCProbeTime            = 5 * time.Second
	defaultRPCReadTimeout          = 5 * time.Second
	defaultConfirmationTimeout     = 2 * time.Minute
	defaultVerificationMetadataDir = "verification"
)

// Deployer executes distributor deployment and verification with one local configuration.
type Deployer struct{ config *config.Config }

// New creates a deployer that uses configuration for all RPC, signing, and explorer access.
func New(configuration *config.Config) *Deployer { return &Deployer{config: configuration} }

// DeploymentResult identifies the transaction and contract created by a deployment attempt.
type DeploymentResult struct {
	TransactionHash string
	Address         string
}

// DeploymentBackend contains the EVM operations needed to deploy and confirm a contract.
type DeploymentBackend interface {
	bind.ContractBackend
	bind.DeployBackend
}

type deploymentRequest struct {
	network         domain.Network
	input           domain.DeploymentInput
	registryPath    string
	backend         DeploymentBackend
	beforeBroadcast func(DeploymentResult) error
	onSubmitted     func(DeploymentResult)
}

// Deploy submits one deployment, records its immutable verification provenance, and verifies it.
func (d *Deployer) Deploy(ctx context.Context, network domain.Network, input *domain.DeploymentInput) (DeploymentResult, VerificationResult, error) {
	return d.deployAndVerify(ctx, network, input, nil)
}

// DeployWithSubmission deploys and calls submitted immediately after broadcast succeeds.
func (d *Deployer) DeployWithSubmission(ctx context.Context, network domain.Network, input *domain.DeploymentInput, submitted func(DeploymentResult)) (DeploymentResult, VerificationResult, error) {
	return d.deployAndVerify(ctx, network, input, submitted)
}

func (d *Deployer) deployAndVerify(ctx context.Context, network domain.Network, input *domain.DeploymentInput, submitted func(DeploymentResult)) (DeploymentResult, VerificationResult, error) {
	if input == nil {
		return DeploymentResult{}, VerificationResult{}, errors.New("deployment input is required")
	}
	if err := validateDeploymentToken(network, input); err != nil {
		return DeploymentResult{}, VerificationResult{}, err
	}
	if d == nil || d.config == nil {
		return DeploymentResult{}, VerificationResult{}, errors.New("configuration is required")
	}
	client, err := utils.SelectRPC(ctx, network, d.config.RPC[network.RPCKey], defaultRPCProbeTime)
	if err != nil {
		return DeploymentResult{}, VerificationResult{}, err
	}
	defer client.Close()

	var metadata VerificationMetadata
	result, err := d.deploy(ctx, &deploymentRequest{
		network: network, input: *input, registryPath: defaultRegistryPath, backend: client,
		onSubmitted: submitted,
		beforeBroadcast: func(candidate DeploymentResult) error {
			metadata, err = BuildVerificationMetadata(defaultArtifactPath, ".", network, candidate.Address, input)
			if err != nil {
				return fmt.Errorf("preserve verification metadata: %w", err)
			}
			path, err := VerificationMetadataPath(defaultVerificationMetadataDir, network, candidate.Address)
			if err != nil {
				return err
			}
			if err := SaveVerificationMetadata(path, &metadata); err != nil {
				return fmt.Errorf("save verification metadata: %w", err)
			}
			return nil
		},
	})
	if err != nil {
		return result, VerificationResult{}, err
	}
	verification, err := d.verify(ctx, &verificationRequest{network: network, address: result.Address, metadata: &metadata})
	if err != nil {
		return result, verification, fmt.Errorf("deployment confirmed at %s (transaction %s), but explorer verification failed: %w; retry with make verify NETWORK=%s CONTRACT=%s", result.Address, result.TransactionHash, err, network.Name, result.Address)
	}
	return result, verification, nil
}

func (d *Deployer) deploy(ctx context.Context, request *deploymentRequest) (DeploymentResult, error) {
	if request == nil {
		return DeploymentResult{}, errors.New("deployment request is required")
	}
	if err := request.input.Validate(); err != nil {
		return DeploymentResult{}, fmt.Errorf("validate deployment input: %w", err)
	}
	if err := validateDeploymentToken(request.network, &request.input); err != nil {
		return DeploymentResult{}, err
	}
	if request.backend == nil {
		return DeploymentResult{}, errors.New("deployment backend is required")
	}
	if d == nil || d.config == nil {
		return DeploymentResult{}, errors.New("configuration is required")
	}
	if request.registryPath == "" {
		return DeploymentResult{}, errors.New("contract registry path is required")
	}
	if _, err := operator.LoadContracts(request.registryPath); err != nil {
		return DeploymentResult{}, fmt.Errorf("read contract registry before deployment: %w", err)
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(d.config.Signing.PrivateKey, "0x"))
	if err != nil {
		return DeploymentResult{}, errors.New("invalid signing private key")
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(request.network.ChainID))
	if err != nil {
		return DeploymentResult{}, fmt.Errorf("create deployment signer: %w", err)
	}
	preparationContext, cancelPreparation := context.WithTimeout(ctx, defaultRPCReadTimeout)
	defer cancelPreparation()
	auth.Context = preparationContext
	auth.NoSend = true
	predictedAddress, transaction, _, err := bindings.DeployUSDTRewardDistributor(
		auth, request.backend, request.input.Token, new(big.Int).SetUint64(uint64(request.input.FeeBps)),
		request.input.Owner, request.input.FeeWallet, deploymentRecipients(request.input.Recipients),
	)
	if err != nil {
		return DeploymentResult{}, errors.New("prepare deployment transaction")
	}
	result := DeploymentResult{TransactionHash: transaction.Hash().Hex()}
	if request.beforeBroadcast != nil {
		candidate := result
		candidate.Address = predictedAddress.Hex()
		if err := request.beforeBroadcast(candidate); err != nil {
			return DeploymentResult{}, fmt.Errorf("prepare deployment before broadcast: %w", err)
		}
	}
	broadcastContext, cancelBroadcast := context.WithTimeout(ctx, defaultRPCReadTimeout)
	defer cancelBroadcast()
	if err := request.backend.SendTransaction(broadcastContext, transaction); err != nil {
		return result, fmt.Errorf("deployment transaction %s submitted; broadcast outcome is uncertain", result.TransactionHash)
	}
	if request.onSubmitted != nil {
		request.onSubmitted(result)
	}
	confirmationContext, cancelConfirmation := context.WithTimeout(ctx, defaultConfirmationTimeout)
	defer cancelConfirmation()
	receipt, err := bind.WaitMined(confirmationContext, request.backend, transaction)
	if err != nil {
		return result, fmt.Errorf("deployment transaction %s submitted; confirmation is uncertain", result.TransactionHash)
	}
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return result, fmt.Errorf("deployment transaction %s failed", result.TransactionHash)
	}
	if receipt.ContractAddress == (common.Address{}) {
		return result, fmt.Errorf("deployment transaction %s succeeded without a contract address", result.TransactionHash)
	}
	readContext, cancelRead := context.WithTimeout(ctx, defaultRPCReadTimeout)
	defer cancelRead()
	code, err := request.backend.CodeAt(readContext, receipt.ContractAddress, nil)
	if err != nil {
		return result, fmt.Errorf("deployment transaction %s confirmed but cannot read deployed bytecode", result.TransactionHash)
	}
	if len(code) == 0 {
		return result, fmt.Errorf("deployment transaction %s confirmed but deployed bytecode is empty", result.TransactionHash)
	}
	if receipt.ContractAddress != predictedAddress {
		return result, fmt.Errorf("deployment transaction %s confirmed at an unexpected contract address", result.TransactionHash)
	}
	result.Address = receipt.ContractAddress.Hex()
	if err := operator.RegisterContractAtPath(request.registryPath, domain.Contract{Name: request.input.Name, Description: request.input.Description, Network: request.network.Name, Address: result.Address}); err != nil {
		return result, fmt.Errorf("deployment confirmed at %s (transaction %s), but cannot save contract registry: %w", result.Address, result.TransactionHash, err)
	}
	return result, nil
}

func validateDeploymentToken(network domain.Network, input *domain.DeploymentInput) error {
	var expected common.Address
	switch network.Name {
	case domain.EthereumNetwork:
		expected = common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
	case domain.ArbitrumNetwork:
		expected = common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9")
	default:
		return fmt.Errorf("unsupported deployment network %q", network.Name)
	}
	if input.Token != expected {
		return fmt.Errorf("deployment token does not match the supported USDT address for %s", network.Name)
	}
	return nil
}

func deploymentRecipients(recipients []domain.Recipient) []bindings.USDTRewardDistributorRecipient {
	converted := make([]bindings.USDTRewardDistributorRecipient, len(recipients))
	for index, recipient := range recipients {
		converted[index] = bindings.USDTRewardDistributorRecipient{Account: recipient.Address, ShareBps: recipient.ShareBps}
	}
	return converted
}
