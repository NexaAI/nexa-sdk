#!/bin/bash
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

echo "--- Preparing file structure for version ${VERSION} ---"

APP_PATH="staging/Applications/NexaCLI.app"

echo "Creating directories..."
mkdir -p "${APP_PATH}/Contents/MacOS"
mkdir -p "${APP_PATH}/Contents/Resources"

echo "Moving build artifacts..."
mv artifacts/* "${APP_PATH}/Contents/Resources/"




echo "Setting permissions..."
chmod +x "${APP_PATH}/Contents/MacOS/launcher"
chmod +x "${APP_PATH}/Contents/Resources/nexa"
chmod +x "${APP_PATH}/Contents/Resources/nexa-cli"
if [ -d "${APP_PATH}/Contents/Resources/nexa_mlx/python_runtime/bin" ]; then
  chmod -R +x "${APP_PATH}/Contents/Resources/nexa_mlx/python_runtime/bin"
fi

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
