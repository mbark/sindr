#!/bin/bash

# sindr installation script
# Usage: curl -sSL https://github.com/mbark/sindr/raw/master/install.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-0.0.6}"
REPO="mbark/sindr"

# Helper functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Detect platform
detect_platform() {
    local platform
    case "$(uname -s)" in
        Darwin) platform="darwin" ;;
        Linux) platform="linux" ;;
        MINGW*|MSYS*|CYGWIN*) platform="windows" ;;
        *) 
            error "Unsupported platform: $(uname -s)"
            exit 1
            ;;
    esac
    echo "$platform"
}

# Detect architecture
detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        i386|i686) arch="386" ;;
        *) 
            error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    echo "$arch"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Download and install sindr
install_sindr() {
    local platform=$(detect_platform)
    local arch=$(detect_arch)
    local filename="sindr_${VERSION}_${platform}_${arch}.tar.gz"
    local url="https://github.com/${REPO}/releases/download/v${VERSION}/${filename}"
    
    log "Detected platform: ${platform}_${arch}"
    log "Downloading sindr v${VERSION} from ${url}"
    
    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT
    
    # Download
    if command_exists curl; then
        curl -sSL "$url" -o "$tmp_dir/$filename"
    elif command_exists wget; then
        wget -q "$url" -O "$tmp_dir/$filename"
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    # Extract
    log "Extracting $filename"
    tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"
    
    # Find the binary (handle potential directory structure)
    local binary_path
    if [[ -f "$tmp_dir/sindr" ]]; then
        binary_path="$tmp_dir/sindr"
    elif [[ -f "$tmp_dir/sindr.exe" ]]; then
        binary_path="$tmp_dir/sindr.exe"
    else
        # Look for binary in subdirectories
        binary_path=$(find "$tmp_dir" -name "sindr" -o -name "sindr.exe" | head -1)
    fi
    
    if [[ -z "$binary_path" ]]; then
        error "Could not find sindr binary in downloaded archive"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Install binary
    log "Installing to ${INSTALL_DIR}"
    
    # Check if we need sudo
    if [[ ! -w "$INSTALL_DIR" ]]; then
        if command_exists sudo; then
            sudo mv "$binary_path" "$INSTALL_DIR/"
        else
            error "No write permission to $INSTALL_DIR and sudo not available"
            error "Try running with: INSTALL_DIR=\$HOME/.local/bin $0"
            exit 1
        fi
    else
        mv "$binary_path" "$INSTALL_DIR/"
    fi
    
    success "sindr v${VERSION} installed successfully!"
}

# Verify installation
verify_installation() {
    if command_exists sindr; then
        local installed_version=$(sindr --help | head -1 | awk '{print $2}' 2>/dev/null || echo "unknown")
        success "sindr is now available in your PATH"
        log "Run 'sindr --help' to get started"
        
        # Check if it's in PATH
        local sindr_path=$(which sindr 2>/dev/null)
        if [[ -n "$sindr_path" ]]; then
            log "Installed at: $sindr_path"
        fi
    else
        warn "sindr was installed but is not in your PATH"
        warn "You may need to add ${INSTALL_DIR} to your PATH:"
        warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

# Main installation process
main() {
    log "Starting sindr installation..."
    
    # Check for required commands
    if ! command_exists tar; then
        error "tar is required but not found"
        exit 1
    fi
    
    # Install
    install_sindr
    
    # Verify
    verify_installation
    
    success "Installation complete!"
}

# Run main function
main "$@"