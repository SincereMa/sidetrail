# ADR-0004: Reasoning trails (gap 1) — kinds, schema, v0 scope

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Constrained by:** ADR-0001 (memory model and I/O surface),
  ADR-0002 (gap 5), ADR-0003 (technology stack)
- **Amends:** ADR-0001 (the `kind` enum; see Amendment A below)

## Context and problem statement

Gap 1 in `docs/scope.md` is the lost-reasoning-trail problem: between
sessions, the host agent cannot recover what was tried, what was
rejected and why, or what the current shape of the code is a
consequence of. The sidecar makes the trail readable.

Scope is explicit that a useful trail must distinguish three kinds of
items — **decisions**, **experiments**, **incidents** — and that the
write side is the harder unsolved problem. This ADR, on top of the
substrate fixed by ADR-0001, chooses:

- whether the three kinds are three independent `kind` values or
  subtypes of a single kind,
- the schema fields that are specific to each kind,
- the v0 slice of gap 1 (a v0 that proves the read/write loop is not
  enough if it ships every kind at once and breaks the substrate),
- the boundary at which the sidecar stops and the host (or human)
  takes over.

## Decision drivers

- Gap 1's primary product surface is `decisions`, with `boundaries` as
  secondary. Experiment and incident are extensions of the same
  surface.
- "Why" is load-bearing (scope dim 2). The schema must keep `reason`
  mandatory and add new fields that carry reasoning, not facts.
- Information decays (scope dim 3). Status fields must be explicit
  per kind.
- The read side is sharp; the write side is harder. v0 must prove the
  read loop, not exhaust the write model.
- AGENTS.md "lightweight": the schema and CLI must stay simple enough
  for the v0 host agent to call without ceremony.

## Considered options

For each decision, the chosen option is first; rejected alternatives
follow.

## Decision

### Amendment A: extend `kind` enum in ADR-0001

The `kind` enum in ADR-0001 (decision | constraint | signal) is
extended to five values:

```
decision | constraint | signal | experiment | incident
```

- **Why three new kinds, not one kind plus `subtype`:** the three
  things have different `status` enums (a decision is active /
  superseded / archived; an experiment is in_progress / succeeded /
  failed / inconclusive / abandoned; an incident is investigating /
  mitigated / resolved). Encoding those as a free-form `subtype`
  field pushes the type discipline into code that has to be written
  in every reader. Three `kind` values keep the read surface honest.
- **Why not more kinds (e.g. `pattern`, `convention`):** the five
  kinds cover the gap-1 surface as written. New kinds are an
  additive change in the schema and require a fresh ADR; they are
  not anticipated by this one.

### 1. New schema fields (gap 1 specific)

The following fields are added to the record schema from ADR-0001.
They are optional unless marked otherwise; readers must handle their
absence.

| Field | Type | Applies to | Required? | Meaning |
| --- | --- | --- | --- | --- |
| `decided_at` | ISO 8601 | decision | yes (when `kind = decision`) | When the decision was made. May predate `created_at` for retrospective capture. |
| `started_at` | ISO 8601 | experiment | yes (when `kind = experiment`) | When the experiment started. |
| `ended_at` | ISO 8601 | experiment | no | When the experiment concluded. |
| `occurred_at` | ISO 8601 | incident | yes (when `kind = incident`) | When the incident happened. |
| `resolved_at` | ISO 8601 | incident | no | When the incident was resolved. |
| `related_to` | array of record ids | all | no | Free-form links. Used by `cortex trail` to walk the reasoning chain. |
| `rejected_alternatives` | array of strings | decision | no | The alternatives considered and not chosen. This is the field that makes gap-1 records distinct from gap-5 records: a decision without rejected alternatives is barely a decision. |

- **Why `decided_at` is required on `decision` but `created_at`
  already exists:** `created_at` is when the record was written;
  `decided_at` is when the decision was actually made. The two
  differ when a team backfills history. Forcing both keeps the
  distinction honest.
- **Why `rejected_alternatives` is a free-text array, not a
  structured set of records:** v0 cannot pay for the cost of
  requiring rejected alternatives to be first-class records. Free
  text inside the decision is enough for an agent to know what was
  considered; it is also enough to be promoted to first-class
  records later if usage demands it.

### 2. Status enums per kind

The shared `status` field in ADR-0001 has its value domain scoped by
`kind`:

- `decision`: `active` (default) | `superseded` | `archived`
- `experiment`: `in_progress` (default) | `succeeded` | `failed` |
  `inconclusive` | `abandoned`
- `incident`: `investigating` (default) | `mitigated` | `resolved`
- `constraint` and `signal` keep ADR-0002 / future-gap definitions.

