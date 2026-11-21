# USDT Reward Distributor

## Purpose

Define an owner-controlled USDT distributor that pays configured participant
shares and a fixed-rate fee atomically, including deterministic rounding and
full-balance recovery under an explicitly trusted owner.

## ADDED Requirements

### Requirement: Independent and fixed deployment configuration

Each deployment SHALL operate independently on its selected Ethereum or Arbitrum
network with exactly one configured USDT token address. Deployment SHALL set the
initial owner, fee wallet, fee rate, and valid initial recipient list. The token
address and fee rate SHALL be immutable, and the contract SHALL expose no logic
upgrade mechanism. Only ownership, the fee wallet, and the full recipient list
with its shares SHALL be mutable configuration.

#### Scenario: Valid deployment
- **WHEN** the contract is deployed with valid initial configuration
- **THEN** it is ready to receive USDT and execute owner-authorized operations
- **AND** it requires no subsequent initialization to establish a valid list

#### Scenario: Independent deployments
- **WHEN** USDT is funded or configuration is changed in one deployment
- **THEN** balances and configuration in other deployments are unaffected

#### Scenario: Immutable configuration
- **WHEN** an owner wants a different USDT token, fee rate, or contract logic
- **THEN** the existing deployment offers no operation to make that change

### Requirement: Percentage precision and fee bounds

All percentage inputs SHALL use integer basis points: one basis point is 0.01%,
and 10000 basis points is 100%. The fee rate SHALL be between 0 and 9999 basis
points inclusive. Token amounts SHALL use the configured token's smallest unit;
percentage precision SHALL NOT restrict amounts to two decimal places.

#### Scenario: Percentage representation
- **WHEN** a recipient is assigned 12.30%
- **THEN** its share input is 1230 basis points

#### Scenario: Fee bounds
- **WHEN** a deployment supplies a fee rate of 0 or 9999 basis points
- **THEN** the fee rate is accepted if the other configuration is valid
- **WHEN** a deployment supplies a fee rate of 10000 or greater
- **THEN** deployment reverts

### Requirement: Valid complete recipient lists

Recipients SHALL be supplied as a list of address/share pairs containing 1–300
entries. Addresses SHALL be unique, nonzero, and different from the distributor
address. Each share SHALL be positive and the total SHALL equal exactly 10000
basis points. These checks SHALL apply both at deployment and on replacement.
An owner or fee wallet SHALL be allowed to also appear once as a recipient.

#### Scenario: Valid list
- **WHEN** a list contains three distinct valid addresses with shares 1000, 2000,
  and 7000 basis points
- **THEN** the list is accepted

#### Scenario: Invalid share total
- **WHEN** a proposed list totals less than or more than 10000 basis points
- **THEN** the configuration operation reverts without installing the list

#### Scenario: Invalid entries
- **WHEN** a proposed list contains a duplicate address, a zero address, the
  distributor address, or a zero share
- **THEN** the configuration operation reverts without installing the list

#### Scenario: Recipient count boundaries
- **WHEN** an otherwise valid list contains exactly 300 recipients
- **THEN** the list is accepted
- **WHEN** a list is empty or contains more than 300 recipients
- **THEN** the configuration operation reverts

### Requirement: Atomic recipient replacement

Only the current owner SHALL be able to replace the recipient list. Replacement
SHALL validate the entire proposed list before changing the active configuration
and SHALL replace all previous entries. A rejected replacement SHALL leave the
previous configuration unchanged. Funds already present SHALL be distributed
using the configuration active when distribution is invoked.

#### Scenario: Previously funded balance uses the new list
- **GIVEN** the contract holds USDT funded while an old list was active
- **WHEN** the owner installs a valid new list and then calls `distribute()`
- **THEN** all participant payouts use the new list and its shares
- **AND** removed recipients receive no payout under their old entries

#### Scenario: Failed replacement preserves the old list
- **WHEN** the owner submits an invalid replacement
- **THEN** the operation reverts
- **AND** subsequent distribution still uses the complete previous list

### Requirement: Owner authorization and immediate ownership transfer

