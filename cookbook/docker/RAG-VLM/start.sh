#!/bin/bash

# Start script for AutoNeural Video Inference Demo
# Starts nexa serve in background and Gradio UI in foreground

set -e

echo "Starting AutoNeural Video Inference Demo..."

# Start nexa serve in background
echo "Starting nexa serve..."
nexa serve --host 0.0.0.0:18181 &
NEXA_PID=$!

# Wait for nexa serve to be ready
echo "Waiting for nexa serve to be ready..."
MAX_WAIT=60
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if curl -s http://127.0.0.1:18181/ > /dev/null 2>&1; then
        echo "nexa serve is ready!"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
    echo "Warning: nexa serve may not be ready, but continuing..."
fi

# Function to cleanup on exit
cleanup() {
    echo "Shutting down..."
    kill $NEXA_PID 2>/dev/null || true
    wait $NEXA_PID 2>/dev/null || true
    exit 0
}

trap cleanup SIGTERM SIGINT

# Start Gradio UI in foreground
echo "Starting Gradio UI..."
cd /app
python3 gradio_ui.py

# Cleanup on exit
cleanup

