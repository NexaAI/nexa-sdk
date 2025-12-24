#!/bin/bash
# Copyright 2024-2025 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

echo "--- Preparing file structure for version ${VERSION} ---"

APP_BASE="staging/Applications"
APP_NAME="NexaCLI"
APP_PATH="${APP_BASE}/${APP_NAME}.app"

echo "Creating directories..."
mkdir -p "${APP_PATH}/Contents/MacOS"
mkdir -p "${APP_PATH}/Contents/Resources"

echo "Moving build artifacts..."
mv artifacts/* "${APP_PATH}/Contents/Resources/"

echo "Copying support files..."
cp runner/release/darwin/scripts/uninstall.sh "${APP_PATH}/Contents/Resources/"
chmod +x "${APP_PATH}/Contents/Resources/uninstall.sh"
cp runner/release/darwin/nexa-icon.icns "${APP_PATH}/Contents/Resources/"

echo "Creating Info.plist..."
sed "s/\${VERSION}/${VERSION}/g" runner/release/darwin/Info.plist > "${APP_PATH}/Contents/Info.plist"

echo "Creating launcher script..."
cat << EOF > "${APP_PATH}/Contents/MacOS/launcher"
#!/usr/bin/osascript
tell application "Terminal"
    activate
    do script "nexa"
end tell
EOF

echo "Setting permissions..."
chmod +x "${APP_PATH}/Contents/MacOS/launcher"
chmod +x "${APP_PATH}/Contents/Resources/nexa"
chmod +x "${APP_PATH}/Contents/Resources/nexa-cli"

if [ -d "${APP_PATH}/Contents/Resources/common/python_runtime/bin" ]; then
  chmod -R +x "${APP_PATH}/Contents/Resources/common/python_runtime/bin"
fi
# Remove joblib test data files to avoid notarization warnings
rm -rf "${APP_PATH}/Contents/Resources/common/python_runtime/lib/python3.10/site-packages/joblib/test/data"

echo "Preparing PKG scripts..."
mkdir -p "pkg_scripts"
cp runner/release/darwin/scripts/preinstall pkg_scripts/
chmod +x pkg_scripts/preinstall
cp runner/release/darwin/scripts/postinstall pkg_scripts/
chmod +x pkg_scripts/postinstall

echo "--- File preparation complete ---"

echo "STAGING_DIR=staging" >> $GITHUB_OUTPUT
echo "SCRIPTS_DIR=pkg_scripts" >> $GITHUB_OUTPUT
echo "APP_PATH=${APP_PATH}" >> $GITHUB_ENV
