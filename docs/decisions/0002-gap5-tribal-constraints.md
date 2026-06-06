# ADR-0002: Tribal constraints (gap 5) — shape, v0 scope, boundaries

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Constrained by:** ADR-0001 (memory model and I/O surface)

## Context and problem statement

Gap 5 in `docs/scope.md` is the constraints problem in its purest
form: explicit "do not do" rules with a reason, held in tribal memory
that the host agent cannot read. The sidecar's first useful job is to
make those constraints readable.

This ADR chooses, on top of the substrate fixed by ADR-0001:

- any schema extensions that are specific to constraints,
- the minimum v0 that proves the read/write loop end to end,
- the boundary at which the sidecar stops and the host agent (or
  human) takes over.

Everything in ADR-0001 is inherited; this ADR only adds gap-5-specific
decisions.

## Decision drivers

- Gap 5's primary product surface is `boundaries/constraints`, with
  `decisions` as secondary (the constraint has a reason that is itself
  a recorded decision).
- Scope calls out the read/write asymmetry: capture is deliberate and
  semi-manual, and any design must assume the constraint set is
  incomplete.
- The hard principles in `AGENTS.md`, especially non-intrusive and
  sidecar-not-replacement, push the sidecar toward "expose the
  constraint, do not enforce it".

## Considered options

For each gap-5 question, the chosen option is listed first; rejected
options are recorded for the record.

## Decision

### 1. Schema extension: `severity` and `valid_until` on the record

Two optional fields are added to the record schema from ADR-0001:

- `severity`: enum, default `hard`. `hard` means a violation has real
  consequences (compliance, security, outage). `soft` means team
  preference or a historical lesson. The read layer is expected to
  present them in this order; the host agent decides whether to act
  on `soft`.
- `valid_until`: ISO 8601 timestamp, optional. When set, the CLI hides
  the record by default after this time. `--include-expired` overrides.

- **Why add `severity` now:** without it, a `soft` constraint
  ("we used to dislike touching this file") reads the same as a
  `hard` constraint ("do not touch — compliance review pending"). The
  agent cannot triage.
- **Why add `valid_until` now:** temporal constraints
  ("do not touch the bridge code until Q3") are common in the gap-5
  examples. Forcing a human to remember to `supersede` at the right
  moment makes temporal constraints unreliable.
- **Why not enforce `severity` more strongly:** the sidecar does not
  enforce. A flag is enough; behavior is the host's responsibility.

### 2. v0 write channels: human CLI and file edit only

v0 ships two write channels:

1. `cortex add constraint --scope ... --reason ... [--severity]
   [--valid-until] [--evidence ...]`
2. Direct edit of `.cortex/constraints/<id>-<slug>.json`, followed by
   `cortex validate`.

The two staged channels from ADR-0001 (`_proposed/` for agent
suggestions, `_seed/` for scrape) are not part of v0 for gap 5.
Capturing tribal constraints is hard enough that any non-human
channel should be opt-in, and v0 is too early to know what the right
opt-in UX is.

- **Why defer `_proposed/`:** the host agent is not yet in the loop
  for v0; this channel becomes interesting when an agent can credibly
  notice a violation and propose a constraint. We have no evidence
  for that loop yet.
- **Why defer `_seed/`:** scope calls out that "the team can usually
  articulate [tribal constraints] when asked", which suggests the
  right first move is a small `cortex ask` ritual, not bulk
  harvesting. Bulk harvesting is gap 4 (architecture drift) territory,
  not gap 5.

### 3. v0 read entry points: `ask` and `context`

Two entry points are first-class in v0:

- `cortex ask --scope <path> --kind constraint [--severity] [--tag]`
  — the lowest-level query; returns matching records sorted by
  `severity` desc, then `last_verified_at` desc.
- `cortex context --file <path> [--radius N]` — the convenience
  aggregate. v0 of this command exists only because it is what the
  host agent will actually call. It returns the file's constraints
  first, then its decisions (gap 1's slot is reserved; the command
  works even if no decisions exist yet). `--radius N` widens the
  match from the file itself to N directory levels up and down.

Both commands accept `--json` for agent consumption.

- **Why `context` is a separate command, not a flag on `ask`:** the
  ergonomics differ. `ask` is for humans exploring; `context` is for
  the host agent. Conflating them blurs the read surface.

### 4. Conflict handling: do not resolve, return everything

v0 does not detect or resolve conflicts between constraints on the
same scope. The CLI returns all matching records; the human or the
host agent arbitrates. The CLI reserves a `--conflict-strategy` hook
for future use; v0 always uses `none`.

- **Why no v0 conflict resolution:** resolving requires a model of
  which constraint "wins" (newer? harder? higher-trust source?). That
  model is not a gap-5 question; it is a cross-gap policy question.
  Premature automation here would be exactly the sidecar-overreach
  AGENTS.md forbids.

### 5. Boundaries the sidecar holds

In addition to the principles in `AGENTS.md`, gap 5 fixes the
following:

- **The sidecar does not enforce constraints.** It surfaces them. The
  host agent decides whether to block an edit. Putting enforcement
  inside the sidecar would make the sidecar a guard, not a memory.
- **The sidecar does not interpret constraints.** The `reason` field
  is human-written. The sidecar does not paraphrase, summarize, or
  rephrase a constraint when returning it. The host agent may do so;
  the sidecar does not.
- **The sidecar does not auto-commit.** Records are written into the
  working tree. A human (or normal git tooling) decides when to
  commit. This avoids sidecar-driven surprises in code review flows.
- **The sidecar does not auto-delete expired records.** A record with
  `valid_until` in the past is hidden from default reads, not removed.
  Auditability beats tidiness.

## Consequences

### Positive

- A working v0 of gap 5 is implementable with the substrate from
  ADR-0001 and two new optional fields. The loop is provable
  end-to-end: write a constraint, query it, see it appear in a
  host-agent context call.
- The sidecar stays a memory, not a guard. The "do not enforce" line
  is sharp and easy to defend.
- Temporal constraints work without humans remembering to clean up.
- Soft constraints are clearly marked, so a host agent can ignore
  them if it wants, without losing access to hard ones.

### Negative

- The conflict surface is unmitigated in v0. Two humans can write
  contradictory hard constraints on the same scope. Mitigated socially
  (constraints go through PR review) rather than technically.
- No automated capture in v0 means the constraint set is whatever
  humans remember to write. This is the deliberate trade-off; gap 5
  explicitly assumes an incomplete set.

### Neutral

- The two new fields make the record schema slightly wider. JSON
  Schema evolution rule (append-only) keeps old records valid.
- A future ADR may add `_proposed/` and `_seed/` for gap 5; this ADR
  is explicit that they are not in v0.

## Open questions deferred

- Cross-repository constraint sharing (one constraint covering
  multiple repos). This is a meta question, not gap-5 specific; it
  lives in a future ADR that revisits the storage layer.
- Two-way sync with issue trackers (write a constraint, create a
  GitHub issue; close the issue, archive the constraint). Out of
  scope for v0; revisit if user demand appears.
- Heuristic capture from PR comments, Slack, post-mortems. Out of
  scope for v0; revisit when an evidence pipeline is designed.
