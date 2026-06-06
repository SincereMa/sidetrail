# ADR-0007: Architecture drift (gap 4) — drift model, v0+ scope, boundaries

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Constrained by:** ADR-0001 (memory model and I/O surface),
  ADR-0002 (gap 5), ADR-0003 (technology stack),
  ADR-0004 (gap 1), ADR-0005 (gap 2), ADR-0006 (gap 3)

## Context and problem statement

Gap 4 in `docs/scope.md` is the architecture-drift problem: the
documented architecture is a snapshot of an earlier moment, and the
actual architecture has drifted. An agent working from current files
and any documentation it can find can construct a mental model
inconsistent with reality — dissolved boundaries treated as
load-bearing, deprecated paths avoided, ownership maps whose owners
have moved on.

Scope is explicit that gap 4 overlaps with gaps 1 and 2 but is
distinct: the remediation is "consistency between two
representations of the same project", and the sidecar needs an
explicit notion of "current" to address it.

This ADR fixes:

- the two views whose difference is the drift (doc-architecture
  drawn from gap-1 decisions, actual-architecture drawn from gap-2
  edges and gap-1 decisions),
- the new data structure (`drift`) that records the differences,
- the taxonomy of drift kinds,
- the v0 scope (gap 4 is not in v0; it requires data the sidecar
  does not yet collect),
- the boundary that keeps the sidecar from rewriting the project's
  documentation on its own.

## Decision drivers

- Gap 4's primary product surface is `decisions` (the original
  architecture was a decision) and `boundaries` (the actual
  boundaries are the binding ones). The sidecar already records
  both; gap 4 is the *consistency check* between the recorded
  views.
- The sidecar is read-dominant. Drift is a *read-side* artifact
  computed from data already in the sidecar; it does not need a
  new write channel on the project's own code or docs.
- AGENTS.md "non-intrusive": the sidecar must not silently rewrite
  documentation. Drift is a notification, not a fix.
- AGENTS.md "lightweight": drift detection must run on demand, not
  as a daemon. CI integration is the right home for "drift runs on
  every change".

## Considered options

For each decision, the chosen option is first; rejected
alternatives follow.

## Decision

### 1. Two views, no new schema for either

The two views are derived from existing data, not new data:

- **doc-architecture view:** a subset of gap-1 `decision` records
  tagged with `architecture`. The tag is the only signal; no
  new record kind is introduced. A team's convention is "any
  decision whose body describes a service, module, boundary, or
  dependency direction gets the `architecture` tag".
- **actual-architecture view:** a union of gap-2 `edge` records
  (all kinds; future filtering by `change_impact` if useful),
  gap-1 `decision` records that have been `superseded` (the trail
  of how the architecture has actually moved), and gap-5
  `constraint` records (the boundaries that are actually being
  held).

A `cortex ask --kind decision --tag architecture` returns the
doc-side view; the actual-side view is composed inside the sidecar
on demand and is not exposed as a standalone query in v0+.

- **Why reuse the tag:** a new `kind` value would force every
  reader to learn a sixth category. The `tag` field already
  exists on every record, and "this is architecture" is a property
  of the content, not a kind of thing.
- **Why no standalone actual-side query:** the actual view is
  defined as the *complement* of the doc view, not as an
  independent dataset. Surfacing it independently would create
  the impression that it has meaning on its own; the meaning only
  emerges in the comparison.

### 2. A new data structure: `drift`

Drift is not a record and not an edge; it is a third shape, the
fourth parallel structure alongside `record`, `edge`, and `signal`.
The shape (informal, v0+):

```
drift {
  id               ulid
  kind             missing | extra | stale | phantom | eroded
  doc_ref          record id (doc side)
  actual_ref       edge id or record id (actual side)
  scope            path glob | module id | service id
  description      string
  detected_at      ISO 8601
  detected_by      manual | derived
  confidence       high | medium | low
  severity         hard | soft
  suggested_action string
  status           active | superseded | archived
  tags             free-form tags
}
```

