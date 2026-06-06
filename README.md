# Cortex SideMark

> A sidecar that gives AI agents long-lived memory of the projects they work on — without modifying them.

Mainstream agents see the current state of a project. The reasoning
that produced that state is lost. The constraints that bind the
project live in the heads of the team, not in the code. The agent
treats every part of the project as equivalent, when in fact some
parts are hard to change safely and some are easy. And when the agent
edits a service, it has no reliable way to see which other services
depend on what it is about to change.

Cortex SideMark is a sidecar that records this missing context —
decisions, constraints, health signals, project state — and makes it
available to the host agent before it acts. It does not touch the
host agent. It does not replace it. It runs alongside it.

## The five gaps

1. **Reasoning trails for project evolution are lost between sessions.**

   > "This area was adjusted three times. The first attempt changed
   > X because of Y. The second attempt found a side effect Z and
   > was rolled back to X. The current state is the result of the
   > third attempt, chosen specifically to avoid Z."

2. **The blast radius of an edit is invisible across services, modules, and subprojects.**

   The agent can see the file it is changing, but not the services,
   modules, subprojects, or shared contracts that depend on that
   file — statically, at runtime, or operationally. Local
   correctness, global regression.

3. **Project difficulty is uneven and shifts, but the agent has no read on it.**

   Some modules are churn hotspots, some are bug clusters, some
   have eroding test coverage, some depend on stale libraries. The
   agent cannot prioritize, warn, slow down, or escalate review
   where it matters.

4. **Documented architecture drifts from actual architecture.**

   > "The architecture diagram shows service A and service B as
   > cleanly separated. Service B has, for the last eight months,
   > been reading service A's primary database directly. Every new
   > feature in B assumes the read. Removing it is a multi-quarter
   > project, not a refactor."

5. **Tribal constraints are nowhere the agent can read them.**

   > "Don't change the code in `legacy/billing/` — there's a
   > compliance review pending."
   >
   > "We use library Z despite its license because alternatives have
   > all been blocked by security."
   >
   > "That migration is paused; do not touch the bridge code until
   > Q3."

The formal definition of each gap — kinds, statuses, sub-cases,
surface mapping, cross-cutting dimensions — is in
[docs/scope.md](docs/scope.md). The non-negotiable principles that
shape how we solve them are in [AGENTS.md](AGENTS.md).

## How it works

Cortex SideMark is a **sidecar**, not a competing agent. It runs
alongside the host agent (Claude Code, Cursor, Aider, …) and
records long-lived context. It does not fork, patch, or inject into
the host.

A few hard constraints shape the design:

- **Non-intrusive.** Observe and record; never edit the host agent.
- **Lightweight.** No heavy runtime, no bundled LLM calls the user
  did not request.
- **Cross-platform and cross-agent.** macOS, Linux, Windows; one
  binary, one config; adding support for a new host agent is a
  localized change.
- **Simple configuration.** Near-zero config to get value;
  everything else opt-in.

The full principle set lives in [AGENTS.md](AGENTS.md).

## Status

Early development. The v0 CLI surface is in place; tagged builds
are published as GitHub releases. The problems the project exists
to address are recorded in [docs/scope.md](docs/scope.md); the
architectural decisions that follow from them live in
[docs/decisions/](docs/decisions/).

## Install

The fastest path is a one-liner. It downloads the latest release
for your OS and architecture, verifies the SHA-256, and drops the
`cortex` binary on disk.

```sh
curl -fsSL https://raw.githubusercontent.com/SincereMa/cortex-sidemark/main/scripts/install.sh | sh
```

On Windows, from PowerShell:

```powershell
iwr https://raw.githubusercontent.com/SincereMa/cortex-sidemark/main/scripts/install.ps1 -useb | iex
```

Both scripts accept flags. See `scripts/install.sh --help` or
`Get-Help ./scripts/install.ps1` for options. The binary lands in
`~/.local/bin` (or `%USERPROFILE%\bin` on Windows) by default; add
that directory to `PATH` if it is not already.

For a pinned version, set `CORTEX_VERSION=v0.1.0` (or pass
`--version` / `-Version`). For a custom install location, set
`CORTEX_INSTALL_DIR` (or pass `--dir` / `-InstallDir`).

If you prefer to install manually, download a release archive from
the [GitHub releases page](https://github.com/SincereMa/cortex-sidemark/releases),
verify it against the matching `*_checksums.txt`, and place the
`cortex` binary anywhere on `PATH`.

## Project layout

| Path | Purpose |
| --- | --- |
| `AGENTS.md` | Mission, principles, and pointers for contributors and host agents. |
| `LICENSE` | MIT license terms. |
| `CODE_OF_CONDUCT.md` | Community expectations. |
| `CONTRIBUTING.md` | How to file issues and submit changes. |
| `docs/scope.md` | The problems Cortex SideMark exists to address; the input to subsequent ADRs. |
| `docs/decisions/` | Architectural decision records. |
| `docs/agents/` | Per-host-agent adapter specifications. |

## CLI surface

The `cortex` binary is a read-dominant command. Most calls ask a
question; a few write a record. The `.cortex/` store is discovered
by walking upward from the current working directory unless
`--root` points elsewhere.

| Command | Purpose |
| --- | --- |
| `cortex add <file>` | Validate a record file and add it to the store. |
| `cortex get <id> [--human]` | Fetch a record by id (exact, then prefix). |
| `cortex list [--kind K] [--limit N] [--json]` | List records, newest first. |
| `cortex ask --scope <s> [--kind] [--tag] [--limit] [--json]` | Query records whose scope matches a pattern. |
| `cortex context --file <path> [--radius N] [--limit] [--json]` | Aggregate records relevant to a file path. |
| `cortex verify <id>` | Refresh a record's `last_verified_at`. |
| `cortex supersede <old-id> --new <file>` | Mark a record superseded and add a replacement. |
| `cortex init [--root <project>] [--no-write]` | Seed a `.cortex/` store from existing project docs. |
| `cortex validate <file>... [--json]` | Validate record files against the schema. |

The first host-agent integration point is `cortex context`: point
it at the file the agent is about to edit, and it returns the
records the team has filed about that file and its enclosing
scopes.

`cortex init` is the cold-start path. It scans the canonical
project paths and writes scrape-derived candidate records to
`.cortex/_seed/`. Skipping init is valid; the sidecar is usable
from empty.

## Contributing

Issues and pull requests are welcome. Please read
[CONTRIBUTING.md](./CONTRIBUTING.md) and
[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) first.

## License

[MIT](./LICENSE). Copyright (c) 2026 Cortex SideMark Authors.
