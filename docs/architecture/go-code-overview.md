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
  - Builds the CLI root command and registers 5 agent-driven subcommands.
  - Sets the `sidetrail` command name, long help text, and version template.

- `cmd/sidetrail/store_root.go`
  - Defines `storeDirName = ".sidetrail"`.
  - Implements `findStoreRoot(start)` to walk upward from CWD and find `.sidetrail/`.
  - Implements `resolveStoreRoot(explicit)` to validate a user-supplied store path.

- `cmd/sidetrail/init.go`
  - Creates a `.sidetrail/` directory at the project root.
  - Uses `resolveProjectRoot(explicit)` because `.sidetrail/` may not exist yet.

- `cmd/sidetrail/add.go`
  - Loads a record file with `record.LoadFile()`.
  - Resolves the store root with `resolveStoreRoot()`.
  - Uses `storage.NewStore(root).Write(r)` to write the record into the appropriate kind directory.
  - Guards idempotency by checking whether the record ID already exists in the store.

- `cmd/sidetrail/update.go`
  - Resolves the store root.
  - Reads the existing record with `storage.Get(id)`.
  - Reads the update JSON file and merges fields.
  - Writes the updated record back with `storage.Write()`.

- `cmd/sidetrail/context.go`
  - Resolves the store root.
  - Requires `--file` and accepts `--radius` and `--limit`.
  - Uses `storage.ContextFor(file, radius, limit)` to include file scope and ancestor scopes.
  - Does not require the file to exist on disk.
  - Provides `writeRecordsJSON` and `writeRecordsTable` helpers for output formatting.

- `cmd/sidetrail/health.go`
  - Resolves the store root.
  - Reads all records with `storage.ListAll()`.
  - Computes counts by kind, counts by status, unique scopes, active supersede chains, stale records, and date range.
  - Supports structured JSON and human-readable output.

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
  - Implements `Write` for atomic writes via temporary files and `os.Rename`.
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
| Command | `cmd/sidetrail/init.go` | Create `.sidetrail/` directory | -> `os.MkdirAll()` |
| Command | `cmd/sidetrail/add.go` | Import records | -> `record.LoadFile()` -> `storage.Write()` |
| Command | `cmd/sidetrail/update.go` | Update records | -> `storage.Get()` -> `json.Merge` -> `storage.Write()` |
| Command | `cmd/sidetrail/context.go` | File context | -> `storage.ContextFor()` |
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
- `context` is the main agent read path; `add` and `update` are the main write paths.
