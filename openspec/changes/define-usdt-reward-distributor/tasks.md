# Tasks

These tasks describe future implementation. They remain unchecked because only
specification and planning work is authorized at this stage.

## 1. Implementation setup

- [ ] 1.1 Select and pin a minimal contract compiler, test runner, and required library versions; verify a clean build and document reproducible build, test, formatting, and static-check commands.
- [ ] 1.2 Record the selected chain and verified USDT contract address and transfer behavior, with authoritative sources; verify that integration tests target that exact deployment rather than a ticker-based assumption.

## 2. Deployment and configuration

- [ ] 2.1 Implement validated deployment with fixed USDT address and fee rate, explicit initial owner, independent fee wallet, and initial recipient pairs; verify successful deployment and rejection of invalid addresses, fee bounds, and lists.
- [ ] 2.2 Implement owner-only full-list replacement with 1–300 unique recipients, positive basis-point shares, and an exact 10000 total; verify replacement removes old entries and every invalid update preserves the old configuration.
- [ ] 2.3 Implement immediate owner transfer and independent fee-wallet update; verify unauthorized calls fail, old ownership loses access immediately, and ownership changes leave the fee wallet and list unchanged.

## 3. Payout operations

- [ ] 3.1 Implement owner-only `distribute()` using the specified balance-snapshot equations and compatible USDT transfers; verify the exact, fractional-residue, tiny-balance, zero-fee, and empty-balance scenarios, including a fee wallet that is also a recipient.
- [ ] 3.2 Implement full-balance USDT withdrawal to an owner-selected valid destination without fees or a previous-failure prerequisite; verify zero balance after success, preserved configuration, authorization, invalid destinations, and empty-balance errors.
- [ ] 3.3 Protect payout and configuration invariants during token interactions; verify participant, fee, and withdrawal failures roll back completely, token-reported failures are not ignored, and nested interactions cannot change allocation or spend twice.

## 4. Integration and acceptance

- [ ] 4.1 Verify direct funding, aggregation of deposits, configuration changes after funding, independent deployments, and repeated calls through integration tests; assert that successful payout operations conserve the full starting balance and leave zero USDT.
- [ ] 4.2 Exercise allocation with varied balances, fee boundaries, and valid share lists; verify integer rounding, nonnegative payouts, total conservation, and safe arithmetic across the supported amount range.
- [ ] 4.3 Measure deployment, replacement, and distribution with 300 recipients against the selected network and token, including 300 nonzero transfers to initially empty recipient balances plus a distinct fee wallet; record gas results and confirm transaction limits without silently reducing capacity or batching payouts.
- [ ] 4.4 Run the focused suite, relevant formatting and static checks, and final diff review against the specification; document commands, results, and any unavailable checks, then provide verified deployment parameters and operating instructions without deploying until authorized.
