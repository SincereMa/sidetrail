# Go Code Overview

This file describes the SideTrail Go code structure, main responsibilities, call chain, and directory hierarchy.

## Directory structure

- `main.go`
- `cmd/sidetrail/*.go`
- `internal/record/*.go`
- `internal/storage/store.go`
- `internal/schema/*.go`
- `internal/version/version.go`

## Main responsibilities

- `main.go`
  - Program entry point; calls `sidetrail.Execute()`.

- `cmd/sidetrail/root.go`
  - Builds the CLI root command and registers 13 subcommands.
  - Sets the `sidetrail` command name, long help text, and version template.

- `cmd/sidetrail/store_root.go`
  - Defines `storeDirName = ".sidetrail"`.
  - Implements `findStoreRoot(start)` to walk upward from CWD and find `.sidetrail/`.
  - Implements `resolveStoreRoot(explicit)` to validate a user-supplied store path.
  - `init` uses a separate `resolveProjectRoot(explicit)` because `.sidetrail/` may not exist yet.

- `cmd/sidetrail/init.go`
  - Scans canonical project paths from `initScanPaths`.
  - Builds seed candidates from README, CONTRIBUTING, AGENTS, CLAUDE, LICENSE, RUNBOOK, docs/adr*, docs/decisions, docs/architecture, and `.github` templates.
  - Uses `collectSeeds` to expand globs, read files, and skip binary or unreadable files.
  - Writes seed records under `.sidetrail/_seed/` via `storage.WriteSeed()`.
  - Supports `--no-write` dry run.

- `cmd/sidetrail/add.go`
  - Loads a record file with `record.LoadFile()`.
  - Resolves the store root with `resolveStoreRoot()`.
  - Uses `storage.NewStore(root).Write(r)` to write the record into the appropriate kind directory.
  - Guards idempotency by checking whether the record ID already exists in the store.

- `cmd/sidetrail/get.go`
  - Resolves the store root.
  - Uses `storage.Get(id)` to look up a record.
  - Supports exact ID and unique prefix matching.
  - Prints raw JSON by default, or a one-line summary with `--human`.

- `cmd/sidetrail/list.go`
  - Resolves the store root.
  - Lists all records or a specific kind.
  - Validates `--kind` and rejects unknown kinds.
  - Supports `--limit` and `--json` output.
  - Default table output is tab-separated for agent parsing.

- `cmd/sidetrail/ask.go`
  - Resolves the store root.
  - Requires `--scope` and queries by scope pattern.
  - Optionally filters by kind and tag.
  - Uses `storage.Ask(scope, kind, tag, limit)`.

- `cmd/sidetrail/context.go`
  - Resolves the store root.
  - Requires `--file` and accepts `--radius` and `--limit`.
  - Uses `storage.ContextFor(file, radius, limit)` to include file scope and ancestor scopes.
  - Does not require the file to exist on disk.

- `cmd/sidetrail/verify.go`
  - Resolves the store root.
  - Reads the record with `storage.Get(id)`.
  - Updates `LastVerifiedAt` to the current UTC second.
  - Writes the record back to the same path.

- `cmd/sidetrail/supersede.go`
  - Loads a replacement record file with `record.LoadFile()`.
  - Reads the old record with `storage.Get(oldID)`.
  - Marks the old record `Status = "superseded"` and sets `SupersededBy`.
  - Sets `newRec.Supersedes = oldID` when missing.
  - Writes both old and new records in one operation.
  - Supports `--dry-run`.

- `cmd/sidetrail/promote.go`
  - Resolves the store root.
  - Reads `.sidetrail/_seed/` entries.
  - Lists available seeds when no args are supplied.
  - Promotes selected IDs or `--all` seeds into the active store by calling `s.Write(r)`.
  - Removes promoted seed files after successful write.

- `cmd/sidetrail/draft.go`
  - Validates the command kind.
  - Generates a new ID with `record.NewID()`.
  - Builds a draft record and writes it under `.sidetrail/_draft/`.
  - Uses `defaultAuthor()` to infer an author from the environment.

- `cmd/sidetrail/status.go`
  - Resolves the store root.
  - Reads the record with `storage.Get(id)`.
  - Validates status transitions against `validTransitions`.
  - Applies one atomic status change with optional `--dry-run`.

- `cmd/sidetrail/health.go`
  - Resolves the store root.
  - Reads all records with `storage.ListAll()`.
  - Computes counts by kind, counts by status, unique scopes, active supersede chains, stale records, and date range.
  - Supports structured JSON and human-readable output.

