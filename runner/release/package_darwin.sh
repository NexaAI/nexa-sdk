#! /bin/sh

set -e

TARGET_DIR="artifacts"
TARGET_APP_PATH="${TARGET_DIR}/Applications/NexaCLI.app"
TARGET_RESOURCES_PATH="${TARGET_APP_PATH}/Contents/Resources"

main() {
    local VERSION=$1
    local ARCH=$2
    if [ -z "$VERSION" ]; then
      echo "Usage: $0 <version> <arch>"
      exit 1
    fi

    # copy files to target directory
    mkdir -p ${TARGET_DIR}
    cp -r release/darwin/Applications ${TARGET_DIR}
    cp -r build/* ${TARGET_RESOURCES_PATH}

    # replace version in Info.plist
    sed -i '' "s/\${VERSION}/$VERSION/g" "${TARGET_APP_PATH}/Contents/Info.plist"

    # fix dylib Linkages (RPATH)
    install_name_tool -add_rpath "@loader_path" "${TARGET_RESOURCES_PATH}/nexa-cli"

    # setting permissions
    chmod +x "${TARGET_APP_PATH}/Contents/MacOS/launcher"
    chmod +x "${TAEGET_RESOURCES_PATH/nexa"
    chmod +x "${TAEGET_RESOURCES_PATH/nexa-cli"
    if [ -d "${TAEGET_RESOURCES_PATH/nexa_mlx/python_runtime/bin" ]; then
      chmod -R +x "${TAEGET_RESOURCES_PATH/nexa_mlx/python_runtime/bin"
    fi

    # Import Code Signing Certificates

    # Sign binaries and libraries

    # Build PKG
    pkgbuild --root "${TARGET_DIR}" \
             --scripts "release/darwin/scripts" \
             --identifier "com.nexaai.nexa-sdk" \
             --version "${VERSION}" \
             --install-location / \
             "${TARGET_DIR}/nexa-cli_macos_${ARCH}-unsigned.pkg"

    # Productsign PKG

    # Notarize & Staple PKG

}

main "$@"
