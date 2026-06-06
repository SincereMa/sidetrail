# OpenCode integration

OpenCode is the first host agent with a documented Cortex SideMark
integration. The integration is intentionally minimal: it ships
as a single `SKILL.md` that the user drops into the project. The
sidecar itself is unchanged; the sidecar never knows which host
agent is calling it.

This document describes:

1. The integration philosophy and how it satisfies the
   non-intrusive principle.
2. The exact files the user creates (and where).
3. Worked examples of a session with and without context.
4. The boundaries: what opencode can and cannot do with cortex
   in v0.
5. Future adapters: MCP and custom tools, deferred to v1.

## 1. Philosophy

The hard constraints in `AGENTS.md` include:

- **Non-intrusive.** Observe and record; never edit the host
  agent. The host agent is opencode here; "editing" includes
  injecting into the system prompt, registering tools without
  the user's consent, or otherwise changing opencode's behaviour
  behind its back.
- **Sidecar, not replacement.** Cortex is a complement, not a
  competitor. The integration must not position itself as
  "opencode plus cortex" but as "opencode that can call cortex".
- **Lightweight.** No daemon, no port, no long-running process,
  no surprise transitive dependencies in the opencode install.
- **Standard install.** The user installs the same `cortex`
  binary they would install for any other host agent.

OpenCode has three extension points that satisfy all of the above:

- **Skills** — on-demand prompt fragments the agent loads via
  the `skill` tool. Per-project (`.opencode/skills/<name>/SKILL.md`)
  or global (`~/.config/opencode/skills/<name>/SKILL.md`).
- **`AGENTS.md` rules** — always-on instructions the agent reads
  on every task in a project. The repo's own `AGENTS.md` is the
  right place for one short line pointing at the skill.
- **MCP servers** — external tools over the Model Context
  Protocol. Powerful but not in v0 (see §5).

Skills are the right primary surface for v0. They are loaded
on-demand, they do not pollute the system prompt when unused, and
they are opt-in by file presence. A `cortex mcp-serve` subcommand
will land in v1; this document treats it as future work.

## 2. Files the user creates

A one-time, per-project setup. The user does not need to edit
opencode's global config.

### `AGENTS.md` (project root)

Add a single paragraph near the top:

```markdown
## Cortex SideMark

This project uses [Cortex SideMark](https://github.com/SincereMa/cortex-sidemark)
to record long-lived context. Before reading or editing a file in
this project, load the `cortex` skill and follow its instructions.
```

This makes the skill discoverable: the agent will see the skill in
the `skill` tool's available list whenever the project is opened,
and the `AGENTS.md` line tells it to load the skill proactively.

### `.opencode/skills/cortex/SKILL.md`

A copy of `docs/agents/opencode/skill/SKILL.md` from this repo.
The file at `docs/agents/opencode/skill/SKILL.md` is the canonical
content; the user copies it into the project. Updating the
canonical file in this repo is not a push to every consumer —
the user re-copies when they want to upgrade.

The skill is also valid as a global skill at
`~/.config/opencode/skills/cortex/SKILL.md` for users who want
the same behaviour across all projects without per-project
copying. The per-project copy takes precedence during opencode's
walk-up discovery.

### Cortex binary

```sh
curl -fsSL https://raw.githubusercontent.com/SincereMa/cortex-sidemark/main/scripts/install.sh | sh
```

The user runs this once. After that, the `cortex` binary is on
`PATH` and the `bash` tool can call it from any project.

## 3. Worked examples

### Reading context for a file

The user asks opencode to fix a bug in
`src/legacy/billing/invoice.go`. The agent loads the `cortex`
skill, then runs:

```sh
cortex context --file src/legacy/billing/invoice.go
```

Output (trimmed, table form):

```
ID                                    KIND        SCOPE                              SUBJECT
01J5H8K...constraint                  src/legacy/billing                    "Do not refactor this module"
01J5H8L...decision                     src/legacy/billing                    "Use the legacy AMQP wire format"
```

The agent now knows: **do not refactor the file**, and **the
existing wire format is a deliberate decision**. Both of these
would have been missed by code-only inspection. The agent
continues the bug fix, aware of the constraint.

