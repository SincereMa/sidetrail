# OpenCode adapter for SideTrail

## 1. Philosophy

The hard constraints from
[AGENTS.md](../../AGENTS.md) apply in full:

- **Non-intrusive.** Observe and record; never edit the host
  agent. Redesign any feature that requires touching the host.
- **Sidecar, not replacement.** SideTrail is a sidecar, not a
  competing agent. It does not replace OpenCode's own tooling and
  does not replace itself as a tool.
- **Lightweight.** No heavy runtime, no bundled LLM calls the user
  did not request, no surprise transitive dependencies.
- **Standard install.** The `sidetrail` binary should install via
  any standard package manager command.

OpenCode's mechanisms that make this adapter possible:

- **Skill files** — on-demand instructions loaded via the `skill`
  tool. Per-project (`.opencode/skills/sidetrail/SKILL.md`) and
  global (`~/.config/opencode/skills/sidetrail/SKILL.md`).
- **`AGENTS.md` linkage** — always-on instructions that OpenCode
  loads at session start. The `AGENTS.md` file at the project root
  tells OpenCode to load the sidetrail skill.
- **MCP (planned)** — a future adapter surface. Powerful but
  deferred to v1 (section 5).

Skills are the high-value, low-friction integration point for v0.
They load on-demand when triggered by the user and do not require
any background process. A `sidetrail-suggest` subcommand will
handle this in v1; the skill + AGENTS.md combination is enough for v0.

## 2. File hierarchy

An OpenCode-aware project uses three files. They are
independent; you do not need all three.

### `AGENTS.md` (project root)

Add a single section that tells OpenCode to load the sidetrail
skill:

```markdown
## SideTrail

This project uses [SideTrail](https://github.com/SincereMa/sidetrail)
for long-lived context. Before editing or planning changes, load the
`sidetrail` skill and follow its instructions.
```

This makes the sidetrail skill available whenever OpenCode reads
`AGENTS.md`. The `AGENTS.md` file is already loaded by OpenCode at
session start, so this linkage is free.

### `.opencode/skills/sidetrail/SKILL.md`

Any file at `docs/agents/opencode/skill/SKILL.md` fits this
purpose. The file at `docs/agents/opencode/skill/SKILL.md` is the
canonical initial version; the human can customize it.

Updating the canonical file is the human's responsibility —
human-only, not agent-initiated.

The skill is validated against the global skill at
`~/.config/opencode/skills/sidetrail/SKILL.md` for consistency when
the project-local file is loaded. The project-local file takes
precedence when both exist.

### SideTrail binary

```sh
curl -fsSL https://raw.githubusercontent.com/SincereMa/sidetrail/main/scripts/install.sh | sh
```

The user must ensure `sidetrail` is on `PATH` and the `sidetrail`
store is initialized (or that they will initialize it).

## 3. Workflow

### Reading relevant files

When the agent is about to edit a file (e.g.,
`/project/billing/init.go`), the skill instructs it to run:

```sh
sidetrail context --file /project/billing/init.go
```

Output (human-readable):

```
ID          KIND        SCOPE              SUBJECT
01J5H8K...  decision    billing/init       Refactor init to use new config
01J5H8L...  constraint  billing/init       Do not remove AMQP config fallback
```

The agent now knows: **do not remove the config fallback** and
**this area was refactored three times**. Both are load-bearing
constraints that the agent would otherwise miss.

### Querying by area

When the agent is asked "what do we know about billing?", it runs:

```sh
sidetrail ask --scope billing --kind decision --json
```

And gets structured output it can reason about.

### Diagnosing dead-whitespace files

While reading `/audit/handler.go`, the agent notices a call to
`Rmq()` but does not know what it does; the commit is from 2024.
The agent asks:

```sh
sidetrail get 01J...  # if it knows the id
```

or:

```sh
sidetrail ask --scope audit/handler --kind constraint --json
```

The agent gets a decision record explaining the `Rmq()` call
and the reason it exists.

## 4. Boundaries

When a human asks the agent to do something, the agent must:

- ✅ Read relevant files via `sidetrail context --file ...`.
- ✅ Query by kind and scope via `sidetrail ask --scope ...`.
- ✅ List recent decisions via `sidetrail list --limit 10`.
- ✅ Validate a draft record via `sidetrail validate <file>`.
- ✅ Check if the binary is installed via `sidetrail --version`.

The agent must never:

- ❌ Inject context automatically. The skill loads on-demand;
  the human initiates.
- ❌ Reload skill files mid-run. Skills are loaded once per
  session.
- ❌ Call `sidetrail add` directly. That is the human's job.
- ❌ Run anything against hashed skill files. The human writes
  the skill file.

These boundaries are enforced by convention, not by code. The
agent respects them because the skill file and `AGENTS.md` say so.

## 5. Future: MCP adapter

A `sidetrail-serve` MCP server would let OpenCode call SideTrail
as a tool. The MCP adapter would:

- Expose `sidetrail_get`, `sidetrail_list`, `sidetrail_ask`,
  `sidetrail_context` as MCP tools.
- Default to human-readable output; `--json` when the tool
  requests it.
- All operations are read-only; no `sidetrail add` via MCP.

This is deferred to v1. The skill + AGENTS.md combination is
enough for v0. The MCP adapter would be a thin wrapper around the
existing CLI, so the implementation cost is low.

### Caveats (not yet implemented)

OpenCode's own configuration layer is not modified by SideTrail.
A `sidetrail serve` is an integration point, not a replacement.

### Multi-agent / sub-agent flow

OpenCode's sub-agent model lets you spawn dedicated skill runners
(agents that only load the sidetrail skill and follow its
instructions). The `skills.list` block in `opencode.json` can
reference a `sidetrail` agent, and the build agent can reference
the sidetrail skill by name.

## 6. References

- `AGENTS.md` — the project's own non-negotiable constraints;
  the skill respects these by construction.
- `ADR-0001` §3 — has an inf field on the read path (CLI first,
  MCP later).
- `ADR-0003` §6 — distribution: single binary, cross-platform.
- OpenCode docs:
  [Skill](https://opencode.ai/docs/skill/)
  [MCP](https://opencode.ai/docs/mcp/)
  [Rules/AGENTS.md](https://opencode.ai/docs/rules/agents).
