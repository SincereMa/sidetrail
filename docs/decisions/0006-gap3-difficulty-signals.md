# ADR-0006: Difficulty signals (gap 3) — signal model, faces, v0+ scope

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Constrained by:** ADR-0001 (memory model and I/O surface),
  ADR-0002 (gap 5), ADR-0003 (technology stack),
  ADR-0004 (gap 1, amendment to `kind`),
  ADR-0005 (gap 2, edge model and source-adapter framework)

## Context and problem statement

Gap 3 in `docs/scope.md` is the uneven-difficulty problem: the host
agent treats all parts of a project as equivalent, but difficulty is
uneven and shifts. The sidecar must give the agent a read on
difficulty at the right granularity, organized along four faces
(change-safely, understand, operate, keep-up), and presented as
actionable signals with causal attribution — never as a single risk
score.

This ADR fixes the data shape, the four-face taxonomy, the rules
that distinguish a `signal` from a `record` or an `edge`, the storage
layout, the boundary that keeps the sidecar from turning into a
monitoring tool, and the v0 scope.

## Decision drivers

- Gap 3's primary product surface is `health data`, which is
  qualitatively different from `decisions` and `boundaries`. Health
  data is computed; the other two are declared. The trust model is
  different.
- The hard principle "the agent is one of several actors" (scope
  dim 6) plus AGENTS.md "lightweight" rules out a long-running
  daemon. Signals must be derivable on demand, not pushed.
- Scope is explicit that signals must be trend-aware, causally
  attributed, actionable, and source-transparent. Each of these
  shows up in the schema and the storage layout.
- ADR-0005 (gap 2) introduces a "machine-derived" write channel and
  a "source adapter" framework. Gap 3 is the second user of that
  framework. Locking the two together early avoids two near-identical
  frameworks.

## Considered options

For each decision, the chosen option is first; rejected alternatives
follow.

## Decision

### 1. A new data structure: `signal`, parallel to `record` and `edge`

A third schema is introduced, distinct from the record and edge
schemas. The shape (informal, v0+):

```
signal {
  id               ulid
  kind             enum, see section 2
  scope            path glob | module id | service id | area id
  difficulty_face  change_safely | understand | operate | keep_up
  value            number (kind-specific)
  unit             string (e.g. "commits/week", "%", "days")
  window           time range (e.g. "30d", "since v1.0")
  direction        better | worse | stable | unknown
  causes           array of { kind, scope, weight }   # causal attribution
  source           { system, ref }                    # source-transparent
  computed_at      ISO 8601
  raw_inputs       links / refs to source data        # reproducibility
  confidence       high | medium | low
  behavioral_hint  string, required                   # actionable
  status           active | superseded | archived
  tags             free-form tags
}
```

- **Why a separate structure, not a `record` with `kind = signal`:**
  the trust model is different. A `record` says "this is what a
  human (or agent) declared, with reasoning". A `signal` says
  "this is what the data says right now, given an algorithm". They
  need different write channels (declared vs derived), different
  freshness rules (declared never goes stale on its own; signals
  do), and different sources of error (typo vs algorithm bug).
- **Why the metadata fields are reused (status, source_type
  analog, freshness):** same concept, two domains. The source-side
  trust value is `confidence` for signals instead of
  `source_type` for records, because a signal is never "human" —
  it is always derived or hand-marked.
- **Why `behavioral_hint` is required and `causes` is optional:**
  scope demands actionability. `behavioral_hint` is the agent's
  pickup line. `causes` is the deeper answer; it is allowed to be
  missing because the data may not support it.

### 2. The four difficulty faces and the kind taxonomy

The four faces from scope are the top-level grouping. Each kind
belongs to exactly one face; a scope may have signals in any number
of faces.

| Face | Kinds |
| --- | --- |
| `change_safely` | `churn`, `coverage`, `defect`, `revert` |
| `understand` | `complexity`, `doc_freshness`, `owner_concentration`, `age` |
| `operate` | `incident_freq`, `test_flakiness`, `deploy_freq`, `mttr` |
| `keep_up` | `commit_rate`, `pr_age`, `dep_freshness`, `migration` |

- **Why one face per signal:** a single signal can be misleading if
  it is read across faces. "High churn" is `change_safely`; "high
  commit rate" is `keep_up`. Putting them on the same axis is what
  produces the "single risk score" anti-pattern scope warns about.
- **Why this is the kind list, not a longer one:** the list mirrors
  the four-face enumeration in scope. New kinds enter the enum when
  a real source adapter needs them; that is an additive schema
  change and a follow-up ADR.

### 3. Trend and decay: history files and superseded-by-rotation

A signal is not a point. The same `(scope, kind)` pair produces a
sequence of signals over time. The history is stored at:

```
.cortex/signals/history/<scope>/<kind>.jsonl
```

Each line is a snapshot of the signal at a point in time. The
"current" view of a `(scope, kind)` is the latest `active` snapshot;
older snapshots are `superseded` but kept in the history file for
reproducibility and trend queries.

- **Why a jsonl history file rather than a database:** a jsonl
  append-only file is git-friendly, diff-friendly, and reviewable.
  This is the same argument as ADR-0001 for plain files: the data
  is too small and too inspectable to justify a database in v0+.
- **Why signals auto-supersede on recompute:** scope dim 3 says
  stale memory presented with the same weight as fresh memory is
  worse than no memory. Recomputing produces a new active snapshot
  and demotes the old one to `superseded`. The history file is
  never rewritten; only appended to.

### 4. Read entry points: `health` and `signals`

- `cortex health --scope <path> [--face <f>] [--kind <k>] [--trend]`
  returns, per face, the top signals for the scope. `--trend`
  reports `direction` and how it changed over the last few windows.
