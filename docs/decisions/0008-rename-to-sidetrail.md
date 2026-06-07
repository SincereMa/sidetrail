# 0008 — Rename project to SideTrail

## Status

Accepted. Implemented in the rename change that lands alongside this ADR.

## Context

The project shipped as **Cortex SideMark** through releases 0.0.x. The
two-word name and the "Cortex" prefix caused three concrete problems:

1. **Search collision.** "Cortex" is a heavily-used word in the LLM
   ecosystem (Anthropic's Cortex, several open-source projects, and
   product categories in general). Queries for the project, the binary,
   or its CLI commands returned noise before signal.
2. **The "SideMark" half was descriptive, not a brand.** It described
   the shape of the data (a sidecar, a mark on the side) but read as
   awkward when said aloud and made the binary's name ambiguous
   (`cortex-sidemark` vs. `cortex`).
3. **The on-disk store directory `.cortex/` was a forced choice driven
   by the binary name.** Renaming the project is the natural moment to
   also rename the directory the user sees in their repo, so the
   name on `PATH`, the directory in the working tree, and the brand in
   docs all line up.

The project is still in the v0 window. There is no published v1.0
release and no external consumers depending on the old name beyond a
small number of installations the maintainer is aware of. A coordinated
rename is cheap now and expensive later.

## Decision

The project is renamed **SideTrail** across every surface at once:

| Surface | Old | New |
| --- | --- | --- |
| Product name | Cortex SideMark | SideTrail |
| Repo slug | `cortex-sidemark` | `sidetrail` |
| Go module | `github.com/SincereMa/cortex-sidemark` | `github.com/SincereMa/sidetrail` |
| Go CLI package | `cmd/cortex` (`package cortex`) | `cmd/sidetrail` (`package sidetrail`) |
| Binary / CLI | `cortex` | `sidetrail` |
| On-disk store dir | `.cortex/` | `.sidetrail/` |
| OpenCode skill | `cortex` (at `.opencode/skills/cortex/SKILL.md`) | `sidetrail` (at `.opencode/skills/sidetrail/SKILL.md`) |
| Install-script env vars | `CORTEX_VERSION`, `CORTEX_INSTALL_DIR`, `CORTEX_REPO` | `SIDETRAIL_VERSION`, `SIDETRAIL_INSTALL_DIR`, `SIDETRAIL_REPO` |
| JSON Schema `$id` | `https://cortex-sidemark/schemas/record.json` | `https://sidetrail.dev/schemas/record.json` |
| LICENSE copyright | `Cortex SideMark Authors` | `SideTrail Authors` |
| GoReleaser `project_name` | `cortex` | `sidetrail` |
| GoReleaser `release.github.name` | `cortex-sidemark` | `sidetrail` |

The CLI help text, every ADR, every adapter guide, the OpenCode skill
file, and the install scripts are updated in the same change so the
repo never ships a half-renamed state.

## Migration

The project has shipped no releases, no tags, and no published
artifacts under the old name. The only data that could exist on
a user's machine is a local `.cortex/` store directory they
created themselves while running an unreleased build. There is
no install path, no skill installer, and no signed release to
migrate from.

Therefore **no backward-compatibility shim is shipped**. The
on-disk layout is `.sidetrail/`, full stop. The old name is not
recognized as a fallback, and a directory called `.cortex/` in
a user's project is not picked up by `findStoreRoot`. Anyone
who somehow has a `.cortex/` directory from an unreleased build
should delete it and start fresh; the records inside have no
schema-versioned history to preserve.

## Consequences

Positive:

- One name, one binary, one directory, one repo slug, one install
  command. The grep footprint is now narrow.
- The OpenCode skill name and the binary name match, so a reader of
  the docs sees one consistent word.
- The JSON Schema `$id` is now a plausible future website, leaving
  room to publish docs there.

Negative / accepted:

- This is a hard cut for anyone who happened to run an
  unreleased build against `.cortex/`. Their local directory
  becomes invisible to the new binary; they delete it and
  start over. With no releases shipped, the population of
  such users is at most the project's own contributors, who
  know to expect this change.
- The repo loses its history association with the old name in
  GitHub search and in `git log` of consumers' clones. The commits
  themselves still contain the old name and remain searchable.

## Notes

- This ADR follows the rules in
  [AGENTS.md](../../AGENTS.md): a name change that touches the
  Go module path, the CLI package, the host-agent adapter, and
  the on-disk data layout must be recorded before merge.
- The `$id` change in `internal/schema/record.schema.json` is a
  one-line tweak; it is not a wire-format break because the schema
  is not fetched by URL anywhere in the codebase. Consumers
  embedding the schema inline are unaffected.