`cortex validate` enforces the `(kind, status)` pair.

### 3. Read entry points: `ask`, `context`, and a new `trail` placeholder

- `cortex ask` gains support for `--kind experiment` and `--kind
  incident`, plus `--status` filters. Existing `--kind decision`
  keeps working and now also accepts `--status active | superseded |
  archived`.
- `cortex context --file <path>` aggregates **constraints →
  decisions (active) → experiments (in_progress) → incidents
  (investigating)** in that order. The aggregate respects the
  status scopes above. This is the read entry point a host agent is
  expected to call.
- `cortex trail <id>` is added as a subcommand. In v0 it returns a
  short message stating the feature is not yet implemented and a
  pointer to the ADR. v0+ will walk `supersedes`, `superseded_by`,
  and `related_to` to assemble the reasoning chain.

### 4. v0 scope: ship only `decision`

v0 implements the gap-1 surface for `decision` only. The schema for
`experiment` and `incident` is locked by this ADR (so the format
does not drift), but the CLI commands to add or list them are
deferred to v0+.

- **Why not ship all three in v0:** the v0 goal is to prove the
  read/write loop end to end for gap 5 (constraints) and a slice of
  gap 1 (decisions). Adding two more kinds with their own
  sub-status enums and time fields at the same time increases the
  regression surface for no extra learning.
- **Why this is safe:** the schema is forward-compatible. Records
  for `experiment` and `incident` written by hand during v0 (or by
  future tooling) will validate against the schema and be readable
  by `cortex ask`. The only thing deferred is the human-facing
  `cortex add experiment` ergonomics and the trail walker.
- **What is in v0's `decision`:** `cortex add decision` with
  `--rejected` (repeatable), `cortex ask --kind decision
  [--status active]`, `cortex context --file` showing constraints
  and active decisions, and `cortex supersede <old-id> --new ...`
  that handles the decision→decision case (already specified in
  ADR-0001).

### 5. Write channels: same as ADR-0002, no gap-1-specific addition

Gap 1 uses the same write channels as gap 5: human CLI, human file
edit, with `_proposed/` (agent) and `_seed/` (scrape) deferred. The
write-model problem scope calls out is shared between the two gaps;
solving it twice is not on the table.

### 6. Boundaries the sidecar holds

In addition to the gap-5 boundaries in ADR-0002:

- **The sidecar does not infer a reasoning chain.** `trail` walks
  only the links that humans (or opt-in tools) have explicitly
  recorded. The sidecar will not synthesize "this commit must
  reflect a decision" by reading git history.
- **The sidecar does not score decisions.** It surfaces the
  `reason` and `rejected_alternatives`. It does not mark a decision
  "good" or "bad" or "still relevant beyond its `last_verified_at`".
- **The sidecar does not auto-archive `superseded` decisions.**
  The human or normal git workflow decides when history moves from
  `superseded` to `archived`. Auto-archiving would silently destroy
  the trail that gap 1 exists to preserve.

## Consequences

### Positive

- The schema for reasoning trails is fully specified, including the
  fields and statuses that gap 1 needs but ADR-0001 deferred. No
  re-litigation later.
- `cortex context --file` is the single host-agent entry point that
  carries both constraints (gap 5) and decisions (gap 1 slice). The
  integration cost for the first host agent stays small.
- Forward-compatible schema means teams that want to hand-write
  `experiment` or `incident` records during v0 can do so and have
  them work the day the CLI catches up.
- The "sidecar does not infer" boundary is sharp. AGENTS.md's
  non-intrusive principle is preserved across the new feature.

### Negative

- v0 ships only one of three new kinds. Teams that need experiments
  or incidents immediately must hand-write records until v0+. This
  is a deliberate cost, not an oversight.
- `cortex trail` is a placeholder in v0. The unique value of gap 1
  (a walkable reasoning chain) is not fully delivered until v0+.
- `rejected_alternatives` is free text, not structured. If teams
  end up wanting to query "how many decisions rejected library X",
  they will need a follow-up ADR to promote it to first-class.

### Neutral

- The five `kind` values are now a stable surface; adding a sixth
  is an additive schema change and a fresh ADR.
- `cortex validate` grows a small rule set for `(kind, status)`
  pairs. This is a strict superset of ADR-0001's behavior.

## Open questions deferred

- Promotion of `rejected_alternatives` to first-class records.
- `cortex trail` v0+: cycles, depth limits, time-window filters.
- A "consideration timeline" view that orders decisions, experiments,
  and incidents on a single time axis. Distinct from `trail` (which
  is a causal walk), this is a chronological view.
- Linking records to specific code symbols (function, class, file)
  beyond the `scope` glob. Useful but a separate modeling problem.
