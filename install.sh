#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Installing junit-to-checklist..."

# Build binary
echo "Building binary..."
go build -o junit-to-checklist ./cmd/junit-to-checklist

# Determine install location
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -w "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
else
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Copy binary
echo "Installing to $INSTALL_DIR..."
cp junit-to-checklist "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/junit-to-checklist"

# Clean up local binary
rm junit-to-checklist

# Check if install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
    echo ""
    echo "Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo -e "${GREEN}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
    echo ""
    echo "Then run: source ~/.bashrc  (or ~/.zshrc)"
fi

# Verify installation
if command -v junit-to-checklist &> /dev/null; then
    echo -e "${GREEN}✓ Successfully installed junit-to-checklist${NC}"
    if ! junit-to-checklist --version; then
        echo -e "${RED}Error: Binary installed but --version check failed${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ Installation complete, but junit-to-checklist not found in PATH${NC}"
    echo "You may need to restart your shell or run: hash -r"
fi
