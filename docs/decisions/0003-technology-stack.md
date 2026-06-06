# ADR-0003: Technology stack (v0)

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Constrained by:** ADR-0001 (memory model and I/O surface),
  ADR-0002 (gap 5 tribal constraints)

## Context and problem statement

ADR-0001 and ADR-0002 fix what v0 does. They do not say what v0 is
written in. This ADR chooses the language, the key libraries, the
build chain, and the distribution model, all under the hard
constraints in `AGENTS.md`: single binary on `PATH` (or a well-known
package manager), no heavy runtime, cross-platform, lightweight,
English-only.

## Decision drivers

- v0 is an experimental product; iteration speed matters more than
  peak performance.
- Cross-platform support (macOS, Linux, Windows) is a non-negotiable,
  not a nice-to-have.
- The sidecar is a CLI first; the only "user" of the library code is
  the sidecar itself. There is no public API surface to stabilize.
- The host-agent ecosystem is moving fast; we want to keep the option
  of MCP / LSP / similar adapters open without rewriting the core.

## Considered options

For each question, the chosen option is listed first; rejected
alternatives are noted.

## Decision

### 1. Primary language: Go (latest stable)

The sidecar binary is written in Go. Static, cross-platform builds are
mature; the CLI ecosystem (`terraform`, `gh`, `kubectl`, `hugo`) sets
the conventions we want to inherit.

- **Why not Rust:** Rust gives the smallest and fastest binary, but
  v0 is not performance-bound and the iteration cost is real. Rust is
  a documented fallback if a future gap (likely gap 3, signal
  collection) makes the binary too slow or too large.
- **Why not Node / Deno / Bun:** every option either requires a
  runtime the user must install, or ships a runtime that is large
  enough to fail the "lightweight" principle.
- **Why not Python:** the only honest cross-platform distribution
  story is PyInstaller, which produces a fat binary with a slow cold
  start. It also pulls in a runtime in spirit even if not in form.
- **Why not Zig:** the ecosystem for CLI tooling and JSON Schema
  validation is too thin to bet a sidecar on.
- **Why not C / C++:** development cost, risk, and the absence of a
  credible JSON Schema story.

### 2. CLI framework: `spf13/cobra`

Subcommands (`ask`, `get`, `list`, `context`, `add`, `init`, `scan`,
`verify`, `supersede`, `mcp-serve`) map cleanly onto Cobra's command
tree. Generator-produced `help` text and shell completion come for
free.

### 3. JSON Schema validation: `santhosh-tekuri/jsonschema/v5`

Draft 2020-12 support, pure Go, no cgo. Used both for `cortex
validate` and for the read path's `cortex ask` filter shape.

### 4. Identifier generation: `oklog/ulid/v2`

22-byte, lexicographically sortable, monotonic. Used for record IDs.
Sortability is a cheap property that makes directory listings
time-ordered and PR diffs stable.

### 5. Glob matching: `bmatcuk/doublestar/v4`

Used to resolve `scope` (path glob) into concrete files for
`cortex context --file`. Cross-platform, supports `**`.

### 6. Cross-platform build: GoReleaser + GitHub Actions

GoReleaser produces a matrix of static binaries (linux/amd64,
linux/arm64, darwin/amd64, darwin/arm64, windows/amd64,
windows/arm64) plus the package-manager artifacts in T7. Tagging a
release triggers the workflow.

### 7. Distribution: single binary + package-manager formulas + install script

- **macOS:** universal2 tar.gz + Homebrew formula in
  `anomalyco/homebrew-tap`. Primary install:
  `brew install anomalyco/tap/cortex`.
- **Linux:** tar.gz (fallback), `.deb`, `.rpm`. Primary install
  varies by distro; for unknown distros the install script
  (`curl -fsSL https://cortex-sidemark.dev/install.sh | sh`) places
  the binary in `~/.local/bin/` or `/usr/local/bin/`.
- **Windows:** tar.gz and zip + Scoop manifest in
  `anomalyco/scoop-bucket`. Primary install: `scoop install cortex`.
  PowerShell install script: `irm https://cortex-sidemark.dev/install.ps1
  | iex`.

