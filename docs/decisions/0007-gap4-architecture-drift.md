# ADR-0007: Gap 4 — Architecture drift

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintained by:** ADR-0001 (Memory model and I/O surface)

## Context and background

Gap 4 in `docs/scope.md` describes the problem of documented
architecture drifting from actual architecture. The architecture
diagram shows service A and service B as cleanly separated, but
service B has been reading service A's primary database directly
for eight months.

This ADR extends the memory model established by ADR-0001 with
gap-4-specific decisions.

## Decision drivers

- Architecture drift is naturally captured by `decision` records
  that document the actual vs. documented state.
- The `scope` field allows mapping drift to specific services or
  modules.
- The `status` field allows tracking whether the drift has been
  resolved.

## Decision

### 1. Drift as decision records

Architecture drift is stored as `decision` records with tags
indicating the type of drift:

- `documented-vs-actual`: the documented architecture differs from
  the actual architecture.
- `implicit-dependency`: an undocumented dependency exists.
- `deprecation-pending`: a component is pending deprecation but
  still in use.

### 2. Drift resolution

When drift is resolved, the record is marked as `superseded` and
a new record is created documenting the resolution.

### 3. Drift detection

Drift detection is out of scope for v0. Future versions may add
automated drift detection by comparing documented architecture
against actual code structure.

## Consequences

### Positive

- Architecture drift is explicitly recorded and tracked.
- The agent can see drift information before making changes.
- The supersession chain preserves the history of drift and
  resolution.

### Negative / accepted

- Drift records must be manually created; the agent cannot
  detect drift automatically in v0.
- The drift detection is manual; tooling may add it later.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): the gap-4-specific changes
  must be recorded before merge.
