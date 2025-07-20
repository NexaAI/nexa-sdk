#!/bin/bash

echo "Uninstalling Nexa SDK..."

if [ "$(id -u)" != "0" ]; then
    echo "This script requires superuser privileges. Please run with sudo."
    exit 1
fi

echo "Removing symbolic links..."
rm -f /usr/local/bin/nexa

echo "Removing launcher applications..."
rm -rf "/Applications/Nexa CLI.app"

pkgutil --forget com.nexaai.nexa-sdk > /dev/null 2>&1 || true

echo "Nexa SDK has been successfully uninstalled."

exit 0