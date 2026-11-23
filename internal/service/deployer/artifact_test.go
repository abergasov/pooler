package deployer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	domain "pooler/internal/entities"
	"pooler/internal/service/deployer"
)

func TestBuildVerificationMetadataChecksEmbeddedCreationBytecode(t *testing.T) {
	// given
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	artifactPath := filepath.Join(root, "out", "USDTRewardDistributor.sol", "USDTRewardDistributor.json")
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)
	input := verificationInput()

	// when
	metadata, err := deployer.BuildVerificationMetadata(artifactPath, root, network, "0x0000000000000000000000000000000000000005", &input)

	// then
	require.NoError(t, err)
	require.NotEmpty(t, metadata.ConstructorArguments)
}

func TestBuildVerificationMetadataRejectsMismatchedCreationBytecode(t *testing.T) {
	// given
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	contents, err := os.ReadFile(filepath.Join(root, "out", "USDTRewardDistributor.sol", "USDTRewardDistributor.json"))
	require.NoError(t, err)
	var artifact map[string]any
	require.NoError(t, json.Unmarshal(contents, &artifact))
	artifact["bytecode"].(map[string]any)["object"] = "0x00"
	modified, err := json.Marshal(artifact)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "USDTRewardDistributor.json")
	require.NoError(t, os.WriteFile(path, modified, 0o600))
	network, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)
	input := verificationInput()

	// when
	_, err = deployer.BuildVerificationMetadata(path, root, network, "0x0000000000000000000000000000000000000005", &input)

	// then
	require.ErrorContains(t, err, "creation bytecode does not match embedded binding")
}

func verificationInput() domain.DeploymentInput {
	return domain.DeploymentInput{
		Name:        "rewards",
		Description: "quarterly",
		FeeBps:      500,
		Token:       common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9"),
		Owner:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
		FeeWallet:   common.HexToAddress("0x0000000000000000000000000000000000000002"),
		Recipients:  []domain.Recipient{{Address: common.HexToAddress("0x0000000000000000000000000000000000000003"), ShareBps: 10_000}},
	}
}
