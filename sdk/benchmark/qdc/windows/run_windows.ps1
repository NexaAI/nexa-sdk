# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause
#
# geniex-bench entry script for QDC Windows (POWERSHELL framework).
#
# QDC extracts the artifact zip to C:\Temp\TestContent\ and runs this via
# PowerShell. Anything under C:\Temp\QDC_Logs\ is auto-uploaded. run_qdc_jobs.py
# substitutes:
#   {MODELS}  → `slug|model_id|csv_compute|spec_type|draft_id|draft_tokens` lines
#   {CHIPSET} → AI Hub chipset slug (e.g. qualcomm-snapdragon-x-elite)
#   {MODE}    → `bench` or `accuracy`
#   {THINK}   → `--think` or `--no-think` (accuracy only)
# Each cell's model_id is resolved on the device by the model-manager C API
# (multi-connection HTTPS, byte-range resume) on first reference; the cached
# copy is reused across the ctx sweep. The plugin column is left empty:
# geniex-bench takes the plugin the manager recorded for the model.
#
# Two modes share the model plan:
#   bench    — ctx sweep {512, 1024, 4096} for speed, one invocation per ctx,
#              prefill from random ids (`-p N`). A spec row (llama_cpp
#              speculative decoding) additionally passes
#              --spec-type/--draft-model/--draft-tokens.
#   accuracy — one pass over the same chat-templated prompt file for every cell,
#              so one TSV covers them all; the `[gen ]` lines geniex-bench
#              prints are the result, not the timings.

$ErrorActionPreference = "Continue"

$LOG = "C:\Temp\QDC_Logs"
$OUT = "$LOG\results"
$MM_CACHE = "C:\Temp\geniex-cache"
$TC = "C:\Temp\TestContent"
$BUNDLE = "$TC\pkg-geniex"
$PROMPTS = "$TC\prompts"

New-Item -ItemType Directory -Force -Path $LOG, $OUT, $MM_CACHE | Out-Null
# QDC reuses the same physical host across jobs, so $OUT can hold stale cell
# JSON files from earlier sessions. Wipe them so log-upload can't ship them
# back and pollute this job's cell set.
Remove-Item -Path "$OUT\*.json" -Force -ErrorAction SilentlyContinue
Start-Transcript -Path "$LOG\script.log" -Force | Out-Null

# Trust the self-signed cert the HTP .cat catalogs are signed with, or the
# Hexagon backends fail their code-integrity check at load.
$cert = "$TC\ggml-htp-v1.cer"
if (Test-Path $cert) {
    & certutil.exe -addstore -f Root $cert | Out-Null
    & certutil.exe -addstore -f TrustedPublisher $cert | Out-Null
}

Set-Location $BUNDLE
$env:GENIEX_PLUGIN_PATH = "$BUNDLE\lib"
$env:PATH = "$BUNDLE\lib;$BUNDLE\lib\llama_cpp;$BUNDLE\lib\qairt;$BUNDLE\lib\qairt\htp-files;$env:PATH"

$rows = @'
{MODELS}
'@ -split "`n" | ForEach-Object { $_.Trim() } | Where-Object { $_ }

$IMG = "$TC/test.png" -replace '\\', '/'
$MODE = "{MODE}"
$THINK = "{THINK}"  # accuracy only
$ACC_TSV = "C:\Temp\matrix-accuracy.tsv"
# Same host is reused across jobs: a stale accuracy.log would be uploaded as if
# this job had generated it.
Remove-Item $ACC_TSV, "$LOG\accuracy.log" -ErrorAction SilentlyContinue

# Sweep dimensions come from workflow inputs (with defaults filled host-side).
# ctxList / ppList / tgList are parallel arrays of equal length.
$ctxList = @({CTX_LIST})
$ppList  = @({PP_LIST})
$tgList  = @({TG_LIST})
$tsvByCtx = @{}
foreach ($ctx in $ctxList) {
    $tsvByCtx["$ctx"] = "C:\Temp\matrix-$ctx.tsv"
    Remove-Item $tsvByCtx["$ctx"] -ErrorAction SilentlyContinue
}
# Spec CLI params are per-invocation, not per-cell; a job carries one entry.
$spec = $null

