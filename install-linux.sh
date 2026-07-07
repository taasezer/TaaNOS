#!/bin/bash
# TaaNOS Linux Installer
# This script installs TaaNOS so you can launch it with a single click and see the icon!

echo "🦦 Installing TaaNOS for Linux..."

# Create directories if they don't exist
mkdir -p ~/.local/bin
mkdir -p ~/.local/share/icons
mkdir -p ~/.local/share/applications

# 1. Copy the executable
echo "-> Copying executable..."
cp bin/taanos-linux-amd64 ~/.local/bin/taanos
chmod +x ~/.local/bin/taanos

# 2. Copy the icon
echo "-> Copying icon..."
if [ -f "icon.png" ]; then
    cp icon.png ~/.local/share/icons/taanos.png
else
    echo "Warning: icon.png not found. The app will install without a custom icon."
fi

# 3. Create the .desktop shortcut
echo "-> Creating desktop shortcut..."
cat << 'EOF' > ~/.local/share/applications/taanos.desktop
[Desktop Entry]
Version=1.0
Name=TaaNOS
Comment=TaaNOS AI Terminal
Exec=taanos
Icon=taanos
Terminal=true
Type=Application
Categories=Utility;TerminalEmulator;
EOF

chmod +x ~/.local/share/applications/taanos.desktop

# Add ~/.local/bin to PATH in bashrc if not present
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    echo "-> Added ~/.local/bin to your PATH in ~/.bashrc"
fi

echo "✅ TaaNOS has been successfully installed!"
echo "You can now find TaaNOS in your applications menu and launch it with a single click!"
