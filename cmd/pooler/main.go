package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/getoptimum/optimum-common/pkg/logger"

	localconfig "pooler/internal/config"
	internal "pooler/internal/entities"
	"pooler/internal/operator"
	"pooler/internal/service/deployer"
	"pooler/internal/service/processor"
)

type config struct {
	command  string
	network  string
	input    string
	contract string
}

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		logger.NewAppSLogger(logger.Error).Fatal("pooler command failed", err)
	}
}

func run(args []string, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: pooler <deploy|set-distribution|show|distribute|distribute-all|verify> [flags]")
	}

	cfg := config{command: args[0]}
	flags := flag.NewFlagSet(cfg.command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&cfg.network, "network", "", "network")
	flags.StringVar(&cfg.input, "input", "", "operation input file")
	flags.StringVar(&cfg.contract, "contract", "", "contract address")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}

	command, err := operator.ParseCommand(cfg.command)
	if err != nil {
		return err
	}
	network, err := internal.ParseNetwork(cfg.network)
	if err != nil {
		return err
	}
	configuration, err := localconfig.Load("config.yml")
	if err != nil {
		return err
	}

	deployerService := deployer.New(configuration)
	processorService := processor.New(configuration)
	var contracts []internal.Contract
	if command == operator.CommandDistributeAll || (operator.RequiresContract(command) && cfg.contract == "") {
		contracts, err = operator.LoadContracts("contracts.json")
		if err != nil {
			return err
		}
	}

	contract := cfg.contract
	if operator.RequiresContract(command) {
		contract, err = operator.ResolveContract(network, contract, contracts, operator.TerminalContractSelector{
			Input:  os.Stdin,
			Output: os.Stdout,
		})
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	switch command {
	case operator.CommandDeploy:
		input, err := operator.LoadDeploymentInput(cfg.input)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "validate deployment input", err)
		}
		result, verification, err := deployerService.DeployWithSubmission(ctx, network, &input, func(submitted deployer.DeploymentResult) {
			fmt.Printf("deployment submitted: %s\n", submitted.TransactionHash)
		})
		if err != nil {
			return mutationError(command, network, cfg.input, contract, "deploy", result.TransactionHash, result.Address, err)
		}
		fmt.Print(deploymentSuccessMessage(network, result, verification))
		return nil
	case operator.CommandSetDistribution:
		input, err := operator.LoadDistributionInput(cfg.input)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "validate distribution input", err)
		}
		address, err := operationAddress(contract)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "replace recipients", err)
		}
		result, err := processorService.ReplaceRecipientsWithSubmission(ctx, network, address, input, printSubmittedTransaction)
		if err != nil {
			return mutationError(command, network, cfg.input, contract, "replace recipients", result.TransactionHash, "", err)
		}
		fmt.Printf("recipient replacement confirmed: %s\n", result.TransactionHash)
		return nil
	case operator.CommandShow:
		address, err := operationAddress(contract)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "show", err)
		}
		result, err := processorService.Show(ctx, network, address)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "show", err)
		}
		printShow(&result)
		return nil
	case operator.CommandDistribute:
		address, err := operationAddress(contract)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "distribute", err)
		}
		result, err := processorService.DistributeWithSubmission(ctx, network, address, printSubmittedTransaction)
		if err != nil {
			return mutationError(command, network, cfg.input, contract, "distribute", result.TransactionHash, "", err)
		}
		fmt.Printf("distribution confirmed: %s\n", result.TransactionHash)
		return nil
	case operator.CommandDistributeAll:
		result, err := processorService.DistributeAllWithSubmission(ctx, network, contracts, printSubmittedTransaction)
		printDistributeAll(result)
		if err != nil {
			for index := range result.Results {
				if result.Results[index].TransactionHash != "" {
					return mutationError(command, network, cfg.input, contract, "distribute all", result.Results[index].TransactionHash, "", err)
				}
			}
			return preflightError(command, network, cfg.input, contract, "distribute all", err)
		}
		return nil
	case operator.CommandVerify:
		verification, err := deployerService.Verify(ctx, network, contract)
		if err != nil {
			return preflightError(command, network, cfg.input, contract, "verify", err)
		}
		fmt.Print(verificationSuccessMessage(verification))
		return nil
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}

func operationAddress(value string) (common.Address, error) {
	address, err := internal.ParseAddress(value)
	if err != nil || address == (common.Address{}) {
		return common.Address{}, errors.New("valid contract address is required")
	}
	return address, nil
}

func preflightError(command operator.Command, network internal.Network, input, contract, action string, err error) error {
	return fmt.Errorf("%s preflight failed (network=%q input=%q contract=%q): %s: %w", command, network.Name, input, contract, action, err)
}

func mutationError(command operator.Command, network internal.Network, input, contract, action, transactionHash, address string, err error) error {
	if transactionHash == "" {
		return preflightError(command, network, input, contract, action, err)
	}
	if address != "" {
		return fmt.Errorf("%s failed after submission (network=%q input=%q contract=%q address=%q transaction=%q): %s: %w", command, network.Name, input, contract, address, transactionHash, action, err)
	}
	return fmt.Errorf("%s failed after submission (network=%q input=%q contract=%q transaction=%q): %s: %w", command, network.Name, input, contract, transactionHash, action, err)
}

func deploymentSuccessMessage(network internal.Network, deployment deployer.DeploymentResult, verification deployer.VerificationResult) string {
	message := fmt.Sprintf("deployment confirmed: address=%s transaction=%s explorer=%s\n", deployment.Address, deployment.TransactionHash, explorerContractURL(network, deployment.Address))
	return message + verificationSuccessMessage(verification)
}

func verificationSuccessMessage(verification deployer.VerificationResult) string {
	if verification.AlreadyKnown {
		return "explorer verification: already verified\n"
	}
	return fmt.Sprintf("explorer verification: confirmed (submission=%s)\n", verification.SubmissionID)
}

func explorerContractURL(network internal.Network, address string) string {
	switch network.Name {
	case internal.EthereumNetwork:
		return "https://etherscan.io/address/" + address
	case internal.ArbitrumNetwork:
		return "https://arbiscan.io/address/" + address
	default:
		return ""
	}
}

func printShow(result *processor.ShowResult) {
	fmt.Printf("owner: %s\nfee wallet: %s\nfee rate: %d.%02d%%\ntoken: %s\nbalance: %s\n", result.Owner, result.FeeWallet, result.FeeBps/100, result.FeeBps%100, result.Token, result.Balance)
	for index := range result.Recipients {
		recipient := &result.Recipients[index]
		fmt.Printf("recipient: %s %d.%02d%%\n", recipient.Address, recipient.ShareBps/100, recipient.ShareBps%100)
	}
}

func printDistributeAll(result processor.DistributeAllResult) {
	if result.NothingToProcess {
		fmt.Println("nothing to process")
		return
	}
	for index := range result.Results {
		entry := &result.Results[index]
		fmt.Println(entry.Terminal())
	}
}

func printSubmittedTransaction(result processor.TransactionResult) {
	fmt.Printf("transaction submitted: %s\n", result.TransactionHash)
}
