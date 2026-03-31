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

echo "Installing Soverstack Launcher (${OS}/${ARCH})..."

# Get latest release (including prereleases)
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases?per_page=1" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
VERSION=${TAG#v}

echo "Latest version: ${VERSION}"

ASSET="soverstack-${VERSION}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

# Download and extract
TMP_DIR=$(mktemp -d)
echo "Downloading ${ASSET}..."
curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${ASSET}"
tar xzf "${TMP_DIR}/${ASSET}" -C "$TMP_DIR"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}/soverstack" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP_DIR}/soverstack" "${INSTALL_DIR}/${BINARY_NAME}"
fi

chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
rm -rf "$TMP_DIR"

echo ""
echo "Soverstack ${VERSION} installed successfully!"
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
