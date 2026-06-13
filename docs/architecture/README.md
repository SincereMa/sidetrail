# Architecture

Reference documentation for the SideTrail codebase as it
actually exists in this repository. The diagrams and feature
inventory in this directory are derived from a full read of
the Go source under `main.go`, `cmd/sidetrail/`, and
`internal/`. They are kept here as a map for new contributors
and for AI agents that need to orient before editing.

The principles this project must respect (non-intrusive,
sidecar, lightweight, cross-platform, cross-agent) live in
[AGENTS.md](../AGENTS.md). The five product-surface
problems are in [scope.md](../scope.md). The architectural
decisions that shaped the current shape are in
[decisions/](../decisions/).

## Contents

| File | Purpose |
| --- | --- |
| [topology.md](./topology.md) | Module and library graph: packages, subcommands, the on-disk store layout, and the external dependencies in `go.mod`. |
| [sequences.md](./sequences.md) | Three end-to-end sequence diagrams for the flows a host agent or a human operator is most likely to run: `ask`, `init`, `supersede`. |
| [go-code-overview.md](./go-code-overview.md) | Go code structure, directory hierarchy, and call chain relationships. |
| [README.md](./README.md) | This file. The feature inventory, organised by subcommand. |

## Feature inventory

The CLI exposes one root command (`sidetrail`) and five
agent-driven subcommands. Every subcommand lives in its own file
under `cmd/sidetrail/` and is wired in
[`cmd/sidetrail/root.go:20`](../../cmd/sidetrail/root.go).

| # | Subcommand | Read / write | Primary behaviour | Key collaborators |
| --- | --- | --- | --- | --- |
| 1 | `init` | write | Create a `.sidetrail/` directory at the project root. The store is usable from empty; `init` is optional. | `resolveProjectRoot`, `os.MkdirAll`. |
| 2 | `add` | write | Read a record file, validate it, and write it to the store. Idempotency-guarded: a second add of the same id is an error. | `record.LoadFile`, `storage.Write`. |
| 3 | `update` | write | Update an existing record with partial JSON fields. Reads the current record, merges provided fields, and writes it back. | `storage.Get`, `json.Unmarshal`, `storage.Write`. |
| 4 | `context` | read | File-anchored aggregate. Returns the records whose scope equals the file path or any of its ancestor directories, up to `--radius` levels. | `storage.ContextFor`. |
| 5 | `health` | read | Scan the store and report project health signals: record counts by kind/status, unique scopes, active supersede chains, stale records. Supports `--json` and `--stale-days`. | `storage.ListAll`. |

### Cross-cutting capabilities

- **Store root resolution.** `store_root.go` defines
  `storeDirName = ".sidetrail"`, `findStoreRoot` (upward walk
  from CWD, no symlink following), and `resolveStoreRoot`
  (`--root` override). Every subcommand except `init` uses
  `resolveStoreRoot`. `init` uses `resolveProjectRoot`
  instead, because the `.sidetrail/` directory does not exist
  yet at the moment of first run.
- **Record identity.** `record.NewID` returns a ULID via
  `oklog/ulid/v2`'s `Timestamp` + `Monotonic` entropy, with
  entropy drawn from `crypto/rand` through
  [`internal/record/rand.go`](../../internal/record/rand.go).
  ULIDs sort lexicographically by creation time.
- **Slug.** `record.Slug` lowercases, retains `[a-z0-9]`,
  compresses runs of any other character into a single dash,
  strips trailing dashes, caps length at 48 characters, and
  falls back to `"record"` for empty input. The slug and the
  ULID compose into a stable filename.
- **On-disk layout.** `<root>/<kinddir>/<id>-<slug>.json`,
  one file per record. The five kind directories are
  `decisions/`, `constraints/`, `signals/`, `experiments/`,
  `incidents/`.
- **Atomic writes.** `storage.writeToDir` writes to
  `<path>.tmp` and `os.Rename`s it into place. A failed
  rename cleans up the temp file.
- **Schema validation.** The record schema
  ([`internal/schema/record.schema.json`](../../internal/schema/record.schema.json),
  Draft 2020-12) is embedded with `//go:embed`, compiled
  once in `init()`, and reused on every call. The
  `LoadFile` helper validates before unmarshalling, so any
  `*Record` it returns is known-valid.
- **Record model.** `Record` is the union of all kind-
  specific fields defined in ADR-0001, ADR-0002, and
  ADR-0004. Optional fields use pointer or empty-slice types
  so "absent" and "zero value" survive JSON round-trips.
- **Enums.** `Kind` (5 values: `decision`, `constraint`,
  `signal`, `experiment`, `incident`), `SourceType` (4
  values), `Severity` (`hard` | `soft`). Each exposes a
  `Valid()` method.
- **Version metadata.** `internal/version.Version` and
  `internal/version.Commit` are set at build time via
  `ldflags` (the injection point is configured in
  `.goreleaser.yml`). The root command wires them into
  `cobra.Command.Version` and a custom
  `SetVersionTemplate`.
