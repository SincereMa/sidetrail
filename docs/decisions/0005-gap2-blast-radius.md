# ADR-0005: Blast radius (gap 2) — edge model, v0+ scope, boundaries

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Constrained by:** ADR-0001 (memory model and I/O surface),
  ADR-0002 (gap 5), ADR-0003 (technology stack),
  ADR-0004 (gap 1, amendment to `kind`)

## Context and problem statement

Gap 2 in `docs/scope.md` is the invisible-blast-radius problem: when
the host agent edits a file, it cannot see which services, modules,
subprojects, or shared contracts depend on that file — statically, at
runtime, or operationally. The agent's need is bidirectional
(outbound consumers; inbound interface consumers), the dependencies
come in four kinds (static, runtime, operational, social), and the
sidecar is asked to provide guidance, not just facts
("breaking / additive / coordinated").

This ADR, on top of the substrate fixed by ADR-0001, chooses:

- the shape of the new data structure needed to express
  relationships (records are not enough),
- the schedule — gap 2 does not enter v0; the first v0+ ADR after
  this one will pick a narrow slice,
- the boundary at which the sidecar stops: it does not evaluate
  impact, does not notify, does not visualize, does not silently
  merge derived edges.

## Decision drivers

- Gap 2's primary product surface is `boundaries/constraints`, with
  `decisions` as secondary. Boundaries are relationships between
  things; records declare properties of one thing. The shapes differ.
- "Information lives outside the code" (scope dim 1): a dependency
  edge is a piece of context the code does not carry in a
  host-agent-readable form.
- v0 already proves the read/write loop for "human- and
  agent-suggested" memory. Gap 2 introduces "machine-derived" memory;
  this is a different write model and should be validated separately.
- AGENTS.md "lightweight": automated dependency analysis is exactly
  the kind of feature that can quietly pull in heavy language
  tooling. v0 must not.

## Considered options

For each decision, the chosen option is first; rejected alternatives
follow.

## Decision

### 1. A new data structure: `edge`, parallel to `record`

A new schema is introduced, separate from the record schema of
ADR-0001. The shape (informal, v0+):

```
edge {
  id              ulid
  kind            static | runtime | operational | social
  from_scope      path glob | module id | service id | area id | person id
  to_scope        same
  from_symbol     string, optional (function / class / field)
  to_symbol       string, optional
  contract_type   api | event | schema | rpc | queue | db_table |
                  deploy | secret | env_var | person, optional
  change_impact   breaking | additive | coordinated
  severity        hard | soft
  description     short
  evidence        links / refs
  source_type     human | derived | scrape
  author          author identity
  created_at      ISO 8601
  last_verified_at ISO 8601
  status          active | superseded | archived
  tags            free-form tags
}
```

- **Why a new structure, not a `record` with a `relates_to` field:**
  records declare properties of one scope. Edges declare a property
  of a *pair* of scopes. A `record` whose body says "X depends on Y"
  forces every reader to parse prose to learn the relationship. A
  first-class `edge` lets `cortex deps` / `cortex rdeps` /
  `cortex impact` answer questions structurally, which is the whole
  point of the gap.
- **Why reuse metadata fields (severity, source_type, status,
  freshness):** they are the same concept across records and edges
  (a stale `edge` is the same kind of bad as a stale `record`). One
  model of metadata, two shapes of payload.
- **Why `change_impact` is required, not optional:** scope explicitly
  asks for guidance, not just facts. An edge without
  `change_impact` is exactly the "facts only" shape the gap says is
  not enough. Default at write time is `additive` (the safe claim);
  the human or upstream agent overrides when they have evidence.
- **Why `from_symbol` and `to_symbol` are optional in v0+:** most
  useful edges are file-level / module-level. Symbol-level edges are
  a v0+ refinement; making them optional now keeps the schema simple
  and the writer fast.

### 2. Storage: `.cortex/edges/`, parallel to other kind directories

```
.cortex/
  decisions/
  constraints/
  signals/
  edges/         # gap 2
  config.json
  index.json
```

Derived edges live under `.cortex/edges/_derived/` before human
review, mirroring the `_seed/` / `_proposed/` convention from
ADR-0001 and ADR-0002. A human accepts them into `edges/`.

### 3. v0 scope: gap 2 is not in v0

v0 does not implement gap 2. The schemas, metadata conventions, and
boundaries are locked by this ADR, but no CLI commands are added and
no scanner is shipped. The four kind-specific sub-projects enter v0+
in this order:

1. **Static edges** — first v0+ ADR. Pick one or two languages to
   prove the scanner model; expand from there. v0+ must justify the
   first language choice in its own ADR rather than this one.
2. **Runtime and operational edges** — second v0+ ADR. Manual writes
   + opt-in scrapers (OpenAPI, protobuf, docker-compose, k8s). The
   runtime scanner is the most likely place LLM-shaped heuristics
   would creep in; the LLM-freeness is preserved by treating these
   as opt-in, never default.
3. **Trail walker** (`cortex trail <id>`) — the gap-1 follow-up; not
   gap-2 strictly, but it depends on the edge store to walk causal
   links between reasoning records.
4. **Impact aggregation in `cortex context`** — the read-side polish:
   `cortex context --file <path>` returns constraints, decisions,
   and a `change_impact`-sorted summary of what would break.

