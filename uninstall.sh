#!/usr/bin/env bash
set -e

echo "╔══════════════════════════════════════════════════════════╗"
echo "║                Uninstalling TaaNOS CLI                   ║"
echo "╚══════════════════════════════════════════════════════════╝"

INSTALL_DIR="/usr/local/bin"

if [ -f "$INSTALL_DIR/taanos" ]; then
    echo "🗑️  Removing binary from $INSTALL_DIR/taanos (requires sudo)..."
    sudo rm -f "$INSTALL_DIR/taanos"
else
    echo "✅ TaaNOS binary not found in $INSTALL_DIR"
fi

CONFIG_DIR="$HOME/.taanos"

if [ -d "$CONFIG_DIR" ]; then
    echo "🗑️  Removing TaaNOS configuration and history ($CONFIG_DIR)..."
    rm -rf "$CONFIG_DIR"
else
    echo "✅ Configuration directory not found"
fi

echo ""
echo "✅ Uninstallation complete!"
