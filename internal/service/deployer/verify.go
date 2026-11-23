package deployer

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"pooler/internal/bindings"
	domain "pooler/internal/entities"
)

const (
	defaultExplorerURL         = "https://api.etherscan.io/v2/api"
	defaultVerificationTimeout = 2 * time.Minute
	defaultVerificationPoll    = 2 * time.Second
)

// VerificationMetadata is the immutable build provenance needed to retry source verification.
type VerificationMetadata struct {
	Network              string          `json:"network"`
	Address              string          `json:"address"`
	ContractName         string          `json:"contractName"`
	CompilerVersion      string          `json:"compilerVersion"`
	CompilerInput        json.RawMessage `json:"compilerInput"`
	ConstructorArguments string          `json:"constructorArguments"`
}

// VerificationResult reports a terminal explorer outcome.
type VerificationResult struct {
	SubmissionID string
	AlreadyKnown bool
}

type verificationRequest struct {
	network      domain.Network
	address      string
	metadata     *VerificationMetadata
	httpClient   *http.Client
	explorerURL  string
	pollInterval time.Duration
	timeout      time.Duration
}

// Verify submits the saved provenance for a deployment to the selected explorer.
func (d *Deployer) Verify(ctx context.Context, network domain.Network, address string) (VerificationResult, error) {
	path, err := VerificationMetadataPath(defaultVerificationMetadataDir, network, address)
	if err != nil {
		return VerificationResult{}, err
	}
	metadata, err := LoadVerificationMetadata(path)
	if err != nil {
		return VerificationResult{}, err
	}
	return d.verify(ctx, &verificationRequest{network: network, address: address, metadata: &metadata})
}

func (d *Deployer) verify(ctx context.Context, request *verificationRequest) (VerificationResult, error) {
	if request == nil || request.metadata == nil {
		return VerificationResult{}, errors.New("verification metadata is required")
	}
	canonicalNetwork, err := canonicalVerificationNetwork(request.network)
	if err != nil {
		return VerificationResult{}, err
	}
	canonicalAddress, err := normalizeAddress(request.address)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := validateVerificationMetadata(request.metadata); err != nil {
		return VerificationResult{}, err
	}
	if request.metadata.Network != canonicalNetwork.Name || request.metadata.Address != canonicalAddress {
		return VerificationResult{}, errors.New("verification metadata does not match the selected network and contract")
	}
	if d == nil || d.config == nil {
		return VerificationResult{}, errors.New("configuration is required")
	}
	explorerURL := request.explorerURL
	if explorerURL == "" {
		explorerURL = defaultExplorerURL
	}
	if _, err := url.ParseRequestURI(explorerURL); err != nil {
		return VerificationResult{}, errors.New("invalid explorer URL")
	}
	client := request.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	timeout := request.timeout
	if timeout == 0 {
		timeout = defaultVerificationTimeout
	}
	if timeout < 0 {
		return VerificationResult{}, errors.New("verification timeout must not be negative")
	}
	pollInterval := request.pollInterval
	if pollInterval == 0 {
		pollInterval = defaultVerificationPoll
	}
	if pollInterval < 0 {
		return VerificationResult{}, errors.New("verification poll interval must not be negative")
	}
	verificationContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	submissionID, alreadyKnown, err := submitVerification(verificationContext, client, explorerURL, canonicalNetwork.ChainID, canonicalAddress, request.metadata, d.config.Explorer.APIKey)
	if err != nil {
		return VerificationResult{}, err
	}
	if alreadyKnown {
		return VerificationResult{AlreadyKnown: true}, nil
	}
	for {
		status, err := checkVerification(verificationContext, client, explorerURL, canonicalNetwork.ChainID, submissionID, d.config.Explorer.APIKey)
		if err != nil {
			return VerificationResult{SubmissionID: submissionID}, err
		}
		switch status {
		case verificationPassed:
			return VerificationResult{SubmissionID: submissionID}, nil
		case verificationFailed:
			return VerificationResult{SubmissionID: submissionID}, errors.New("explorer verification failed")
		}
		select {
		case <-verificationContext.Done():
			return VerificationResult{SubmissionID: submissionID}, errors.New("explorer verification remained pending until timeout")
		case <-time.After(pollInterval):
		}
	}
}