**Social edges** are deferred past v0+. They bring privacy and
attribution concerns (linking a `person_id` in the repo) that need
their own ADR with a different threat model.

- **Why static first:** static edges have the strongest "machine
  can derive it" story and the weakest "human must articulate it"
  story. They are the cheapest to add and the most likely to be
  useful on day one of v0+. The risk is they make v0+ "feel like a
  linter"; the boundary in section 5 is what keeps that from
  happening.
- **Why runtime/operational second:** the data sources (OpenAPI
  specs, compose files, k8s manifests) exist in real projects but
  are inconsistent. The scanner design has to be more opinionated
  than the static case, so it gets its own ADR with a real
  reference input.
- **Why social last:** out of v0+ scope. Linking people to code in
  a versioned file is a policy decision, not a technical one.

### 4. Write channels: same as records, with a derived channel for scans

The three write channels from ADR-0001 and ADR-0002 carry over
unchanged:

- **Human via CLI:** `cortex add edge --from ... --to ... --kind
  static --change-impact breaking` produces an edge file.
- **Human via file edit:** direct JSON write, `cortex validate`
  enforces the schema.
- **Derived via scanner** (v0+ only): the scanner writes to
  `.cortex/edges/_derived/` with `source_type = derived`. A human
  must accept it before it joins the canonical set.

`_proposed/` (agent suggestion) and `_seed/` (scrape) channels from
ADR-0001 are inherited for gap 2 as well. The "machine-derived"
channel is new and is the one most likely to need careful UX
design; that UX is the v0+ static-edges ADR's problem, not this
one's.

### 5. Boundaries the sidecar holds

In addition to the principles in `AGENTS.md` and the boundaries in
ADR-0002:

- **The sidecar does not evaluate impact size.** It returns edges
  sorted by `change_impact`; it does not say "this is a small
  change" or "this is a large change". The host agent decides.
- **The sidecar does not notify the affected owners.** It surfaces
  edges and their `change_impact`; the host agent (or a future
  out-of-scope tool) decides whether to mention `@alice` or open a
  PR in `service-billing`.
- **The sidecar does not visualize the graph.** It returns
  structured data over the CLI. Visualization is a host concern.
- **The sidecar does not silently merge derived edges.** A scanner
  output lands in `_derived/` and stays there until a human moves
  it. This is the gap-2 analog of the `_seed/` discipline in
  ADR-0001: machine claims are not records until a human says so.
- **The sidecar does not enforce a `change_impact` value.** It
  requires the field (because without it the edge is not useful)
  but does not second-guess the value. The human is the source of
  truth.
- **The sidecar does not introduce language-specific heavy
  tooling.** Each v0+ scanner is its own ADR and must justify its
  dependency footprint. The default is "no scanner, manual
  edges only".

## Consequences

### Positive

- The data shape for gap 2 is fixed. The first v0+ ADR for static
  edges only has to argue about scope (which languages, what
  counts as a "from" symbol), not the schema.
- A clear "machine-derived" channel exists; the trust model from
  ADR-0001 (source_type down-weighting) extends to edges with no
  new work.
- The `change_impact` field is the thing that turns the sidecar
  from "yet another dependency tool" into "the thing the host
  agent actually wants to consult". A future ADR can refine the
  enum without a schema break.
- The "v0 does not include gap 2" line is sharp. v0 ships with
  fewer surprises and v0+ has a clear first ADR to write.

### Negative

- Until v0+, a host agent working on a project that has no records
  but rich dependencies gets no blast-radius help from the sidecar.
  This is the deliberate v0 scope, but it is worth saying out loud.
- `change_impact` being required at write time adds friction. The
  writer is asked to make a judgment the data does not always
  support. The CLI default of `additive` is the safety valve; the
  `severity` field is the lever for "I'm not sure, mark it soft".
- Cycle handling in `cortex impact` (when `--depth > 1` traverses
  a cycle) is not specified here. The first v0+ static-edges ADR
  must pick a strategy.

### Neutral

- The four `kind` values are stable; adding a fifth
  (e.g. `temporal` for time-windowed deploy order) is additive.
- The schema for `edge` is its own JSON Schema file, living next
  to the record schema. The storage layout treats them as
  siblings.

## Risks and rollback triggers

| Risk | Trigger | Action |
| --- | --- | --- |
| The first v0+ static-edges ADR picks a language whose scanner pulls in heavy tooling | Dependency footprint of the chosen scanner > 1 MB or requires cgo | The ADR is rejected; pick a different first language, or defer to "manual edges only" |
| `change_impact` friction makes people write fewer edges | Empirical observation in v0+ | Make `change_impact` optional with default `additive`; re-evaluate after a release |
| Cycle handling in `--depth > 1` becomes a footgun | A real report of a misleading transitive result | Cap `--depth` at 1 in `cortex context`; require an explicit flag for deeper walks |
| Social edges become a privacy liability | A request to add them arrives before the threat model exists | Open a new ADR; do not absorb into gap 2 |
| The `_derived/` workflow is too heavy and people ignore it | Empirical observation | Move derived edges into a separate `edges.derived.jsonl` and let the CLI auto-merge low-confidence edges; re-evaluate the trust model |

## Open questions deferred to the first v0+ ADR

- Which language(s) does the first static scanner support?
- What is the cycle-handling policy in `cortex impact`?
- What is the `--depth` default for `cortex context`?
- Does the static scanner produce symbol-level edges by default or
  only file-level?