Only the current owner SHALL be able to distribute, withdraw the full balance,
replace recipients, change the fee wallet, or change ownership. The initial owner
SHALL be an explicit deployment parameter, not necessarily the deployer. Ownership
transfer SHALL take effect in one successful call by the current owner, without
acceptance by the new owner. New owner addresses SHALL be nonzero and different
from the distributor address; no ownership renunciation operation SHALL exist.

#### Scenario: Unauthorized operation
- **WHEN** a caller other than the current owner attempts any privileged operation
- **THEN** the operation reverts with an authorization error
- **AND** configuration and token balances remain unchanged by that operation

#### Scenario: Immediate transfer
- **WHEN** the current owner appoints a valid new owner
- **THEN** the new owner immediately has all owner permissions
- **AND** the previous owner loses those permissions without a second transaction
- **AND** the fee wallet and recipient list remain unchanged

#### Scenario: Invalid owner address
- **WHEN** deployment or ownership transfer specifies the zero address or the
  distributor address as owner
- **THEN** the operation reverts without changing ownership

### Requirement: Independent fee wallet

The owner SHALL be able to set the fee wallet independently of ownership and the
recipient list. The fee wallet SHALL be nonzero and different from the distributor
address. It SHALL be allowed to equal the owner or a participant address. Updating
it SHALL NOT change the fee rate. Ownership transfer SHALL NOT update it implicitly.

#### Scenario: Redirect fees
- **WHEN** the owner replaces the fee wallet with a valid address
- **THEN** the next distribution sends the fee and rounding residue to that address
- **AND** this applies to the full balance, including funds received before the change

#### Scenario: Invalid fee destination
- **WHEN** deployment or a fee-wallet update specifies the zero address or the
  distributor address as fee wallet
- **THEN** the operation reverts and any previous fee wallet remains unchanged

### Requirement: Funding and explicit full-balance distribution

Ordinary USDT transfers to the distributor SHALL fund its balance without starting
distribution. An owner call to `distribute()` SHALL allocate the entire current
USDT balance using the current recipient list and fee wallet. Multiple incoming
transfers SHALL be treated as one balance, without depositor-specific entitlements,
deposit snapshots, or participant claims.

#### Scenario: Funding does not pay recipients
- **WHEN** someone transfers USDT to the distributor without calling `distribute()`
- **THEN** the received USDT remains in the distributor until an owner operation

#### Scenario: Multiple deposits
- **GIVEN** 40 USDT and 60 USDT have been transferred to the distributor
- **WHEN** the owner calls `distribute()` with no other balance changes
- **THEN** allocation is calculated from the combined 100 USDT

### Requirement: Deterministic allocation and zero remaining balance

For balance B in token base units, fee rate f in basis points, and recipient shares
s_i, distribution SHALL produce these exact amounts:

- Base fee F = floor(B * f / 10000).
- Participant pool D = B - F.
- Each participant payout P_i = floor(D * s_i / 10000).
- Fee-wallet payout C = B - sum(P_i).

C SHALL include F and all participant-rounding residue. Successful distribution
SHALL transfer all P_i and C to their configured destinations and leave the
distributor with zero USDT. Zero-valued transfers SHALL be skipped. Residue SHALL
NOT be reserved on the contract or attributed to a future distribution. When a
fee wallet is also a participant, its total receipt SHALL include both amounts.

#### Scenario: Exact standard allocation
- **GIVEN** B is 100 USDT, f is 500 basis points, and recipient shares are 1000,
  2000, and 7000 basis points
- **WHEN** distribution succeeds
- **THEN** the recipients receive 9.5, 19, and 66.5 USDT respectively
- **AND** the fee wallet receives 5 USDT and the distributor balance is zero

#### Scenario: Fractional residue goes to the fee wallet
- **GIVEN** B is 100000001 base units, f is 500, and shares are 1000, 2000, and 7000
- **WHEN** distribution succeeds
- **THEN** F is 5000000 and D is 95000001 base units
- **AND** participant payouts are 9500000, 19000000, and 66500000 base units
- **AND** the fee wallet receives 5000001 base units and the balance is zero

