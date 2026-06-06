# ADR-0001: Memory model and I/O surface

- **Status:** Accepted
- **Date:** 2026-06-06
- **Scope:** Meta-layer decisions that constrain every per-gap ADR that follows.

## Context and problem statement

The five gaps in `docs/scope.md` all share a common substrate: a record
of project context that an agent can read before it acts. Before any
gap-specific design is meaningful, we must lock down:

- the shape of a single memory unit (so records from different gaps
  compose cleanly),
- where the memory physically lives (so the read and write models are
  possible),
- how the host agent retrieves context (the only behavior that
  actually delivers value),
- who is allowed to write and how (scope calls this out as the harder
  problem),
- how a zero-history project gets its first useful memory.

This ADR is the first ADR and also establishes the ADR format used by
this directory.

## Decision drivers

- The five non-negotiable principles in `AGENTS.md` (non-intrusive,
  sidecar-not-replacement, lightweight, simple config, cross-platform,
  cross-agent, standard install, English-only content).
- The cross-cutting dimensions in `docs/scope.md`: "why" is
  load-bearing, information decays, granularity must match the
  project's natural unit, read-dominant with imperfect writes, shared
  surface across actors.

## Considered options

For each meta question the considered options are listed in the
following subsections. The chosen option is the first one; the rest
are recorded for the record.

## Decision

### 1. Memory unit: strongly typed records on a shared schema

Every record conforms to the following shape (informal, v0):

```
record {
  id               ulid/UUID
  kind             decision | constraint | signal
  scope            path glob | area id
  subject          one-line statement
  body             long-form detail
  reason           required; "why" is load-bearing
  evidence         links / refs (ADR, PR, issue, incident, log line)
  source_type      human | agent-suggested | scrape | derived
  author           author identity
  created_at       ISO 8601
  last_verified_at ISO 8601
  status           active | superseded | archived
  supersedes       id of older record
  superseded_by    id of newer record
  tags             free-form tags
  freshness        derived; recomputed from last_verified_at
}
```

- **Why A over B (free-form note with tags):** scope already partitions
  the surface into three categories; strong typing maps directly and
  makes the read API predictable, which the principles require.
- **Why A over C (entity-relationship graph):** the graph is the
  underlying model — every record hangs off a `scope` and may be
  linked by `evidence` / `supersedes` — but we do not yet need full
  graph traversal. Gap 2 (blast radius) is where the graph becomes a
  first-class store; until then, flat records are enough.

Schema is expressed in JSON Schema. Additions are append-only; old
records keep working.

### 2. Storage: project-local git-tracked directory by default

Default: `.cortex/` inside the project repo, git-tracked. Layout:

```
.cortex/
  decisions/      0001-<slug>.json, 0002-<slug>.json, ...
  constraints/    same shape
  signals/        reserved for gap 3
  index.json      optional cache, regeneratable
  config.json     project-level config (trusted sources, scope aliases)
```

Optional personal overlay: `~/.config/cortex/projects/<project-hash>/`,
gitignored, same layout. Read precedence: project layer first, then
overlay.

- **Why A over B (gitignored project dir):** scope dim 6 calls for a
  shared surface across actors; team-visible records belong with the
  code.
- **Why A over C (global store):** records are versioned with the code
  they describe; a decision made at v0.5 should not silently disappear
  at v1.0.
- **Why A over D (embedded DB):** DBs are hard to diff, hard to
  review, and violate "near-zero config" and "lightweight".

One record per file, identified by ULID, is chosen to keep PR diffs
clean and merges cheap.

### 3. Read interface: CLI first, MCP as opt-in adapter

The sidecar ships a single binary. Read surface is exposed as CLI
subcommands; an MCP server is an opt-in adapter reachable as a
subcommand.

```
cortex ask    --scope <path> [--kind <type>] [--tag <t>] [--limit N]
cortex get    <id>
cortex list   [--kind] [--status]
cortex context --file <path> [--radius N]      # convenience aggregate
cortex add    <kind> --scope ...               # write
cortex init                                     # cold start
cortex scan                                     # opt-in scrape
cortex verify  <id>                             # refresh last_verified_at
cortex supersede <old-id> --new ...
cortex mcp-serve                                # opt-in MCP adapter
```

Output defaults to human-readable text; `--json` is for agents.

- **Why A over B (MCP primary):** not every host agent supports MCP;
  CLI is the minimum common denominator. MCP is a per-agent adapter
  that comes later, consistent with "adding support for a new agent is
  a localized change".
