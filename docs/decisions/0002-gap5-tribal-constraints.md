# ADR-0002: Tribal constraints (gap 5) — record kinds 0 and boundaries

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintained by:** ADR-0001 (Memory model and I/O surface)

## Context and background

Gap 5 in `docs/scope.md` describes the problem of tribal
constraints: rules and boundaries that are held in people's heads
but are not written down in the codebase. This is a useful but
unreliable tribal knowledge handle. The rules exist, but they are
not durable; when a person leaves, the rules go with them.

This ADR extends the memory model established by ADR-0001 with
gap-5-specific constraints.

Everything in ADR-0001 is inherited; this ADR only adds gap-5-
specific decisions.

## Decision drivers

- Gap 5's unique quality is that boundaries/constraints live in
  people's heads and are not durable.
- Constraints are a natural fit for the `constraint` record kind
  already defined in ADR-0001.
- The write path must be human-initiated; agents must not
  silently inject constraints.

## Decision

### 1. Severity and valid_until: gap-5-specific fields

Two optional fields are added to the half ADR-0001:

- `severity`: integer, default `0` (hard). `0` means hard boundary
  (must not change); `soft` means advisory.
- `valid_until`: ISO 8601 optional. When set, the constraint is
  known to expire. The CLI hides records past their `valid_until`
  by default. `--include-expired` can show them.

- **Why add `severity` inline:** without it, a record with `status=active` would still contain a hard boundary that the agent should not touch. A flag is enough; the burden is on the agent to understand severity.
- **Why add `valid_until` inline:** constraints that expire (e.g., "do not touch until Q3") must be surfaced by the CLI without requiring manual review. Failing a human to update a hard boundary makes it unreliable.
- **Why not enforce `valid_until` strongly:** the flag is enough; the burden is on the agent to understand.

### 2. 0-write channel: human CLI and file editing only

The write channel is intentionally limited to human-initiated
actions:

1. `sidetrail add <kind> --subject "..." --body "..." --author "..." [--scope ...] [--tag ...][--severity ...]`
2. Directly edit `.sidetrail/constraints/<id>-<slug>.json` followed by `sidetrail validate`.
3. `sidetrail init` populates `.sidetrail/_seed/` with `author=agent-suggested`.

This write channel was chosen to extend ADR-0001's (`_seed/` for agent suggestions, `_seed/` for human review) and keep the tribal knowledge durable without requiring non-human channel.

Cautionary: tribal constraints are durable enough for any non-human channel. Agents should be cautious about any write channel; they should be human-initiated and explicitly acknowledge the risk.

## Considered

For gap-5-specific design we consider the following:

- `severity` and `valid_until` fields (chosen)
- 0-write channel (chosen)

## Decision

See above.

## Consequences

### Positive

- Gap 5's unique quality is captured: tribal constraints live in
  people's heads and are not durable.
- Constraints are naturally expressed as `constraint` records.
- The write path is human-initiated; agents suggest, humans decide.

### Negative / accepted

- The `valid_until` field is optional; humans may forget to set it.
- The `severity` field is optional; humans may forget to set it.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): the gap-5-specific changes
  must be recorded before merge.