Explicitly **not** doing:

- DMG / MSI / `.pkg` installers. They are not "standard install" for
  CLI tools and bring signing / notarization overhead that v0 does
  not need.
- AppImage / Flatpak / Snap. Not "well-known" for a developer CLI
  sidecar.
- A custom update server. Distribution channels handle upgrades.
- Code signing in v0. The release notes state the binary is unsigned
  and the install script prints a checksum. Signing is a v1 concern.

### 8. Things we deliberately do not introduce

- Embedded databases of any kind. ADR-0001 already ruled this out;
  restated here because every language choice makes it easy to slip
  one in.
- HTTP framework or RPC stack. v0 is a local CLI; no network.
- LLM client SDKs. v0 does not call any model.
- Any dependency that requires cgo. It breaks the clean static
  cross-platform build chain.
- A long-running daemon. Each invocation is a process that exits
  when the command returns.

### 9. Cross-platform discipline (binding for all future code)

These are not preferences; they are how ADR-0001's "cross-platform"
principle is enforced in code:

- **No shelling out.** `os/exec` is allowed only when invoking a
  binary directly, never via `sh -c` / `cmd / c` / PowerShell. Every
  pipeline must compose Go-side or call a binary with explicit args.
- **No raw syscalls.** All platform work goes through the Go standard
  library. If a need appears, it is wrapped behind an interface and
  the interface gets a stub on unsupported platforms.
- **No hard-coded path separators.** Use `filepath.Join`,
  `filepath.Separator`, and `path/filepath` everywhere. Even though
  Go accepts `/` on Windows, the discipline is what we want.
- **No terminal-color hardcoding.** `fatih/color` (when used) honors
  `NO_COLOR` and TTY detection automatically.
- **No file-locking assumptions across platforms.** Where concurrent
  access matters (v0: probably nowhere; gap 2: maybe), the design
  uses a portable mechanism, not `flock` / `LockFileEx`.

## Consequences

### Positive

- A v0 binary can be built and shipped today with the chosen toolchain.
- Cross-platform support is not a future task; it is a property of the
  Go toolchain, applied to every build.
- The library set is small (six direct dependencies for the core).
  Lightweight is enforceable, not aspirational.
- Falling back to Rust is a known escape hatch with a clear trigger,
  not a vague option.

### Negative

- The Go binary is 5–10 MB. Acceptable under "lightweight" but not
  the smallest possible. The fallback path is documented.
- Cobra brings its own conventions. Future contributors must learn
  them. This is the same cost as `clap` in Rust; the trade-off is
  not specific to Go.
- No code signing in v0. Users installing the binary from the
  install script get a checksum but not a signature. This is honest
  v0 scope, not a security claim.
- The cross-platform discipline is enforceable in review, not at
  compile time. A future ADR may add a `go vet`-style check for the
  most common violations (`os/exec` with shell, hard-coded
  separators).

### Neutral

- CI is GitHub Actions by default because the project is on GitHub.
  Switching later is a one-time cost, not a recurring one.
- Dependency upgrades follow `go get -u` and `go mod tidy`. No
  separate dependency manager is needed.

## Risks and rollback triggers

| Risk | Trigger | Action |
| --- | --- | --- |
| Binary size becomes a user complaint | Explicit user feedback or installer metrics | Evaluate `upx` (note AV false-positive risk) or a Rust rewrite of the hot path |
| JSON Schema validation becomes a hot spot on large repos | > 10k records with noticeable query latency | Move to lazy validation at the read/write boundary |
| GoReleaser cross-platform build fails for one matrix cell | First release where any cell fails | Patch the build; if recurring, drop the cell and document |
| A future gap demands long-running behavior (gap 3 signals) | Gap 3 ADR concludes the sidecar must watch continuously | Revisit; possibly fork a `cortex-watch` subcommand, not the main binary |
| Cobra dependency churns | Breaking change in a major version | Pin; do not chase majors without an ADR |
