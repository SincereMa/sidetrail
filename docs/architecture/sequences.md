# Sequences

Three end-to-end sequence diagrams for the flows a host
agent or a human operator is most likely to run. Each
diagram names the source files the messages come from so
that the trace can be followed in the code.

The three flows are:

- [`ask`](#ask---read-dominant-primary-path): the read-
  dominant path a host agent calls before acting.
- [`init`](#init---write-path): the heaviest write path,
  used once per project to seed the store.
- [`supersede`](#supersede---two-record-transaction): the
  only path that mutates two records in lockstep.

## `ask` — read-dominant primary path

Source: [`cmd/sidetrail/ask.go`](../../cmd/sidetrail/ask.go),
[`cmd/sidetrail/list.go`](../../cmd/sidetrail/list.go),
[`cmd/sidetrail/store_root.go`](../../cmd/sidetrail/store_root.go),
[`internal/storage/store.go`](../../internal/storage/store.go),
[`internal/record/match.go`](../../internal/record/match.go).

```mermaid
sequenceDiagram
  autonumber
  participant U as Host agent / shell
  participant CLI as cmd/sidetrail (ask.go)
  participant SR as store_root.go
  participant FS as Filesystem
  participant STO as storage.Store
  participant REC as internal/record (MatchScope)

  U->>CLI: sidetrail ask --scope <pat> [--kind K] [--tag T] [--limit N] [--json]
  CLI->>SR: resolveStoreRoot(opts.root)
  SR->>FS: os.Getwd / filepath.Abs / os.Stat (walk up to .sidetrail/)
  FS-->>SR: store path
  SR-->>CLI: root
  CLI->>STO: NewStore(root)
  CLI->>STO: Ask(scope, kind, tag, limit)
  STO->>STO: ListAll() -> walk 5 kind dirs -> Read each file
  STO->>REC: MatchScope(r.Scope, scope) per record
  REC-->>STO: bool
  STO->>STO: filter kind / tag -> sortByCreatedAtDesc -> truncate to limit
  STO-->>CLI: []*Record
  alt --json
    CLI->>U: writeRecordsJSON (json.MarshalIndent)
  else default
    CLI->>U: writeRecordsTable (TSV)
  end
```

## `init` — write path

Source: [`cmd/sidetrail/init.go`](../../cmd/sidetrail/init.go),
[`cmd/sidetrail/store_root.go`](../../cmd/sidetrail/store_root.go),
[`internal/record/record.go`](../../internal/record/record.go),
[`internal/storage/store.go`](../../internal/storage/store.go).

```mermaid
sequenceDiagram
  autonumber
  participant U as Operator
  participant CLI as cmd/sidetrail (init.go)
  participant PR as resolveProjectRoot
  participant FS as Filesystem
  participant REC as internal/record
  participant STO as storage.Store

  U->>CLI: sidetrail init [--root <project>] [--no-write]
  CLI->>PR: resolveProjectRoot(opts.root)
  PR->>FS: os.Getwd / filepath.Abs / os.Stat
  FS-->>PR: projectRoot
  PR-->>CLI: projectRoot
  CLI->>CLI: collectSeeds(projectRoot) -> walk initScanPaths
  loop Each scan path
    alt contains glob meta
      CLI->>FS: filepath.Glob
    else is a directory
      CLI->>FS: os.ReadDir (non-recursive)
    else regular file
      CLI->>FS: os.Stat
    end
    CLI->>FS: os.Open + io.ReadFull (<= 500 B)
    FS-->>CLI: bytes (skip NUL, treat as binary)
  end
  CLI->>CLI: sort by path

  alt --no-write
    CLI->>U: reportPlan (would-scan / would-write list)
  else real write
    CLI->>FS: os.MkdirAll(<root>/.sidetrail)
    CLI->>STO: NewStore(storeDir)
    loop Each seedCandidate
      CLI->>CLI: buildSeedRecord (Kind=decision, Status=active, SourceType=scrape)
      CLI->>REC: NewID() -> ULID
      REC-->>CLI: id
      CLI->>STO: WriteSeed(r)
      STO->>FS: mkdir .sidetrail/_seed/ + atomic write <id>-<slug>.json
    end
    CLI->>U: scanned N paths, wrote M seeds under .sidetrail/_seed
  end
```

## `supersede` — two-record transaction

Source: [`cmd/sidetrail/supersede.go`](../../cmd/sidetrail/supersede.go),
[`cmd/sidetrail/store_root.go`](../../cmd/sidetrail/store_root.go),
[`internal/record/load.go`](../../internal/record/load.go),
[`internal/storage/store.go`](../../internal/storage/store.go).

```mermaid
sequenceDiagram
  autonumber
  participant U as Operator / agent
  participant CLI as cmd/sidetrail (supersede.go)
  participant LOAD as record.LoadFile
  participant SCH as schema.ValidateRecord
  participant SR as store_root.go
  participant STO as storage.Store
  participant FS as Filesystem

  U->>CLI: sidetrail supersede <old-id> --new <file> [--dry-run]
  CLI->>LOAD: LoadFile(opts.new)
  LOAD->>FS: os.ReadFile
  LOAD->>SCH: ValidateRecord(bytes)
  SCH-->>LOAD: ok / error
  LOAD-->>CLI: newRec
  CLI->>SR: resolveStoreRoot(opts.root)
  SR-->>CLI: root
  CLI->>STO: NewStore(root)
  CLI->>STO: Get(old-id)
  STO->>FS: walk 5 kind dirs for <id>-.json exact -> prefix
  STO-->>CLI: oldRec
  CLI->>CLI: assert newRec.ID != oldRec.ID
  CLI->>CLI: if newRec.Supersedes == "" -> = oldRec.ID
  CLI->>CLI: oldRec.Status = "superseded"<br/>oldRec.SupersededBy = newRec.ID

  alt --dry-run
    CLI->>U: print would-update-old / would-write-new
  else real commit
    CLI->>STO: Write(oldRec) -> atomic rename overwrites in place
    STO->>FS: write tmp + rename
    CLI->>STO: Write(newRec) -> new file
    STO->>FS: mkdir kindDir + write tmp + rename
    CLI->>U: print oldRec.ID \n newRec.ID
  end
```