- **Why A over C (file drop):** file drops make the sidecar push into
  the host's context window, which violates non-intrusive. Pull-based
  access leaves the host in control.
- **Why A over D (local socket / HTTP):** introduces a daemon, port
  management, lifecycle. Violates lightweight.

The sidecar does not inject itself into the host agent's system
prompt. Host agents are pointed at the CLI by a single line of
user-managed config (e.g. "before editing, run `cortex context --file
<path>`").

### 4. Write model: human-first, agent-proposed, opt-in scrape

Four write channels, ordered by trust:

1. **Human via CLI:** `cortex add <kind> ...` produces a record file.
2. **Human via file edit:** direct JSON edit; `cortex validate` checks
   it against the schema.
3. **Agent-proposed:** the host agent emits a structured block; `cortex
   ingest` writes it to `.cortex/_proposed/` with `source_type =
   agent-suggested`. A human must `accept` to move it to the canonical
   directory.
4. **Scrape:** `cortex scan` reads README, CONTRIBUTING, AGENTS,
   CLAUDE.md, docs/adr*, docs/decisions, .github templates, runbooks.
   Candidates are written to `.cortex/_seed/` with `source_type =
   scrape` and `confidence: low`. A human must `accept` to promote.

`source_type` is mandatory on every record. The read layer down-weights
non-`human` sources so that the imperfect-write assumption from scope
dim 5 is honored at the product level.

- **Why B over A (human-only):** the host agent is one of the actors
  per scope dim 6; forbidding it from proposing makes the sidecar
  useless precisely when the agent notices something interesting.
- **Why B over C (aggressive scrape as default):** scraped content is
  noisy and often wrong; leading with it poisons the trust of the
  read layer.

The sidecar never `git commit` on its own. Commits are made by humans
through their normal git workflow.

### 5. Cold start: `cortex init` seeds from existing docs

`cortex init` is idempotent and run once per project. It scans a fixed
list of paths (README*, CONTRIBUTING*, AGENTS*, CLAUDE*, LICENSE*,
docs/adr*, docs/decisions, docs/architecture, .github templates,
RUNBOOK*, docs/runbooks) and emits candidate records to
`.cortex/_seed/`, all marked `source_type = scrape`. The user reviews
candidates with `accept` / `edit` / `reject`, or accepts in bulk with
`--min-confidence`.

Parsing is local and rule-based. No LLM call is made. `cortex init
--no-write` is supported for a dry run. Skipping `init` is valid; the
sidecar is usable from empty.

- **Why B+C over A (empty):** the first-day experience decides whether
  the tool gets used. A new project with no memory gives no value.
- **Why rule-based over LLM scraping:** violates "lightweight" and
  "no bundled LLM calls the user did not request". LLM-assisted seed
  is a future opt-in, not the default.

## Consequences

### Positive

- Every per-gap ADR (gap 1–5) can now be written against a fixed
  substrate; ADR-by-ADR we only argue about gap-specific shape.
- The read layer (CLI) is identical for every gap, so adapters per
  host agent are thin and local.
- "Why" is forced into the schema, directly addressing scope dim 2.
- Records travel with code versions, addressing the team-shared
  surface in scope dim 6.
- Imperfect writes are handled at the product level via `source_type`
  and the `_seed/` / `_proposed/` staging dirs.

### Negative

- One file per record means the directory grows with usage. Mitigated
  by periodic `cortex prune` of `archived` records; the directory is
  still inspectable by `ls`, which is a feature not a bug.
- Scraped candidates can be wrong. Mitigated by mandatory human
  review before promotion.
- The CLI-only read surface requires host-agent owners to add one
  line of config; we cannot escape this without violating
  non-intrusive.

### Neutral

- The schema will grow. Backwards compatibility is by append-only
  JSON Schema.
- Per-host-agent adapters (Claude Code, Cursor, Aider) are a future
  concern; this ADR explicitly does not pick their surface.

## ADR format (established by this ADR)

Subsequent ADRs in this directory follow the MADR shape:

- **Status:** Proposed | Accepted | Superseded | Deprecated
- **Date:**
- **Context and problem statement:**
- **Decision drivers:**
- **Considered options:**
- **Decision:** (with subsections per question when the ADR covers
  multiple questions)
- **Consequences:** Positive / Negative / Neutral
- **Supersedes / Superseded by:** (when applicable)

Numbering: `NNNN-kebab-case-title.md`, sequential, no gaps.
