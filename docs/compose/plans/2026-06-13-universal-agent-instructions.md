# Universal Agent Instructions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create universal CLAUDE.md and SKILL.md files that instruct any AI agent on when/how to use SideTrail's record system.

**Architecture:** Two files with identical behavioral rules in different formats. CLAUDE.md is the canonical source for project-level instructions. SKILL.md is a reformatted copy for OpenCode's skill system. No code changes — only documentation.

**Tech Stack:** Markdown, JSON templates

---

## File Structure

```
CLAUDE.md                          # Create: project-level instructions
docs/agents/skill/
  SKILL.md                         # Create: OpenCode skill file
docs/agents/README.md              # Modify: update adapter documentation
```

---

### Task 1: Create CLAUDE.md

**Covers:** [S2], [S3], [S4], [S5], [S6]

**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: Create CLAUDE.md with project overview and behavioral rules**

```markdown
# CLAUDE.md

## Project: SideTrail

SideTrail is a sidecar CLI tool that gives AI coding agents long-lived
project memory — decisions, constraints, health signals — without modifying
the host agent. It stores records as JSON files in `.sidetrail/` under the
project root.

## Quick Reference

| Command | Purpose | Example |
|---------|---------|---------|
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
```

- [ ] **Step 2: Verify CLAUDE.md renders correctly**

Read the file back and verify all sections are present and properly formatted.

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add CLAUDE.md with universal agent instructions"
```

---

### Task 2: Create OpenCode SKILL.md

**Covers:** [S2], [S3], [S4], [S5], [S6]

**Files:**
- Create: `docs/agents/skill/SKILL.md`

- [ ] **Step 1: Create SKILL.md with OpenCode-compatible format**

```markdown
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
```

- [ ] **Step 2: Verify SKILL.md renders correctly**

Read the file back and verify all sections are present and properly formatted.

- [ ] **Step 3: Commit**

```bash
git add docs/agents/skill/SKILL.md
git commit -m "docs: add OpenCode SKILL.md with universal agent instructions"
```

---

### Task 3: Update Adapter Documentation

**Covers:** [S7], [S8]

**Files:**
- Modify: `docs/agents/README.md`

- [ ] **Step 1: Update README.md to reflect new instruction files**

Read the current `docs/agents/README.md` and update it to reference the new CLAUDE.md and SKILL.md files. Add a section explaining the universal instruction approach.

```markdown
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
```

- [ ] **Step 2: Commit**

```bash
git add docs/agents/README.md
git commit -m "docs: update agent adapter documentation"
```

---

### Task 4: Verify All Files

**Covers:** [S7], [S8]

**Files:**
- Verify: `CLAUDE.md`
- Verify: `docs/agents/skill/SKILL.md`
- Verify: `docs/agents/README.md`

- [ ] **Step 1: Read all three files and verify consistency**

Check that:
- CLAUDE.md and SKILL.md have identical behavioral rules
- README.md correctly references both files
- No contradictions between files
- All JSON templates are valid

- [ ] **Step 2: Run git status to confirm clean state**

```bash
git status
```

Expected: No uncommitted changes.

- [ ] **Step 3: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "docs: fix consistency issues in agent instructions"
```

(Only if fixes were needed — skip if everything was clean.)