The same call with `--json` is what the agent parses
programmatically:

```sh
cortex context --file src/legacy/billing/invoice.go --json
```

### Querying by area

The user asks "is there a reason we use library X?". The agent
runs:

```sh
cortex ask --scope "**" --tag library-choice --json
```

and gets the relevant decision records.

### Discovering a record-worthy fact

While reading `src/auth/session.go`, the agent notices a comment
saying "must call `Revoke()` before the response is written;
see incident #4231 for the 2024 outage". This is a constraint
the team has not recorded. The agent drafts a record file
matching the JSON Schema (the schema is at
`internal/schema/record.schema.json` in the cortex-sidemark
repo) and tells the user:

> I'd like to propose a constraint record. Place this at
> `.cortex/_proposed/01J...auth-revoke.json` after you review:
>
> ```json
> { "kind": "constraint", "scope": "src/auth/session.go",
>   "subject": "Revoke must run before the response is written",
>   "reason": "Order matters; incident #4231 (2024 outage)",
>   "source_type": "agent-suggested", ... }
> ```
>
> To accept: move it to `.cortex/constraints/`. To reject:
> delete it. `cortex validate <file>` will check the schema
> first.

The agent does not run `cortex add` itself. The write side is
human-first per `ADR-0001` §4.

## 4. Boundaries (v0)

What opencode can do today with the integration above:

- ✅ Read context for a file before editing.
- ✅ Query records by scope, kind, or tag.
- ✅ List recent records to refresh its working memory.
- ✅ Validate a record draft it is about to propose.
- ✅ Surface a missing install to the user.

What opencode cannot do today, by design:

- ❌ Inject context into its own system prompt automatically.
  The skill teaches the *agent* to query; the user still has to
  trust the model to follow the skill.
- ❌ Receive live push updates from the sidecar. The sidecar has
  no daemon and does not watch files; every read is a fresh
  process invocation.
- ❌ Call `cortex add` directly. The write side is human-first.
  Proposals go to `.cortex/_proposed/` and a human promotes.
- ❌ Run the integration in any project that does not have the
  skill file. Per-project opt-in is the point.

These are the same boundaries that apply to every other host
agent. They are properties of the sidecar, not of opencode.

## 5. Future adapters

### `cortex mcp-serve` (v1)

A `cortex mcp-serve` subcommand will expose the sidecar's read
surface as MCP tools: `cortex_get`, `cortex_list`, `cortex_ask`,
`cortex_context`. The user's `opencode.json` would then add:

```json
{
  "mcp": {
    "cortex": {
      "type": "local",
      "command": ["cortex", "mcp-serve"],
      "enabled": true
    }
  }
}
```

The MCP adapter gives the agent typed tool calls instead of
stringly-typed shell invocations. It is a strict superset of the
shell-based skill — every shell call has an MCP equivalent, plus
a few ergonomic wins (no `--json` parsing, no stderr handling).
It is not in v0 because it is more code and a v0 release is
useful without it. The skill + bash integration is the
deliberately minimal v0.

### Custom tool (alternative to MCP)

OpenCode supports user-defined tools via TypeScript modules. A
cortex custom tool is equivalent to MCP for read-only commands
and adds nothing the MCP adapter would not give us. The MCP path
is the right one; the custom tool route is a fallback if the MCP
spec drifts.

### Multi-agent / subagent flow

OpenCode's subagents could be given different skill subsets
("the test agent gets `cortex` and the test-runner skill, the
build agent gets `cortex` and the build skill"). The
`permission.skill` block in `opencode.json` makes this
configurable per agent. The v0 integration does not need this;
it is noted for completeness.

## 6. References

- `AGENTS.md` — the project's own non-negotiable principles;
  the integration satisfies them by construction.
- `ADR-0001` §3 — the read interface decision; the opencode
  integration is one concrete realisation of "CLI first, MCP as
  opt-in adapter".
- `ADR-0003` §6 — distribution; the install scripts this
  document points at.
- OpenCode docs:
  [Skills](https://opencode.ai/docs/skills/),
  [MCP servers](https://opencode.ai/docs/mcp-servers/),
  [Rules / AGENTS.md](https://opencode.ai/docs/rules/).
