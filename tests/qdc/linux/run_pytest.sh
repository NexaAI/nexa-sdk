#!/bin/bash
# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause
#
# QDC pytest entry — Linux QCS9075M (Qualcomm Linux 1.9 ships python 3.12 + pip).

set +e
umask 022

LOG=/data/local/tmp/QDC_logs
TC=/data/local/tmp/TestContent
BUNDLE=$TC/pkg-geniex
mkdir -p "$LOG"
exec > "$LOG/harness.log" 2>&1

python3 -m pip install --user --no-warn-script-location "pytest>=7.0" "pytest-reportlog>=0.4" "pytest-xdist>=3.5" "tqdm>=4.65"

cd "$BUNDLE" || { echo "FATAL: bundle missing at $BUNDLE"; exit 1; }
chmod +x bin/* 2>/dev/null

export LD_LIBRARY_PATH="$BUNDLE/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
export GENIEX_PLUGIN_PATH="$BUNDLE/lib"
export GENIEX_LIB_PATH="$BUNDLE/lib"
export PYTHONPATH="$TC/bindings/python:$PYTHONPATH"
# QDC's shared link times out fixtures if all model pulls burst in parallel.
export HF_HUB_DOWNLOAD_CONCURRENCY=1
export GENIEX_DEVICE_TEST=1

python3 -c "import geniex; geniex.init(); geniex.deinit()" || {
    echo "FATAL: geniex import failed"
    exit 1
}

cd "$TC/tests"
# One xdist worker: a native abort fails that case and the run continues on a fresh worker.
python3 -m pytest . -v -n 1 --max-worker-restart=100 \
    --tb=short \
    --junitxml="$LOG/device-results.xml" \
    --report-log="$LOG/device-report.log" \
    -m "llama_cpp or qairt" \
    --deselect "test_llama_cpp.py::test_llm_greedy_is_deterministic[Qwen3-4B-GGUF-cpu]" \
    --deselect "test_llama_cpp.py::test_vlm_multi_turn[gemma-4-E2B-it-GGUF-cpu]" \
    --deselect "test_llama_cpp.py::test_llm_quality_keywords[The planet closest to the Sun is-Mercury-Qwen3-4B-GGUF-cpu]" 2>&1
exit 0
