# Agent Adapters

SideTrail serves AI coding agents through universal instruction files.
The behavioral rules are identical across agents — only the file format differs.

## Files

| File | Purpose | Audience |
|------|---------|----------|
| `CLAUDE.md` | Project-level instructions | Any agent that reads project files |
| `skill/SKILL.md` | OpenCode skill file | OpenCode agents |
| `opencode.md` | OpenCode-specific notes | OpenCode users |

## How It Works

1. Agent reads `CLAUDE.md` (or loads `SKILL.md` as a skill)
2. Instructions tell the agent when/how to use SideTrail commands
3. Agent calls `sidetrail` CLI commands via shell
4. Records are stored in `.sidetrail/` as JSON files

## Adding Support for a New Agent

To add SideTrail support for a new agent:

1. Check if the agent reads `CLAUDE.md` — if yes, it already works
2. If the agent uses a different instruction format, create a derived file
   in `docs/agents/` that reformats the CLAUDE.md content
3. Update this README to list the new adapter

## Migration

For projects without SideTrail:
1. Run `sidetrail init`
2. Agent reads CLAUDE.md and starts following instructions

For projects with SideTrail but no instructions:
1. Copy CLAUDE.md to the project root
2. Agent picks up instructions on next session
