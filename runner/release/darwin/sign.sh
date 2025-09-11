#!/bin/bash
set -e

APP_PATH="$1"
SIGNING_IDENTITY="$2"

if [ -z "$APP_PATH" ] || [ -z "$SIGNING_IDENTITY" ]; then
  echo "Usage: $0 <app_path> <signing_identity>"
  exit 1
fi

echo "--- Signing binaries and libraries in ${APP_PATH} ---"

RESOURCES_PATH="${APP_PATH}/Contents/Resources"

echo "Signing dylibs and executables..."
find "$RESOURCES_PATH" -type f \( -name "*.dylib" -o -name "*.so" \) -exec codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" {} \;

if [ -d "$RESOURCES_PATH/nexa_mlx/python_runtime/bin" ]; then
  find "$RESOURCES_PATH/nexa_mlx/python_runtime/bin" -type f -name "python*" -exec codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist {} \;
fi

find "$RESOURCES_PATH" -type f -name "nexa*" -maxdepth 1 -exec codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist {} \;
codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist "${APP_PATH}/Contents/MacOS/launcher"

echo "Signing main app bundle..."
codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist "$APP_PATH"

echo "Verifying signatures..."
codesign --verify --deep --strict --verbose=4 "$APP_PATH"

echo "--- Signing complete ---"