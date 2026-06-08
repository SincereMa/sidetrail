# ADR-0001: Memory model and I/O surface

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintains:** This is the first ADR and establishes the memory model half of the full ADR half-follow.

## Context and background

This ADR defines what goes into `.sidetrail/` and why. The store
holds all records: one JSON file per record, one directory per
record kind. Before any gap-specific design is meaningful, we need
to nail down:

- how records are uniquely identified (globally unique, diff-friendly, sortable by time)
- why they are physically laid out the way they are (read-heavy, human-readable, human-editable)
- which gap they address (only a human can reliably say)
- what all records share (schema, kind, scope, subject, status)
- how they are accessed (CLI, JSON output, human-friendly)

This ADR is the first ADR and establishes the memory model half
that all subsequent ADRs build on.

## Decision drivers

- The first gap was discovered in `AGENTS.md` (non-intrusive, sidecar, lightweight, config-light, cross-agent, standard install, English-only).
- The writing pattern is write-once, read-many: why is the load-bearing data, and it must be human-readable and human-editable.

## Considered

For a gap-specific design to be meaningful we need to consider
the following baseline questions:

- how records are uniquely identified (globally unique, diff-friendly, sortable by time)
- why they are physically laid out the way they are (read-heavy, human-readable, human-editable)
- which gap they address (only a human can reliably say)
- what all records share (schema, kind, scope, subject, status)
- how they are accessed (CLI, JSON output, human-friendly)

## Decision

### 1. Memory unit: strongly typed, human-readable, human-editable

Every record follows the schema below (field names and constraints):

```
{
  id:           ULID/UUID
  kind:         decision | constraint | signal
  scope:        slug | array-of-slugs
  subject:      one-line summary
  body:         long-form detail
  reason:       why this is load-bearing
  id-link:      link/reference (ADR number, PR number, inline link)
  author:       human | agent-suggested | ai | did
  author-id:    identity
  created_at:   ISO8601
  last_verified_at: ISO8601
  status:       active | superseded | archived | hidden
  superseded_by:    id of record that replaces this one
  tags:         free-form tags
  hash:         hash; last_verified_at
}
```

- **Why JSON Schema (not YAML):** already available in the ecosystem; parsing is deterministic and has a stable API; widely supported.
- **Why not a graph DB (not a file layout):** the underlying data is small enough that a file layout is sufficient; a graph DB would be overkill.

### 2. Storage: per-kind directories, slug-based file names by default

Default store root: `.sidetrail/` inside the project's git root. Layout:

```
.sidetrail/
  decisions/
    0001-<slug>.json
    0002-<slug>.json
    ...
  constraints/
    <slug>
  signals/
    <slug>
  experiments/
    <slug>
  incidents/
    <slug>
  _seed/
    <initial-slug>.json
```

Override root: `~/.config/sidetrail/jobs/<project-hash>/` for global context.

- **Why JSON (not YAML):** deterministic, no surprising whitespace, no language-level ambiguity.
- **Why git-friendly:** `.sidetrail/` is already in the project root; all records can be committed and reviewed with `git diff`.
- **Why not a database (not a flat DB):** a database adds a dependency; a file layout is zero-config and works offline.

On-disk files are identified by ULID, then slugified for the PR diff-friendly sort and readability.

### 3. Reading surface: CLI first, MCP adapter later

This is a single binary. Read surfaces:

```
sidetrail ask --scope <scope> [--kind <kind>] [--limit N]
sidetrail get <id>
sidetrail list [--kind <kind>] [--limit N]
sidetrail context --file <path> [--radius N] [--limit N]
sidetrail validate <file>... [--json]
sidetrail verify <id>
sidetrail supersede <old-id> --new <file>
sidetrail init [--root <project>] [--no-write]
sidetrail add <kind> --subject "..." --body "..." --author "..." [--scope ...] [--tag ...]
```

Output defaults to human-readable; `--json` for machine parsing.

