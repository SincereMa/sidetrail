---
name: cortex
description: Load the project's cortex-sidemark context. Use before reading, editing, or planning changes to a file in this project to surface relevant decisions, constraints, and rules recorded in `.cortex/`.
license: MIT
compatibility: opencode
metadata:
  audience: host-agents
  surface: read
---

## What this skill is

This skill teaches you to use **Cortex SideMark**, a sidecar
binary (`cortex`) that records long-lived project context. Context
lives in a project-local `.cortex/` directory; records cover
decisions, constraints, and other tribal knowledge the team has
filed about this project. Cortex is a sidecar, not a replacement
for you, and it never edits your system prompt, your tools, or
your behaviour. It only answers questions when you ask.

The first time you see this skill in the `skill` tool's available
list, load it once. After that, treat it as background knowledge;
the same content will be available to you on every task in this
project.

## When to use the cortex CLI

The shell `cortex` binary is your only contact with the sidecar.
The cases below are mandatory; everything else is optional.

### Before reading or editing a file

Run `cortex context --file <path>` before the first time you touch
a file. The output is a JSON array of records whose `scope` covers
that file or one of its ancestors. If the array is empty, the
project has no filed context for that file — proceed normally. If
the array has entries, **read them**. Decisions and constraints in
the output are load-bearing; "why" is in the `reason` field, not
the `body`. The point of this skill is to make sure you do not
miss them.

```sh
cortex context --file path/to/file.go
cortex context --file path/to/file.go --radius 1   # one parent
cortex context --file path/to/file.go --limit 20  # cap
```

### Before answering "what should I do here" questions

When the user asks a question that depends on tribal knowledge
rather than code (`is X safe to change?`, `where do we keep Y?`,
`why is the code shaped this way?`), query first:

```sh
cortex ask --scope src/legacy      # records whose scope matches
cortex ask --scope src/legacy --kind constraint
cortex ask --scope src/legacy --tag compliance --json
```

`--scope` is required. It is a path-like pattern; records are
returned if the file or area described matches.

### When you discover a decision-worthy fact

If during a task you notice something that should become a record
(a hidden constraint, a non-obvious reason for a design, a tribal
rule), propose it — **do not write it**. The write side is
human-first; you only suggest:

1. Compose a record JSON file matching the schema.
2. Tell the user: "I'd like to propose a record. Place this file
   under `.cortex/_proposed/` after review."
3. Wait. The user moves it to a canonical kind dir
   (`.cortex/decisions/`, `.cortex/constraints/`, …) to promote
   it.

`cortex validate <file>` is a useful pre-check before you hand a
draft to the user. `cortex add <file>` is the human's tool, not
yours.

### At session start, refresh your context

If the user says something like "let's get to work" or "where were
we", a quick `cortex list --limit 10` can give you a working
summary. This is optional and only useful for long-lived sessions.

## Output conventions

- Default output is human-readable. Use `--json` when you intend
  to parse the result. The two forms are stable; the table form
  is for humans, the JSON form is for you.
- `cortex get <id>` returns one record. Prefix matching is
  supported; a unique prefix is enough.
- All commands exit non-zero on error. Read stderr before retrying.

## What this skill explicitly does not do

- It does not modify your system prompt. Your behaviour is still
  governed by opencode and by this project's own `AGENTS.md`.
- It does not run cortex in the background. Each call is a fresh
  process.
- It does not require an MCP server, a daemon, or a port. The
  sidecar is just a binary.
- It does not push context into you. You pull, on demand, with
  the bash tool.

## Quick reference

| You want to | Run |
| --- | --- |
| Read context for a file | `cortex context --file <path>` |
| Find records by area | `cortex ask --scope <pattern>` |
| Get one record by id | `cortex get <id>` |
| List recent records | `cortex list --limit 10` |
| Validate a draft record | `cortex validate <file>` |
| Check the binary is installed | `cortex --version` |

## If cortex is not installed

```sh
curl -fsSL https://raw.githubusercontent.com/SincereMa/cortex-sidemark/main/scripts/install.sh | sh
```

On Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/SincereMa/cortex-sidemark/main/scripts/install.ps1 -useb | iex
```

If the install cannot reach the network, tell the user. Do not
pretend the sidecar is working.
