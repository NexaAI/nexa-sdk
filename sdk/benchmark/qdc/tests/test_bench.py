# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

"""On-device geniex-bench run for QDC Android phones.

The host (this pytest process) builds the matrix.tsv with model-manager ids
in column 4 and runs geniex-bench on-device; the benchmark resolves each id
via the model-manager C API and downloads to GENIEX_DATADIR on first use,
replacing the host-side urllib + on-device curl that earlier revisions used.
The per-cell JSON is written straight to the device's QDC_logs/results,
which QDC auto-collects — keeping run_qdc_jobs.py's download_cells path
identical to Linux.

There is no entry script to substitute a mode into, so `params.json` carries it
and pytest skips whichever test the mode did not ask for: `test_bench` sweeps ctx
for speed, `test_accuracy` makes one pass per prompt and pushes the generated
text back as a QDC log, since stdout of this process is not collected.
"""

from __future__ import annotations

import json
import os
import subprocess
from pathlib import Path

import pytest
from utils import (
    BUNDLE_PATH,
    HOST_CHIPSET,
    HOST_IMAGE,
    HOST_PARAMS,
    HOST_PROMPTS,
    HOST_ROWS,
    IMAGE_PATH,
    MM_CACHE_PATH,
    PROMPTS_PATH,
    RESULTS_PATH,
    push_bundle_if_needed,
    run_adb_command,
    write_qdc_log,
)

PARAMS = (
    json.loads(Path(HOST_PARAMS).read_text())
    if Path(HOST_PARAMS).exists()
    else {"mode": "bench"}
)
ENV = f"LD_LIBRARY_PATH={BUNDLE_PATH}/lib GENIEX_PLUGIN_PATH={BUNDLE_PATH}/lib"


def stage_device() -> tuple[str, list[str]]:
    """Push the bundle and inputs; return the chipset slug and matrix rows."""
    push_bundle_if_needed()
    run_adb_command(f"mkdir -p {MM_CACHE_PATH} {RESULTS_PATH} {PROMPTS_PATH}")
    subprocess.run(["adb", "push", HOST_IMAGE, IMAGE_PATH], check=True)
    subprocess.run(["adb", "push", f"{HOST_PROMPTS}/.", PROMPTS_PATH], check=True)
    return (
        Path(HOST_CHIPSET).read_text().strip(),
        [r for r in Path(HOST_ROWS).read_text().splitlines() if r.strip()],
    )


def push_tsv(path: str, rows: list[str]) -> None:
    run_adb_command(
        "printf '%s\\n' " + " ".join(f"'{ln}'" for ln in rows) + f" > {path}"
    )


def cell_rows(rows: list[str], ctx: int) -> list[str]:
    """Matrix TSV lines for one ctx. The plugin and columns 5/6
    (tokenizer/mmproj) are blank: geniex-bench takes them from the model
    manager. The image rides on every row; only VLM cells use it."""
    out = []
    for row in rows:
        slug, model_id, computes, *_spec = row.split("|")
        out += [
            f"{slug}-{d}-c{ctx}\t\t{d}\t{model_id}\t\t\t{IMAGE_PATH}\t"
            for d in computes.split(",")
        ]
    return out


@pytest.mark.skipif(PARAMS["mode"] != "bench", reason="accuracy mode requested")
def test_bench():
    chipset, rows = stage_device()
    assert rows, "no model rows produced"

    failures: list[int] = []
    for ctx, pp, tg in zip(PARAMS["ctx"], PARAMS["pp"], PARAMS["tg"]):
        tsv_path = f"/data/local/tmp/matrix-{ctx}.tsv"
        push_tsv(tsv_path, cell_rows(rows, ctx))
        res = run_adb_command(
            f"cd {BUNDLE_PATH} && {ENV} ./bin/geniex-bench "
            f"--matrix-file {tsv_path} --output-json-dir {RESULTS_PATH} -r 3 "
            f"-c {ctx} -p {pp} -n {tg} "
            f"--mm-data-dir {MM_CACHE_PATH} --chipset '{chipset}'",
            check=False,
        )
        if res.returncode != 0:
            failures.append(ctx)
    listing = run_adb_command(f"ls {RESULTS_PATH}", check=False).stdout
    n_json = sum(1 for ln in listing.splitlines() if ln.strip().endswith(".json"))
    assert not failures, f"geniex-bench failed for ctx {failures}"
    assert (
        n_json > 0
    ), f"no cell JSON produced on device (ls {RESULTS_PATH}: {listing!r})"


@pytest.mark.skipif(PARAMS["mode"] != "accuracy", reason="bench mode requested")
def test_accuracy():
    chipset, rows = stage_device()
    ctx, tg = PARAMS["ctx"][0], PARAMS["tg"][0]
    think = "--think" if PARAMS["think"] else "--no-think"

    cells = cell_rows(rows, ctx)
    assert cells, "no model rows produced"

    tsv_path = "/data/local/tmp/matrix-accuracy.tsv"
    push_tsv(tsv_path, cells)
    # --accuracy pins --warmup 0 -r 1. adb stdout is this process's stdout, which
    # QDC does not collect, so the generated text has to be pushed back as a log.
    res = run_adb_command(
        f"cd {BUNDLE_PATH} && {ENV} ./bin/geniex-bench "
        f"--matrix-file {tsv_path} --output-json-dir {RESULTS_PATH} "
        f"--accuracy --prompt-file {PROMPTS_PATH}/accuracy_prompts.txt "
        f"-c {ctx} -n {tg} {think} "
        f"--mm-data-dir {MM_CACHE_PATH} --chipset '{chipset}'",
        check=False,
    )
    write_qdc_log("script.log", res.stdout)
    assert res.returncode == 0, f"geniex-bench --accuracy failed ({res.returncode})"
    assert "[gen ]" in res.stdout, "no generated text in the device output"


if __name__ == "__main__":
    import pytest

    raise SystemExit(
        pytest.main(["-s", "--junitxml=results.xml", os.path.realpath(__file__)])
    )
