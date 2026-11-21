# Design

## Context

The repository currently contains agent instructions, OpenSpec setup, and a minimal
`go.mod`, with no contract, executable CLI, Makefile, tests, or deployment tooling. See
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
- Provide a Go CLI with a small Makefile interface for independently deployed
  contracts, using the chain as the source of current configuration.

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

### Go CLI with Makefile defaults

Implement the operator commands in Go within the existing module. Makefile targets
forward arguments to that CLI rather than containing RPC, signing, JSON, or menu
logic. Define `NETWORK ?= arbitrum` and `INPUT ?= input.json` at the top of the
Makefile. Pass paths as single arguments so input paths containing spaces work.
Invocation overrides take precedence. No separate state daemon is needed.

| Target | Behavior | Input |
| --- | --- | --- |
| `make deploy` | Deploy, register, then verify source | Effective `NETWORK` and `INPUT` |
| `make set-distribution` | Replace the full recipient list | `INPUT`, selected or explicit contract |
| `make show` | Read current configuration and USDT balance | Selected or explicit contract |
| `make distribute` | Call one contract's `distribute()` | Selected or explicit contract |
| `make distribute-all` | Visit known contracts of the selected network sequentially | Registry and `NETWORK` |
| `make verify` | Retry explorer source verification without deploy | Selected or explicit contract, build metadata |

`CONTRACT=<address>` bypasses interactive selection for single-contract commands.
Otherwise, a terminal menu filters the registry by network and supports arrow keys
and Enter for one selection. Display long names and descriptions as labels, with
addresses to distinguish duplicates. Cancelling sends nothing. Missing terminal
input must fail promptly with guidance to supply `CONTRACT`; never guess a target.
Reuse a small suitable terminal component at implementation time rather than
building a full terminal UI framework.

### Registry metadata and operation inputs

Use `contracts.json` as a JSON array of objects with `name`, `description`, `network`,
and `address`. Names need no slug and are not unique identifiers. Validate addresses
and use normalized network/address pairs for uniqueness; preserve human-readable
labels. Missing registry on the first deployment means an empty registry, whereas
malformed existing contents are an error. Write updates using a temporary file and
replacement so a failed save does not truncate the previous registry. No database,
discovery service, or local mirror of allocation state is needed.

Resolve the canonical network name to its expected chain ID and RPC settings, and
check the RPC chain ID before signing. An explicit address can target an existing
instance without first adding it to the registry. Registry operations and batch
selection remain scoped to the effective network.

Use one operation-input JSON format with a `recipients` array of address/shareBps
pairs. Deployment additionally requires `name`, `description`, `owner`, `feeWallet`,
`feeBps`, and `token`. Recipient replacement consumes and validates only `recipients`
and never attempts to change other parameters from that file. The Makefile selects
the network; it is not duplicated in the operation input. These files are requests,
not a mandatory `pools/` directory or the authority for current configuration.

### Read-only inspection and local signing

Expose contract getters for recipients and shares, owner, fee wallet, fee basis
points, and token. Read these plus the token balance through RPC using the ABI
produced by the build. The `show` command must not fetch its ABI from an explorer
or need an explorer API key. Format shares as decimal percentages for the user,
while retaining integer basis points in transaction inputs.

Load `.env` in the Go process as configuration data, not as shell commands. Include
the signing private key, per-network RPC endpoints, and the explorer API key in its
documented settings. Add `.env` to `.gitignore` and a secret-free `.env.example`
during implementation. Do not pass secrets as Make arguments or echo them in
commands, error messages, or diagnostics. Public reads need only RPC configuration;
load and require signing/explorer credentials only for operations that use them.

Before an existing-contract mutation, read its owner and compare it with the signer.
Show the chosen network and contract in output. Validate shared settings before
starting a batch, then handle contract-specific problems as individual outcomes.

### Confirmed deployment and independently retryable verification

Validate deployment input and registry readability before broadcasting. Wait for a
successful deployment receipt and verify deployed code, then register the address
before starting explorer verification. Report the address and transaction hash
even if a subsequent registry save or verification fails; never turn either failure
into an automatic second deployment.

Preserve the exact compiler input, version, settings, and encoded constructor
arguments in build/deployment artifacts sufficient for later `make verify`. They
describe deployment provenance, not current participant configuration, and contain
no private keys. Register only the four agreed identity/description fields in
`contracts.json`; do not use it for compiler metadata or live allocations.

Submit the matching source/build input to the explorer for the selected chain and
check the result of verification. A submission identifier alone is not success.
If verification fails or remains pending at timeout, keep the confirmed deployment,
report its status separately, and provide the retry command. The command returns
a nonzero outcome when registration or required verification remains unsuccessful.
An already verified matching contract counts as successful verification.

The protocol distinction is documented by
[Ethereum's read-only RPC](https://ethereum.org/developers/docs/apis/json-rpc/)
and [Etherscan's verification workflow](https://docs.etherscan.io/contract-verification/verify-with-api).
Read-only methods work independently of verification; publishing matching source
also enables explorer inspection. Select supported API details during implementation.

### Sequential distribution and terminal outcomes

Use one sequential loop over the selected network's registry entries. Read each
balance, skip empty contracts, and submit owner-authorized distribution for each
funded contract. Track transaction confirmation before reporting `OK`; retain the
transaction hash if confirmation cannot be established, and do not label an
unknown outcome as a confirmed revert or automatically resubmit it.

Print a result per contract and continue past individual errors, for example:

```text
pool A: OK
pool B: distribution err: ABC
pool C: SKIPPED: empty balance
pool D: OK
```

Include address context where names collide, and make submitted transaction hashes
available in the output. Return nonzero if any entry has an error; successful and
skipped entries alone produce zero. With no matching entries, report that there is
nothing to process. This is a sequence of independent transactions, not an atomic
operation across pools. No concurrency, retry queue, or cross-pool rollback layer
is needed. Individual direct `distribute()` calls retain the on-chain empty-balance
error; batch skipping is an operator-tool behavior.

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
- One signing key can control several independent pools → Keep it out of Git and
  output, and check current ownership before mutations; compromise affects every
  pool owned by that address.
- Deploy can succeed while local registration or explorer verification fails →
  Preserve existing registry data, report the deployed address/hash, and retry only
  the failed local or verification step without sending a new deployment.
- A batch can partially succeed → Print per-contract results and an aggregate exit
  status; never imply that successful independent transactions were rolled back.

## Migration Plan

There is no existing contract or data migration. After implementation is separately
authorized, verify the selected USDT deployment, run focused tests and full-size
gas checks, and deploy with validated initial configuration. Register the confirmed
deployment, verify its source, and inspect live configuration before funding. If a
deployed version must be replaced, the owner
can withdraw USDT and fund a new deployment; no proxy upgrade is available.

## Open Questions

- Which USDT address and RPC endpoint will each supported network use? Verify them
  before integration tests and deployment. Arbitrum is the default network, but no
  token address or RPC endpoint is asserted in this change.
- Which contract compiler, test runner, and compatible library versions will be
  used? Select them during implementation setup and pin reproducible versions.

The unmeasured 300-recipient gas requirement is an outstanding verification item,
not permission to change the agreed behavior or a claim of deployment readiness.