- **Why a separate structure, not a `record`:** a `record` is a
  declaration about one scope. A `drift` is a *comparison* between
  two scopes' representations. The same argument as
  edge-vs-record applies: a free-text body saying "X's doc says Y
  but actual says Z" forces every reader to parse prose to learn
  the relationship. First-class structure lets `cortex drift`
  answer questions cleanly.
- **Why not an `edge`:** an edge is a relationship between two
  scopes. A drift is a relationship between two *representations*
  of scopes, plus a verdict on their mismatch. The unit is
  different.
- **Why `suggested_action` is required:** scope dim 2 says "why"
  is load-bearing. A drift without a suggested next action is
  trivia; the sidecar's job is to make the next step obvious.

### 3. The five drift kinds

The taxonomy is closed at five kinds. Each is a single, recognizable
mismatch; new kinds enter only via a new ADR.

| Kind | Doc says | Actual says |
| --- | --- | --- |
| `missing` | X depends on Y | the edge X→Y is not present |
| `extra` | (nothing about X→Y) | the edge X→Y is present |
| `stale` | the doc cites decision D | D has been `superseded` |
| `phantom` | the actual cites decision D | D is not in the doc |
| `eroded` | constraint C is `hard` | the surrounding evidence (edge churn, multiple `soft` supersedes, ...) suggests C is no longer held |

- **Why five, not three:** collapsing `missing` and `extra` into
  "asymmetric" loses the information about which side is the
  surprise. Collapsing `stale` and `phantom` into "uncited" loses
  the difference between "your doc is out of date" and "your
  code references a ghost". Five is the smallest taxonomy that
  preserves both distinctions.
- **Why `eroded` is a kind, not a derived signal:** the sidecar
  records human-stated `hard` constraints; the actual state is
  observable in the edge store. The combination is a *kind of
  drift*, not a health signal, because the right response is "go
  verify the boundary", not "act on the data".

### 4. Read entry points: `drift` and `context`

- `cortex drift --scope <path> [--kind <k>]` returns the
  active drifts for the scope, sorted by `severity` desc, then
  `detected_at` desc.
- `cortex context --file <path>` is extended to include a "drift
  summary" section: the top three active drifts for the file's
  scope, with a count of how many more exist. This is the entry
  point a host agent is expected to call.
- `cortex drift --detect` (v0+) and `--action <verb>` (v0+) are
  subcommand verbs; they do not exist in v0. The subcommand
  itself is added in v0+ so the discoverability is there, but it
  returns "not yet implemented" in v0.

### 5. Write channels: manual plus a derived channel for the detector

- **Human via CLI:** `cortex add drift --kind missing --doc-ref
  <id> --actual-ref <id> --scope ...` produces a drift file.
- **Human via file edit:** direct JSON write, `cortex validate`
  enforces the schema.
- **Derived via detector** (v0+): the detector writes to
  `.cortex/drift/_derived/`, with `detected_by = derived` and a
  `confidence` derived from the rule that fired. A human must
  accept it before it joins the canonical set. The discipline
  matches `_seed/` and `_derived/` from earlier ADRs: machine
  claims are not records until a human says so.

### 6. Boundaries the sidecar holds

In addition to the principles in `AGENTS.md` and the boundaries
fixed by the earlier ADRs:

- **The sidecar does not rewrite documentation.** A drift is a
  notification; rewriting `README.md` or an ADR to match the
  actual is the human's job. Doing this automatically would
  invert the trust model: the sidecar would be editing files
  in the project, which is exactly the "edits the host agent"
  violation AGENTS.md forbids.
- **The sidecar does not adjudicate doc vs actual.** A drift is
  a statement of mismatch. Which side is right is a team
  decision; the sidecar surfaces the mismatch and stays out of
  the verdict.
- **The sidecar does not delete old documentation.** Stale docs
  stay where they are; the way to express "this is no longer
  true" is to write a `superseded` decision, not to delete the
  doc.
