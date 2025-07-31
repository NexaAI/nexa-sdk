#!/bin/sh
# This script installs Nexa on Linux.
# It detects the current operating system architecture and installs the appropriate version of Nexa.

set -eu

red="$( (/usr/bin/tput bold || :; /usr/bin/tput setaf 1 || :) 2>&-)"
plain="$( (/usr/bin/tput sgr0 || :) 2>&-)"

status() { echo ">>> $*" >&2; }
error() { echo "${red}ERROR:${plain} $*"; exit 1; }
warning() { echo "${red}WARNING:${plain} $*"; }

TEMP_DIR=$(mktemp -d)
cleanup() { rm -rf $TEMP_DIR; }
trap cleanup EXIT

# Global variables
ARCH=""
IS_WSL2=false
SUDO=""
NEXA_INSTALL_DIR=""
BINDIR=""
BACKEND=""

# Check if a command is available
available() { command -v $1 >/dev/null; }

# Check required tools
require() {
    local MISSING=''
    for TOOL in $*; do
        if ! available $TOOL; then
            MISSING="$MISSING $TOOL"
        fi
    done
    echo $MISSING
}

# Detect system environment
detect_system_environment() {
    # Detect system architecture
    ARCH=$(uname -m)
    status "Detected architecture: $ARCH"
    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        # aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    # Detect WSL2 environment
    KERN=$(uname -r)
    case "$KERN" in
        *icrosoft*WSL2 | *icrosoft*wsl2) IS_WSL2=true; status "WSL2 environment detected" ;;
        *icrosoft) error "Microsoft WSL1 is not currently supported. Please use WSL2 with 'wsl --set-version <distro> 2'" ;;
        *) ;;
    esac
}

# Check and setup sudo
setup_sudo() {
    if [ "$(id -u)" -ne 0 ]; then
        if ! available sudo; then
            error "This script requires superuser permissions. Please re-run as root."
        fi
        SUDO="sudo"
    fi
}

# Validate required tools
validate_requirements() {
    NEEDS=$(require curl gcc)
    if [ -n "$NEEDS" ]; then
        error "The following tools are required but missing: $NEEDS"
    fi
    status "All required tools are available"
}

# Let user choose backend
select_backend() {
    echo "Available backends:"
    echo "1) llama-cpp-cpu (CPU only)"
    echo "2) llama-cpp-cuda (CUDA GPU acceleration)"
    echo ""

    while true; do
        read -p "Please select backend (1 or 2): " choice
        case $choice in
            1)
                BACKEND="llama-cpp-cpu"
                status "Selected CPU backend"
                break
                ;;
            2)
                BACKEND="llama-cpp-cuda"
                status "Selected CUDA backend"
                break
                ;;
            *)
                echo "Invalid choice. Please enter 1 or 2."
                ;;
        esac
    done
}

# Install Nexa SDK
install_nexa_sdk() {
    # Determine binary directory for symlinks
    for BINDIR in /usr/local/bin /usr/bin /bin; do
        echo $PATH | grep -q $BINDIR && break || continue
    done
    NEXA_INSTALL_DIR="/opt/nexa-sdk"
    status "Installation directory: $NEXA_INSTALL_DIR"
    status "Binary directory: $BINDIR"

    # Clean up old installation
    if [ -d "$NEXA_INSTALL_DIR" ] ; then
        status "Cleaning up old version at $NEXA_INSTALL_DIR"
        $SUDO rm -rf "$NEXA_INSTALL_DIR"
    fi

    # Create necessary directories
    status "Creating installation directories"
    $SUDO install -o0 -g0 -m755 -d "$NEXA_INSTALL_DIR"

    # Download and extract Nexa
    : "${NEXA_VERSION:=latest}"
    if [ "$NEXA_VERSION" = "latest" ]; then
        NEXA_VERSION=$(curl -sSfL "https://api.github.com/repos/NexaAI/nexa-sdk/releases/latest" | \
            grep '"tag_name":' | cut -d '"' -f 4)
    fi
    : "${NEXA_BASE_URL:=https://github.com/NexaAI/nexa-sdk/releases/download}"
    NEXA_DOWNLOAD_URL="${NEXA_BASE_URL}/${NEXA_VERSION}/nexa-cli_ubuntu_22.04_${BACKEND}_${NEXA_VERSION}.tar.gz"
    status "Downloading Nexa bundle from $NEXA_DOWNLOAD_URL"
    curl --fail --show-error --location --progress-bar \
        "$NEXA_DOWNLOAD_URL" | $SUDO tar -xz -C "$NEXA_INSTALL_DIR"

    # Create symbolic links
    status "Creating symbolic links in $BINDIR"
    $SUDO ln -sf "$NEXA_INSTALL_DIR/nexa" "$BINDIR/nexa"
}

# Create system user and groups
create_system_user() {
    if ! id nexa >/dev/null 2>&1; then
        status "Creating nexa user..."
        $SUDO useradd -r -s /bin/false -U -m -d /usr/share/nexa nexa
    fi
    if getent group render >/dev/null 2>&1; then
        status "Adding nexa user to render group..."
        $SUDO usermod -a -G render nexa
    fi
    if getent group video >/dev/null 2>&1; then
        status "Adding nexa user to video group..."
        $SUDO usermod -a -G video nexa
    fi

    status "Adding current user to nexa group..."
    $SUDO usermod -a -G nexa $(whoami)
}

# Create systemd service
create_systemd_service() {
    status "Creating nexa systemd service..."
    cat <<EOF | $SUDO tee /etc/systemd/system/nexa.service >/dev/null
[Unit]
Description=Nexa Service
After=network-online.target

[Service]
ExecStart=$BINDIR/nexa serve
User=nexa
Group=nexa
Restart=always
RestartSec=3
Environment="PATH=$PATH"

[Install]
WantedBy=default.target
EOF
}

# Enable and start systemd service
enable_systemd_service() {
    SYSTEMCTL_RUNNING="$(systemctl is-system-running || true)"
    case $SYSTEMCTL_RUNNING in
        running|degraded)
            status "Enabling and starting nexa service..."
            $SUDO systemctl daemon-reload
            $SUDO systemctl enable nexa

            start_service() { $SUDO systemctl restart nexa; }
            trap start_service EXIT
            ;;
        *)
            warning "systemd is not running"
            if [ "$IS_WSL2" = true ]; then
                warning "see https://learn.microsoft.com/en-us/windows/wsl/systemd#how-to-enable-systemd to enable it"
            fi
            ;;
    esac
}

# Configure systemd service
configure_systemd() {
    create_system_user
    create_systemd_service

    enable_systemd_service
}

# Installation success message
install_success() {
    status 'The Nexa is now available at 127.0.0.1:18181.'
    status 'Install complete. Run "nexa" from the command line.'
}

# Main installation function
main() {
    [ "$(uname -s)" = "Linux" ] || error 'This script is intended to run on Linux only.'

    detect_system_environment
	validate_requirements
    setup_sudo

    select_backend
    install_nexa_sdk

    if available systemctl; then
        configure_systemd
    fi

    status 'Install complete. Run "nexa" from the command line.'
}

main