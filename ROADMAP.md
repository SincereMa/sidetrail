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
- Twelve subcommands are present and wired:
  - `init`
  - `validate`
  - `add`
  - `get`
  - `list`
  - `ask`
  - `context`
  - `verify`
  - `supersede`
  - `promote`
  - `draft`
  - `status`
  - `health`
- `init` seeds candidate records under `.sidetrail/_seed/` from canonical doc paths.
- `promote` moves seed records from `_seed/` into the active store by kind.
- Record kinds supported: `decisions`, `constraints`, `signals`, `experiments`, `incidents`.
- Embedded architecture/reference docs exist in `docs/architecture/README.md`.
- Scope and product problem statements exist in `docs/scope.md`.
- Contributor and agent design principles exist in `AGENTS.md`.
- Tests exist for command and internal packages.

### Verified implementation coverage

- `cmd/sidetrail/init.go`: project-root discovery, seed creation, `--no-write` dry run.
- `cmd/sidetrail/validate.go`: record schema validation.
- `cmd/sidetrail/add.go`: record import validation and store write.
- `cmd/sidetrail/get.go`: exact and prefix record lookup.
- `cmd/sidetrail/list.go`: record enumeration, kind filtering, JSON/table output.
- `cmd/sidetrail/ask.go`: scope-based record querying.
- `cmd/sidetrail/context.go`: file-anchored context aggregation.
- `cmd/sidetrail/verify.go`: refresh record freshness timestamp.
- `cmd/sidetrail/supersede.go`: two-record supersede workflow.
- `cmd/sidetrail/promote.go`: seed promotion from `_seed/` to active store.
- `cmd/sidetrail/draft.go`: draft record creation in `_draft/` for human review.
- `cmd/sidetrail/status.go`: record status transitions with validation.
- `cmd/sidetrail/health.go`: project health signals, stale detection, JSON output.
- `internal/storage/store.go`: atomic write, list, get, ask, context, prefix resolution, WriteDraft.
- `internal/record/record.go`: record model, enums, slug hashing, ULID entropy.

## Outstanding tasks

### 1. Seed lifecycle and promotion

Status: implemented

- `init` writes seed candidates to `.sidetrail/_seed/`.
- `promote` moves seed records from `_seed/` into the appropriate kind directory (decisions/, constraints/, etc.).
- Remaining items:
  - Decide whether `_seed` files should be included in `list`, `ask`, or `context` searches as candidates.
  - Document the seed lifecycle in `README.md` and `docs/architecture/README.md`.

### 2. Host-agent adapter implementation

Status: not implemented

- Adapter guidance exists only in documentation (`docs/agents/README.md`, `docs/agents/opencode.md`).
- There is no actual adapter code or sample host integration in this repository.
- Action items:
  - Select a first host-agent integration target (OpenCode, Claude Code, Cursor, or other).
  - Implement a minimal adapter/service that calls `sidetrail context` and `sidetrail ask` for the host agent.
  - Add end-to-end examples or sample `SKILL`/adapter files for the target agent.
  - Ensure integration respects non-intrusive, cross-platform, lightweight principles.

### 3. Record write UX and verification

Status: implemented

- `add` validates and writes records; `supersede` handles two-record swap.
- `draft` creates schema-valid draft records in `_draft/` for human review.
- `status` transitions records between statuses (active ↔ archived, active ↔ hidden).
- Valid transitions are enforced per kind (decisions/constraints/signals: active/superseded/archived; experiments: in_progress/succeeded/failed/inconclusive/abandoned; incidents: investigating/mitigated/resolved).
- `--dry-run` flag on status and supersede for safe preview.

### 4. Health-data and architecture-drift support

Status: implemented

- `health` command scans the store and reports project health signals.
- Reports: record counts by kind/status, unique scopes, active supersede chains, stale records (configurable `--stale-days`), date range.
- `--json` flag for structured output suitable for agent consumption.
- Stale records are flagged when `last_verified_at` exceeds the threshold and status is active/in_progress/investigating.

### 5. Documentation cleanup and alignment

Status: mostly implemented

- `README.md` accurately summarizes the current CLI surface.
- `docs/architecture/README.md` aligns with the codebase.
- `docs/agents/opencode.md` has been cleaned up and contains a complete OpenCode adapter guide.
- `docs/decisions/` ADRs 0001-0007 have been cleaned up (corruption tokens removed).
- Action items:
  - Add or update adapter docs for the actual host integration target.
  - Add a roadmap or milestone section to `README.md` once the next work phases are confirmed.

### 6. Testing and quality verification

Status: improved

- Tests cover commands and internal packages with 84.4% overall coverage.
- Added tests for `validate`, `promote`, `draft`, `status`, `health`, and `Kind.Valid`/`SourceType.Valid`/`Severity.Valid`.
- `go vet ./...` passes clean.
- Cross-platform builds verified for all 4 targets.

### 7. Release and packaging

Status: verified

- `scripts/install.sh` and `scripts/install.ps1` exist.
- `go.mod`, `go.sum`, and `.goreleaser.yml` exist.
- `goreleaser check` passes (configuration validated).
- Cross-platform builds verified: linux/amd64, linux/arm64, darwin/arm64, windows/amd64.

## Next milestones

1. **Seed documentation**
   - Document the seed lifecycle and promote workflow in `README.md`.
   - Decide whether seeds should appear in `list`/`ask`/`context`.

2. **Host-agent adapter scaffolding**
   - Choose first integration target.
   - Add adapter integration docs and sample configuration.
   - Optionally add a minimal adapter binary or script.

3. **Record model refinement**
   - Validate health and drift record shapes.
   - Improve query coverage for `ask` and `context`.

4. **Documentation and onboarding**
   - Add a short contributor onboarding section in `README.md`.

5. **Quality and testing**
   - Add regression tests for conflict cases and host integration flows.

## References

- `README.md`
- `AGENTS.md`
- `docs/scope.md`
- `docs/architecture/README.md`
- `cmd/sidetrail/*.go`
- `internal/storage/store.go`
- `internal/record/record.go`
- `internal/schema/record.schema.json`

---

*Generated from current repository audit on 2026-06-07.*