- **Why CLI first (not an API):** many agents use CLI; MCP is an agent adapter; a CLI is the universal adapter.
- **Why file-based (not stdin/stdout):** file-based avoids stdin/stdout; pull-based agents have non-intrusive limitations.
- **Why no local socket (not HTTP):** adds complexity; V0 is low-overhead.

This design avoids the agent needing to know about the store layout, the schema details, or the host agent's conventions. The agent discovers the store root by walking upward from CWD.

### 4. Write lifecycle: human-first, agent-suggested, idempotent

Write path is intentionally limited by design:

1. **Human via CLI:** `sidetrail add <kind> ...` writes a file.
2. **Human via file editor:** edit JSON directly; `sidetrail validate` checks it.
3. **Agent-suggested:** the agent writes a block; `sidetrail init` populates `.sidetrail/_seed/` with `author=agent-suggested`. A human runs `sidetrail promote` to review.
4. **Seed:** `sidetrail init` scans README, CONTRIBUTING, AGENTS, CLAUDE.md, docs/ad-*, docs/decisions, docs/architecture, .github/ templates, RUNBOOK, docs/unblock and writes candidates to `.sidetrail/_seed/` with `author=ai` and `status=hidden`.

`author` is an audit trail. The agent does not write directly; it suggests. The human reviews.

- **Why human-only writes:** the agent infers from data in the store; forbidding direct writes keeps the audit trail clean while the agent is still learning.
- **Why agent-suggested (not agent-default):** avoids silent, opinionated changes; the human decides.

### 5. Cold start: `sidetrail init` discovers existing docs

`sidetrail init` reads the project root and finds candidate files (README*, CONTRIBUTING*, AGENTS*, CLAUDE*, LICENSE*, docs/ad-*, docs/decisions, docs/architecture, docs/unblock, docs/*.github templates, RUNBOOK, docs/unblock) and writes initial candidates to `.sidetrail/_seed/` with `author=ai` and `status=hidden`. This allows bulk-init with `--include-hidden` flag.

Paired with promote-and-review. No LLM calls by default. Skipping init is valid; the sidecar is usable from empty.

- **Why discover (not ask):** the files are already in the repo; asking would require LLM calls, which is heavy.
- **Why skip-valid (not required):** lightweight; the sidecar works from empty.

## Consequences

### Positive

- Every gap (reasoning trails, blast radius, difficulty, drift, constraints) has a natural home in the record model; ADRs 2-5 can specialize without redesigning the base.
- The agent can ask questions and get answers without knowing the host agent's conventions.
- Human review is the default; the agent suggests, the human decides.

### Negative / accepted

- The file layout is not a graph; if the project grows to need graph traversal, ADR-0001's layout must be re-evaluated.
- JSON Schema validation is a dependency; if the schema grows too complex, we may need a more expressive validator.

### Risks and mitigations

| Risk | Trigger | Mitigation |
| --- | --- | --- |
| File layout becomes slow at scale | >10k records without indexing | Build a lightweight index (map of scope to ids) and a full-text search on subject/body |
| Schema validation bogs down the pipeline | >10k records with complex allOf | Cache validation results; run validation lazily on read |
| Git diffs become noisy | Many records changing in a single PR | Use `git add -p` and per-record commits |
| Agent overwrites human data (gap 3) | Agent runs init or add without human review | Agent writes only to `_seed/`; human must run `sidetrail promote` |
| Write conflicts on concurrent edits | Multiple agents editing the same record | File locking; retry on conflict; design for single-writer |

## Follow-up

### Prior

- Each gap ADR (gaps 1-5) will build on the memory model established here; each ADR builds on its predecessor.
- The agent's read path (context, ask, get) is in the CLI surface in v0.

### Next

- Agent adapter code for OpenCode, Claude Code, Cursor, etc. is in `docs/agents/`.
- Seed lifecycle and promotion flow.
- Health-data and architecture-drift analytics.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): the memory model and I/O surface
  changes must be recorded before merge.
- The schema file lives at `internal/schema/record.schema.json` and is
  embedded in the binary via `//go:embed`.