- `cortex signals --scope <path> [--face <f>] [--kind <k>]`
  returns the raw list, ordered by `computed_at` desc.
- `cortex context --file <path>` is extended to include the top
  three signals per face. This is the entry point a host agent is
  expected to call; the host agent should not need to know about
  `cortex health` directly to act on signals.

`cortex health` and `cortex signals` are subcommands added in v0+;
they do not exist in v0.

### 5. Compute trigger: explicit, never automatic

The sidecar does not run a daemon. The compute path is:

- `cortex compute signals [--since 30d] [--source git,ci]` is the
  one and only entry point. It is invoked by a human, by a CI job,
  or by an opt-in `git` hook installed by the user. The first
  v0+ ADR for gap 3 must justify the recommended trigger in its own
  ADR rather than this one.
- A host agent that calls `cortex context` and finds the signals
  stale gets a warning, not a recompute. Stale-by-design is the
  contract.
- CI integration is the recommended default in v0+; the project
  ships a sample workflow. Manual and lazy-via-host are equally
  valid; the choice is the user's.

- **Why no daemon:** AGENTS.md "lightweight" plus "no surprise
  transitive dependencies". A signal-computing daemon is exactly
  the kind of thing a project will quietly turn off six months in.
- **Why `cortex context` warns on staleness rather than
  recomputing:** the host agent's call should remain fast. A
  stale-but-correctly-flagged signal is more useful than a slow
  call. Recomputation is a deliberate action.

### 6. Boundaries the sidecar holds

In addition to the principles in `AGENTS.md` and the boundaries in
ADR-0002 and ADR-0005:

- **The sidecar does not produce a risk score.** It surfaces
  multiple signals and a `behavioral_hint` per scope. The host
  agent combines them; the sidecar does not. This is the sharpest
  application of the gap-3 wording "causally attributed, not a
  single risk score".
- **The sidecar does not silently recompute.** Compute is a
  command, not a side effect. CI integration is the
  recommendation; the sidecar does not start a process on its
  own.
- **The sidecar does not build a dashboard.** Output is CLI
  text and `--json`. Visualization is a host concern.
- **The sidecar does not phone home.** Source data is read
  locally (git, local CI artifacts, local monitoring exports);
  the sidecar does not contact an external service. Privacy of
  operational data is preserved by construction.
- **The sidecar does not let signals stay active past their
  `window`.** A signal whose `computed_at` is older than the
  window's end is auto-`superseded` by the read layer. The host
  agent sees it as missing, not as fresh.
- **The sidecar does not invent a source.** If a source adapter
  is not implemented, the kind is simply unavailable. The CLI
  returns "no data" rather than making something up.

## Consequences

### Positive

- The data shape for health is fixed. The first v0+ gap-3 ADR
  only argues about the source adapter and the first set of kinds.
- The four-face taxonomy is exactly the one scope asks for; future
  contributions can be reviewed against it.
- Trend-aware and source-transparent are properties of the schema
  and storage layout, not of the source adapters. New adapters get
  this for free.
- Auto-`superseded` on staleness means a host agent cannot act on
  a signal that has outlived its window. The trust model from
  ADR-0001 is honored.
- Daemon-free compute keeps AGENTS.md "lightweight" enforceable
  rather than aspirational.

### Negative

- Without a daemon, the host agent must either accept stale
  signals or trigger a recompute. Both are acceptable, but neither
  is invisible. This is the deliberate cost of "lightweight".
- The "no risk score" rule is real: an agent that wants one will
  have to do its own combination. Some agents will do this badly.
  The boundary is still the right one; the cost is named.
- The first v0+ gap-3 ADR has to pick a first source and a first
  set of kinds. The choice is constrained by which source adapters
  are stable; if no good option exists, gap 3 enters v0+ later
  rather than as planned.

### Neutral

- The history file format (jsonl per (scope, kind)) is a stable
  choice. A future ADR can add a compact format if a project's
  history grows too large; the append-only property is preserved.
- The `confidence` field defaults to `medium` for derived signals
  and `high` for signals whose `raw_inputs` can be reproduced
  bit-for-bit. Defaults are not a controversy; the writer picks.

## Risks and rollback triggers

| Risk | Trigger | Action |
| --- | --- | --- |
| Compute trigger friction makes signals go stale in practice | Empirical observation: median `computed_at` age > 1 week in a CI-using project | Reconsider the CI recommendation; consider a `cortex watch` opt-in that is a single-purpose process, not a generic daemon |
| Source adapter dependency footprint is heavy | First v0+ adapter pulls in > 1 MB of new code or requires cgo | Reject the adapter; fall back to manual `cortex add signal` for those kinds |
| `causes` is set to nonsense weights and the agent trusts them | Empirical observation of misleading `causes` | Demote `causes` to advisory: host agents must display them as such or ignore them |
| Stale-but-still-`active` signals in spite of the auto-`superseded` rule | A bug in the read layer | Hotfix the read layer; the rule is part of the contract |
| Gap 2 and gap 3 source-adapter frameworks drift apart | Two near-identical frameworks in v0+ | Refactor into a shared `internal/source` package; one ADR |

## Open questions deferred to the first v0+ gap-3 ADR

- Which source system is the first adapter (git, CI, issue
  tracker, monitoring)?
- Which kinds ship in the first adapter?
- What is the recommended compute trigger (CI job, manual, hook)?
- How are conflicting signals from multiple sources resolved?
- How is `causes` populated in v0+ — declared by the source
  adapter, or inferred by the sidecar?
- What is the staleness threshold per face, and is the default
  one window or one window plus a grace period?
