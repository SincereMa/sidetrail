// Package version exposes build-time version metadata.
//
// The Version and Commit variables are set at build time via
// ldflags:
//
//	-X github.com/SincereMa/sidetrail/internal/version.Version=v0.1.0
//	-X github.com/SincereMa/sidetrail/internal/version.Commit=abc1234
package version

// Version is the human-readable release tag. Defaults to "dev"
// for builds that do not pass an ldflags override.
var Version = "dev"

// Commit is the short git SHA. Defaults to "none" for builds
// that do not pass an ldflags override.
var Commit = "none"
