# ADR-0006: Gap 3 — Difficulty signals

- **Status:** Accepted (see note below)
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintained by:** ADR-0001 (Memory model and I/O surface)

> **Note (2026-06-13):** The `ask` command has been removed. Use
> `context --file <path>` to query records for a file area.
> See [README.md](../../README.md).

## Context and background

Gap 3 in `docs/scope.md` describes the problem of project
difficulty being uneven and shifting, but the agent has no read
on it. Some modules are churn hotspots, some are bug clusters,
some have eroding test coverage, some depend on stale libraries.

This ADR extends the memory model established by ADR-0001 with
gap-3-specific decisions.

## Decision drivers

- Health signals are naturally captured by `signal` records.
- The `signal` record kind is already defined in ADR-0001.
- Signals need a `severity` field to indicate the level of concern.

## Decision

### 1. Signal record shape

A `signal` record captures:

- `subject`: one-line summary of the signal
- `body`: detailed explanation
- `severity`: integer (0-10) indicating the level of concern
- `source_type`: how the signal was detected (manual, automated, inferred)
- `tags`: free-form tags for categorization

### 2. Signal categories

Signals are categorized by tags:

- `churn`: high change frequency
- `bugs`: high bug density
- `coverage`: eroding test coverage
- `stale`: stale dependencies
- `complexity`: high complexity

### 3. Signal aggregation

`sidetrail ask --scope <scope> --kind signal` returns all signals
for a given scope. The agent can use this to understand the
difficulty profile of a project area.

## Consequences

### Positive

- The agent can see difficulty signals before making changes.
- Tags allow filtering by signal category.
- The severity field allows prioritization.

### Negative / accepted

- Signals must be manually created or inferred by tooling;
  the agent cannot detect them from code alone.
- The severity field is subjective.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): the gap-3-specific changes
  must be recorded before merge.
