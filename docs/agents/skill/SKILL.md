---
name: sidetrail
description: Project memory for AI agents — decisions, constraints, health signals
usage: on-demand
---

# SideTrail Skill

SideTrail gives you long-lived project memory. Use it to record decisions,
constraints, and lessons — and to recall them in future sessions.

## Quick Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `sidetrail init` | Initialize store | `sidetrail init` |
| `sidetrail context --file <path>` | Get records for a file | `sidetrail context --file src/auth/handler.go` |
| `sidetrail add <file.json>` | Create a record | `sidetrail add decision.json` |
| `sidetrail update <id> --file <file.json>` | Update a record | `sidetrail update 01HXYZ... --file update.json` |
| `sidetrail health` | Check store health | `sidetrail health --stale-days 60` |

## Workflow: Before Editing

Before modifying any file, check for relevant records:

```bash
sidetrail context --file <path> --limit 10 --max-tokens 2000
```

If records exist, incorporate their guidance. Constraints are binding —
do not violate them. Decisions explain why the code is the way it is.

## Workflow: After Editing

After a significant edit, consider recording what you learned. Only record
meaningful choices, constraints, or lessons — not routine edits.

### Record a decision

```json
{
  "kind": "decision",
  "scope": "src/billing/",
  "subject": "Use Stripe for payment processing",
  "reason": "Stripe has better API docs and lower fees",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "active",
  "decided_at": "2026-01-15T00:00:00Z"
}
```

### Record a constraint

```json
{
  "kind": "constraint",
  "scope": "src/auth/",
  "subject": "Never log raw passwords",
  "reason": "Security policy: passwords must never appear in logs",
  "severity": "hard",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "active"
}
```

### Record an experiment

```json
{
  "kind": "experiment",
  "scope": "src/cache/",
  "subject": "Test Redis vs Memcached",
  "reason": "Evaluate caching options for session data",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "in_progress",
  "started_at": "2026-01-15T00:00:00Z"
}
```

### Record an incident

```json
{
  "kind": "incident",
  "scope": "src/payments/",
  "subject": "Payment webhook timeout",
  "reason": "Handler exceeded 30s timeout under load",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "resolved",
  "occurred_at": "2026-01-15T00:00:00Z",
  "resolved_at": "2026-01-16T00:00:00Z"
}
```

## Workflow: New Session

When starting a new session, check for stale records:

```bash
sidetrail health --stale-days 60
```

If stale records exist, note them for potential refresh.

## Token Budgets

| Context | Budget | Strategy |
|---------|--------|----------|
| Pre-edit lookup | 2000 tokens | `--max-tokens 2000` |
| Health check | 500 tokens | Summary only |
| Task-relevant lookup | 3000 tokens | `--max-tokens 3000` |

Prioritization when over budget:
1. Constraints over decisions
2. Recent over old
3. Matching scope over unrelated
4. Drop experiments/incidents unless relevant

## Rules

- Always use the CLI — never edit `.sidetrail/` directly
- Never create records for trivial changes
- Never exceed budgets without user approval
- Validate JSON before running `sidetrail add`
