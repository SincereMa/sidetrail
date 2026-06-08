# Per-host-agent adapter specifications

This directory holds the adapter specifications for host agents
that SideTrail integrates with. Each adapter is a self-contained
document that describes how the host agent should use SideTrail.

See [AGENTS.md](../../AGENTS.md) for the universal constraints
that apply to all adapters.

## Available adapters

| Host agent | Adapter | Adapter type |
| --- | --- | --- |
| [OpenCode](opencode.md) | Skill (on-demand) + `AGENTS.md` linkage | CLI-based, read-only |

Future adapters will follow the same pattern: a document in
`docs/agents/<host>.md` and, where applicable, a skill or
configuration file under `docs/agents/<host>/`.

## What each adapter includes

Each adapter document describes:

- The integration point (skill, MCP server, AGENTS.md linkage, etc.)
- How the host agent discovers and loads the SideTrail context
- The commands the host agent is allowed to run (`sidetrail ask`,
  `sidetrail get`, `sidetrail list`, `sidetrail context`)
- What the host agent must never do (write records, inject context
  automatically, etc.)

The adapter does not modify the host agent. It runs alongside it.
