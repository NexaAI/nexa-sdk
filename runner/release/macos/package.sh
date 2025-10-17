#!/bin/bash

set -euo pipefail

readonly VERSION="$1" ARCH="$2"
readonly APP_PATH="artifacts/Applications/NexaCLI.app"
readonly RESOURCES_PATH="$APP_PATH/Contents/Resources"
readonly KEYCHAIN="build.keychain"

readonly RED='\033[0;31m' GREEN='\033[0;32m' YELLOW='\033[1;33m' NC='\033[0m'
log() { echo -e "${GREEN}[INFO]${NC} $*" >&2; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
die() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

setup_bundle() {
    log "Setting up application bundle..."
    mkdir -p "$RESOURCES_PATH"
    cp -r release/macos/Applications artifacts/
    cp -r build/* "$RESOURCES_PATH"
    sed -i '' "s/\${VERSION}/$VERSION/g" "$APP_PATH/Contents/Info.plist"
}

fix_libs() {
    log "Fixing library paths..."
    [[ -f "$RESOURCES_PATH/nexa-cli" ]] && install_name_tool -add_rpath "@loader_path" "$RESOURCES_PATH/nexa-cli"
}

set_perms() {
    log "Setting permissions..."
    chmod +x "$APP_PATH/Contents/MacOS/launcher" "$RESOURCES_PATH/nexa" "$RESOURCES_PATH/nexa-cli" 2>/dev/null || true
    [[ -d "$RESOURCES_PATH/metal/python_runtime/bin" ]] && chmod -R +x "$RESOURCES_PATH/metal/python_runtime/bin"
    log "Permissions set successfully"
}

import_certs() {
    log "Setting up signing environment..."
    security delete-keychain "$KEYCHAIN" 2>/dev/null || true
    security create-keychain -p "" "$KEYCHAIN"
    security default-keychain -s "$KEYCHAIN"
    security unlock-keychain -p "" "$KEYCHAIN"

    log "Importing certificates..."
    [[ -z "${APP_CERTIFICATE_BASE64:-}" ]] && die "APP_CERTIFICATE_BASE64 not set"
    [[ -z "${APP_CERTIFICATE_PASSWORD:-}" ]] && die "APP_CERTIFICATE_PASSWORD not set"
    echo "$APP_CERTIFICATE_BASE64" | base64 --decode > "codesign_cert.p12"
    security import "codesign_cert.p12" -k "$KEYCHAIN" -P "$APP_CERTIFICATE_PASSWORD" -T "/usr/bin/codesign"
    rm "codesign_cert.p12"

    [[ -z "${INSTALLER_CERTIFICATE_BASE64:-}" ]] && die "INSTALLER_CERTIFICATE_BASE64 not set"
    [[ -z "${INSTALLER_CERTIFICATE_PASSWORD:-}" ]] && die "INSTALLER_CERTIFICATE_PASSWORD not set"
    echo "$INSTALLER_CERTIFICATE_BASE64" | base64 --decode > "productsign_cert.p12"
    security import "productsign_cert.p12" -k "$KEYCHAIN" -P "$INSTALLER_CERTIFICATE_PASSWORD" -T "/usr/bin/productsign"
    rm "productsign_cert.p12"

    security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "" "$KEYCHAIN"
    security list-keychains -s "$KEYCHAIN"
}

sign_app() {
    log "Signing application..."
    [[ -z "${APP_SIGNING_IDENTITY:-}" ]] && die "APP_SIGNING_IDENTITY not set"

    local entitlements="release/macos/entitlements.plist"

    local lib_files=() macho_files=()
    while IFS= read -r f; do
        case "$f" in
            *.dylib|*.so) lib_files+=("$f") ;;
            *)
                if file "$f" | grep -q "Mach-O"; then
                    macho_files+=("$f")
                fi
                ;;
        esac
    done < <(find "$RESOURCES_PATH" -type f)

    for f in "${lib_files[@]}"; do
        (log "Signing dependency: $f"; codesign -s "$APP_SIGNING_IDENTITY" --force --options runtime --timestamp "$f") &
    done

    for f in "${macho_files[@]}"; do
        (log "Signing executable: $f"; codesign -s "$APP_SIGNING_IDENTITY" --force --options runtime --timestamp --entitlements "$entitlements" "$f") &
    done

    wait

    codesign -s "$APP_SIGNING_IDENTITY" --force --options runtime --timestamp --entitlements "$entitlements" "$APP_PATH/Contents/MacOS/launcher"
    codesign -s "$APP_SIGNING_IDENTITY" --force --options runtime --timestamp --entitlements "$entitlements" "$APP_PATH"
    codesign --verify --deep --strict --verbose=4 "$APP_PATH"
    log "App signing complete."
}

build_pkg() {
    log "Building PKG installer..."
    local pkg="artifacts/nexa-cli_macos_${ARCH}-unsigned.pkg"
    pkgbuild --root artifacts --scripts "release/macos/scripts" \
             --identifier "com.nexaai.nexa-sdk" --version "$VERSION" \
             --install-location / "$pkg" >/dev/null
    echo "$pkg"
}

sign_pkg() {
    local unsigned="$1" signed="artifacts/nexa-cli_macos_${ARCH}.pkg"
    log "Signing PKG installer..."
    [[ -z "${INSTALLER_SIGNING_IDENTITY:-}" ]] && die "INSTALLER_SIGNING_IDENTITY not set"
    productsign --sign "$INSTALLER_SIGNING_IDENTITY" "$unsigned" "$signed" >/dev/null || die "PKG signing failed"
    pkgutil --check-signature "$signed" >/dev/null || die "PKG signature verification failed"
    rm "$unsigned"
    echo "$signed"
}


notarize() {
    local pkg="$1"
    log "Submitting PKG for notarization..."
    [[ -z "${APPLE_ID:-}" ]] && die "APPLE_ID not set"
    [[ -z "${APPLE_PASSWORD:-}" ]] && die "APPLE_PASSWORD not set"
    [[ -z "${TEAM_ID:-}" ]] && die "TEAM_ID not set"

    local output
    output=$(xcrun notarytool submit "$pkg" --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$TEAM_ID" --wait)

    local submission_id
    submission_id=$(echo "$output" | grep -oE 'id: [0-9a-f-]+' | head -n 1 | awk '{print $2}')
    [[ -z "$submission_id" ]] && die "Failed to extract submission ID"

    local submission_info
    submission_info=$(xcrun notarytool info "$submission_id" --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$TEAM_ID")

    local status
    status=$(echo "$submission_info" | grep "status:" | awk '{print $2}')
    log "Notarization status: $status"
    [[ "$status" != "Accepted" ]] && {
        xcrun notarytool log "$submission_id" --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$TEAM_ID"
        die "Notarization failed"
    }

    log "Stapling PKG..."
    xcrun stapler staple "$pkg"
    log "Notarization completed successfully"
}

cleanup() {
    [[ -n "${APP_CERTIFICATE_BASE64:-}" ]] && {
        log "Cleaning up keychain..."
        security delete-keychain "$KEYCHAIN" 2>/dev/null || true
    }
}

main() {
    [[ -z "$VERSION" ]] && die "Usage: $0 <version> <arch>"
    log "Creating macOS installer package for version $VERSION, arch $ARCH"

    setup_bundle
    fix_libs
    set_perms

    import_certs
    sign_app

    local pkg_file
    pkg_file=$(build_pkg)
    pkg_file=$(sign_pkg "$pkg_file")
    notarize "$pkg_file"
    log "Package created successfully: $pkg_file"
}

trap 'cleanup || true' EXIT
main "$@"