- `cmd/sidetrail/validate.go`
  - Reads one or more files.
  - Calls `schema.ValidateRecord(data)` for each file.
  - Emits per-file success or failure results.
  - Returns non-zero when any validation fails.

- `internal/record/record.go`
  - Defines the canonical `Record` type.
  - Declares `Kind`, `SourceType`, and `Severity` enums with `Valid()` methods.
  - Implements `NewID()` using ULID with `crypto/rand` entropy.
  - Implements `Slug()` for file-safe slugs.

- `internal/record/load.go`
  - Implements `LoadFile(path)`.
  - Reads bytes, validates with `schema.ValidateRecord`, then unmarshals JSON.
  - Ensures returned records are schema-valid.

- `internal/record/match.go`
  - Implements `MatchScope(recordScope, pattern)`.
  - Matches exact scope or strict descendant paths.
  - Does not interpret glob syntax.

- `internal/record/rand.go`
  - Provides `randReader` to adapt `crypto/rand` for ULID.

- `internal/storage/store.go`
  - Defines `Store` rooted at a store directory.
  - Implements `Write`, `WriteSeed`, and `WriteDraft`.
  - Uses atomic write via temporary files and `os.Rename`.
  - Implements `Read`, `List`, `ListAll`, `ListKind`, `Get`, `Ask`, and `ContextFor`.
  - `Get` first tries exact ID match, then prefix match with ambiguity detection.

- `internal/schema/schema.go`
  - Embeds `record.schema.json` with `//go:embed`.
  - Compiles the schema once in `init()`.
  - Exposes `ValidateRecord(data)`.

- `internal/schema/json.go`
  - Wraps `json.Unmarshal` for the schema package.

- `internal/version/version.go`
  - Exposes build-time `Version` and `Commit` variables.
  - These values are injected via ldflags.

## Call chain and directory hierarchy

| Layer | File | Responsibility | Call direction |
| --- | --- | --- | --- |
| Entry | `main.go` | Program entry point | -> `sidetrail.Execute()` |
| CLI root | `cmd/sidetrail/root.go` | Build root command, register subcommands | -> `newXCmd()` |
| Common resolution | `cmd/sidetrail/store_root.go` | Resolve `.sidetrail/` | Used by most commands |
| Command | `cmd/sidetrail/init.go` | Create `_seed/` seeds | -> `storage.WriteSeed()` |
| Command | `cmd/sidetrail/add.go` | Import records | -> `record.LoadFile()` -> `storage.Write()` |
| Command | `cmd/sidetrail/get.go` | Fetch records | -> `storage.Get()` |
| Command | `cmd/sidetrail/list.go` | List records | -> `storage.ListAll()` / `storage.ListKind()` |
| Command | `cmd/sidetrail/ask.go` | Query by scope | -> `storage.Ask()` |
| Command | `cmd/sidetrail/context.go` | File context | -> `storage.ContextFor()` |
| Command | `cmd/sidetrail/verify.go` | Refresh verification | -> `storage.Get()` -> `storage.Write()` |
| Command | `cmd/sidetrail/supersede.go` | Supersede records | -> `record.LoadFile()` -> `storage.Get()` -> `storage.Write()` |
| Command | `cmd/sidetrail/promote.go` | Promote seeds | -> `storage.Read()` -> `storage.Write()` |
| Command | `cmd/sidetrail/draft.go` | Create drafts | -> `record.NewID()` -> `storage.WriteDraft()` |
| Command | `cmd/sidetrail/status.go` | Transition status | -> `storage.Get()` -> `storage.Write()` |
| Command | `cmd/sidetrail/health.go` | Health reporting | -> `storage.ListAll()` |
| Record model | `internal/record/record.go` | Data model, enums, ULID, slug | Used by `storage` / `cmd` |
| Record loading | `internal/record/load.go` | File load + validation | -> `schema.ValidateRecord()` |
| Storage layer | `internal/storage/store.go` | Disk I/O and query | Called directly by commands |
| Schema | `internal/schema/schema.go` | JSON Schema validation | Embeds schema |
| Version | `internal/version/version.go` | Version metadata | Injected into CLI by `root.go` |

## Key relationship summary

- `main.go` -> `root.go` -> `runX()`.
- The command layer primarily depends on `internal/storage`, `internal/record`, and `internal/schema`.
- `storage` depends on `record` for ID/slug and scope matching.
- `record.LoadFile()` depends on `schema.ValidateRecord()`.
- `schema` handles schema validation; `internal/schema/json.go` only wraps JSON decoding.
- `init` / `promote` form the seed lifecycle; `ask` / `context` are the main agent read paths.
