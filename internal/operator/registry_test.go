package operator_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	internal "pooler/internal/entities"
	"pooler/internal/operator"
)

const contractAddress = "0x1111111111111111111111111111111111111111"

func TestLoadContractsMissingRegistryIsEmpty(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "contracts.json")

	// when
	contracts, err := operator.LoadContracts(path)
	// then
	if err != nil {
		t.Fatalf("LoadContracts() error = %v", err)
	}
	if len(contracts) != 0 {
		t.Fatalf("LoadContracts() = %#v, want empty registry", contracts)
	}
}

func TestSaveAndLoadContractsPreserveLabelsAndNetworkIsolation(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "contracts.json")
	contracts := []internal.Contract{
		{
			Name:        "A long pool name with spaces that is intentionally duplicated",
			Description: "Ethereum deployment",
			Network:     " Ethereum ",
			Address:     "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		},
		{
			Name:        "A long pool name with spaces that is intentionally duplicated",
			Description: "Arbitrum deployment",
			Network:     "arbitrum",
			Address:     "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	// when
	err := operator.SaveContracts(path, contracts)
	loaded, loadErr := operator.LoadContracts(path)

	// then
	if err != nil {
		t.Fatalf("SaveContracts() error = %v", err)
	}
	if loadErr != nil {
		t.Fatalf("LoadContracts() error = %v", loadErr)
	}
	want := []internal.Contract{
		{
			Name:        contracts[0].Name,
			Description: contracts[0].Description,
			Network:     "ethereum",
			Address:     "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			Name:        contracts[1].Name,
			Description: contracts[1].Description,
			Network:     "arbitrum",
			Address:     "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	if !reflect.DeepEqual(loaded, want) {
		t.Fatalf("LoadContracts() = %#v, want %#v", loaded, want)
	}
}

func TestRegisterContractRejectsDuplicateCanonicalIdentity(t *testing.T) {
	// given
	contracts := []internal.Contract{{Network: "ethereum", Address: contractAddress}}

	// when
	registered, err := operator.RegisterContract(contracts, internal.Contract{
		Name:    "a different label",
		Network: "ETHEREUM",
		Address: "0x1111111111111111111111111111111111111111",
	})

	// then
	if err == nil {
		t.Fatalf("RegisterContract() = %#v, want duplicate identity error", registered)
	}
}

func TestRegisterContractAtPathPreservesConcurrentEntries(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "contracts.json")
	contracts := []internal.Contract{
		{Name: "ethereum", Network: "ethereum", Address: "0x1111111111111111111111111111111111111111"},
		{Name: "arbitrum", Network: "arbitrum", Address: "0x2222222222222222222222222222222222222222"},
	}
	commands := make([]*exec.Cmd, len(contracts))
	for index, contract := range contracts {
		commands[index] = exec.Command(os.Args[0], "-test.run=TestRegisterContractAtPathHelper")
		commands[index].Env = append(os.Environ(), "POOLER_TEST_REGISTRY="+path, "POOLER_TEST_NETWORK="+contract.Network, "POOLER_TEST_ADDRESS="+contract.Address)
		require.NoError(t, commands[index].Start())
	}

	// when
	for _, command := range commands {
		require.NoError(t, command.Wait())
	}
	registered, err := operator.LoadContracts(path)

	// then
	require.NoError(t, err)
	require.ElementsMatch(t, contracts, registered)
}

func TestRegisterContractAtPathHelper(t *testing.T) {
	path := os.Getenv("POOLER_TEST_REGISTRY")
	if path == "" {
		return
	}
	require.NoError(t, operator.RegisterContractAtPath(path, internal.Contract{
		Name:    os.Getenv("POOLER_TEST_NETWORK"),
		Network: os.Getenv("POOLER_TEST_NETWORK"),
		Address: os.Getenv("POOLER_TEST_ADDRESS"),
	}))
}

func TestLoadContractsRejectsMalformedRegistry(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "contracts.json")
	if err := os.WriteFile(path, []byte(`[{"network":"ethereum","address":"`+contractAddress+`","secret":"no"}]`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// when
	contracts, err := operator.LoadContracts(path)

	// then
	if err == nil {
		t.Fatalf("LoadContracts() = %#v, want malformed registry error", contracts)
	}
}

func TestRegistryRejectsControlLabelsWithoutEchoingThem(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "contracts.json")
	const maliciousName = "Pool\\u001b]8;;https://example.invalid\\u0007"
	require.NoError(t, os.WriteFile(path, []byte(`[{"name":"`+maliciousName+`","description":"safe","network":"ethereum","address":"`+contractAddress+`"}]`), 0o600))

	// when
	_, loadErr := operator.LoadContracts(path)
	_, registerErr := operator.RegisterContract(nil, internal.Contract{
		Name:        "safe",
		Description: "description\nwith control",
		Network:     "ethereum",
		Address:     contractAddress,
	})
	registered, validErr := operator.RegisterContract(nil, internal.Contract{
		Name:        "Récompenses — pool",
		Description: "Normal Unicode description",
		Network:     "ethereum",
		Address:     contractAddress,
	})

	// then
	require.EqualError(t, loadErr, "validate contract 0: contract name contains control characters")
	require.NotContains(t, loadErr.Error(), "example.invalid")
	require.EqualError(t, registerErr, "validate contract 0: contract description contains control characters")
	require.NotContains(t, registerErr.Error(), "description\nwith control")
	require.NoError(t, validErr)
	require.Equal(t, "Récompenses — pool", registered[0].Name)
}

func TestSaveContractsKeepsExistingRegistryWhenTemporaryWriteFails(t *testing.T) {
	// given
	directory := t.TempDir()
	path := filepath.Join(directory, "contracts.json")
	original := []byte("[\n  {\"name\": \"existing\", \"description\": \"kept\", \"network\": \"ethereum\", \"address\": \"" + contractAddress + "\"}\n]\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chmod(directory, 0o500); err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(directory, 0o700); err != nil {
			t.Errorf("restore directory permissions: %v", err)
		}
	})

	// when
	err := operator.SaveContracts(path, []internal.Contract{{Network: "arbitrum", Address: contractAddress}})

	// then
	if err == nil {
		t.Fatal("SaveContracts() succeeded despite an unwritable registry directory")
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatalf("restore directory permissions: %v", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("registry = %q, want unchanged %q", got, original)
	}
}
