# USDT deployment verification

This document records the compatibility checks for the two exact production
deployments selected by `define-usdt-reward-distributor`. It does not make a
claim about tokens with the same ticker on any other chain.

## Authoritative deployment references

| Network | Chain ID | Token | Address | Authoritative reference | Claim used here |
| --- | --- | --- | --- | --- | --- |
| Ethereum mainnet | 1 | USD₮ (USDT) | `0xdAC17F958D2ee523a2206206994597C13D831ec7` | [Tether supported protocols](https://tether.to/en/supported-protocols/) | Tether identifies this as its Ethereum USD₮ contract and says its older ERC-20 implementation does not return a Boolean from `transfer`. |
| Arbitrum One | 42161 | USD₮0 (USDT0) | `0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9` | [USDT0 deployments](https://docs.usdt0.to/technical-documentation/deployments) | USDT0 identifies this address as the Arbitrum One token deployment. |

[The USDT0 developer guide](https://docs.usdt0.to/technical-documentation/developer/)
states that USDT0 has six decimals. Tether's Ethereum integration guidance is
the authoritative source for the Ethereum no-Boolean-return compatibility
requirement. The contract tests verify both claims on the deployed code instead
of inferring behavior from the ticker.

## Pinned RPC evidence

Read-only JSON-RPC checks completed on 2026-11-22 used these pinned blocks:

| Network | Block | Hex block | RPC observations |
| --- | ---: | --- | --- |
| Ethereum mainnet | 26,469,961 | `0x193e649` | `eth_chainId` returned `0x1`; `eth_getCode` for the specified address was non-empty; `decimals()` returned `0x...06`; an `eth_call` of `transfer(owner, 0)` from the token owner returned `0x`. |
| Arbitrum One | 526,486,711 | `0x1f618cb7` | `eth_chainId` returned `0xa4b1`; `eth_getCode` for the specified address was non-empty; `decimals()` returned `0x...06`; a zero-value `transfer(owner, 0)` call returned a 32-byte ABI-encoded `true`. |

The Ethereum transfer probe was a read-only `eth_call`; it did not broadcast a
transaction or modify live state. The Arbitrum transfer is exercised in the
fork test below and must return the observed 32-byte ABI-encoded `true`; empty
return data or a reported `false` is rejected.

## Reproduction

The test is self-contained and does not rely on a token mock or a third-party
Solidity library. It first creates each pinned fork, checks its chain ID,
checks that the exact address has deployed code, and checks `decimals() == 6`.
It then impersonates the on-chain token owner in the pinned fork and executes a
zero-value self-transfer through the deployed token's `transfer`
implementation. The probe exercises the actual return-data behavior without
requiring a live transaction or altering either live network.

Provide read-only archive-capable endpoints and run:

```sh
ETHEREUM_RPC_URL=https://your-ethereum-rpc \
ARBITRUM_RPC_URL=https://your-arbitrum-rpc \
forge test --match-path test/USDTDeployment.t.sol -vvv
```

`testEthereumUSDTTransferSucceedsWithoutBooleanReturnValue` additionally
asserts that the successful Ethereum transfer returns exactly zero bytes. This
is deliberately a low-level call: a Solidity call compiled as `returns (bool)`
would attempt to decode return data that the deployed Ethereum token does not
provide.

On 2026-11-22, the command above completed successfully with Foundry v1.0.0:
four tests passed (both chain/code/decimal checks and both transfer checks).