foreach ($row in $rows) {
    $slug, $model_id, $computes, $spec_type, $draft_model_id, $draft_tokens = $row -split '\|'
    Write-Output "=== plan $slug id=$model_id ==="
    if ($spec_type) {
        $spec = @{ type = $spec_type; draft = $draft_model_id; tokens = $draft_tokens }
    }
    foreach ($d in $computes -split ',') {
        # accuracy collects every cell in one TSV at a single ctx; bench spreads
        # them over per-ctx files.
        $planCtx = if ($MODE -eq "accuracy") { @($ctxList[0]) } else { $ctxList }
        foreach ($ctx in $planCtx) {
            # Columns 5/6 (tokenizer/mmproj) intentionally blank: the model
            # manager fills both from the resolved manifest. The image goes on
            # every row; geniex-bench only feeds it to VLM cells.
            $tsv = if ($MODE -eq "accuracy") { $ACC_TSV } else { $tsvByCtx["$ctx"] }
            "{0}-{1}-c{2}`t`t{1}`t{3}`t`t`t{4}`t" -f $slug, $d, $ctx, $model_id, $IMG `
                | Add-Content $tsv
        }
    }
}

if ($MODE -eq "accuracy") {
    # --accuracy pins --warmup 0 -r 1.
    $ctx = $ctxList[0]
    $tg = $tgList[0]
    Write-Output "=== matrix accuracy ctx=$ctx tg=$tg $THINK (prompt-file) ==="
    Get-Content $ACC_TSV
    # Start-Transcript does not capture a native command's stdout, so the
    # generated text is collected into its own log. UTF-8 both ways: the
    # console's ANSI codepage would mojibake non-ASCII output, and the rubric
    # scores mojibake.
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $gen = & "$BUNDLE\bin\geniex-bench.exe" --matrix-file $ACC_TSV --output-json-dir "$OUT" `
        --accuracy --prompt-file "$PROMPTS\accuracy_prompts.txt" `
        -c $ctx -n $tg $THINK `
        --mm-data-dir $MM_CACHE --chipset "{CHIPSET}" 2>&1
    Write-Output "rc=$LASTEXITCODE ($($gen.Count) lines -> accuracy.log)"
    $gen | Out-File -FilePath "$LOG\accuracy.log" -Encoding utf8
}
else {
    for ($i = 0; $i -lt $ctxList.Count; $i++) {
        $ctx = $ctxList[$i]
        $pp  = $ppList[$i]
        $tg  = $tgList[$i]
        $tsv = $tsvByCtx["$ctx"]
        if (-not ((Test-Path $tsv) -and ((Get-Item $tsv).Length -gt 0))) { continue }

        $runTg = $tg
        $reps = @("-r", "3")
        $extra = @()
        if ($spec) {
            # Spec-decoding needs `--draft-tokens` extra KV slots on the last
            # decode step (target + draft), or llama.cpp trips
            # "decode: failed to find a memory slot for batch of size N+1".
            # Bench defaults pp+tg = ctx exactly, so trim tg by that margin.
            $draftHeadroom = if ($spec.tokens) { [int]$spec.tokens + 1 } else { 4 }
            $runTg = [Math]::Max(1, [int]$tg - $draftHeadroom)
            $reps = @("-r", "1", "--no-warmup")
            $extra += @("--spec-type", $spec.type, "--draft-model", $spec.draft)
            if ($spec.tokens) { $extra += @("--draft-tokens", $spec.tokens) }
        }
        Write-Output "=== matrix ctx=$ctx pp=$pp tg=$runTg ==="
        Get-Content $tsv
        & "$BUNDLE\bin\geniex-bench.exe" --matrix-file $tsv --output-json-dir "$OUT" @reps `
            -c $ctx -p $pp -n $runTg @extra `
            --mm-data-dir $MM_CACHE --chipset "{CHIPSET}"
        Write-Output "rc=$LASTEXITCODE  ($((Get-ChildItem $OUT).Count) cell json files so far)"
    }
}
Write-Output "=== done ==="
Stop-Transcript | Out-Null
exit 0
