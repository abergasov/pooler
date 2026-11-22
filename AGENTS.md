# Project instructions

Read this document before discussing specifications, writing documentation, or
implementing changes in this repository.

## Project purpose

Pooler is a future smart contract project for EVM networks that distributes
rewards to participants. The project is currently in the specification phase.

Participant eligibility, reward calculation, funding, payout mechanics, supported
networks and assets, permissions, and upgradeability are not yet decided. Do not
treat assumptions about these topics as agreed requirements.

## Language

- Write all repository documentation and OpenSpec artifacts in English,
  including proposals, specifications, designs, tasks, and decision records.
- Use the user's language for conversation unless they request otherwise.

## Behavior specifications

- Use OpenSpec to describe and evolve project behavior.
- Read `openspec/config.yaml`, relevant specifications in `openspec/specs/`, and
  the relevant change artifacts in `openspec/changes/` before proposing changes.
- Keep proposed behavior in change artifacts distinct from accepted behavior in
  the main specifications. Use the OpenSpec workflow to synchronize them.
- Record unresolved decisions explicitly. Recommendations and unanswered
  questions are not requirements.
- Keep this document focused on project context and working rules; detailed
  contract behavior belongs in OpenSpec.

## Specification discussions

- Use the `grill-me` skill when discussing or refining specifications and
  documentation. Read its instructions and follow its delegated `grilling` skill.
- If the skill is absent from the available catalog, look for an installed copy.
  If it cannot be found, report that limitation rather than claiming to use it.
- Explore decisions in dependency order. Ask the questions whose prerequisites
  are settled in each round, give recommendations, and wait for the user's
  answers before asking dependent questions.
- Investigate repository facts using tools; ask the user for product decisions.
- Challenge ambiguous requirements, assumptions, edge cases, and failure modes.
  For reward distribution, clarify eligibility, accounting, authorization, and
  the handling of funds as the relevant decisions arise.
- Record confirmed answers in English and keep open questions visibly unresolved.
  Confirm shared understanding before treating the design as settled.

## Go coding standards

- For Go code changes, follow [Go coding standards](docs/go-coding-standards.md).
- Use the repository-root `.golangci.yaml` for Go formatting and linting.

## Implementation

- When implementing an OpenSpec change, use both `openspec-apply-change` and
  `task`: the former manages artifacts and task checkboxes; the latter governs
  investigation, implementation, regression tests, verification, and diff review.
- Prefer the smallest readable implementation that satisfies agreed behavior.
  Do not add speculative features or silently choose unresolved product behavior.
