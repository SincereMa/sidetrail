# SideTrail Roadmap

## Purpose

This document records the current implementation status, the work that remains, and the next development milestones for SideTrail. It is grounded in the existing codebase, the product scope in `docs/scope.md`, the architectural and agent principles in `AGENTS.md`, and the current Go implementation.

## Current status (as of this workspace)

### Completed

- Core CLI implemented in Go using Cobra.
- Store root discovery and `.sidetrail/` layout implemented.
- On-disk record model implemented with one JSON file per record.
- JSON schema validation implemented and embedded in `internal/schema`.
- Record ID generation using ULID and stable slug-based file naming.
- Agent-driven CLI surface with seven commands:
  - `context` — read records relevant to a file
  - `add` — validate and store a record
  - `update` — update an existing record with partial JSON
  - `health` — report project health signals
  - `init` — create a `.sidetrail/` directory
  - `seed --files` — generate agent prompt from project documents
  - `seed --apply` — apply agent-generated records with conflict detection
- Record kinds supported: `decisions`, `constraints`, `signals`, `experiments`, `incidents`.
- Scope and product problem statements exist in `docs/scope.md`.
- Contributor and agent design principles exist in `AGENTS.md`.
- Tests exist for command and internal packages.

### Verified implementation coverage

- `cmd/sidetrail/context.go`: file-anchored context aggregation.
- `cmd/sidetrail/add.go`: record validation and store write.
- `cmd/sidetrail/update.go`: partial record updates via JSON merge.
- `cmd/sidetrail/health.go`: project health signals, stale detection, JSON output.
- `cmd/sidetrail/init.go`: create `.sidetrail/` directory.
- `cmd/sidetrail/seed.go`: seed command with --files and --apply modes.
- `internal/seed/prompt.go`: prompt generation from project documents.
- `internal/seed/conflict.go`: conflict detection by scope+subject matching.
- `internal/storage/store.go`: atomic write, list, get, context, prefix resolution.
- `internal/record/record.go`: record model, enums, slug hashing, ULID entropy.

## Outstanding tasks

### 1. Host-agent adapter implementation

Status: not implemented

- Adapter guidance exists only in documentation (`docs/agents/README.md`, `docs/agents/opencode.md`).
- There is no actual adapter code or sample host integration in this repository.
- Action items:
  - Select a first host-agent integration target (OpenCode, Claude Code, Cursor, or other).
  - Implement a minimal adapter/service that calls `sidetrail context` and `sidetrail add` for the host agent.
  - Add end-to-end examples or sample `SKILL`/adapter files for the target agent.
  - Ensure integration respects non-intrusive, cross-platform, lightweight principles.

### 2. Documentation cleanup and alignment

Status: mostly implemented

- `README.md` accurately summarizes the current CLI surface.
- `docs/agents/opencode.md` has been cleaned up and contains a complete OpenCode adapter guide.
- Action items:
  - Add or update adapter docs for the actual host integration target.

### 3. Testing and quality verification

Status: improved

- Tests cover commands and internal packages.
- `go vet ./...` passes clean.

### 4. Release and packaging

Status: verified

- `scripts/install.sh` and `scripts/install.ps1` exist.
- `go.mod`, `go.sum`, and `.goreleaser.yml` exist.

## Next milestones

1. **Host-agent adapter scaffolding**
   - Choose first integration target.
   - Add adapter integration docs and sample configuration.
   - Optionally add a minimal adapter binary or script.

2. **Documentation and onboarding**
   - Add a short contributor onboarding section in `README.md`.

3. **Quality and testing**
   - Add regression tests for conflict cases and host integration flows.

## References

- `README.md`
- `AGENTS.md`
- `docs/scope.md`
- `cmd/sidetrail/*.go`
- `internal/storage/store.go`
- `internal/record/record.go`
- `internal/schema/record.schema.json`

---

*Updated for simplified CLI surface on 2026-06-13.*
