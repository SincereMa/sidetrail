# Sequences

Three end-to-end sequence diagrams for the flows a host
agent or a human operator is most likely to run. Each
diagram names the source files the messages come from so
that the trace can be followed in the code.

The three flows are:

- [`context`](#context---read-dominant-primary-path): the read-
  dominant path a host agent calls before acting.
- [`add`](#add---write-path): the primary write path for
  recording decisions, constraints, and signals.
- [`update`](#update---partial-update): the path for updating
  existing records with partial JSON fields.

## `context` — read-dominant primary path

Source: [`cmd/sidetrail/context.go`](../../cmd/sidetrail/context.go),
[`cmd/sidetrail/store_root.go`](../../cmd/sidetrail/store_root.go),
[`internal/storage/store.go`](../../internal/storage/store.go).

```mermaid
sequenceDiagram
  autonumber
  participant U as Host agent / shell
  participant CLI as cmd/sidetrail (context.go)
  participant SR as store_root.go
  participant FS as Filesystem
  participant STO as storage.Store

  U->>CLI: sidetrail context --file <path> [--radius N] [--limit N] [--json]
  CLI->>SR: resolveStoreRoot(opts.root)
  SR->>FS: os.Getwd / filepath.Abs / os.Stat (walk up to .sidetrail/)
  FS-->>SR: store path
  SR-->>CLI: root
  CLI->>STO: NewStore(root)
  CLI->>STO: ContextFor(file, radius, limit)
  STO->>STO: ancestorScopes(file, radius) -> walk parent dirs
  STO->>STO: ListAll() -> walk 5 kind dirs -> Read each file
  STO->>STO: scopeMatchesAny(r.Scope, patterns) per record
  STO->>STO: sortByCreatedAtDesc -> truncate to limit
  STO-->>CLI: []*Record
  alt --json
    CLI->>CLI: writeRecordsJSON (json.MarshalIndent)
  else default
    CLI->>CLI: writeRecordsTable (TSV)
  end
  CLI-->>U: records
```

## `add` — write path

Source: [`cmd/sidetrail/add.go`](../../cmd/sidetrail/add.go),
[`cmd/sidetrail/store_root.go`](../../cmd/sidetrail/store_root.go),
[`internal/record/load.go`](../../internal/record/load.go),
[`internal/storage/store.go`](../../internal/storage/store.go).

```mermaid
sequenceDiagram
  autonumber
  participant U as Agent / shell
  participant CLI as cmd/sidetrail (add.go)
  participant LOAD as record.LoadFile
  participant SCH as schema.ValidateRecord
  participant SR as store_root.go
  participant STO as storage.Store
  participant FS as Filesystem

  U->>CLI: sidetrail add <file>
  CLI->>LOAD: LoadFile(args[0])
  LOAD->>FS: os.ReadFile
  LOAD->>SCH: ValidateRecord(bytes)
  SCH-->>LOAD: ok / error
  LOAD-->>CLI: record
  CLI->>SR: resolveStoreRoot(opts.root)
  SR-->>CLI: root
  CLI->>STO: NewStore(root)
  CLI->>STO: Get(r.ID) -- idempotency check
  STO-->>CLI: exists / not found
  alt already exists
    CLI-->>U: error: record already exists
  else new record
    CLI->>STO: Write(r)
    STO->>FS: mkdir kindDir + write tmp + rename
    STO-->>CLI: path
    CLI-->>U: print id + path
  end
```

## `update` — partial update

Source: [`cmd/sidetrail/update.go`](../../cmd/sidetrail/update.go),
[`cmd/sidetrail/store_root.go`](../../cmd/sidetrail/store_root.go),
[`internal/storage/store.go`](../../internal/storage/store.go).

```mermaid
sequenceDiagram
  autonumber
  participant U as Agent / shell
  participant CLI as cmd/sidetrail (update.go)
  participant SR as store_root.go
  participant STO as storage.Store
  participant FS as Filesystem

  U->>CLI: sidetrail update <id> --file <json-file>
  CLI->>CLI: os.ReadFile(json-file) -> parse updates
  CLI->>SR: resolveStoreRoot(opts.root)
  SR-->>CLI: root
  CLI->>STO: NewStore(root)
  CLI->>STO: Get(id)
  STO->>FS: walk 5 kind dirs for <id>-.json
  STO-->>CLI: existing record
  CLI->>CLI: json.Marshal(existing) -> merge updates -> json.Unmarshal
  CLI->>STO: Write(updated)
  STO->>FS: write tmp + rename (overwrite in place)
  STO-->>CLI: path
  CLI-->>U: print id
```
