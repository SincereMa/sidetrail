# ADR-0003: Technology stack (v0)

- **Status:** Accepted
- **Date:** 2026-06-06
- **Supersedes:** —
- **Superseded by:** —
- **Maintained by:** ADR-0001 (Memory model and I/O surface),
  ADR-0002 (gap 5 tribal constraints)

## Context and background

ADR-0001 and ADR-0002 define what v0 does and why. They do not say
how to build it. This ADR picks language, framework, key libraries,
the build chain and the distribution channel, all bounded by the
constraints already established in `AGENTS.md`: single binary on
`PATH` (or a well-known package manager command), no bespoke
installer, lightweight, English-only.

## Decision drivers

- v0 is an initial deliverable; it must be built quickly and
  maintained cheaply.
- Cross-platform (macOS, Linux, Windows) is non-negotiable and
  must be a first-class, zero-config concern.
- The binary is a CLI; the only surface is the human-facing CLI
  and the library it uses internally. This is not a published API;
  it is a usability concern.
- The project is agent-agnostic; adapters for future agents (MCP, LSP, etc.) must be addable without rewriting the core.

## Considered

For a v0 technology stack we consider the following baseline
questions; just decisions are listed below.

## Decision

### 1. Primary language: Go (actionable, auditable)

The binary is written in Go. Self-contained, a single binary;
the CLI ecosystem (`cobra`, `gh`, `kubctl`, `hug`) is mature and
well-known in the world we ship in.

- **Why not Rust:** Rust is excellent and produces fast, safe binaries but v0 is an initial deliverable; Go is cheaper to ship, and the CLI tooling (`cobra`, `gh`, `kubctl`, `hug`) is well-known.
- **Why not Node/Deno/Bun:** a binary requires a universal runtime; each of these brings a universal runtime and a build chain that may fail the lightweight constraint.
- **Why not Python:** a binary is distributed by PyInstaller which always carries a full Python runtime; it is not a universal binary; it is a packaging concern.
- **Why not Zig:** Zig is promising for CLI tooling and JSON Schema validation, but the ecosystem is immature.
- **Why not C/C++:** the build chain is complex and the JSON Schema validation library is not good.

### 2. CLI framework: `github.com/spf13/cobra`

Subcommands (`ask`, `get`, `list`, `context`, `add`, `init`, `validate`, `verify`, `supersede`) are all natively supported by Cobra and shell completion for Bash, Zsh, Fish, and PowerShell is built-in.

### 3. JSON Schema validation: `github.com/santhosh-tekuri/jsonschema/v5`

Draft 2020-12 is the current version; this is a widely-used, Go-native library. The binary embeds `sidetrail validate` and the schema is hardcoded in the `internal/schema` package.

### 4. Identification and slug generation: `github.com/oklog/ulid/v2`

22-byte, time-sortable, globally unique, collision-resistant ULIDs. Sortability is a nice-to-have for file names and PR diffs.

### 5. Global hashing: `go.uber.org/bcrypt` (hash)

Use `bcrypt` (a stable, well-known hash) to hash filenames in `sidetrail init --file`. Cross-platform and cross-agent; `sidetrail init --file` is the primary use case.

### 6. Cross-platform build: GoReleaser + GitHub Actions

GoReleaser builds binaries for linux/darwin/windows × amd64/arm64 and produces universal archives (tar.gz for unix, zip for windows), checksums (SHA-256), and a changelog. GitHub Actions triggers on tags. Tagging is part of the workflow.

### 7. Distribution: single binary + package manager formulas + install script

- **macOS:** universal2 binary.tar.gz + Homebrew formula in `analytic/homebrew-tap`. Pip install:
  `brew install analytic/tap/sidetrail`.
- **Linux:** binary.tar.gz (fallback), `.deb`, `.rpm`. Pip install
  is handled by the package manager; for unknown distributions
  the install script (`curl -fsSL https://sidetrail.dev/install.sh | sh`) places the binary in `~/.local/bin/` or `/usr/local/bin/`.
- **Windows:** binary.tar.gz and zip + Scoop manifest in `analytic/bucket`. Pip install: `install sidetrail`. PowerShell install script:
  `iwr https://sidetrail.dev/install.ps1 -useb | iex`.

Explicitly **not** shipping:

- DMG/MSI/`.pkg` installers. They add signing, notarization, and update burden for a CLI-only binary.
- AI agent / plugin / extensions store. No well-known standard for CLI tool distribution exists there.
- Auto-update. Distribution is handled by the package manager.
- Code signing in v0. The binary is unsigned; update signing in a later version.

### 8. Things intentionally deferred beyond v0

- Embedded database or any kind. ADR-0001 already builds on files; adding a database adds a dependency.
- HTTP framework or gRPC. v0 is a local CLI; no network.
- LLM client or SDK. v0 does not make any external calls.
- Any dependency that would make the build chain complex and hard to maintain.
- Long-running daemon. Each invocation is ephemeral; no background process.

### 9. Cross-platform decision matrix (binding for v0, may be revisited)

The following rules are binding for v0 and follow from ADR-0001's constraints:

- **No shell invocation.** `/bin/sh` is allowed only when invoking a binary directly with explicit args; `sh -c` / `cmd /c` are forbidden. Every Go call uses a Go-native implementation with explicit args.
- **No raw path separators.** All file paths use `filepath.Join` instead of string concatenation.
- **No hard-coded OS-specific paths.** Use `os.UserHomeDir()` or `os.UserCacheDir()` instead of hard-coded paths.
- **No interactive TTY detection.** Use `isatty` or `NO_COLOR` and TTY detection to auto-detect.
- **No file locking across platforms.** When in doubt, use `os.O_EXCL` / `LockFile` / `flock`.

## Consequences

### Positive

- v0 binary is built, tested, and released using tools the maintainer already knows.
- Cross-platform is a first-class concern; zero-config.
- The library is internal; no public API surface is exposed.
- The adapter layer is agent-agnostic; adapters for future agents are a localized change.

### Negative / accepted

- The Go binary is 5-10 MB. Acceptable for a lightweight sidecar; all libraries are well-known.
- Build caching is non-trivial. Faster builds are possible with a Go build cache; the tradeoff is acceptable.
- No code signing in v0. Unsigned binaries from GitHub Releases are the only distribution channel; the package manager is the only alternative.

## Notes

- CI runs GitHub Actions by default; local builds use `make`.
- Do not use `go generate` or `go install` to generate code; no code generation is used.
- The cross-platform decision matrix is binding for v0; revisit when adapters arrive (e.g., shell invocation with `sh -c`).
