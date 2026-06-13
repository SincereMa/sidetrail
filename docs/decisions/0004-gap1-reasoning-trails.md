# ADR-0004: Gap 1 — Reasoning trails

- **Status:** Accepted (see note below)
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintained by:** ADR-0001 (Memory model and I/O surface)

> **Note (2026-06-13):** The `supersede` command has been replaced
> by `update`. To supersede a record: add a new record with
> `supersedes` field, then update the old record's status.
> See [README.md](../../README.md).

## Context and background

Gap 1 in `docs/scope.md` describes the problem of reasoning
trails being lost between sessions. When an agent makes changes
to a project, the reasoning behind those changes is not recorded.
This means that future sessions (human or agent) cannot understand
why a particular approach was taken, what alternatives were
considered, and what tradeoffs were made.

This ADR extends the memory model established by ADR-0001 with
gap-1-specific decisions.

## Decision drivers

- Decisions are the natural home for reasoning trails.
- The `decision` record kind is already defined in ADR-0001.
- The write path must support capturing the reasoning at the
  time the decision is made.

## Decision

### 1. Decision record shape

A `decision` record captures:

- `subject`: one-line summary of the decision
- `body`: long-form explanation of the reasoning
- `reason`: why this decision is load-bearing
- `decided_at`: when the decision was made
- `rejected_alternatives`: what was considered and why it was rejected
- `status`: `active`, `superseded`, `archived`
- `superseded_by`: id of the decision that replaces this one

### 2. Status lifecycle

Decisions follow a natural lifecycle:

1. **active**: the decision is in effect.
2. **superseded**: a newer decision has replaced this one.
3. **archived**: the decision is no longer relevant but is kept for historical context.

### 3. Supersession workflow

`sidetrail supersede <old-id> --new <file>` marks the old record
as superseded and adds the new record. This preserves the chain
of reasoning while making it clear which decision is current.

## Consequences

### Positive

- Decisions are captured at the time they are made, not retroactively.
- The supersession chain preserves the full history of reasoning.
- The agent can query decisions by scope and kind to understand
  the project's decision history.

### Negative / accepted

- Decisions must be manually created; the agent cannot infer
  them from code changes alone.
- The `rejected_alternatives` field is optional; humans may
  forget to fill it in.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): the gap-1-specific changes
  must be recorded before merge.
