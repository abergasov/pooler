# Design

## Context

The repository currently contains agent instructions and OpenSpec setup, with no
contract, build toolchain, tests, or deployment configuration. See
[proposal.md](proposal.md) for motivation and
[the specification](specs/usdt-reward-distributor/spec.md) for agreed behavior.

This design is needed because the contract controls funds, performs up to 301
token transfers per distribution, and must preserve exact allocation and rollback
semantics. It describes future implementation; this change does not implement it.

## Goals / Non-Goals

**Goals:**
- Use a single contract with local ownership of configuration and payout logic.
- Make allocation reproducible using integer arithmetic and one balance snapshot.
- Preserve atomicity and a zero USDT balance after successful payout operations.

**Non-Goals:**
- No per-deposit history, participant claim balances, rounding reserves, or
  prerequisite failure state for owner withdrawal.
- No keeper service, deposit-and-distribute entry point, cross-chain messaging,
  proxy, upgrade admin, or ownership acceptance workflow.
- No support for arbitrary assets, native-currency distribution, or recovery of
  unrelated tokens in this version. The full-withdrawal guarantee covers USDT.

## Decisions

### One fixed token and fee rate per deployment

Deployment supplies the USDT token address, fee basis points, initial owner, fee
wallet, and complete initial recipient list. Store the token and fee as immutable
configuration. Validate the token address as a nonzero contract address; the
deployment process must separately verify that it is the intended USDT on the
selected chain. Generic token discovery or switching would add unsupported scope.

Do not infer the intended chain or token address from a ticker. Exact network,
token address, compiler, and test-tool versions are deployment/implementation
inputs still to be selected and checked; they do not change allocation behavior.

### Complete recipient list with integer basis points

Use an enumerable list of `(address, shareBps)` pairs so a distribution can visit
every participant. For example, 12.30% is supplied as 1230, not a floating-point
number. Keep the denominator fixed at 10000.

Replacement validates length, address validity, uniqueness, positive shares, and
the exact total before replacing the active list. Individual add/remove or
share-edit methods are unnecessary and would introduce intermediate invalid
states. The same validation applies at deployment. The owner and fee wallet may
overlap with recipient addresses; uniqueness applies within the recipient list.

### One balance snapshot and exact conservation

At the start of distribution, read the configured token balance once and reject
zero. Use the allocation equations in the specification with integer floor
rounding and arithmetic that does not overflow its supported input domain. Freeze
the effective recipient list and fee destination for the operation through
appropriate control of reentrant state changes, without requiring copies of the
entire list or historical snapshots.

Send each nonzero participant payout, then send the remainder of the original
balance to the fee wallet. The amount owed to that wallet is calculated from the
original balance minus the sum of participant allocations. Do not recompute
participant percentages from a shrinking balance.

This approach assigns every base unit exactly once. At small balances, rounding
can make the actual fee-wallet payout exceed the nominal fee percentage; it may
receive the full balance even when the configured fee rate is zero. This is the
explicitly selected residue rule, not an additional configurable charge.

### Atomic token operations

Use token-transfer handling compatible with the selected USDT deployment and
propagate all failures. Confirm its success/return-value behavior before choosing
the transfer API; a reverted or failed transfer must never be silently accepted.
Skip zero transfers. Check the final USDT balance invariant and revert if it is
not zero rather than reporting a partial success.

Protect payout entry points and configuration from reentrant operations that
could invalidate allocation. Prefer established library primitives available in
the selected toolchain over a custom security layer. Do not catch failures and
continue, create a retry queue, or convert a failed transfer into a pending claim.
The transaction itself supplies rollback of earlier token transfers.

### Direct owner operations

Provide owner-only operations for `distribute()`, full USDT withdrawal, full-list
replacement, fee-wallet update, and immediate owner update. Only `distribute()`
has a user-agreed method name; remaining ABI names and error identifiers can be
chosen consistently during implementation.

Ownership changes in one step and does not redirect fees. No renunciation method
is needed. Full withdrawal takes a destination and transfers the whole balance
without applying shares or a fee. It needs no record of a previous failure.

As boundary-validation defaults, reject zero or self addresses for the owner,
fee wallet, and withdrawal destination. These extend the confirmed recipient
validation to prevent unusable control or destinations that cannot empty the
contract. An update to the existing owner or fee-wallet address can be a harmless
no-op; no pending-owner or confirmation state is introduced.

### Preserve the agreed capacity

Keep the 300-recipient limit and atomic payout requirement explicit. Measure
distribution with 300 nonzero payouts and a distinct fee wallet, including
recipients with no prior token balance. Also check full-list replacement at the
limit. Do not introduce batching or lower the limit unless measurements show a
problem and the specification is explicitly revised.

## Risks / Trade-offs

- Trusted owner can redirect allocations or withdraw all USDT → This is the
  agreed trust model, not a participant custody guarantee.
- One-step ownership transfer can assign an inaccessible address → Reject zero
  and self addresses; require careful address verification before submission.
  No acceptance handshake or recovery authority is promised.
- One failed token transfer blocks all distribution → Preserve atomicity; the
  owner can replace the list, change the fee wallet, or attempt full withdrawal.
  Withdrawal cannot bypass restrictions imposed by the token itself.
- Large lists may exceed a network transaction limit or be expensive → Require
  measured validation of 300-recipient operations before deployment readiness.
- A wrong immutable token address or fee cannot be corrected in place → Verify
  deployment inputs; recovery requires owner withdrawal and a new deployment.
- Funding alone does not trigger payouts → The owner must submit a transaction
  and fund its network execution fee separately.

## Migration Plan

There is no existing contract or data migration. After implementation is separately
authorized, verify the selected USDT deployment, run focused tests and full-size
gas checks, and deploy with validated initial configuration. Verify the deployed
configuration before funding. If a deployed version must be replaced, the owner
can withdraw USDT and fund a new deployment; no proxy upgrade is available.

## Open Questions

- Which exact network and USDT address will the first deployment use? Verify them
  before integration tests and deployment; no address is asserted in this change.
- Which contract compiler, test runner, and compatible library versions will be
  used? Select them during implementation setup and pin reproducible versions.

The unmeasured 300-recipient gas requirement is an outstanding verification item,
not permission to change the agreed behavior or a claim of deployment readiness.
