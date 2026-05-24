#!/bin/sh
set -e

REPO="codeatlasdev/domain-hunter"
BINARY="domain-hunter"
INSTALL_DIR="/usr/local/bin"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_${OS}_${ARCH}.tar.gz"

echo "◆ Domain Hunter installer"
echo "  Version: $LATEST"
echo "  OS: $OS/$ARCH"
echo "  URL: $URL"
echo ""

# Download and install
TMP=$(mktemp -d)
curl -fsSL "$URL" -o "$TMP/archive.tar.gz"
tar -xzf "$TMP/archive.tar.gz" -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMP"

echo "  ✓ Installed to $INSTALL_DIR/$BINARY"
echo ""
echo "  Run: domain-hunter"
