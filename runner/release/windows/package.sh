#！/bin/bash

# Usage:
#   ./pack.sh <arch> [version]

# Setup:
#   winget install git
#   pip install pyinstaller
#   winget install -e --id JRSoftware.InnoSetup

# In case above cannot work, install from `https://files.jrsoftware.org/is/6/innosetup-6.4.3.exe` manually then add the path to the executable in the path environment variable

# Example, in git bash
#   bash package.sh  arm64 v0.2.22-rc6

# Export necessary environment variables for Azure signing
# Keep the secrets!!!


export AZURE_KEY_VAULT_URI=""
export AZURE_CERT_NAME=""
export AZURE_CLIENT_ID=""
export AZURE_CLIENT_SECRET=""
export AZURE_TENANT_ID=""

# --- Configuration ---
TARGET_ARCH=$1
VERSION_TAG=$2
REPO_ROOT=$(git rev-parse --show-toplevel)
ARTIFACTS_DIR="$REPO_ROOT/artifacts"
RUNNER_DIR="$REPO_ROOT/runner"

# --- Helper Functions ---
log_info() {
    echo "INFO: $1"
}

log_step() {
    echo ""
    echo "=============================================================================="
    echo " STEP: $1"
    echo "=============================================================================="
}



build_cli() {
    export VERSION="$VERSION_TAG"
    export ARCH="$TARGET_ARCH"

    log_step "Building Nexa CLI"
    cd "$REPO_ROOT/runner"
    make build ARCH="$TARGET_ARCH" VERSION="$VERSION_TAG"

    log_info "Moving build output to artifacts directory..."
    mkdir -p "$ARTIFACTS_DIR"
    cp -rf  "$RUNNER_DIR/build"/* "$ARTIFACTS_DIR/"
    rm "$ARTIFACTS_DIR/ml.h"
    log_info "Build complete. Artifacts are in $ARTIFACTS_DIR"
}

build_installer_package() {
    log_step "Building Installer Package"


    log_info "Building launcher with PyInstaller..."
    pyinstaller --onefile --noconsole --distpath "$ARTIFACTS_DIR" \
        --name "nexa-cli-launcher" \
        --icon="$RUNNER_DIR/release/windows/nexa_logo.ico" \
        "$RUNNER_DIR/release/windows/nexa_launcher.py"
}

sign_artifacts() {
    log_step "Signing artifacts with AzureSignTool"

    if [ -z "$AZURE_KEY_VAULT_URI" ]; then
        log_info "Azure signing environment variables not set. Skipping signing."
        return
    fi

    log_info "Installing AzureSignTool..."
    dotnet tool install --global AzureSignTool

    log_info "Recursively signing all executables and DLLs in artifacts directory..."

    find "$ARTIFACTS_DIR" -type f \( -name "*.exe" -o -name "*.dll" \) | while IFS= read -r file; do
        log_info "Signing $file..."
        azuresigntool sign \
            -kvu "$AZURE_KEY_VAULT_URI" \
            -kvc "$AZURE_CERT_NAME" \
            -kvi "$AZURE_CLIENT_ID" \
            -kvs "$AZURE_CLIENT_SECRET" \
            -kvt "$AZURE_TENANT_ID" \
            -tr http://timestamp.globalsign.com/tsa/advanced \
            -td sha256 \
            "$file"
    done
}

build_installer() {
    log_info "Compiling Inno Setup installer..."
    # The Inno Setup script likely looks for files in the script's directory or a relative path.
    # We pass the artifacts path and version as defines.
    powershell.exe -Command "& 'ISCC.exe' \
        '${RUNNER_DIR}/release/windows/nexa_installer.iss' \
        /O'${ARTIFACTS_DIR}' \
        /F'nexa-cli_windows_${TARGET_ARCH}'"

    log_info "Installer created in $ARTIFACTS_DIR/nexa-cli_windows_${TARGET_ARCH}.exe"
}


sign_installer() {
    local installer_file="$ARTIFACTS_DIR/nexa-cli_windows_${TARGET_ARCH}.exe"

    log_step "Signing Inno Setup Installer: $installer_file"

    if [ ! -f "$installer_file" ]; then
        log_info "Installer file '$installer_file' not found. Skipping signing."
        return
    fi

    log_info "Signing installer '$installer_file'..."
    azuresigntool sign \
        -kvu "$AZURE_KEY_VAULT_URI" \
        -kvc "$AZURE_CERT_NAME" \
        -kvi "$AZURE_CLIENT_ID" \
        -kvs "$AZURE_CLIENT_SECRET" \
        -kvt "$AZURE_TENANT_ID" \
        -tr http://timestamp.globalsign.com/tsa/advanced \
        -td sha256 \
        "$installer_file"

    log_info "Installer signed successfully."
}



create_github_release() {
    log_step "Creating GitHub Release"

    if [ -z "$GITHUB_TOKEN" ]; then
        log_info "GITHUB_TOKEN not set. Skipping release."
        return
    fi

    if [[ ! "$VERSION_TAG" == v* ]]; then
        log_info "Version tag '$VERSION_TAG' does not start with 'v'. Skipping release."
        return
    fi

    log_info "Checking for gh-cli..."
    if ! command -v "gh" &> /dev/null; then
        echo "ERROR: GitHub CLI ('gh') not found. Please install it to create releases."
        return
    fi

    local installer_file="$ARTIFACTS_DIR/nexa-cli_windows_${TARGET_ARCH}.exe"
    log_info "Uploading '$installer_file' to release '$VERSION_TAG'..."

    # Check if release already exists, if so, upload file to it. Otherwise, create it.
    gh release view "$VERSION_TAG" >/dev/null 2>&1 || gh release create "$VERSION_TAG" --title "$VERSION_TAG" --notes "Release for version $VERSION_TAG"

    gh release upload "$VERSION_TAG" "$installer_file" --clobber

    log_info "GitHub release updated successfully."
}


# --- Main Execution Logic ---

main() {
    if [ -z "$TARGET_ARCH" ]; then
        echo "Usage: $0 <arch> [version]"
        echo "  <arch>: x86_64 or arm64"
        exit 1
    fi

    if [ -z "$VERSION_TAG" ]; then
        VERSION_TAG=$(git rev-parse --short HEAD)
        log_info "Version not provided, using short commit hash: $VERSION_TAG"
    fi

    # check_deps
    # setup_environment
    build_cli
    build_installer_package
    sign_artifacts
    build_installer
    sign_installer
    # create_github_release

    log_step "Build process finished successfully!"
    log_info "Final artifacts are in the '$ARTIFACTS_DIR' directory."
}

# Run the main function with all arguments passed to the script
main "$@"