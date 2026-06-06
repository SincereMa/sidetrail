# AGENTS.md — Cortex SideMark

> A sidecar tool that adds long-lived auxiliary memory to mainstream AI agents — without modifying them.

## Mission

Cortex SideMark is an **auxiliary / sidecar tool** for mainstream AI agents. It runs *alongside* an existing host agent (e.g. Claude Code, Cursor, Aider) and adds long-lived memory — decisions, constraints, health signals, project context — without forking, patching, or injecting into the host.

The word *side* is load-bearing. Cortex SideMark is an **addition**, never a replacement.

## Non-negotiable principles

These are hard constraints, not aspirations. Any change that violates them is a regression.

### Architectural

- **Non-intrusive.** Observe and record; never edit the host agent. If a feature can only be done by touching the host, redesign the feature.
- **Sidecar, not replacement.** Never positioned, documented, or architected as a competing agent.
- **Lightweight.** No heavy runtime, no bundled LLM calls the user did not request, no surprise transitive dependencies. "Doesn't overpower" is the bar.
- **Simple configuration.** Near-zero config to get value; sensible defaults; everything else opt-in.
- **Cross-platform.** macOS, Linux, Windows. No platform-specific paths, syscalls, or shell calls without an abstraction layer.
- **Cross-agent.** Adapters for multiple host agents are first-class. Adding support for a new agent is a localized change, not a rewrite.
- **Standard install.** Single binary on `PATH` or a well-known package manager command. No bespoke installer.

### Process

- **English-only content.** Documentation, code, comments, and resource files are written in English. The only exception is user-supplied text intentionally in another language (e.g. a quoted message or a manually entered translation). No mixed-language commits.
- **Protect the mission.** The principles in this file are the source of truth. If a proposed change, refactor, or task could conflict with the mission — in any way, however minor — the agent must stop, surface the conflict in detail (what the change does, which principle it touches, and why it is at risk), and wait for explicit developer confirmation before proceeding.

## Product surface

The categories of auxiliary memory Cortex SideMark records. New categories need a concrete use case, not just an idea.

- **Decisions** — choices the user has made, with reasoning, so future sessions do not re-litigate them.
- **Boundaries / constraints** — explicit *do not do* rules or hard limits on a project.
- **Health data** — project health signals (test status, lint status, stale files, etc.) that a host agent can pull before acting.

## Repository status

- **Greenfield.** Source is intentionally absent. The first PR establishes the layout — do not scaffold a large tree speculatively.
- **Undecided.** Language, framework, storage backend, and IPC mechanism are not set yet. Record such decisions as ADRs in `docs/decisions/` (create the directory on the first decision).
- **Adapter designs.** Per-host-agent adapter specifications go under `docs/agents/` (create when the first adapter is designed).

## Workflow conventions

- **Do not invent install steps.** Until install is implemented, link to the actual command in `README.md` or a script. Never write instructions that do not run.
- **Do not document features that do not exist.** If a markdown file claims a capability, the code in this repo must run it.
- **Update this file when the picture changes.** New agent adapter, refined principle, or major decision (language, storage, IPC) — reflect it here.

## Pointers

| Path | Purpose |
| --- | --- |
| `README.md` | User-facing description and the real install command (once it exists). |
| `docs/scope.md` | The problems Cortex SideMark exists to address; the input to subsequent ADRs. |
| `docs/decisions/` | Architectural decision records (ADRs). |
| `docs/agents/` | Per-host-agent adapter specifications. |
