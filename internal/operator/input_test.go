package operator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"pooler/internal/operator"
)

func TestParseDeploymentInputValidatesDeploymentFields(t *testing.T) {
	// given
	input := validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`)

	// when
	deployment, err := operator.ParseDeploymentInput(strings.NewReader(input))

	// then
	require.NoError(t, err)
	require.Equal(t, "Reward pool", deployment.Name)
	require.Equal(t, "Quarterly rewards", deployment.Description)
	require.Equal(t, uint16(9999), deployment.FeeBps)
	require.Len(t, deployment.Recipients, 1)
	require.Equal(t, "0x0000000000000000000000000000000000000004", deployment.Recipients[0].Address.String())
}

func TestParseDeploymentInputRejectsInvalidFields(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"missing name": {
			input: strings.Replace(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), `"name":"Reward pool",`, "", 1),
			want:  "name is required",
		},
		"zero owner": {
			input: strings.Replace(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), "0x0000000000000000000000000000000000000001", "0x0000000000000000000000000000000000000000", 1),
			want:  "zero address",
		},
		"fee out of range": {
			input: strings.Replace(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), `"feeBps":9999`, `"feeBps":10000`, 1),
			want:  "feeBps must be less than 10000",
		},
		"null fee": {
			input: strings.Replace(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), `"feeBps":9999`, `"feeBps":null`, 1),
			want:  "feeBps must be an unsigned integer",
		},
		"duplicate recipient": {
			input: validDeploymentInput(`[
                {"address":"0x0000000000000000000000000000000000000004","shareBps":5000},
                {"address":"0x0000000000000000000000000000000000000004","shareBps":5000}
            ]`),
			want: "duplicate recipient",
		},
		"shares do not total 10000": {
			input: validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":9999}]`),
			want:  "recipient shares must total 10000 basis points",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// when
			_, err := operator.ParseDeploymentInput(strings.NewReader(test.input))

			// then
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestParseDeploymentInputRejectsControlLabelsAndPreservesUnicode(t *testing.T) {
	for name, test := range map[string]struct {
		input string
		want  string
	}{
		"name": {
			input: strings.Replace(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), `"name":"Reward pool"`, `"name":"Reward\u001bpool"`, 1),
			want:  "deployment name contains control characters",
		},
		"description": {
			input: strings.Replace(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), `"description":"Quarterly rewards"`, `"description":"Quarterly\u007frewards"`, 1),
			want:  "deployment description contains control characters",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// when
			_, err := operator.ParseDeploymentInput(strings.NewReader(test.input))

			// then
			require.EqualError(t, err, test.want)
		})
	}

	// given
	input := strings.ReplaceAll(validDeploymentInput(`[{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]`), "Reward pool", "Récompenses — pool")

	// when
	deployment, err := operator.ParseDeploymentInput(strings.NewReader(input))

	// then
	require.NoError(t, err)
	require.Equal(t, "Récompenses — pool", deployment.Name)
}

func TestParseDistributionInputConsumesOnlyRecipients(t *testing.T) {
	// given
	input := `{
        "owner": false,
        "feeBps": 10000,
        "recipients": [{"address":"0x0000000000000000000000000000000000000004","shareBps":10000}]
    }`

	// when
	distribution, err := operator.ParseDistributionInput(strings.NewReader(input))

	// then
	require.NoError(t, err)
	require.Len(t, distribution.Recipients, 1)
}

func validDeploymentInput(recipients string) string {
	return `{
        "name":"Reward pool",
        "description":"Quarterly rewards",
        "owner":"0x0000000000000000000000000000000000000001",
        "feeWallet":"0x0000000000000000000000000000000000000002",
        "feeBps":9999,
        "token":"0x0000000000000000000000000000000000000003",
        "recipients":` + recipients + `
    }`
}
