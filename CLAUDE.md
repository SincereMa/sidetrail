# CLAUDE.md

## Project: SideTrail

SideTrail is a sidecar CLI tool that gives AI coding agents long-lived
project memory — decisions, constraints, health signals — without modifying
the host agent. It stores records as JSON files in `.sidetrail/` under the
project root.

## Quick Reference

| Command | Purpose | Example |
|---------|---------|--------|
| `sidetrail init` | Initialize store | `sidetrail init` |
| `sidetrail context --file <path>` | Get records for a file | `sidetrail context --file src/auth/handler.go` |
| `sidetrail add <file.json>` | Create a record | `sidetrail add decision.json` |
| `sidetrail update <id> --file <file.json>` | Update a record | `sidetrail update 01HXYZ... --file update.json` |
| `sidetrail health` | Check store health | `sidetrail health --stale-days 60` |

## When to Use SideTrail

### Before editing a file

Run the context command to check for relevant records:

```bash
sidetrail context --file <path> --limit 10 --max-tokens 2000
```

If records exist, incorporate their guidance into your edit. Constraints
are binding — do not violate them. Decisions explain why the code is
the way it is — respect the reasoning.

### After a significant edit

Consider recording what you learned. Only record when there is a
meaningful choice, constraint, or lesson — not for routine edits.

**Record a decision** when you chose approach A over B for reason C:

```json
{
  "kind": "decision",
  "scope": "src/billing/",
  "subject": "Use Stripe for payment processing",
  "reason": "Stripe has better API docs and lower fees than alternatives",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "active",
  "decided_at": "2026-01-15T00:00:00Z"
}
```

**Record a constraint** when you learned "don't do X because Y":

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

**Record an experiment** when you tried something that worked or failed:

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

**Record an incident** when you hit a production issue:

```json
{
  "kind": "incident",
  "scope": "src/payments/",
  "subject": "Payment webhook timeout in production",
  "reason": "Webhook handler exceeded 30s timeout under load",
  "source_type": "agent-suggested",
  "author": "agent",
  "status": "resolved",
  "occurred_at": "2026-01-15T00:00:00Z",
  "resolved_at": "2026-01-16T00:00:00Z"
}
```

### When starting a new session

Check for stale records:

```bash
sidetrail health --stale-days 60
```

If stale records exist, note them for potential refresh during your work.

## Token Budgets

Respect these token limits when injecting records into context:

| Context | Budget | Strategy |
|---------|--------|----------|
| Pre-edit lookup | 2000 tokens | Use `--max-tokens 2000` |
| Health check | 500 tokens | Summary only |
| Task-relevant lookup | 3000 tokens | Use `--max-tokens 3000` |

When the total context exceeds the budget:
1. Prioritize constraints over decisions (constraints are binding)
2. Prioritize recent records over old ones
3. Prioritize records matching the current file's scope
4. Drop experiments/incidents unless directly relevant

## Forbidden Actions

- **Never** auto-inject records into context without running sidetrail commands
- **Never** modify `.sidetrail/` directory directly — always use the CLI
- **Never** create records for trivial changes (typos, formatting)
- **Never** record information that belongs in code comments
- **Never** exceed token budgets without explicit user approval
- **Never** run `sidetrail add` without validating the JSON first
