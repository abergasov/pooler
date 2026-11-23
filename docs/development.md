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
compiler warnings. No separate Solidity analyzer is configured.

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
