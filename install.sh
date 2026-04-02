#!/bin/bash
# Soverstack Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/soverstack/soverstack/main/install.sh | bash

set -e

REPO="soverstack/soverstack"
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

echo "Installing Soverstack (${OS}/${ARCH})..."

# Get latest stable release
RELEASE_JSON=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")
TAG=$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
VERSION=${TAG#v}

echo "Latest version: ${VERSION}"

TMP_DIR=$(mktemp -d)

# Try tar.gz archive first, then fall back to raw binary
ARCHIVE="soverstack-${VERSION}-${OS}-${ARCH}.tar.gz"
ARCHIVE_URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"

RAW_BINARY="soverstack-${OS}-${ARCH}"
RAW_URL="https://github.com/${REPO}/releases/download/${TAG}/${RAW_BINARY}"

if curl -fsSL --head "$ARCHIVE_URL" > /dev/null 2>&1; then
  echo "Downloading ${ARCHIVE}..."
  curl -fsSL "$ARCHIVE_URL" -o "${TMP_DIR}/${ARCHIVE}"
  tar xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"
else
  echo "Downloading ${RAW_BINARY}..."
  curl -fsSL "$RAW_URL" -o "${TMP_DIR}/soverstack"
fi

# Install
chmod +x "${TMP_DIR}/soverstack"
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}/soverstack" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP_DIR}/soverstack" "${INSTALL_DIR}/${BINARY_NAME}"
fi

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
