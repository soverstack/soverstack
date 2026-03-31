#!/bin/bash
# Soverstack Launcher Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/soverstack/cli-launcher/main/install.sh | bash

set -e

REPO="soverstack/cli-launcher"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="soverstack"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Error: Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)             echo "Error: Unsupported architecture: $ARCH"; exit 1 ;;
esac

ASSET="soverstack-${OS}-${ARCH}"
echo "Installing Soverstack Launcher (${OS}/${ARCH})..."

# Get latest release download URL
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

# Download
TMP=$(mktemp)
echo "Downloading ${DOWNLOAD_URL}..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMP"

# Install
chmod +x "$TMP"
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo ""
echo "Soverstack installed successfully!"
echo ""
$BINARY_NAME --version
echo ""

# Check Docker
if command -v docker &> /dev/null; then
  echo "Docker: found"
else
  echo "WARNING: Docker is not installed."
  echo "  Install it from: https://www.docker.com/products/docker-desktop/"
fi
