# Development

## Solidity toolchain

Install the pinned Foundry toolchain using the [README](../README.md#installation).
`foundry.toml` manages Solidity `0.8.28`. No Solidity libraries are currently
required or installed; add a library only with an exact version as part of the
task that introduces its use.

## Solidity commands

Run these commands from the repository root:

```sh
forge build
forge fmt --check
forge build --deny-warnings
```

The final command is the current static compiler check: it fails on Solidity
compiler warnings.

## Lint

Install the pinned Solidity linter with Node.js 24 and npm, then run both linters:

```sh
npm ci --ignore-scripts
make lint
```

The target runs `go tool golangci-lint run ./...` using `.golangci.yaml`, then
the locally installed Solhint against `contracts/**/*.sol` using `.solhint.json`.
Solhint uses the recommended rules: errors fail the command; warnings are reported
without failing it. Constructor visibility is ignored because Solidity 0.8.28
does not require it. Use `make lint-go` or `make lint-solidity` to run one language.
CI runs `make --keep-going lint` so both linters report results even if one fails.

## Test entry point

Run the complete Forge suite, including the pinned Ethereum and Arbitrum fork
tests, with:

```sh
make test
```

The target uses credential-free public read-only RPC defaults. It does not skip
fork tests: the command fails if either endpoint cannot serve the pinned blocks.
Use archive-capable endpoints when necessary, without placing credentials in the
Makefile or command output:

```sh
ETHEREUM_RPC_URL=https://your-ethereum-archive-rpc \
ARBITRUM_RPC_URL=https://your-arbitrum-archive-rpc \
make test
```

## Local integration node

Start Anvil in one terminal for future Go CLI integration tests:

```sh
anvil --chain-id 31337 --block-time 1
```

It listens on `http://127.0.0.1:8545` by default. Future integration tests
should use that endpoint explicitly and must not target a live network.

## Operator configuration

Copy `config.example.yml` to the ignored `config.yml`. It combines required RPC
endpoint lists with required signing and explorer credentials:

```yaml
rpc:
  eth:
    - https://ethereum-rpc-1.example
  arb:
    - https://arbitrum-rpc-1.example
signing:
  privateKey: ""
explorer:
  apiKey: ""
```

After valid command flags are parsed, Pooler reads and validates this complete
file once before contract selection or command execution. Both `rpc.eth` and
`rpc.arb` need at least one nonblank endpoint, and both credential fields are
required for every command. Do not commit the runtime file: its endpoint URLs
can contain credentials.

## Explorer verification

A confirmed deployment stores its compiler input, compiler version, and encoded
constructor arguments in `verification/<network>-<address>.json`. This metadata
is separate from `contracts.json` and contains no credentials. To retry only
source verification for an existing contract, use:

```sh
make verify NETWORK=arbitrum CONTRACT=0x...
```

Verification failure or a pending timeout does not send another deployment
transaction.
