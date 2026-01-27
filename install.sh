#!/bin/bash
set -e

# feedme Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/JonathanWThom/feedme/main/install.sh | bash

REPO="JonathanWThom/feedme"
BINARY_NAME="fm"
INSTALL_DIR="${HOME}/.local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin|linux)
        EXT="tar.gz"
        ;;
    mingw*|msys*|cygwin*)
        OS="windows"
        EXT="zip"
        BINARY_NAME="fm.exe"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Get latest release tag
echo "Detecting latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "Could not detect latest release."
    echo ""
    echo "Alternative installation methods:"
    echo ""
    echo "1. Install with Go:"
    echo "   go install github.com/$REPO@latest"
    echo ""
    echo "2. Build from source:"
    echo "   git clone https://github.com/$REPO.git"
    echo "   cd feedme && go build -o fm ."
    echo ""
    exit 1
fi

# Download URL (matches goreleaser naming: fm_darwin_arm64.tar.gz)
FILENAME="fm_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

echo "Downloading $BINARY_NAME $LATEST for $OS/$ARCH..."
echo "URL: $DOWNLOAD_URL"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download archive
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME"; then
    echo ""
    echo "Download failed."
    echo ""
    echo "Alternative installation methods:"
    echo ""
    echo "1. Install with Go:"
    echo "   go install github.com/$REPO@latest"
    echo ""
    echo "2. Build from source:"
    echo "   git clone https://github.com/$REPO.git"
    echo "   cd feedme && go build -o fm ."
    echo ""
    exit 1
fi

# Extract
cd "$TMP_DIR"
if [ "$EXT" = "tar.gz" ]; then
    tar xzf "$FILENAME"
else
    unzip -q "$FILENAME"
fi

# Create install directory if needed
mkdir -p "$INSTALL_DIR"

# Install
mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"

echo ""
echo "Successfully installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"

# Check if install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "NOTE: $INSTALL_DIR is not in your PATH."
    echo "Add it by running:"
    echo ""
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    echo ""
    echo "Or for zsh:"
    echo ""
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
    echo "  source ~/.zshrc"
fi

echo ""
echo "Run 'fm' to start browsing tech news!"
echo "  fm           - Browse Hacker News"
echo "  fm -s lobsters - Browse Lobste.rs"
