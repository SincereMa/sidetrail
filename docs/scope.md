# Scope

The problems Cortex SideMark exists to address. This is the working
document that drives development; subsequent ADRs and adapter designs
will be derived from it.

The user-facing version of these problems, with concrete examples, is
in [README.md](../README.md). The non-negotiable principles that any
solution must respect are in [AGENTS.md](../AGENTS.md).

These problems sit inside the product surface defined in
[AGENTS.md](../AGENTS.md): decisions, boundaries/constraints, and
health data. Each is mapped to the surface categories it primarily
exercises, but the mapping is illustrative, not exclusive.

Nothing in this document is a feature spec. Solutions are constrained
by the principles in AGENTS.md and will be designed against them.

## Cross-cutting dimensions

Every problem below is shaped by a small set of dimensions that apply
across all of them. The sidecar must handle these even when a specific
problem is the immediate focus.

1. **Information lives outside the code.** The codebase is the
   artifact; the sidecar is the context. The problems below are all
   about non-code knowledge that an agent needs at edit time or
   planning time. A sidecar that lives next to the code is not
   optional — it is the only place this knowledge can be held.
2. **"Why" is load-bearing, not "what".** Knowing a decision was made
   is less useful than knowing why; knowing a dependency exists is
   less useful than knowing what its contract implications are;
   knowing a module is risky is less useful than knowing why. A
   sidecar that records facts without reasoning has roughly half the
   value of one that records both.
3. **Information decays.** A history from 2018 is less relevant than
   a history from last week. A health snapshot from a year ago can be
   misleading. The sidecar must record a freshness signal (when was
   this last verified?) and, where possible, a direction signal (is
   this getting better or worse?). Stale memory presented with the
   same weight as fresh memory is worse than no memory.
4. **Granularity must match the project's natural unit.** Project-
   level signals are too coarse to act on. The sidecar must localize
   to whatever unit the project itself uses — subproject, service,
   module, package — and must work even when those boundaries are
   fuzzy or only implicit.
5. **The problem is read-dominant.** The pain in every case is that
   the agent cannot discover the relevant information. The write
   side — how information gets into the sidecar, by whom, with what
   verification — is a separate, harder problem and is not described
   here, but the read-side designs must assume imperfect or
   incomplete writes. Where a problem depends on a write model (e.g.
   tribal constraints), that dependence is called out.
6. **The agent is one of several actors.** Humans, other agents, and
   automation all work on the same project. The sidecar's memory
   should be a shared surface, not an agent-private one, and must
   remain useful to actors with very different levels of context (a
   fresh teammate, a long-running agent, a CI script).

## 1. Reasoning trails for project evolution are lost between sessions

The reasoning trail that produced the project's current state is lost
between sessions. The agent cannot recover what was tried, what was
rejected and why, or what the current shape is a consequence of.

A useful trail must distinguish:

- Kinds of items: **decisions** (we chose X for reasons A, B, C),
  **experiments** (we tried X; the result was…), **incidents** (this
  code is shaped by a production failure on date D).
- Status of items: **active** (still binding), **superseded**
  (replaced), **open** (under discussion).

The read side of this problem is sharp: a new session opens on a
result the agent did not witness. The write side is the harder
unsolved problem: who records the trail, when, and with what
reasoning? Any design that touches this problem must acknowledge this
gap explicitly.

Primary product surface: **decisions**, with **boundaries** as a
secondary use (rejected approaches are also constraints).

## 2. The blast radius of an edit is invisible across services, modules, and subprojects

When a host agent edits a project, the edit's blast radius is
invisible. The agent can see the file it is changing, but not the
services, modules, subprojects, or shared contracts that depend on
that file — statically, at runtime, or operationally.

The agent's need is bidirectional:

- **Outbound** — downstream consumers of the change.
- **Inbound** — recipients of the changed interface.

Dependency kinds that must be distinguished:

- **Static** — imports, types, schemas, codegen inputs.
- **Runtime** — HTTP, RPC, message queues, shared databases.
- **Operational** — deploy order, capacity, env config, secrets.
- **Social** — maintainers, reviewers, on-call, domain experts.

