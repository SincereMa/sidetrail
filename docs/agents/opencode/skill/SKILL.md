---
name: sidetrail
description: Load this skill when the user asks about project context, decisions, constraints, or health signals. Use before editing or planning changes to understand load-bearing context.
license: MIT
affinity: none
author: host-agent
usage: on-demand
---

## What this skill does

This skill tells you to use the **SideTrail** CLI binary
(`sidetrail`) for long-lived context. Context is stored in a
local `.sidetrail/` directory; each record is a JSON file with a
ULID, a kind (decision, constraint, signal), a scope, and a
subject. SideTrail is a sidecar — it does not modify your host
and it does not replace you. It only surfaces context when you ask.

The skill is available as a tool via the `skill` tool in your
tool list. After you load it, the human will be able to trigger
it in their project.

## When to use the SideTrail CLI

The `sidetrail` binary is your read interface. The human handles
writes; you handle reads. Everything below is read-only.

### Before editing a file

Run `sidetrail context --file <path>` before you edit a file.
This returns a JSON array of records whose scope matches the file
path. If the array is empty, there is no context for this file —
proceed normally. If the array has results, read them. They
contain load-bearing constraints and decisions that you must not
violate. Pay special attention to the `reason` field on each
record — that is the "why" you need to understand.

```sh
sidetrail context --file path/to/file.go
sidetrail context --file path/to/file.go --radius 1    # narrower
sidetrail context --file path/to/file.go --limit 20     # more results
```

### Before answering "what do we know about X?"

When the human asks about a topic and you suspect there is
project-specific context (e.g., "what do we know about billing?",
"why is auth done this way?", "why did we choose library X?"),
run:

```sh
sidetrail ask --scope <topic> --kind decision --json
sidetrail ask --scope <topic> --kind constraint --json
sidetrail ask --scope billing --limit 10 --json
```

`--scope` is required. `--kind` and `--limit` are optional. The
`--json` flag gives you structured output you can reason about.

### When you need a specific record

If you know a record ID (from a previous context result or a
list), you can fetch it directly:

```sh
sidetrail get <id>
sidetrail get <id> --human   # human-readable
```

### When you want to see recent decisions

```sh
sidetrail list --limit 10
sidetrail list --kind decision --limit 10
```

## When to suggest a record (not write one)

If you identify a decision, constraint, or health signal that is
not yet recorded — **do not write it yourself**. Instead:

1. Create a JSON file with the correct schema.
2. Tell the human: "I'd like to suggest a record. Please review
   the file at `.sidetrail/_seed/<suggested-name>.json`."
3. Wait. The human will run `sidetrail add <file>` or
   `sidetrail promote` when they are ready.

`sidetrail validate <file>` is a useful pre-flight check before
the human commits. `sidetrail add <file>` is the human's
approval gate.

## What you must not do

- **Do not call `sidetrail add` directly.** That is the human's
  job. You suggest; they decide.
- **Do not run `sidetrail init`.** That overwrites existing
  records.
- **Do not modify `.sidetrail/` files directly.** That bypasses
  validation and the audit trail.
- **Do not run any command that writes to disk.** The CLI is
  read-only from your perspective.

## Conventions

- Default output is human-readable. Use `--json` when you need
  structured data.
- `sidetrail get <id>` is fast. Prefetch a record if you suspect
  you will need it.
- All data is local. No network calls. No LLM calls. No
  background process.

## Quick reference

| You want to... | Run |
| --- | --- |
| Read context for a file | `sidetrail context --file <path>` |
| Find decisions by topic | `sidetrail ask --scope <topic> --kind decision` |
| Get a specific record | `sidetrail get <id>` |
| List recent records | `sidetrail list --limit 10` |
| Validate a draft file | `sidetrail validate <file>` |
| Check if binary is installed | `sidetrail --version` |

## If sidetrail is not installed

```sh
curl -fsSL https://raw.githubusercontent.com/SincereMa/sidetrail/main/scripts/install.sh | sh
```

On Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/SincereMa/sidetrail/main/scripts/install.ps1 -useb | iex
```

If the install fails or the human asks you not to install software,
tell them and stop. Do not attempt further configuration.
