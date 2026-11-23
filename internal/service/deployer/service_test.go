package deployer

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	domain "pooler/internal/entities"
)

func TestValidateDeploymentTokenMatchesNetwork(t *testing.T) {
	input := domain.DeploymentInput{}
	input.Token = mustAddress(t, "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9")
	arbitrum, err := domain.ParseNetwork("arbitrum")
	require.NoError(t, err)

	require.NoError(t, validateDeploymentToken(arbitrum, &input))

	ethereum, err := domain.ParseNetwork("ethereum")
	require.NoError(t, err)
	require.ErrorContains(t, validateDeploymentToken(ethereum, &input), "does not match the supported USDT address")
}

func mustAddress(t *testing.T, value string) common.Address {
	t.Helper()
	address, err := domain.ParseAddress(value)
	require.NoError(t, err)
	return address
}
