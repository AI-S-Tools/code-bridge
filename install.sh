#!/bin/bash
set -e

# Code-Bridge Installer
# Usage: curl -sSL https://raw.githubusercontent.com/AI-S-Tools/code-bridge/master/install.sh | bash

REPO="AI-S-Tools/code-bridge"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo "Installing code-bridge..."

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    linux)
        BINARY="code-bridge-linux-${ARCH}"
        ;;
    darwin)
        BINARY="code-bridge-darwin-${ARCH}"
        ;;
    *)
        echo "Error: Unsupported OS: $OS"
        exit 1
        ;;
esac

# Get latest release version
echo "Fetching latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not fetch latest release"
    exit 1
fi

echo "Latest version: $LATEST_RELEASE"

# Download binary
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${BINARY}"
echo "Downloading from: $DOWNLOAD_URL"

TMP_FILE=$(mktemp)
if ! curl -sSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    echo "Error: Failed to download binary"
    rm -f "$TMP_FILE"
    exit 1
fi

# Make executable
chmod +x "$TMP_FILE"

# Install (may need sudo)
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "$INSTALL_DIR/code-bridge"
else
    echo "Installing to $INSTALL_DIR (requires sudo)"
    sudo mv "$TMP_FILE" "$INSTALL_DIR/code-bridge"
fi

echo "✓ code-bridge installed successfully to $INSTALL_DIR/code-bridge"
echo ""
echo "Usage:"
echo "  code-bridge init         # Initialize in current directory"
echo "  code-bridge index        # Index your codebase"
echo "  code-bridge search <q>   # Search for code"
echo ""
echo "Run 'code-bridge --help' for more information"