- **The sidecar does not force drift resolution.** A drift can
  stay `active` indefinitely; closing it is a deliberate human
  action (`cortex resolve drift <id>`).
- **The sidecar does not run drift detection on a timer.**
  Detection is `cortex drift --detect` (manual), a CI job, or a
  user-installed hook. Daemon-free compute per ADR-0006.

## Consequences

### Positive

- Gap 4 uses data the sidecar is already collecting (gap-1
  decisions, gap-2 edges, gap-5 constraints). No new write
  channels on the project's own code or docs.
- The five-kind taxonomy is small and stable. New kinds enter by
  ADR; existing readers do not break.
- `cortex context --file` becomes the single host-agent entry
  point that surfaces everything: constraints (gap 5), active
  decisions (gap 1), derived signals (gap 3, v0+), and active
  drift (gap 4, v0+). The integration cost for the first host
  agent stays small.
- The "no doc rewrite" boundary is sharp. AGENTS.md
  non-intrusive is preserved across the new feature.

### Negative

- Until v0+, the host agent gets no drift help from the
  sidecar. This is the deliberate v0 scope.
- Drift detection depends on the doc-side being tagged. If a
  team does not tag architecture decisions, detection has
  nothing to compare against. The CLI's `cortex drift --scope X`
  can still list manually-added drifts, so the feature is
  useful even when tagging is sparse.
- `eroded` is the most rule-laden kind. Its `suggested_action`
  is "verify with team", which is the honest answer — the
  sidecar cannot know whether a boundary is intentionally soft
  now. The boundary is preserved by not pretending otherwise.

### Neutral

- The four data structures (`record`, `edge`, `signal`, `drift`)
  share metadata conventions (severity, source-type analog,
  status, freshness) and live in sibling directories. A future
  schema-evolution rule is straightforward: append-only, per
  ADR.
- `cortex drift` is added as a subcommand in v0+ with a
  placeholder. The placeholder is the right shape for the first
  v0+ ADR to fill in.

## Risks and rollback triggers

| Risk | Trigger | Action |
| --- | --- | --- |
| Drift detection noise drowns the signal | Empirical: median project accumulates > 100 `active` drifts in a month | Demote `confidence: medium` to default-hide; tighten the detector rules |
| A team rewrites docs based on unverified drift | External: a real report | The boundary is the sidecar's contract; the fix is documentation, not product. Reiterate in README. |
| `eroded` triggers arguments instead of resolutions | Empirical: most `eroded` drifts sit open > 90 days | Drop `eroded` to a manual-only kind (no derived detection); keep the schema entry for human use |
| Drift detection introduces LLM-shaped heuristics | The first v0+ detector ADR pulls in any model API | Reject the ADR; rule-based only |
| The four-data-structure design (`record`, `edge`, `signal`, `drift`) becomes a usability tax | A new contributor cannot figure out where to put something | A `docs/data-model.md` index is required when the third structure lands; we are at four, so write it |

## Open questions deferred to the first v0+ gap-4 ADR

- What is the first detector's rule set? (`missing` is the
  easiest; `eroded` is the hardest.)
- Where does detection run? (CI job is the default
  recommendation; manual and hook are also options.)
- How are `doc_ref` and `actual_ref` resolved? (record id and
  edge id are obvious; could the `scope` glob be enough?)
- Is `cortex drift --action` worth shipping, or is a free-form
  `suggested_action` in the drift record enough?
- How do drift records age out? (`active` forever, or `superseded`
  by a new `decision` that resolves the mismatch?)

## Final note: v0 scope reaffirmed

This ADR explicitly reaffirms that v0 does not include gap 4. The
detection depends on data the sidecar does not collect in v0
(gap-2 edges), and the schema and CLI for `drift` are not needed
to prove the v0 read/write loop. The first v0+ ADR for gap 4 will
inherit this ADR's decisions and is expected to be the third
v0+ ADR (after the static-edges ADR for gap 2 and the
first-source-adapter ADR for gap 3).
