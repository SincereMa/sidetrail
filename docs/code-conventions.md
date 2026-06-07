# Code conventions

These rules are binding for every contribution to this repository,
whether from a human or an AI agent. They are enforced by `gofmt`,
`go vet`, `golangci-lint`, the test suite, and review.

The principles are short. The detail below is the application.

## Principles

1. **Read by humans first.** A change is merged when the next
   contributor can understand it without asking. Comments earn
   their place by saying *why*, not *what*.
2. **Small and obvious beats clever and short.** A 60-line file
   with three helpers beats a 20-line file with a clever trick.
3. **Errors carry context.** A returned error that does not name
   what failed, where, and why is a bug.
4. **Boundaries are real.** The architectural rules in
   [AGENTS.md](../AGENTS.md) and the ADRs in
   [docs/decisions/](./decisions/) are reflected in the code, not
   in a separate document.
5. **The code matches the test.** A test that does not exercise
   the code it claims to is worse than no test.

## Go

These rules are written for Go 1.22+ and align with Effective Go,
the Go Code Review Comments, and the Google Go style guide. Where
this document disagrees with those, this document wins for this
repository.

### Tooling (must pass on every PR)

- `gofmt -l .` is silent.
- `goimports -l .` is silent.
- `go vet ./...` is silent.
- `golangci-lint run ./...` is silent. The enabled linter set lives
  in `.golangci.yml`.
- `go test -race -count=1 ./...` passes.

### Package layout

- One package per directory. The directory name is the package
  name, lowercase, no underscores.
- `internal/` packages are not importable from outside the module.
  Use it for anything the sidecar does not expose as a public API.
- `cmd/sidetrail/` is the only command entry point. Subcommands live
  as files alongside `root.go`.
- No `pkg/` directory. `internal/` is enough.

### Imports

- Three groups, separated by blank lines: standard library, third
  party, project. `goimports` enforces this.
- No dot imports, no underscore imports except in test files for
  side-effect packages (and we have none yet).
- Module path is `github.com/SincereMa/sidetrail` per
  [ADR-0003](./decisions/0003-technology-stack.md).

### Naming

- Packages: short, lowercase, single word. `storage`, `record`,
  `schema`. No stutter: `record.Record` is fine; `record.Record`
  is not. (Yes, the project's central type is `record.Record`.
  The trade-off is documented and accepted.)
- Exported names: `PascalCase`. Unexported: `camelCase`.
- Constants: `PascalCase` for exported, `camelCase` for
  unexported. Acronyms keep their case: `URL`, `ID`, `HTTP`, not
  `Url`, `Id`, `Http`. Types use string-based enums whose values
  are lowercase: `Kind = "decision"`, never `"DECISION"`.
- Interfaces: single-method interfaces end in `-er` (`Reader`,
  `Writer`). Multi-method interfaces get a noun (`Store`).
- Errors: sentinel values are `ErrXxx`; custom types are `XxxError`.

### Comments and documentation

This is the part that needs to be said out loud. Comments are
expected, not optional.

- **Every package has a `// Package x ...` doc comment** in one
  file. It is the entry point for `go doc` and the first thing a
  reader sees.
- **Every exported name has a doc comment.** It starts with the
  name itself, in a complete sentence. `// ValidateRecord
  reports whether data conforms to the record schema.`
- **Unexported helpers do not need doc comments** unless their
  behavior is non-obvious.
- **Comments explain *why*, not *what*.** A line that restates
  the code is a line to delete. A line that names the constraint
  the code is satisfying is a line to keep.
- **No noisy banners.** Section dividers, author tags, "TODO
  rewrite this" comments, and `////` separators are out.
- **TODO comments use the form `// TODO(name): ...`** and are
  tracked. A TODO without a name and a tracker is a forgotten
  promise.

### Functions

- Aim for ≤ 50 lines. A function over 100 lines is a refactor
  candidate.
- One job per function. The function name is a verb phrase that
  fits in one breath.
- Return errors as the last value. Never return both a value and
  an error for the same outcome.
- Use guard clauses for the unhappy path. The "happy" body of a
  function is its last block, not buried under nested `if`s.

### Errors

- Wrap with context: `fmt.Errorf("read %q: %w", path, err)`. The
  `%w` verb preserves the chain for `errors.Is` / `errors.As`.
- Return errors; do not log-and-return. Logging is the caller's
  choice.
- Do not panic in business code. `init` may panic on programmer
  error (a schema that does not compile, a constant that is
  wrong); runtime code must not.
