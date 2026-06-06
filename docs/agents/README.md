# Per-host-agent adapter specifications

This directory holds adapter specifications for the host agents
that Cortex SideMark can attach to. Each adapter is a localised
change: the sidecar binary is unchanged, the host agent is
unchanged, and the integration is one or more user-placed files
plus a documented workflow.

See [AGENTS.md](../../AGENTS.md) for the architectural constraints
any adapter must respect.

## Available adapters

| Host agent | Adapter | Adapter surface |
| --- | --- | --- |
| [OpenCode](opencode.md) | Skill (on-demand prompt) + `AGENTS.md` line | `cortex` CLI via the bash tool |

Future adapters will follow the same shape: a single integration
guide under `docs/agents/<host>.md` and, where the artifact is
more than prose, a copyable file under
`docs/agents/<host>/<artifact>/`.

## When a host agent is not listed

A host agent not listed here still works: the sidecar is a CLI
binary and any host agent that can run a shell command can call
`cortex ask`, `cortex get`, `cortex list`, and `cortex context`.
The adapter documents the host-specific glue (where to put the
prompt, which configuration keys to set, which caveats apply).
