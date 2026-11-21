# Tasks

These tasks describe future implementation. They remain unchecked because only
specification and planning work is authorized at this stage.

## 1. Implementation setup

- [ ] 1.1 Set up Foundry/Forge for Solidity builds and tests and Anvil for Go CLI integration tests; pin compatible Foundry, Solidity compiler, and required library versions, verify a clean build, and document reproducible build, test, formatting, and static-check commands.
- [ ] 1.2 Verify deployed code, six decimals, and transfer behavior for the specified Ethereum mainnet USD₮ and Arbitrum One USD₮0 addresses; retain the authoritative references and test both exact deployments, including Ethereum USD₮ transfers without a Boolean return value.

## 2. Deployment and configuration

- [ ] 2.1 Implement validated deployment with fixed USDT address and fee rate, explicit initial owner, independent fee wallet, and initial recipient pairs; verify successful deployment and rejection of invalid addresses, fee bounds, and lists.
- [ ] 2.2 Implement owner-only full-list replacement with 1–100 unique recipients, positive basis-point shares, and an exact 10000 total; verify replacement removes old entries and every invalid update preserves the old configuration.
- [ ] 2.3 Implement immediate owner transfer and independent fee-wallet update; verify unauthorized calls fail, old ownership loses access immediately, and ownership changes leave the fee wallet and list unchanged.
- [ ] 2.4 Expose public getters for the complete recipient list, owner, fee wallet, fee rate, and USDT address; verify callers without owner privileges can read initial and updated state without a transaction.

## 3. Payout operations

- [ ] 3.1 Implement owner-only `distribute()` using the specified balance-snapshot equations and compatible USDT transfers; verify the exact, fractional-residue, tiny-balance, zero-fee, and empty-balance scenarios, including a fee wallet that is also a recipient.
- [ ] 3.2 Implement full-balance USDT withdrawal to an owner-selected valid destination without fees or a previous-failure prerequisite; verify zero balance after success, preserved configuration, authorization, invalid destinations, and empty-balance errors.
- [ ] 3.3 Protect payout and configuration invariants during token interactions; verify participant, fee, and withdrawal failures roll back completely, token-reported failures are not ignored, and nested interactions cannot change allocation or spend twice.

## 4. Go operator tooling

- [ ] 4.1 Add a Go CLI and Makefile targets `deploy`, `set-distribution`, `show`, `distribute`, `distribute-all`, and `verify`; verify top-level `NETWORK ?= arbitrum` and `INPUT ?= input.json`, command-line overrides, and paths containing spaces.
- [ ] 4.2 Add operation-input parsing and network/RPC validation, including RPC YAML lists and canonical network mapping (`ethereum` to `eth`, `arbitrum` to `arb`), with basis-point inputs and owner checks before mutations; test ordered first-working endpoint selection with bounded timeouts, skipped unreachable/wrong-chain endpoints, no probes after success, and empty/unusable lists. Verify invalid inputs and wrong signers cannot broadcast, uncertain submissions are not automatically retried, and non-input commands work without `input.json`.
- [ ] 4.3 Load signing/explorer secrets from `.env` and RPC lists from `rpc.yml` as data; provide `.env.example` and `rpc.example.yml`, exclude actual local configuration through `.gitignore`, and verify missing credentials fail only operations that need them and secrets (including credentials in RPC URLs) never appear in command arguments, logs, or error output.
- [ ] 4.4 Add `contracts.json` loading and safe updates using name, description, network, and address; verify long/spaced/duplicate names, duplicate identity rejection, cross-network isolation, first deployment without a registry, and preservation of existing entries on write failure.
- [ ] 4.5 Add the arrow-key/Enter selector for single-contract commands and the explicit `CONTRACT` override; verify network filtering, one selected destination, cancellation, no candidates, and actionable noninteractive errors without unintended transactions.
- [ ] 4.6 Implement deploy with validated parameters, receipt/code confirmation, and registry insertion; verify failed deployment creates no entry and post-deploy save failure exposes the address/hash without triggering redeployment.
- [ ] 4.7 Implement live `show`, recipient replacement, and individual distribution; verify display reflects external on-chain changes, needs no private/explorer key, replacement changes only shares/recipients, and submitted transactions are reported with their actual outcomes.
- [ ] 4.8 Implement sequential `distribute-all` and per-contract terminal results; verify empty balances send no transaction, one failed contract does not stop later contracts, duplicate names remain identifiable, hashes are available, and exit status reflects partial failure.
- [ ] 4.9 Preserve reproducible build and constructor metadata and implement deploy-time explorer verification plus `make verify`; verify submission is not mistaken for success, wrong-chain verification is prevented, and failure/pending/already-verified cases never cause another deployment.

## 5. Integration and acceptance

- [ ] 5.1 Use Forge integration tests on pinned Ethereum and Arbitrum forks with the specified tokens to verify direct funding, aggregation of deposits, configuration changes after funding, independent deployments, and repeated calls; record block numbers and assert that successful payout operations conserve the full starting balance and leave zero USDT.
- [ ] 5.2 Use Solidity unit, fuzz, and invariant tests in Forge to exercise varied balances, fee boundaries, and valid share lists; verify integer rounding, nonnegative payouts, total conservation, and safe arithmetic across the supported amount range, with controlled token mocks for failure and reentrancy cases.
- [ ] 5.3 Measure deployment, replacement, and distribution with 100 recipients on both pinned network forks using their specified tokens, including 100 nonzero transfers to initially empty recipient balances plus a distinct fee wallet; record gas results and check actual network limits and network-specific RPC estimates with final transactions and suitable contract state before deployment readiness. State fork limitations, including Arbitrum L1 data-posting costs, without silently reducing capacity or batching payouts.
- [ ] 5.4 Exercise the Go CLI against Anvil deployments and a controlled explorer API stub; verify deploy/register/inspect/update/distribute/verify flows and mixed batch outcomes, then record any separately authorized real explorer checks and limitations.
- [ ] 5.5 Run the focused contract and Go tests, relevant formatting/static checks, and final diff review against the specification; document commands, results, Makefile usage, and unavailable checks without deploying to a live network until authorized.
