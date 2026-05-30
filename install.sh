#!/usr/bin/env bash
set -euo pipefail

# tide installer — one-line install for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

OWNER="0xfig521"
REPO="tide"
BINARY="tide"
INSTALL_DIR="${TIDE_INSTALL_DIR:-/usr/local/bin}"

# ── detect os/arch ──────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      echo "tide: unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)            echo "tide: unsupported architecture: $ARCH"; exit 1 ;;
esac

# ── binary name pattern: tide-darwin-arm64, tide-linux-amd64, etc. ──
TARBALL="${BINARY}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${OWNER}/${REPO}/releases/latest/download/${TARBALL}"

# ── Go install fallback ─────────────────────────────────────
if command -v go &>/dev/null; then
    echo "tide: installing via Go toolchain..."
    go install "github.com/${OWNER}/${REPO}@latest"
    echo "✓ tide installed to \$(go env GOPATH)/bin/tide"
    exit 0
fi

# ── download binary ─────────────────────────────────────────
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "tide: downloading ${TARBALL}..."
if command -v curl &>/dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"
elif command -v wget &>/dev/null; then
    wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$TARBALL"
else
    echo "tide: need curl or wget to download. Install one and retry."
    exit 1
fi

# ── extract and install ─────────────────────────────────────
cd "$TMP_DIR"
tar xzf "$TARBALL"

if [ ! -f "$BINARY" ]; then
    echo "tide: binary not found in tarball"
    exit 1
fi

if [ -w "$INSTALL_DIR" ]; then
    install -m 755 "$BINARY" "$INSTALL_DIR/$BINARY"
else
    sudo install -m 755 "$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo "✓ tide installed to $INSTALL_DIR/$BINARY"
echo "  Run 'tide --help' to get started."
