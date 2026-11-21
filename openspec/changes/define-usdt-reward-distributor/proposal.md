# Proposal

## Why

Pooler needs a precise contract for distributing a funded USDT balance among
participants by configured percentages. The owner must be able to perform one
atomic payout, collect a fixed fee plus rounding residue, and recover funds
without distribution when needed. A Go command-line tool and Makefile will let
the operator deploy, inspect, and manage multiple independent instances.

## What Changes

- Define an independently deployed, non-upgradeable USDT distributor for Ethereum
  or Arbitrum, with the token address and fee rate fixed at deployment.
- Accept ordinary USDT funding and distribute the full balance only when the
  owner calls `distribute()`.
- Configure 1–300 unique recipients using address/share pairs with 0.01%
  precision and a total share of exactly 100%.
- Pay participants from the balance after the base fee; send the base fee and all
  rounding residue to a separately configured fee wallet, leaving zero USDT.
- Revert the entire distribution if any required transfer fails.
- Allow only the owner to replace the full recipient list, change the fee wallet,
  transfer ownership in one step, or withdraw all USDT to a chosen address.
- Expose public read-only configuration methods and read live state from the
  blockchain instead of maintaining local copies of recipient configurations.
- Add a Go CLI behind Makefile targets for deployment, recipient replacement,
  inspection, individual and sequential batch distribution, and source verification.
- Default Makefile settings to `NETWORK=arbitrum` and `INPUT=input.json`, allow
  invocation overrides, and offer arrow-key contract selection when no address
  is supplied interactively.
- Register confirmed deployments in `contracts.json` with name, description,
  network, and address; keep names independent from contract identity.
- Load signing and connection secrets from an ignored `.env`, provide a secret-free
  `.env.example`, and verify source on the selected network's explorer after deploy.

## Capabilities

### New Capabilities

- `usdt-reward-distributor`: Deployment configuration, ownership, recipient
  validation, full-balance allocation, fee collection, unrestricted owner
  withdrawal of USDT, public inspection, and operator tooling for independent instances.

### Modified Capabilities

None. The repository has no existing behavior specifications.

## Impact

This change produces planning artifacts only. No contract implementation,
deployment, or executable tests are included at this stage. Future implementation
will require a contract, Go CLI, Makefile, deployment and verification artifacts,
registry, environment examples, regression tests, and a gas check for 300 recipients
against the selected network and USDT deployment. Registry metadata is local;
current contract configuration remains on-chain. Explorer verification failures
must be recoverable without another deployment.

Participants explicitly trust the owner, who may change future allocations or
withdraw all USDT without distributing it. There is no automatic execution on
funding, cross-chain accounting, proxy upgrade, mutable fee rate, or participant
claim mechanism in this scope.
