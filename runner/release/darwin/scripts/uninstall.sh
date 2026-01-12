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

echo "Uninstalling Nexa SDK..."

if [ "$(id -u)" != "0" ]; then
    echo "This script requires superuser privileges. Please run with sudo."
    exit 1
fi

# Force kill if still running
pkill -9 -f "nexa" 2>/dev/null || true
pkill -9 -f "nexa-cli" 2>/dev/null || true

echo "Removing symbolic links..."
rm -f /usr/local/bin/nexa

echo "Removing launcher applications..."
rm -rf "/Applications/Nexa CLI.app"
rm -rf "/Applications/NexaCLI.app"

pkgutil --forget com.nexaai.nexa-sdk > /dev/null 2>&1 || true

echo "Nexa SDK has been successfully uninstalled."

exit 0