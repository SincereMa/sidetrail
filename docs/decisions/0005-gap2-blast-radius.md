# ADR-0005: Gap 2 — Blast radius

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintained by:** ADR-0001 (Memory model and I/O surface)

## Context and background

Gap 2 in `docs/scope.md` describes the problem of blast radius
being invisible across services, modules, and subprojects. When an
agent edits a file, it cannot see which other services, modules,
subprojects, or shared contracts depend on that file — statically,
at runtime, or operationally.

This ADR extends the memory model established by ADR-0001 with
gap-2-specific decisions.

## Decision drivers

- Blast radius is naturally captured by `constraint` records with
  a specific focus on dependencies.
- The `scope` field on records allows mapping blast radius to
  specific parts of the project.
- The `context` command aggregates records relevant to a file path,
  which is the primary use case for blast radius visibility.

## Decision

### 1. Blast radius as constraint records

Blast radius information is stored as `constraint` records with
tags indicating the type of dependency:

- `static`: compile-time dependencies
- `runtime`: runtime dependencies
- `operational`: operational dependencies (monitoring, deployment)
- `social`: human dependencies (code ownership, review requirements)

### 2. Context aggregation

`sidetrail context --file <path>` returns all records whose scope
matches the file path or any enclosing scope. This allows the agent
to see the full blast radius of a file before editing it.

### 3. Radius flag

`--radius N` controls how many levels of enclosing scopes to
include. Default is 2 (immediate parent and grandparent).

## Consequences

### Positive

- The agent can see blast radius information before making changes.
- The scope-based matching allows natural expression of dependency
  boundaries.
- Tags allow filtering by dependency type.

### Negative / accepted

- Blast radius records must be manually created; the agent cannot
  infer dependencies from code alone.
- The radius flag is a heuristic; it may miss deep dependencies.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): the gap-2-specific changes
  must be recorded before merge.
