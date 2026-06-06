# Contributing to Cortex SideMark

Thanks for your interest in contributing. Cortex SideMark is an early-stage
project, and clear, well-scoped contributions are very welcome.

## Before you start

Read [AGENTS.md](./AGENTS.md) first. The non-negotiable principles there
shape every decision in this repository, and any contribution that
contradicts them will be redirected rather than merged.

All participants are expected to follow the
[Code of Conduct](./CODE_OF_CONDUCT.md).

## How to file an issue

Use the issue templates under
[.github/ISSUE_TEMPLATE/](./.github/ISSUE_TEMPLATE/):

- **Bug report** for things that are broken or behave incorrectly.
- **Feature request** for new capabilities or notable changes.

If neither template fits, open a blank issue and explain the situation.

## How to submit a change

1. Fork the repository and create a short-lived branch off `main`:
   - `feat/<short-topic>` for features
   - `fix/<short-topic>` for bug fixes
   - `docs/<short-topic>` for documentation-only changes
   - `chore/<short-topic>` for tooling and housekeeping
2. Keep the change focused. One logical change per pull request.
3. Use the [pull request template](./.github/PULL_REQUEST_TEMPLATE.md).
4. Confirm in the PR description that the change does not violate any
   principle in `AGENTS.md`.
5. Squash-merge is the default. The PR title becomes the squashed commit
   subject, so make it descriptive.

## Commit messages

Write commits in English. Use a short subject line (50 characters or
fewer) written in the imperative mood ("Add X", not "Added X"). Add a
body when more context is useful.

## Architectural decisions

If your change introduces or revises a significant design choice
(language, framework, storage, IPC, host-agent adapter, etc.), add or
update an ADR in [docs/decisions/](./docs/decisions/) and link it from
the pull request.

## License

By contributing, you agree that your contributions will be licensed
under the [MIT License](./LICENSE).
