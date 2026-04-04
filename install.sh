#!/usr/bin/env bash
set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║                Installing TaaNOS CLI                     ║"
echo "╚══════════════════════════════════════════════════════════╝"

# Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
fi

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    echo "❌ Unsupported OS: $OS"
    exit 1
fi

# Define URLs (Points to the Latest GitHub Release)
REPO="taasezer/TaaNOS"
BINARY_URL="https://github.com/$REPO/releases/latest/download/taanos-${OS}-${ARCH}"

TMP_FILE="/tmp/taanos"

echo "📥 Downloading TaaNOS for ${OS}/${ARCH}..."
curl -fsSL -o "$TMP_FILE" "$BINARY_URL"

echo "⚙️  Making binary executable..."
chmod +x "$TMP_FILE"

# Move to /usr/local/bin
INSTALL_DIR="/usr/local/bin"
echo "📦 Installing to $INSTALL_DIR (requires sudo)..."
sudo mv "$TMP_FILE" "$INSTALL_DIR/taanos"

echo ""
echo "✅ Installation complete!"
echo "🚀 Run 'taanos init' to set up your system."
