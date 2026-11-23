# Pooler

Pooler is a future EVM rewards-distribution project. Its contract behavior is
still being specified.

## Installation

Install the prerequisites for your platform.

### macOS

```sh
xcode-select --install
```

### Ubuntu

```sh
sudo apt update
sudo apt install --yes ca-certificates curl git make tar
```

### Arch Linux

```sh
sudo pacman -Syu --needed ca-certificates curl git make tar
```

Arch does not support partial upgrades; do not replace `-Syu` with `-Sy`.

Install the pinned Foundry release, then add it to the current shell's `PATH`:

```sh
curl -L https://foundry.paradigm.xyz | bash
export PATH="$HOME/.foundry/bin:$PATH"
foundryup --install v1.0.0
forge --version
```

Foundry manages Solidity `0.8.28` from [`foundry.toml`](foundry.toml).

Sources: [Foundry installation](https://github.com/foundry-rs/foundryup#usage),
[Apple Command Line Tools](https://developer.apple.com/documentation/xcode/installing-the-command-line-tools),
and [Arch partial-upgrade policy](https://wiki.archlinux.org/title/System_maintenance#Partial_upgrades_are_unsupported).

## Tests

Run the complete Forge suite, including the pinned Ethereum and Arbitrum fork
tests:

```sh
make test
```

The default read-only endpoints are
`https://eth.drpc.org` and
`https://arbitrum-one.public.blastapi.io`. Supply archive-capable endpoints when
they cannot serve the pinned blocks; keep credentials out of the Makefile:

```sh
ETHEREUM_RPC_URL=https://your-ethereum-archive-rpc \
ARBITRUM_RPC_URL=https://your-arbitrum-archive-rpc \
make test
```
