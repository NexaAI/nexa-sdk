#!/bin/sh
# This script installs a custom executable on Linux.
# It detects the current operating system architecture and CUDA availability,
# then installs the appropriate version of the executable.

set -eu

status() { echo ">>> $*" >&2; }
error() { echo "ERROR $*"; exit 1; }
warning() { echo "WARNING: $*"; }

TEMP_DIR=$(mktemp -d)
cleanup() { rm -rf $TEMP_DIR; }
trap cleanup EXIT

available() { command -v $1 >/dev/null; }
require() {
    local MISSING=''
    for TOOL in $*; do
        if ! available $TOOL; then
            MISSING="$MISSING $TOOL"
        fi
    done

    echo $MISSING
}

[ "$(uname -s)" = "Linux" ] || error 'This script is intended to run on Linux only.'

ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

SUDO=
if [ "$(id -u)" -ne 0 ]; then
    if ! available sudo; then
        error "This script requires superuser permissions. Please re-run as root."
    fi
    SUDO="sudo"
fi

NEEDS=$(require curl awk grep sed tee xargs)
if [ -n "$NEEDS" ]; then
    status "ERROR: The following tools are required but missing:"
    for NEED in $NEEDS; do
        echo "  - $NEED"
    done
    exit 1
fi

INSTALL_DIR="/opt"
BINARY_NAME="nexa"

# Function to check for CUDA availability
check_cuda() {
    if available nvidia-smi && [ -n "$(nvidia-smi | grep -o "CUDA Version: [0-9]*\.[0-9]*")" ]; then
        return 0
    else
        return 1
    fi
}

# Determine which version to download based on CUDA availability
if check_cuda; then
    status "CUDA GPU detected. Downloading CUDA version."
    DOWNLOAD_URL="https://public-storage.nexa4ai.com/nexa-sdk-executable-installer/nexa-0082-cuda-amd64.tar.gz"
    FOLDER_NAME="nexa-cuda"
else
    status "No CUDA GPU detected. Downloading CPU version."
    DOWNLOAD_URL="https://public-storage.nexa4ai.com/nexa-sdk-executable-installer/nexa-0082-cpu-amd64.tar.gz"
    FOLDER_NAME="nexa"
fi

# Download and extract the application
status "Downloading application..."
curl --fail --show-error --location --progress-bar "$DOWNLOAD_URL" | $SUDO tar -xzf - -C "$INSTALL_DIR"

# Create a symbolic link in /usr/local/bin
status "Creating symbolic link for easy access..."
$SUDO ln -sf "$INSTALL_DIR/$FOLDER_NAME/$FOLDER_NAME" "/usr/local/bin/$BINARY_NAME"

# Ensure /usr/local/bin is in PATH for all users
if ! grep -q '/usr/local/bin' /etc/environment; then
    status "Adding /usr/local/bin to PATH in /etc/environment..."
    echo 'PATH="/usr/local/bin:$PATH"' | $SUDO tee -a /etc/environment > /dev/null
fi

# Set up a systemd service (if needed)
setup_systemd() {
    status "Setting up systemd service..."
    cat <<EOF | $SUDO tee /etc/systemd/system/nexa.service >/dev/null
[Unit]
Description=Nexa Application Service
After=network.target

[Service]
ExecStart=$INSTALL_DIR/$FOLDER_NAME/$FOLDER_NAME
Restart=always
User=nobody

[Install]
WantedBy=multi-user.target
EOF

    $SUDO systemctl daemon-reload
    $SUDO systemctl enable nexa.service
    $SUDO systemctl start nexa.service
}

if available systemctl; then
    setup_systemd
fi

status "Installation complete. You can now run '$BINARY_NAME' from the command line."
