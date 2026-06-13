# Universal Agent Instructions Design

## [S1] Problem

Agents (Claude, OpenCode, Codex, etc.) work on SideTrail projects but have
no guidance on when/how to use SideTrail's record system. The current state:
CLI commands exist, the store works, but agents don't know when to add
records, when to look up records, or how to manage token budgets when
injecting records into context.

The instruction system must:
- Work across different agent platforms (Claude, OpenCode, Codex, others)
- Be a single source of truth (no duplicated content)
- Be low-friction (agents read it automatically or with minimal setup)
- Respect token budgets when injecting records into context

## [S2] Architecture: CLAUDE.md + Universal SKILL.md

Two files, one canonical source:

**`CLAUDE.md`** — Project-level instructions that any agent reads automatically.
Contains behavioral rules for record management. This is the canonical source.

**`docs/agents/skill/SKILL.md`** — Agent-specific skill file for platforms
that support skills (OpenCode, etc.). Contains the same behavioral rules
but formatted for the skill system. Derived from CLAUDE.md content.

The key insight: **behavioral rules are identical across agents**. Only the
file format differs. CLAUDE.md is the source; SKILL.md is a reformatted copy.

## [S3] Behavioral Rules

The instructions define three behavioral phases:

### Phase 1: Before editing (read-side)

Before modifying any file, the agent should:
1. Run `sidetrail context --file <path> --limit 10 --max-tokens 2000`
2. Read the returned records
3. If records exist, incorporate their guidance into the edit
4. If no records, proceed normally

This ensures the agent has project context (decisions, constraints) before
making changes.

### Phase 2: During editing (write-side)

After completing a significant edit, consider recording:
- A decision was made → `sidetrail add decision.json`
- A constraint was discovered → `sidetrail add constraint.json`
- An experiment was run → `sidetrail add experiment.json`
- An incident occurred → `sidetrail add incident.json`

When to record:
- You chose approach A over B for reason C → record the decision
- You learned "don't do X because Y" → record as constraint
- You tried something that worked/failed → record as experiment
- You hit a production issue → record as incident

When NOT to record:
- Routine edits with no meaningful choice
- Temporary workarounds that won't persist
- Information already in code comments
- Trivial changes (typos, formatting)

### Phase 3: Periodic health check

When starting a new session or when idle:
1. Run `sidetrail health --stale-days 60`
2. If stale records exist, note them for potential refresh
3. If superseded chains are long, consider cleanup

## [S4] Token Budget Guidelines

The instructions specify token budgets for different contexts:

| Command | Default budget | Purpose |
|---------|---------------|---------|
| `context` | 2000 tokens | Pre-edit record lookup |
| `health` | 500 tokens | Summary only |
| `suggest` | 3000 tokens | Task-relevant records (future) |

When the total context exceeds the budget:
1. Prioritize constraints over decisions (constraints are binding)
2. Prioritize recent records over old ones
3. Prioritize records matching the current file's scope
4. Drop experiments/incidents unless directly relevant

## [S5] Record Format Templates

The instructions include JSON templates for each record kind:

### Decision template

```json
{
  "kind": "decision",
  "scope": "src/billing/",
  "subject": "Use Stripe for payment processing",
  "reason": "Stripe has better API docs and lower fees than alternatives",
  "source_type": "human",
  "author": "user",
  "status": "active",
  "decided_at": "2026-01-15T00:00:00Z"
}
```

### Constraint template

```json
{
  "kind": "constraint",
  "scope": "src/auth/",
  "subject": "Never log raw passwords",
  "reason": "Security policy: passwords must never appear in logs",
  "severity": "hard",
  "source_type": "human",
  "author": "user",
  "status": "active"
}
```

### Experiment template

```json
{
  "kind": "experiment",
  "scope": "src/cache/",
  "subject": "Test Redis vs Memcached for session cache",
  "reason": "Need to evaluate caching options for session data",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "in_progress",
  "started_at": "2026-01-15T00:00:00Z"
}
```

### Incident template

```json
{
  "kind": "incident",
  "scope": "src/payments/",
  "subject": "Payment webhook timeout in production",
  "reason": "Webhook handler exceeded 30s timeout under load",
  "source_type": "human",
  "author": "user",
  "status": "resolved",
  "occurred_at": "2026-01-15T00:00:00Z",
  "resolved_at": "2026-01-16T00:00:00Z"
}
```

## [S6] Forbidden Actions

The instructions explicitly prohibit:
- Auto-injecting records into context without running sidetrail commands
- Modifying `.sidetrail/` directory directly (always use CLI)
- Creating records for trivial changes (typos, formatting)
- Recording information that belongs in code comments
- Exceeding token budgets without explicit user approval
- Running `sidetrail add` without validating the JSON first

## [S7] File Layout

```
CLAUDE.md                          # Project-level instructions (canonical)
docs/agents/
  skill/
    SKILL.md                       # OpenCode skill file (derived)
  README.md                        # Adapter documentation
  opencode.md                      # OpenCode-specific notes
```

## [S8] Migration Path

Existing projects without SideTrail:
1. `sidetrail init` creates `.sidetrail/`
2. Agent reads CLAUDE.md and starts following instructions
3. No manual record creation needed — agent creates records as it works

Existing projects with SideTrail but no instructions:
1. Add CLAUDE.md with the behavioral rules
2. Agent picks up instructions on next session

## [S9] Future Extensions

This design deliberately excludes:
- MCP server integration (deferred to v1)
- Intelligent suggest command with configurable retrieval strategies
- Token-aware output truncation in CLI commands

These are natural next steps but are out of scope for the initial
instruction system. The instruction text can reference future commands
(e.g., `sidetrail suggest`) without implementing them now.