#### Scenario: Tiny balance and zero-valued payouts
- **GIVEN** B is 7 base units, f is 500, and shares are 1000, 2000, and 7000
- **WHEN** distribution succeeds
- **THEN** participant payouts are 0, 1, and 4 base units, with no transfer for zero
- **AND** the fee wallet receives the remaining 2 base units
- **AND** the distributor balance is zero even though the base fee rounds to zero

#### Scenario: Zero fee still assigns residue
- **GIVEN** B is 1 base unit, f is 0, and two recipients each have 5000 basis points
- **WHEN** distribution succeeds
- **THEN** both participant payouts round to zero
- **AND** the fee wallet receives the 1 base unit of residue

### Requirement: Atomic and protected transfers

If any required participant or fee-wallet transfer fails, the entire distribution
SHALL revert, including earlier transfers in that transaction. A token-reported
failure SHALL NOT be treated as a successful transfer. Calls made during token
interactions SHALL NOT permit duplicate spending or configuration changes that
alter the operation's allocation. Withdrawal SHALL likewise revert if its transfer
fails. No failed operation SHALL leave partial payouts or configuration changes.

#### Scenario: Participant transfer failure
- **WHEN** any participant transfer fails after earlier transfers were attempted
- **THEN** every transfer in that distribution is rolled back
- **AND** the distributor and destination balances match their pre-call values

#### Scenario: Fee transfer failure
- **WHEN** the fee-wallet transfer fails
- **THEN** participant payouts are also rolled back

#### Scenario: Reentrant interaction
- **WHEN** a token interaction attempts a nested payout, withdrawal, or mutation
  that would alter the active operation
- **THEN** it cannot spend the balance twice or change the active allocation
- **AND** the outer operation either satisfies its full invariants or reverts

### Requirement: Unconditional owner withdrawal of all USDT

The owner SHALL be able to withdraw the entire current USDT balance to a chosen
nonzero destination other than the distributor address at any time. This operation
SHALL bypass participant allocation and fees and SHALL NOT require a previous
distribution failure. Success SHALL leave zero USDT and preserve configuration.

#### Scenario: Withdrawal without prior failure
- **GIVEN** the distributor holds 100 USDT and no distribution has failed
- **WHEN** the owner requests full withdrawal to a valid destination
- **THEN** the destination receives 100 USDT, no participant or fee payout occurs,
  and the distributor balance becomes zero

#### Scenario: Recovery after a failed distribution
- **GIVEN** distribution reverted because a participant could not receive USDT
- **WHEN** the owner withdraws to another destination that can receive USDT
- **THEN** the full balance is transferred without participant allocation

#### Scenario: Invalid or failing withdrawal
- **WHEN** the destination is the zero address or the distributor itself, or the
  required USDT transfer fails
- **THEN** withdrawal reverts without transferring any funds

### Requirement: Explicit empty-balance errors

Both `distribute()` and full-balance withdrawal SHALL revert with an identifiable
empty-balance error when the configured USDT balance is zero. Neither SHALL issue
token transfers in that case.

#### Scenario: No funds available
- **WHEN** the owner calls either payout operation with a zero USDT balance
- **THEN** the call reverts with an empty-balance error and no transfer occurs

#### Scenario: Repeated distribution
- **GIVEN** a previous distribution succeeded and no additional USDT has arrived
- **WHEN** the owner calls `distribute()` again
- **THEN** the call reverts with an empty-balance error

### Requirement: Full-size atomic distribution

The implementation SHALL support a valid list of 300 recipients in a single atomic
distribution on the selected deployment network. Deployment readiness SHALL include
a measured gas check using representative worst-case token transfers. If this check
fails, the implementation SHALL NOT silently lower the limit or split payouts across
transactions; the requirement SHALL be revisited before release.

#### Scenario: Maximum-size distribution
- **GIVEN** a valid 300-recipient list and sufficient USDT for every payout to be nonzero
- **WHEN** the owner distributes under the selected network's transaction limits
- **THEN** all payouts and the fee complete in one transaction with zero remaining USDT
- **AND** recorded validation demonstrates that execution fits those limits
