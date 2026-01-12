#!/bin/bash
# Copyright 2024-2026 Nexa AI, Inc.
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

if [ -d "$RESOURCES_PATH/common/python_runtime/bin" ]; then
  find "$RESOURCES_PATH/common/python_runtime/bin" -type f -name "python*" -exec codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist {} \;

fi

if [ -d "$RESOURCES_PATH/common/python_runtime/lib/python3.10/site-packages/torch/bin" ]; then
  find "$RESOURCES_PATH/common/python_runtime/lib/python3.10/site-packages/torch/bin" -type f -exec codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist {} \;
fi

find "$RESOURCES_PATH" -type f -name "nexa*" -maxdepth 1 -exec codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist {} \;
codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist "${APP_PATH}/Contents/MacOS/launcher"

echo "Signing main app bundle..."
codesign --force --options runtime --timestamp --verify -s "$SIGNING_IDENTITY" --entitlements runner/release/darwin/entitlements.plist "$APP_PATH"

echo "Verifying signatures..."
codesign --verify --deep --strict --verbose=4 "$APP_PATH"

echo "--- Signing complete ---"
