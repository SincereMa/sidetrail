#!/usr/bin/env bash
# Install the cortex CLI binary.
#
# Usage: install.sh [options]
#
# Options:
#   -V, --version <ver>    Install a specific version (default: latest)
#   -d, --dir <path>       Install directory (default: ~/.local/bin)
#   -r, --repo <owner/name> GitHub repo (default: SincereMa/cortex-sidemark)
#   -h, --help             Show this help
#
# Environment variables (overridden by flags):
#   CORTEX_VERSION         Same as --version
#   CORTEX_INSTALL_DIR     Same as --dir
#   CORTEX_REPO            Same as --repo

set -euo pipefail

print_usage() {
  sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
}

REPO="${CORTEX_REPO:-SincereMa/cortex-sidemark}"
VERSION="${CORTEX_VERSION:-}"
INSTALL_DIR="${CORTEX_INSTALL_DIR:-$HOME/.local/bin}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -V|--version)
      VERSION="$2"
      shift 2
      ;;
    -d|--dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -r|--repo)
      REPO="$2"
      shift 2
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      echo "install.sh: unknown flag: $1" >&2
      print_usage >&2
      exit 64
      ;;
  esac
done

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "install.sh: required tool not found: $1" >&2
    exit 69
  fi
}

fetch() {
  local url="$1"
  local out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$out" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$out" "$url"
  else
    echo "install.sh: required tool not found: curl or wget" >&2
    exit 69
  fi
}

resolve_latest_version() {
  local url="https://api.github.com/repos/${REPO}/releases/latest"
  local body
  body="$(mktemp)"
  fetch "$url" "$body"
  # Parse tag_name without invoking a shell pipeline.
  local tag
  tag="$(grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' "$body" | head -n1 | sed 's/.*"\([^"]*\)"$/\1/')"
  rm -f "$body"
  if [[ -z "$tag" ]]; then
    echo "install.sh: could not resolve latest version from $url" >&2
    exit 69
  fi
  echo "$tag"
}

detect_platform() {
  local os arch
  case "$(uname -s)" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *)      echo "install.sh: unsupported OS: $(uname -s)" >&2; exit 69 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)            echo "install.sh: unsupported arch: $(uname -m)" >&2; exit 69 ;;
  esac
  echo "$os $arch"
}

verify_sha256() {
  local archive="$1"
  local checksums="$2"
  local archive_name
  archive_name="$(basename "$archive")"
  local expected
  expected="$(awk -v name="$archive_name" '$2 == name { print $1 }' "$checksums")"
  if [[ -z "$expected" ]]; then
    echo "install.sh: checksum not found for $archive_name" >&2
    return 1
  fi
  local actual
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$archive" | awk '{ print $1 }')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$archive" | awk '{ print $1 }')"
  else
    echo "install.sh: required tool not found: sha256sum or shasum" >&2
    return 1
  fi
  if [[ "$expected" != "$actual" ]]; then
    echo "install.sh: checksum mismatch" >&2
    echo "  expected: $expected" >&2
    echo "  actual:   $actual" >&2
    return 1
  fi
}

require tar

if [[ -z "$VERSION" ]]; then
  VERSION="$(resolve_latest_version)"
fi
if [[ "$VERSION" != v* ]]; then
  VERSION="v${VERSION}"
fi

read -r OS ARCH < <(detect_platform)

PROJECT="cortex"
EXT="tar.gz"
if [[ "$OS" == "windows" ]]; then
  EXT="zip"
fi
ARCHIVE_NAME="${PROJECT}_${VERSION#v}_${OS}_${ARCH}.${EXT}"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE_NAME}"
CHECKSUMS_URL="${BASE_URL}/${PROJECT}_${VERSION#v}_checksums.txt"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "install.sh: downloading $ARCHIVE_URL"
fetch "$ARCHIVE_URL" "$WORK/$ARCHIVE_NAME"

echo "install.sh: downloading $CHECKSUMS_URL"
fetch "$CHECKSUMS_URL" "$WORK/checksums.txt"

echo "install.sh: verifying checksum"
verify_sha256 "$WORK/$ARCHIVE_NAME" "$WORK/checksums.txt"

echo "install.sh: extracting"
mkdir -p "$WORK/extract"
case "$EXT" in
  tar.gz) tar -xzf "$WORK/$ARCHIVE_NAME" -C "$WORK/extract" ;;
  zip)    unzip -q "$WORK/$ARCHIVE_NAME" -d "$WORK/extract" ;;
  *)      echo "install.sh: unsupported archive: $EXT" >&2; exit 69 ;;
esac

mkdir -p "$INSTALL_DIR"
echo "install.sh: installing to $INSTALL_DIR/$PROJECT"
install -m 0755 "$WORK/extract/$PROJECT" "$INSTALL_DIR/$PROJECT"

echo
echo "cortex installed at: $INSTALL_DIR/$PROJECT"
"$INSTALL_DIR/$PROJECT" --version
echo
if ! command -v cortex >/dev/null 2>&1; then
  echo "Note: $INSTALL_DIR is not on PATH. Add it to your shell profile." >&2
fi
