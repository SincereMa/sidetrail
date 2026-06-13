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

The `sidetrail` binary has five commands:

| Command | Purpose |
|---------|---------|
| `context` | Read records relevant to a file |
| `add` | Validate and store a record |
| `update` | Update an existing record |
| `health` | Report project health signals |
| `init` | Create a `.sidetrail/` directory |

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
use `context` with a file path in that area:

```sh
sidetrail context --file billing/handler.go --json
sidetrail context --file auth/login.go --json
```

### When you want to record a decision or constraint

If you identify a decision, constraint, or health signal that is
not yet recorded, create a JSON file and use `sidetrail add`:

```sh
# Create the record file
cat > /tmp/record.json << 'EOF'
{
  "kind": "decision",
  "scope": "src/auth/login.go",
  "subject": "Use bcrypt for password hashing",
  "reason": "OWASP recommended, good compatibility",
  "source_type": "human",
  "author": "agent",
  "created_at": "2026-06-13T00:00:00Z",
  "last_verified_at": "2026-06-13T00:00:00Z",
  "status": "active",
  "decided_at": "2026-06-13T00:00:00Z"
}
EOF

# Add it to the store
sidetrail add /tmp/record.json
```

### When you need to update a record

```sh
# Create update file
cat > /tmp/update.json << 'EOF'
{"status": "superseded", "superseded_by": "new-record-id"}
EOF

# Update the record
sidetrail update <id> --file /tmp/update.json
```

## What you must not do

- **Do not run `sidetrail init`.** That creates the store directory;
  it is usually already there.
- **Do not modify `.sidetrail/` files directly.** That bypasses
  validation and the audit trail.
- **Do not create records without required fields.** Each kind has
  required fields: `decided_at` for decisions, `started_at` for
  experiments, `occurred_at` for incidents.

## Conventions

- Default output is human-readable. Use `--json` when you need
  structured data.
- All data is local. No network calls. No LLM calls. No
  background process.

## Quick reference

| You want to... | Run |
| --- | --- |
| Read context for a file | `sidetrail context --file <path>` |
| Add a new record | `sidetrail add <json-file>` |
| Update an existing record | `sidetrail update <id> --file <json>` |
| Check project health | `sidetrail health` |
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
