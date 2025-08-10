#!/bin/bash
set -e

PKG_PATH="$1"
APPLE_ID="$2"
APPLE_PASSWORD="$3"
TEAM_ID="$4"

if [ -z "$PKG_PATH" ] || [ -z "$APPLE_ID" ] || [ -z "$APPLE_PASSWORD" ] || [ -z "$TEAM_ID" ]; then
  echo "Usage: $0 <pkg_path> <apple_id> <app_specific_password> <team_id>"
  exit 1
fi

echo "--- Starting notarization for ${PKG_PATH} ---"

NOTARIZATION_OUTPUT=$(xcrun notarytool submit "$PKG_PATH" --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$TEAM_ID" --wait)
echo "$NOTARIZATION_OUTPUT"

SUBMISSION_ID=$(echo "$NOTARIZATION_OUTPUT" | grep -oE 'id: [0-9a-f-]+' | head -n 1 | awk '{print $2}')

if [ -z "$SUBMISSION_ID" ]; then
  echo "::error::Failed to extract submission ID. Notarization failed."
  exit 1
fi

NOTARIZATION_INFO=$(xcrun notarytool info "$SUBMISSION_ID" --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$TEAM_ID")
STATUS=$(echo "$NOTARIZATION_INFO" | grep "status:" | awk '{print $2}')
echo "Final notarization status: $STATUS"

if [ "$STATUS" != "Accepted" ]; then
  echo "::error::Notarization was not successful. Fetching log..."
  xcrun notarytool log "$SUBMISSION_ID" --apple-id "$APPLE_ID" --password "$APPLE_PASSWORD" --team-id "$TEAM_ID"
  exit 1
fi

echo "Stapling notarization ticket..."
xcrun stapler staple "$PKG_PATH"

echo "--- Notarization and stapling complete ---"