#!/bin/bash
#
# macOS Package Builder
# Creates signed and notarized PKG installer for NexaCLI
#

set -euo pipefail

# Configuration
readonly TARGET_DIR="artifacts"
readonly APP_PATH="${TARGET_DIR}/Applications/NexaCLI.app"
readonly RESOURCES_PATH="${APP_PATH}/Contents/Resources"
readonly ENTITLEMENTS="release/darwin/entitlements.plist"
readonly KEYCHAIN="build.keychain"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${GREEN}[INFO]${NC} $*" >&2; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# Utility functions
check_file() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        log_warn "$file not found"
        return 1
    fi
    return 0
}

check_dir() {
    local dir="$1"
    if [[ ! -d "$dir" ]]; then
        log_warn "$dir not found"
        return 1
    fi
    return 0
}

# Setup functions
setup_app_bundle() {
    log_info "Setting up application bundle..."
    mkdir -p "$TARGET_DIR" "$RESOURCES_PATH"
    cp -r release/darwin/Applications "$TARGET_DIR"
    cp -r build/* "$RESOURCES_PATH"
    
    # Update version in Info.plist
    sed -i '' "s/\${VERSION}/$VERSION/g" "$APP_PATH/Contents/Info.plist"
}

fix_dylib_linkages() {
    log_info "Fixing dylib linkages..."
    check_file "$RESOURCES_PATH/nexa-cli" && \
        install_name_tool -add_rpath "@loader_path" "$RESOURCES_PATH/nexa-cli"
}

set_permissions() {
    log_info "Setting permissions..."
    local files=(
        "$APP_PATH/Contents/MacOS/launcher"
        "$RESOURCES_PATH/nexa"
        "$RESOURCES_PATH/nexa-cli"
    )
    
    for file in "${files[@]}"; do
        check_file "$file" && chmod +x "$file"
    done
    
    if check_dir "$RESOURCES_PATH/nexa_mlx/python_runtime/bin"; then
        chmod -R +x "$RESOURCES_PATH/nexa_mlx/python_runtime/bin"
    fi
}

# Certificate management
import_certificate() {
    local cert_b64="$1" cert_pass="$2" cert_type="$3"
    
    log_info "Importing $cert_type certificate..."
    echo "$cert_b64" | base64 --decode > "${cert_type}_certificate.p12"
    
    if [[ "$cert_type" == "app" ]]; then
        security create-keychain -p "" "$KEYCHAIN"
        security default-keychain -s "$KEYCHAIN"
        security unlock-keychain -p "" "$KEYCHAIN"
        security set-key-partition-list -S apple-tool:,apple: -s -k "" "$KEYCHAIN"
    fi
    
    security import "${cert_type}_certificate.p12" -k "$KEYCHAIN" -P "$cert_pass" -T "/usr/bin/$cert_type"
    rm "${cert_type}_certificate.p12"
}

cleanup_keychain() {
    [[ -n "${APP_CERTIFICATE_BASE64:-}" ]] && {
        log_info "Cleaning up keychain..."
        security delete-keychain "$KEYCHAIN" 2>/dev/null || true
    }
}

# Signing functions
sign_files() {
    local pattern="$1" entitlements="${2:-}"
    local sign_args=(-s "$SIGNING_IDENTITY" --force --options runtime --timestamp --verify)
    [[ -n "$entitlements" ]] && sign_args+=(--entitlements "$entitlements")
    
    find "$RESOURCES_PATH" -type f -name "$pattern" -exec codesign "${sign_args[@]}" {} \;
}

sign_application() {
    log_info "Signing application..."
    
    # Sign libraries
    sign_files "*.dylib"
    sign_files "*.so"
    
    # Sign Python runtime
    if check_dir "$RESOURCES_PATH/nexa_mlx/python_runtime/bin"; then
        sign_files "python*" "$ENTITLEMENTS"
    fi
    
    # Sign main executables
    sign_files "nexa*" "$ENTITLEMENTS"
    codesign -s "$SIGNING_IDENTITY" --force --options runtime --timestamp --verify \
             --entitlements "$ENTITLEMENTS" "$APP_PATH/Contents/MacOS/launcher"
    
    # Sign app bundle
    codesign -s "$SIGNING_IDENTITY" --force --options runtime --timestamp --verify \
             --entitlements "$ENTITLEMENTS" "$APP_PATH"
    
    # Verify signatures
    codesign --verify --deep --strict --verbose=4 "$APP_PATH"
    log_info "Code signing completed successfully"
}

# PKG functions
build_pkg() {
    log_info "Building PKG installer..."
    local unsigned_pkg="$TARGET_DIR/nexa-cli_macos_${ARCH}-unsigned.pkg"
    
    pkgbuild --root "$TARGET_DIR" \
             --scripts "release/darwin/scripts" \
             --identifier "com.nexaai.nexa-sdk" \
             --version "$VERSION" \
             --install-location / \
             "$unsigned_pkg" >/dev/null
    
    echo "$unsigned_pkg"
}

sign_pkg() {
    local unsigned_pkg="$1"
    local signed_pkg="$TARGET_DIR/nexa-cli_macos_${ARCH}.pkg"
    
    log_info "Signing PKG installer..."
    productsign --sign "$SIGNING_IDENTITY" "$unsigned_pkg" "$signed_pkg"
    rm "$unsigned_pkg"
    echo "$signed_pkg"
}

# Notarization functions
notarize_pkg() {
    local pkg_file="$1"
    
    log_info "Submitting PKG for notarization..."
    local output
    output=$(xcrun notarytool submit "$pkg_file" \
        --apple-id "$APPLE_ID" \
        --password "$APPLE_PASSWORD" \
        --team-id "$TEAM_ID" \
        --wait)
    
    echo "$output"
    
    local submission_id
    submission_id=$(echo "$output" | grep -oE 'id: [0-9a-f-]+' | head -n 1 | awk '{print $2}')
    
    [[ -z "$submission_id" ]] && {
        log_error "Failed to extract submission ID"
        exit 1
    }
    
    local status
    status=$(xcrun notarytool info "$submission_id" \
        --apple-id "$APPLE_ID" \
        --password "$APPLE_PASSWORD" \
        --team-id "$TEAM_ID" | grep "status:" | awk '{print $2}')
    
    log_info "Notarization status: $status"
    
    [[ "$status" != "Accepted" ]] && {
        log_error "Notarization failed. Fetching log..."
        xcrun notarytool log "$submission_id" \
            --apple-id "$APPLE_ID" \
            --password "$APPLE_PASSWORD" \
            --team-id "$TEAM_ID"
        exit 1
    }
    
    log_info "Stapling notarization ticket..."
    xcrun stapler staple "$pkg_file"
    log_info "Notarization completed successfully"
}

# Main execution
main() {
    local VERSION="${1:-}"
    local ARCH="${2:-}"
    
    [[ -z "$VERSION" ]] && {
        log_error "Usage: $0 <version> <arch>"
        exit 1
    }
    
    log_info "Creating macOS installer package for version $VERSION, arch $ARCH"
    
    # Setup
    setup_app_bundle
    fix_dylib_linkages
    set_permissions
    
    # Certificate and signing
    if [[ -n "${APP_CERTIFICATE_BASE64:-}" && -n "${APP_CERTIFICATE_PASSWORD:-}" ]]; then
        import_certificate "$APP_CERTIFICATE_BASE64" "$APP_CERTIFICATE_PASSWORD" "app"
    else
        log_warn "App certificate not provided, skipping certificate import"
    fi
    
    if [[ -n "${SIGNING_IDENTITY:-}" ]]; then
        sign_application
    else
        log_warn "SIGNING_IDENTITY not provided, skipping code signing"
    fi
    
    # PKG creation
    local pkg_file
    pkg_file=$(build_pkg)
    
    # PKG signing
    if [[ -n "${INSTALLER_CERTIFICATE_BASE64:-}" && -n "${INSTALLER_CERTIFICATE_PASSWORD:-}" && -n "${SIGNING_IDENTITY:-}" ]]; then
        import_certificate "$INSTALLER_CERTIFICATE_BASE64" "$INSTALLER_CERTIFICATE_PASSWORD" "productsign"
        pkg_file=$(sign_pkg "$pkg_file")
    else
        log_warn "Installer certificate not provided, keeping unsigned PKG"
        mv "$pkg_file" "$TARGET_DIR/nexa-cli_macos_${ARCH}.pkg"
        pkg_file="$TARGET_DIR/nexa-cli_macos_${ARCH}.pkg"
    fi
    
    # Notarization
    if [[ -n "${APPLE_ID:-}" && -n "${APPLE_PASSWORD:-}" && -n "${TEAM_ID:-}" ]]; then
        notarize_pkg "$pkg_file"
    else
        log_warn "Apple ID credentials not provided, skipping notarization"
    fi
    
    # Cleanup
    cleanup_keychain
    
    log_info "Package created successfully: $pkg_file"
}

# Trap for cleanup
trap cleanup_keychain EXIT

main "$@"
