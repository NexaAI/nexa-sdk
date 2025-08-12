#!/bin/bash
#
# Self-contained installer for the Nexa SDK on Linux.
# This script includes an embedded binary payload.
#
# Usage:
#   chmod +x ./install.sh
#   sudo ./install.sh

set -eu

# --- Shell UI Helper Functions ---
red="$( (tput bold 2>/dev/null || :) && (tput setaf 1 2>/dev/null || :) )"
plain="$( tput sgr0 2>/dev/null || : )"

status() { echo ">>> $*" >&2; }
error() { echo "${red}ERROR:${plain} $*" >&2; exit 1; }
warning() { echo "${red}WARNING:${plain} $*" >&2; }

# --- Cleanup handler ---
# Ensures temporary files are removed on exit
TEMP_DIR=$(mktemp -d)
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# --- Global Variables ---
SUDO=""
NEXA_INSTALL_DIR="/opt/nexa-cli"
BINDIR="/usr/local/bin"
IS_WSL2=false

# --- Prerequisite and Environment Checks ---

# Checks if a command is available on the system
available() {
    command -v "$1" >/dev/null 2>&1
}

# Checks for a list of required tools and reports missing ones
require_tools() {
    local missing_tools=""
    for tool in "$@"; do
        if ! available "$tool"; then
            missing_tools="${missing_tools} ${tool}"
        fi
    done
    echo "$missing_tools"
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

# Sets up the SUDO variable if not running as root
setup_sudo() {
    if [ "$(id -u)" -ne 0 ]; then
        if ! available sudo; then
            error "This script requires superuser permissions, but 'sudo' is not available. Please re-run as root."
        fi
        SUDO="sudo"
        status "Using 'sudo' for privileged operations."
    fi
}

# Validates all system requirements before proceeding
validate_requirements() {
    local needs
    needs=$(require_tools tar base64)
    if [ -n "$needs" ]; then
        error "The following required tools are missing:$needs"
    fi
    status "All required tools are available."
}

# --- Core Installation Logic ---

# Main installation function
install_nexa_sdk() {
    status "Starting Nexa SDK installation..."

    # --- 1. Locate and extract the embedded payload ---
    local payload_line
    payload_line=$(awk '/^__PAYLOAD_BELOW__/ {print NR + 1}' "$0")
    if [ -z "$payload_line" ]; then
        error "Could not find payload in the script. The installer appears to be corrupted."
    fi

    local temp_extract_dir
    temp_extract_dir=$(mktemp -d)

    status "Extracting embedded payload to a temporary directory..."
    tail -n "+$payload_line" "$0" | base64 --decode | tar -xzf - -C "$temp_extract_dir"
    if [ $? -ne 0 ]; then
        rm -rf "$temp_extract_dir"
        error "Failed to extract payload. The installer might be corrupted or incomplete."
    fi

    # --- 2. Clean up previous installations ---
    if [ -d "$NEXA_INSTALL_DIR" ]; then
        status "Removing existing installation at $NEXA_INSTALL_DIR"
        $SUDO rm -rf "$NEXA_INSTALL_DIR"
    fi

    # --- 3. Install new files ---
    status "Creating installation directory: $NEXA_INSTALL_DIR"
    $SUDO install -o root -g root -m 755 -d "$NEXA_INSTALL_DIR"

    status "Installing Nexa SDK files..."
    $SUDO mv "$temp_extract_dir"/* "$NEXA_INSTALL_DIR/"

    # --- 4. Create symbolic links ---
    status "Creating symbolic links in $BINDIR..."
    $SUDO mkdir -p "$BINDIR"
    $SUDO ln -sf "$NEXA_INSTALL_DIR/nexa" "$BINDIR/nexa"

    # --- 5. Clean up ---
    rm -rf "$temp_extract_dir"
    status "Nexa SDK files installed successfully."
}

# --- Systemd Service Configuration ---

# Creates the nexa system user and adds to relevant groups
create_system_user() {
    if ! id nexa >/dev/null 2>&1; then
        status "Creating system user 'nexa'..."
        $SUDO useradd -r -s /bin/false -U -m -d /usr/share/nexa nexa
    fi
    if getent group render >/dev/null 2>&1; then
        status "Adding 'nexa' user to 'render' group..."
        $SUDO usermod -a -G render nexa
    fi
    if getent group video >/dev/null 2>&1; then
        status "Adding 'nexa' user to 'video' group..."
        $SUDO usermod -a -G video nexa
    fi

    status "Adding current user ($(whoami)) to 'nexa' group..."
    $SUDO usermod -a -G nexa "$(whoami)"
}

# Creates the systemd service unit file
create_systemd_service() {
    status "Creating nexa systemd service..."
    # Using a heredoc with sudo tee to write the file as root
    $SUDO tee /etc/systemd/system/nexa.service >/dev/null <<EOF
[Unit]
Description=Nexa Background Service
After=network-online.target

[Service]
Type=simple
ExecStart=$BINDIR/nexa serve
User=nexa
Group=nexa
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
EOF
}

# Enables and starts the systemd service if systemd is running
enable_systemd_service() {
    local systemctl_running
    systemctl_running="$(systemctl is-system-running 2>/dev/null || echo 'unknown')"

    case "$systemctl_running" in
        running|degraded)
            status "Enabling and starting nexa service via systemd..."
            $SUDO systemctl daemon-reload
            $SUDO systemctl enable nexa.service
            $SUDO systemctl restart nexa.service
            ;;
        *)
            warning "systemd does not appear to be running."
            if [ "$IS_WSL2" = true ]; then
                warning "To enable systemd in WSL2, see: https://devblogs.microsoft.com/commandline/systemd-support-is-now-available-in-wsl/"
            fi
            warning "The nexa service has been installed but not started."
            ;;
    esac
}

# --- Main Execution ---

# Main function to orchestrate the installation
main() {
    if [ "$(uname -s)" != "Linux" ]; then
        error "This script is intended to run on Linux only."
    fi

    status "Starting Nexa SDK installer..."

    setup_sudo
    detect_system_environment
    validate_requirements

    install_nexa_sdk

    create_system_user
    create_systemd_service
    enable_systemd_service

    status "${plain}Install complete! The Nexa SDK is now installed."
    status "You can use the 'nexa' commands from your terminal."
    status "You may need to start a new terminal session for the 'nexa' group membership to take effect."
    status "The background service is running. Check its status with: sudo systemctl status nexa"
}

# Run the main function with all arguments passed to the script
main "$@"

# --- IMPORTANT ---
# The script MUST exit before the payload marker.
# The CI/CD process will append the base64 encoded payload below this line.
exit 0
__PAYLOAD_BELOW__