// BuildVerificationMetadata snapshots compiler input and constructor arguments for a confirmed deployment.
func BuildVerificationMetadata(artifactPath, sourceRoot string, network domain.Network, address string, input *domain.DeploymentInput) (VerificationMetadata, error) {
	if input == nil {
		return VerificationMetadata{}, errors.New("deployment input is required")
	}
	canonicalNetwork, err := canonicalVerificationNetwork(network)
	if err != nil {
		return VerificationMetadata{}, err
	}
	if err := input.Validate(); err != nil {
		return VerificationMetadata{}, fmt.Errorf("validate deployment input: %w", err)
	}
	canonicalAddress, err := normalizeAddress(address)
	if err != nil {
		return VerificationMetadata{}, err
	}
	if sourceRoot == "" {
		return VerificationMetadata{}, errors.New("source root is required")
	}
	contents, err := os.ReadFile(artifactPath)
	if err != nil {
		return VerificationMetadata{}, fmt.Errorf("read deployment artifact: %w", err)
	}
	var artifact struct {
		ABI         json.RawMessage `json:"abi"`
		RawMetadata string          `json:"rawMetadata"`
		Bytecode    struct {
			Object string `json:"object"`
		} `json:"bytecode"`
	}
	if err := json.Unmarshal(contents, &artifact); err != nil {
		return VerificationMetadata{}, errors.New("parse deployment artifact")
	}
	if err := validateCreationBytecode(artifact.Bytecode.Object); err != nil {
		return VerificationMetadata{}, err
	}
	if artifact.RawMetadata == "" {
		return VerificationMetadata{}, errors.New("deployment artifact has no compiler metadata")
	}
	parsedABI, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
	if err != nil {
		return VerificationMetadata{}, errors.New("parse deployment ABI")
	}
	var build struct {
		Compiler struct {
			Version string `json:"version"`
		} `json:"compiler"`
		Language string          `json:"language"`
		Settings json.RawMessage `json:"settings"`
		Sources  map[string]struct {
			Keccak256 string `json:"keccak256"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(artifact.RawMetadata), &build); err != nil {
		return VerificationMetadata{}, errors.New("parse compiler metadata")
	}
	if build.Compiler.Version == "" || build.Language == "" || len(build.Settings) == 0 || len(build.Sources) == 0 {
		return VerificationMetadata{}, errors.New("deployment artifact has incomplete compiler metadata")
	}
	sources := make(map[string]struct {
		Content string `json:"content"`
	}, len(build.Sources))
	for sourceName, sourceMetadata := range build.Sources {
		sourcePath, err := verificationSourcePath(sourceRoot, sourceName)
		if err != nil {
			return VerificationMetadata{}, err
		}
		source, err := os.ReadFile(sourcePath)
		if err != nil {
			return VerificationMetadata{}, fmt.Errorf("read compiler source %q: %w", sourceName, err)
		}
		if err := verifySourceHash(sourceName, source, sourceMetadata.Keccak256); err != nil {
			return VerificationMetadata{}, err
		}
		sources[sourceName] = struct {
			Content string `json:"content"`
		}{Content: string(source)}
	}
	settings, contractName, err := verificationCompilerSettings(build.Settings)
	if err != nil {
		return VerificationMetadata{}, err
	}
	compilerInput, err := json.Marshal(struct {
		Language string `json:"language"`
		Sources  map[string]struct {
			Content string `json:"content"`
		} `json:"sources"`
		Settings json.RawMessage `json:"settings"`
	}{Language: build.Language, Sources: sources, Settings: settings})
	if err != nil {
		return VerificationMetadata{}, fmt.Errorf("encode compiler input: %w", err)
	}
	constructorArguments, err := parsedABI.Constructor.Inputs.Pack(input.Token, new(big.Int).SetUint64(uint64(input.FeeBps)), input.Owner, input.FeeWallet, deploymentRecipients(input.Recipients))
	if err != nil {
		return VerificationMetadata{}, fmt.Errorf("encode constructor arguments: %w", err)
	}
	return VerificationMetadata{Network: canonicalNetwork.Name, Address: canonicalAddress, ContractName: contractName, CompilerVersion: build.Compiler.Version, CompilerInput: compilerInput, ConstructorArguments: hex.EncodeToString(constructorArguments)}, nil
}

func validateCreationBytecode(artifactBytecode string) error {
	artifactBytes, err := hex.DecodeString(strings.TrimPrefix(artifactBytecode, "0x"))
	if err != nil || len(artifactBytes) == 0 {
		return errors.New("deployment artifact has invalid creation bytecode")
	}
	bindingBytecode, err := hex.DecodeString(strings.TrimPrefix(bindings.USDTRewardDistributorBin, "0x"))
	if err != nil || len(bindingBytecode) == 0 {
		return errors.New("embedded deployment binding has invalid creation bytecode")
	}
	if !bytes.Equal(artifactBytes, bindingBytecode) {
		return errors.New("deployment artifact creation bytecode does not match embedded binding")
	}
	return nil
}

// VerificationMetadataPath returns the network- and contract-scoped metadata path.
func VerificationMetadataPath(directory string, network domain.Network, address string) (string, error) {
	canonicalNetwork, err := canonicalVerificationNetwork(network)
	if err != nil {
		return "", err
	}
	canonicalAddress, err := normalizeAddress(address)
	if err != nil {
		return "", err
	}
	if directory == "" {
		return "", errors.New("verification metadata directory is required")
	}
	return filepath.Join(directory, canonicalNetwork.Name+"-"+strings.TrimPrefix(canonicalAddress, "0x")+".json"), nil
}

// SaveVerificationMetadata atomically stores one deployment's verification provenance.
func SaveVerificationMetadata(path string, metadata *VerificationMetadata) error {
	if err := validateVerificationMetadata(metadata); err != nil {
		return err
	}
	contents, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode verification metadata: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create verification metadata directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".verification-*.json")
	if err != nil {
		return fmt.Errorf("create temporary verification metadata: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary verification metadata: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary verification metadata permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary verification metadata: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace verification metadata: %w", err)
	}
	return nil
}

// LoadVerificationMetadata loads saved provenance for a later verification-only retry.
func LoadVerificationMetadata(path string) (VerificationMetadata, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return VerificationMetadata{}, fmt.Errorf("read verification metadata: %w", err)
	}
	var metadata VerificationMetadata
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metadata); err != nil {
		return VerificationMetadata{}, errors.New("decode verification metadata")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return VerificationMetadata{}, errors.New("decode verification metadata")
	}
	if err := validateVerificationMetadata(&metadata); err != nil {
		return VerificationMetadata{}, err
	}
	return metadata, nil
}

type verificationStatus uint8

const (
	verificationPending verificationStatus = iota
	verificationPassed
	verificationFailed
)

type explorerResponse struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

func submitVerification(ctx context.Context, client *http.Client, explorerURL string, chainID uint64, address string, metadata *VerificationMetadata, apiKey string) (submissionID string, alreadyKnown bool, err error) {
	form := url.Values{"apikey": {apiKey}, "chainid": {fmt.Sprint(chainID)}, "module": {"contract"}, "action": {"verifysourcecode"}, "contractaddress": {address}, "sourceCode": {string(metadata.CompilerInput)}, "contractname": {metadata.ContractName}, "compilerversion": {compilerVersion(metadata.CompilerVersion)}, "codeformat": {"solidity-standard-json-input"}, "constructorArguments": {metadata.ConstructorArguments}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, explorerURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", false, errors.New("create explorer verification request")
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return "", false, errors.New("submit explorer verification request")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", false, errors.New("explorer verification submission returned an unsuccessful status")
	}
	decoded, err := decodeExplorerResponse(response.Body)
	if err != nil {
		return "", false, err
	}
	if isAlreadyVerified(decoded.Result) {
		return "", true, nil
	}
	if decoded.Status != "1" || decoded.Result == "" {
		return "", false, errors.New("explorer verification submission failed")
	}
	return decoded.Result, false, nil
}

func checkVerification(ctx context.Context, client *http.Client, explorerURL string, chainID uint64, submissionID, apiKey string) (verificationStatus, error) {
	endpoint, err := url.Parse(explorerURL)
	if err != nil {
		return verificationFailed, errors.New("invalid explorer URL")
	}
	query := endpoint.Query()
	query.Set("apikey", apiKey)
	query.Set("chainid", fmt.Sprint(chainID))
	query.Set("module", "contract")
	query.Set("action", "checkverifystatus")
	query.Set("guid", submissionID)
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return verificationFailed, errors.New("create explorer verification status request")
	}
	response, err := client.Do(request)
	if err != nil {
		return verificationFailed, errors.New("check explorer verification status")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return verificationFailed, errors.New("explorer verification status returned an unsuccessful status")
	}
	decoded, err := decodeExplorerResponse(response.Body)
	if err != nil {
		return verificationFailed, err
	}
	return explorerVerificationStatus(decoded.Result), nil
}

func decodeExplorerResponse(reader io.Reader) (explorerResponse, error) {
	var response explorerResponse
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&response); err != nil {
		return explorerResponse{}, errors.New("decode explorer verification response")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return explorerResponse{}, errors.New("decode explorer verification response")
	}
	return response, nil
}

func explorerVerificationStatus(result string) verificationStatus {
	lower := strings.ToLower(result)
	if strings.Contains(lower, "pass - verified") || isAlreadyVerified(result) {
		return verificationPassed
	}
	if strings.Contains(lower, "pending") || strings.Contains(lower, "queue") || strings.Contains(lower, "in progress") {
		return verificationPending
	}
	return verificationFailed
}

func isAlreadyVerified(result string) bool {
	return strings.Contains(strings.ToLower(result), "already verified")
}

func compilerVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func canonicalVerificationNetwork(network domain.Network) (domain.Network, error) {
	canonical, err := domain.ParseNetwork(network.Name)
	if err != nil {
		return domain.Network{}, err
	}
	if network != canonical {
		return domain.Network{}, errors.New("verification network does not match its canonical chain")
	}
	return canonical, nil
}

func normalizeAddress(value string) (string, error) {
	address, err := domain.ParseAddress(value)
	if err != nil || address == (common.Address{}) {
		return "", fmt.Errorf("invalid contract address %q", value)
	}
	return address.String(), nil
}

func verificationSourcePath(root, sourceName string) (string, error) {
	cleanSource := filepath.Clean(sourceName)
	if filepath.IsAbs(cleanSource) || cleanSource == ".." || strings.HasPrefix(cleanSource, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid compiler source path %q", sourceName)
	}
	return filepath.Join(root, cleanSource), nil
}

func verificationContractName(settings json.RawMessage) (string, error) {
	var decoded struct {
		CompilationTarget map[string]string `json:"compilationTarget"`
	}
	if err := json.Unmarshal(settings, &decoded); err != nil {
		return "", errors.New("parse compiler settings")
	}
	if len(decoded.CompilationTarget) != 1 {
		return "", errors.New("compiler metadata must have one compilation target")
	}
	for source, contract := range decoded.CompilationTarget {
		if source == "" || contract == "" {
			return "", errors.New("compiler metadata has an invalid compilation target")
		}
		return source + ":" + contract, nil
	}
	return "", errors.New("compiler metadata must have one compilation target")
}

func verificationCompilerSettings(settings json.RawMessage) (compilerSettings json.RawMessage, contractName string, err error) {
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(settings, &decoded); err != nil {
		return nil, "", errors.New("parse compiler settings")
	}
	contractName, err = verificationContractName(settings)
	if err != nil {
		return nil, "", err
	}
	delete(decoded, "compilationTarget")
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return nil, "", errors.New("encode compiler settings")
	}
	return encoded, contractName, nil
}

func verifySourceHash(sourceName string, source []byte, expected string) error {
	expectedHash, err := hex.DecodeString(strings.TrimPrefix(expected, "0x"))
	if err != nil || len(expectedHash) != 32 {
		return fmt.Errorf("compiler source %q has an invalid keccak256", sourceName)
	}
	if !bytes.Equal(crypto.Keccak256(source), expectedHash) {
		return fmt.Errorf("compiler source %q does not match its recorded keccak256", sourceName)
	}
	return nil
}

func validateVerificationMetadata(metadata *VerificationMetadata) error {
	if metadata == nil {
		return errors.New("verification metadata is required")
	}
	if _, err := domain.ParseNetwork(metadata.Network); err != nil {
		return err
	}
	if _, err := normalizeAddress(metadata.Address); err != nil {
		return err
	}
	if metadata.ContractName == "" || metadata.CompilerVersion == "" || len(metadata.CompilerInput) == 0 {
		return errors.New("verification metadata is incomplete")
	}
	if _, err := hex.DecodeString(metadata.ConstructorArguments); err != nil {
		return errors.New("verification metadata has invalid constructor arguments")
	}
	var compilerInput struct {
		Language string                     `json:"language"`
		Sources  map[string]any             `json:"sources"`
		Settings map[string]json.RawMessage `json:"settings"`
	}
	if err := json.Unmarshal(metadata.CompilerInput, &compilerInput); err != nil || compilerInput.Language == "" || len(compilerInput.Sources) == 0 || len(compilerInput.Settings) == 0 {
		return errors.New("verification metadata has invalid compiler input")
	}
	if _, found := compilerInput.Settings["compilationTarget"]; found {
		return errors.New("verification metadata has invalid compiler input")
	}
	return nil
}