The most consequential sub-case is **contract changes** (public APIs,
event schemas, shared types). The sidecar should aim for guidance,
not just facts — is this breaking, additive, or coordinated?

The pain covers projects with fuzzy or implicit boundaries. The
sidecar must work when boundaries are conventions, not enforced
structures.

Primary product surface: **boundaries/constraints**, with
**decisions** as a secondary use.

## 3. Project difficulty is uneven and shifts, but the agent has no read on it

The host agent treats all parts of a project as equivalent, but a
project's difficulty is uneven and shifts over time. The agent has
no read on this at the right granularity (subproject, service,
module) and so cannot prioritize, warn, slow down, or escalate
review where it matters.

Four distinct faces of "difficulty", each with different signals:

- **Hard to change safely** — churn, coverage, defect rate, revert
  frequency.
- **Hard to understand** — complexity, doc freshness, owner
  concentration, time-since-meaningful-edit.
- **Hard to operate** — incident frequency, test flakiness, deploy
  frequency, MTTR.
- **Hard to keep up with** — commit rate, PR age, dependency
  freshness, in-flight migrations.

Signals should be:

- **Trend-aware**, not point-in-time.
- **Causally attributed** where possible (complexity vs. test gap
  vs. unclear requirements, not a single risk score).
- **Actionable** — paired with a behavioral implication, not emitted
  as trivia.
- **Source-transparent** — git, CI, issue tracker, monitoring,
  manual annotation.

Primary product surface: **health data**.

## 4. Documented architecture drifts from actual architecture

A project's documented architecture is a snapshot of an earlier
moment; the actual architecture has drifted. The agent working from
current files and any documentation it can find can easily
construct a mental model inconsistent with reality (a dissolved
boundary treated as load-bearing; a deprecated path avoided; an
ownership map whose owners have moved on).

The sidecar should hold a **current** picture of the architecture,
not the picture the project published last. Where documented and
actual differ, the agent should see the actual one with the
documented one flagged as stale.

This overlaps with items 1 and 2 (the drift is lost reasoning; the
gap is invisible dependency) but is distinct because the remediation
is different: the problem is "consistency between two
representations of the same project", and the sidecar needs an
explicit notion of "current" to address it.

Primary product surface: **decisions** (the original architecture
was a decision) and **boundaries** (the actual boundaries are the
binding ones).

## 5. Tribal constraints are nowhere the agent can read them

The most binding constraints on a project are often not in the code,
not in the documentation, and not in the issue tracker. They live in
the heads of the team's senior members, in chat threads that have
scrolled off, in post-mortems linked once and forgotten, in runbooks
that no one updates, and in the unspoken "don't-do-that" accumulated
across years of incidents.

The host agent, left to itself, will violate these constraints. It
cannot infer them, and it does not know to ask. The team can usually
articulate them when asked.

This is the **boundaries** surface at its purest — explicit "do not
do" rules with a reason, exactly what the sidecar exists to
remember.

Capture will, by necessity, be a deliberate semi-manual activity.
The access pattern (agent pulls constraints before acting) should
feel effortless so that the capture cost is repaid many times over
on the read side. Per the cross-cutting dimension on read vs. write
asymmetry, any design in this area must assume the constraint set
is incomplete and degrade gracefully.

Primary product surface: **boundaries/constraints**, with
**decisions** as a secondary use (the constraint has a reason that
is itself a recorded decision).

## What this document is not

- It is not a feature spec. Solutions are constrained by
  [AGENTS.md](../AGENTS.md) and will be designed against those
  constraints.
- It is not a roadmap. The order in which these are addressed is
  decided separately.
- It is not an ADR. The architectural decisions that follow from
  these problems will be filed in [docs/decisions/](./decisions/).
- It is not a complete enumeration of auxiliary memory problems.
  Some real problems are deferred (ongoing migration state as a
  first-class concern, ownership gaps, stale conventions as a
  standalone signal, operational signals not encoded in code).