- `errors.Is(err, fs.ErrNotExist)` is preferred over
  `os.IsNotExist(err)`.

### Context and concurrency

- The first parameter of any function that may block, do I/O, or
  take a request is `ctx context.Context`.
- Do not store `context.Context` in a struct field. Pass it.
- Never pass `nil` as a context. `context.TODO()` is acceptable
  for in-progress code; `context.Background()` is acceptable at
  the entry point.
- Goroutines must have an explicit exit. There is no implicit
  shutdown.

### Resources and the filesystem

- The sidecar reads and writes only inside `.sidetrail/` relative
  to the project root. Writes outside this tree are a bug.
- Use `filepath.Join` for paths. Hard-coded `/` or `\` is a
  portability bug. `path/filepath` everywhere; `path` only for
  slash-only paths (URLs, import paths).
- File writes are atomic: write to `path + ".tmp"`, then rename.
  The `Store` in `internal/storage` follows this rule.
- No `os/exec` calls that go through a shell. Invoking a binary
  with explicit arguments is allowed; `sh -c` and `cmd /c` are
  not.

### Testing

- Tests live next to the code they test (`foo.go` →
  `foo_test.go`).
- Table-driven tests are the default when there is more than one
  case. The table is a slice of anonymous structs with `name`
  string fields.
- Subtests use `t.Run(name, func(t *testing.T) { ... })`.
- `t.Helper()` is the first line of any helper that itself
  reports an error.
- `t.TempDir()` for filesystem tests; never write to a fixed
  path.
- Race detector is part of `make test`. Tests that fail under
  `-race` are bugs.
- A test that depends on another test's state is removed on
  sight.

### Logging and output

- `fmt.Fprintln(os.Stderr, err)` from `main`; nothing else.
- The CLI library is Cobra. Subcommands emit to
  `cmd.OutOrStdout()` and `cmd.ErrOrStderr()`; do not grab
  `os.Stdout` directly.
- No log levels in v0. When the user needs a flag to control
  verbosity, add `--verbose` and stop.

## Modern software engineering

The Go rules above imply these general rules. They are stated
explicitly so that AI agents do not need to infer them.

- **Small commits, small PRs.** One logical change per commit,
  one logical change per PR. A PR that touches seven subsystems
  is seven PRs.
- **Commit subject is the user-facing summary.** "Add X" not
  "Added X". 50 characters or less. Conventional-commit prefix
  is encouraged.
- **No dead code.** Unused exports, commented-out code, and
  `_ = x` are reviewed as bugs.
- **No speculative generality.** A feature built for a future
  user is a feature the current user pays for. Build for the
  actual next step.
- **No magic numbers.** A literal in a function that is not
  obviously 0/1/-1 is a constant.
- **Reproducibility is part of correctness.** `go.sum` is
  committed. `go mod tidy` is part of the pre-commit check.
- **Secrets do not appear in code, comments, or tests.** Not
  even in fixtures. Use placeholders.
- **English-only content.** Documentation, code, comments,
  commit messages, PR descriptions, and resource files are in
  English. The only exception is user-supplied text intentionally
  in another language.

## Project-specific rules

Beyond the ADRs, these are conventions that span the codebase:

- The sidecar's public CLI surface is what `sidetrail --help` shows.
  Adding a subcommand is a deliberate act; discuss in the PR.
- File-on-disk formats are versioned by the directory layout
  itself. Adding a new field to a record is non-breaking; adding
  a new directory is a new ADR.
- The four data structures (`record`, `edge`, `signal`, `drift`)
  each get their own package and their own `.sidetrail/`
  subdirectory. They share metadata conventions but not schema.
- A `.sidetrail/_seed/`, `_proposed/`, or `_derived/` subdirectory
  holds machine-claimed records. They never appear in
  `sidetrail ask` results without an explicit `--include-seed`
  flag.
- Cross-platform discipline is enforced in review. A PR that
  uses `os/exec` with a shell, hard-codes a path separator, or
  reaches for `syscall` is rejected even if the tests pass.

## What is not in this document

- Project pitch and install instructions. See [README.md](../README.md).
- Architectural decisions. See [docs/decisions/](./decisions/).
- The five problems the sidecar exists to solve. See
  [docs/scope.md](./scope.md).
- How to file issues and submit changes. See
  [CONTRIBUTING.md](../CONTRIBUTING.md).

Update this file when a rule proves unworkable, not when a
contributor pushes back on a single PR. The rules are binding
because the consistency is the point.
