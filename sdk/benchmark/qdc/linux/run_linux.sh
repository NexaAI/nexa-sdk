#!/bin/bash
# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause
#
# geniex-bench entry script for QDC Linux IoT (BASH framework).
#
# QDC extracts the artifact zip to /data/local/tmp/TestContent/ and runs this
# via /bin/bash. Anything under /data/local/tmp/QDC_logs/ is auto-uploaded.
# run_qdc_jobs.py substitutes:
#   {MODELS}  → `slug|model_id|csv_compute|spec_type|draft_id|draft_tokens` lines
#   {CHIPSET} → AI Hub chipset slug (e.g. qualcomm-qcs9075)
#   {MODE}    → `bench` or `accuracy`
#   {THINK}   → `--think` or `--no-think` (accuracy only)
# Each cell's model_id is resolved on the device by the model-manager C API
# (multi-connection HTTPS, byte-range resume) on first reference; the cached
# copy is reused across the ctx sweep. The plugin column is left empty:
# geniex-bench takes the plugin the manager recorded for the model.
#
# Two modes share the model plan:
#   bench    — ctx sweep {512, 1024, 4096} for speed; prefill from random ids
#              (`-p N`, like llama-bench `pp{N}`).
#   accuracy — one pass over the same chat-templated prompt file for every cell;
#              the `[gen ]` lines in this log are the result, not the timings.

set +e
umask 022

LOG=/data/local/tmp/QDC_logs
OUT=$LOG/results
MM_CACHE=/data/local/tmp/geniex-cache
TC=/data/local/tmp/TestContent
BUNDLE=$TC/pkg-geniex
PROMPTS=$TC/prompts

mkdir -p "$LOG" "$OUT" "$MM_CACHE"
# QDC reuses the same physical host across jobs, so $OUT can hold stale cell
# JSON files from earlier sessions. Wipe them so log-upload can't ship them
# back and pollute this job's cell set.
rm -f "$OUT"/*.json 2>/dev/null || true
exec > "$LOG/script.log" 2>&1
date -u
uname -a

cd "$BUNDLE" || { echo "FATAL: missing $BUNDLE"; exit 1; }
chmod +x bin/* 2>/dev/null
export LD_LIBRARY_PATH="$BUNDLE/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
export GENIEX_PLUGIN_PATH="$BUNDLE/lib"

IMG=$TC/test.png

# Sweep dimensions come from workflow inputs (with defaults filled host-side).
# CTX / PP / TG are parallel arrays of equal length.
IFS=',' read -ra CTX_ARR <<< "{CTX_LIST}"
IFS=',' read -ra PP_ARR  <<< "{PP_LIST}"
IFS=',' read -ra TG_ARR  <<< "{TG_LIST}"

MODE="{MODE}"
THINK="{THINK}" # --think / --no-think, accuracy only
ACC_TSV=/data/local/tmp/matrix-accuracy.tsv

# ------------------------------- plan ---------------------------------------
: > "$ACC_TSV"
for ctx in "${CTX_ARR[@]}"; do
  : > "/data/local/tmp/matrix-${ctx}.tsv"
done

# Columns 5/6 (tokenizer/mmproj) blank: the model manager fills both. The image
# goes on every row; geniex-bench only feeds it to VLM cells.
emit_row() { # $1 = cell ctx, $2 = target tsv
  printf '%s-%s-c%s\t\t%s\t%s\t\t\t%s\t\n' \
    "$slug" "$d" "$1" "$d" "$model_id" "$IMG" >> "$2"
}

while IFS='|' read -r slug model_id computes _spec_type _draft_id _draft_tokens; do
  [ -z "$slug" ] && continue
  echo "=== plan $slug id=$model_id ==="
  IFS=','
  for d in $computes; do
    if [ "$MODE" = accuracy ]; then
      emit_row "${CTX_ARR[0]}" "$ACC_TSV"
    else
      for ctx in "${CTX_ARR[@]}"; do
        emit_row "$ctx" "/data/local/tmp/matrix-${ctx}.tsv"
      done
    fi
  done
  IFS='|'
done <<'EOF'
{MODELS}
EOF

# ------------------------------- run ----------------------------------------
if [ "$MODE" = accuracy ]; then
  # --accuracy pins --warmup 0 -r 1.
  ctx="${CTX_ARR[0]}"
  tg="${TG_ARR[0]}"
  echo "=== matrix accuracy ctx=$ctx tg=$tg $THINK (prompt-file) ==="
  cat "$ACC_TSV"
  ./bin/geniex-bench --matrix-file "$ACC_TSV" --output-json-dir "$OUT" \
    --accuracy --prompt-file "$PROMPTS/accuracy_prompts.txt" \
    -c "$ctx" -n "$tg" "$THINK" \
    --mm-data-dir "$MM_CACHE" --chipset "{CHIPSET}"
  echo "rc=$?"
else
  for i in "${!CTX_ARR[@]}"; do
    ctx="${CTX_ARR[$i]}"
    pp="${PP_ARR[$i]}"
    tg="${TG_ARR[$i]}"
    tsv="/data/local/tmp/matrix-${ctx}.tsv"
    [ -s "$tsv" ] || continue

    echo "=== matrix ctx=$ctx pp=$pp tg=$tg ==="
    cat "$tsv"
    ./bin/geniex-bench --matrix-file "$tsv" --output-json-dir "$OUT" -r 3 \
      -c "$ctx" -p "$pp" -n "$tg" \
      --mm-data-dir "$MM_CACHE" --chipset "{CHIPSET}"
    echo "rc=$?  ($(ls "$OUT" | wc -l) cell json files so far)"
  done
fi
echo "=== done ==="
exit 0
