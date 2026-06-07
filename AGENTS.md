# AGENTS.md — SideTrail

Mission and pitch: [README.md](./README.md). This file is the
"don't miss this" guide for AI agents working in the repo.

## Status

- **Greenfield.** No source, no build, no tests, no CI. The first
  PR establishes the layout — do not scaffold a large tree
  speculatively.
- **Undecided stack.** Language, framework, storage, IPC, and
  host-agent target are not chosen. When chosen, record as ADRs
  in [docs/decisions/](./decisions/).

## Hard constraints (non-negotiable)

These are not aspirational. A change that violates them is a
regression. If a task asks for one, stop, surface the conflict,
and wait for explicit confirmation.

- **Non-intrusive.** Observe and record; never edit the host
  agent. Redesign any feature that requires touching the host.
- **Sidecar, not replacement.** Never position, document, or
  architect this as a competing agent.
- **Lightweight.** No heavy runtime, no bundled LLM calls the
  user did not request, no surprise transitive dependencies.
- **Simple config.** Near-zero config to get value; everything
  else opt-in.
- **Cross-platform.** macOS, Linux, Windows. No platform-specific
  paths, syscalls, or shell calls without an abstraction layer.
- **Cross-agent.** Adapters for multiple host agents are first-
  class; adding support for a new agent is a localized change.
- **Standard install.** Single binary on `PATH` or a well-known
  package manager command. No bespoke installer.
- **English-only content.** Documentation, code, comments, and
  resource files are in English. The only exception is user-
  supplied text intentionally in another language. No mixed-
  language commits.

## Scope and product surface

The project exists to address the five problems in
[docs/scope.md](./scope.md). A feature that does not address at
least one of them is out of scope.

Auxiliary memory is recorded in three categories. A new category
needs a concrete use case, not just an idea.

- **Decisions** — choices the user has made, with reasoning.
- **Boundaries / constraints** — explicit *do not do* rules or
  hard limits.
- **Health data** — project health signals an agent can pull
  before acting.

## Workflow

- **Commits.** English, imperative mood, 50-character subject
  ("Add X", not "Added X"). Conventional-commit prefix is
  encouraged (`docs:`, `feat:`, `fix:`, `chore:`, `refactor:`,
  `test:`).
- **Branches.** Short-lived, off `main`: `feat/<topic>`,
  `fix/<topic>`, `docs/<topic>`, `chore/<topic>`. One logical
  change per PR.
- **PRs.** Squash-merge default. The PR title becomes the
  squashed commit subject, so make it descriptive. The PR
  template has a checkbox confirming AGENTS.md principles are
  not violated — do not tick without genuinely checking.
- **Architectural decisions.** A change that picks a language,
  framework, storage, IPC, or host-agent adapter must add or
  update an ADR in [docs/decisions/](./decisions/) and link it
  from the PR.
- **Do not document features that do not exist.** A markdown
  claim must be backed by code that runs.
- **Do not invent install steps.** Until install is implemented,
  link to the actual command in `README.md` or a script. Never
  write instructions that do not run.

## Code conventions (binding)

Code in this repository must follow
[docs/code-conventions.md](./code-conventions.md). That document
is binding for every contributor, human or AI. Highlights that
AI agents most often get wrong:

- **Every exported name has a doc comment** starting with the
  name itself. `// ValidateRecord reports whether ...`, not
  `// Validates a record.`
- **Every package has a `// Package x ...` doc comment.**
- **Errors carry context**: `fmt.Errorf("read %q: %w", path, err)`.
  Use `%w`, not `%v`, when wrapping.
- **No shell calls**: `os/exec` invokes a binary with explicit
  args; `sh -c` and `cmd /c` are forbidden.
- **No raw path separators**: `filepath.Join` everywhere.
- **English-only** in code, comments, commit messages, and PR
  descriptions.

A PR that violates these rules is a regression even if its
tests pass.

## Pointers

| Path | Purpose |
| --- | --- |
| `README.md` | Project pitch; the real install command (once one exists). |
| `docs/scope.md` | The five problems this project addresses; input to ADRs. |
| `docs/code-conventions.md` | Binding code style and review rules. |
| `docs/decisions/` | Architectural decision records. |
| `docs/agents/` | Per-host-agent adapter specifications. |
| `CONTRIBUTING.md` | How to file issues and submit changes. |
| `.github/ISSUE_TEMPLATE/` | Bug report and feature request templates. |
| `.github/PULL_REQUEST_TEMPLATE.md` | PR template; includes the AGENTS.md compliance checkbox. |

Update this file when the picture changes: a new agent adapter
appears, a principle is refined, a major decision is made, or the
project leaves greenfield.
