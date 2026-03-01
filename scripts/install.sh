#!/usr/bin/env bash
set -euo pipefail

# AgentMux Installer
# Usage: curl -sSL https://raw.githubusercontent.com/HubQC/my-agentmux/main/scripts/install.sh | bash

REPO="HubQC/my-agentmux"
BINARY="agentmux"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="armv7" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "🔧 AgentMux Installer"
echo "  OS:   $OS"
echo "  Arch: $ARCH"
echo ""

# Get latest release
LATEST=$(curl -sS "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "⚠ No releases found. Building from source..."
    
    # Check Go is installed
    if ! command -v go &> /dev/null; then
        echo "Error: Go is required to build from source. Install it first."
        exit 1
    fi

    TMPDIR=$(mktemp -d)
    trap "rm -rf $TMPDIR" EXIT

    echo "  Cloning repository..."
    git clone --depth 1 "https://github.com/${REPO}.git" "$TMPDIR/agentmux" 2>/dev/null

    echo "  Building..."
    cd "$TMPDIR/agentmux"
    go build -o "$BINARY" .

    echo "  Installing to $INSTALL_DIR..."
    sudo mv "$BINARY" "$INSTALL_DIR/$BINARY"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}_${OS}_${ARCH}.tar.gz"
    
    echo "  Version: $LATEST"
    echo "  URL: $DOWNLOAD_URL"
    echo ""

    TMPDIR=$(mktemp -d)
    trap "rm -rf $TMPDIR" EXIT

    echo "  Downloading..."
    curl -sSL "$DOWNLOAD_URL" | tar -xz -C "$TMPDIR"

    echo "  Installing to $INSTALL_DIR..."
    sudo mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "✓ Installed $BINARY to $INSTALL_DIR/$BINARY"
echo ""
echo "  Run 'agentmux init' to get started."
echo "  Run 'agentmux --help' for usage."
