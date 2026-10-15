# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

"""Host-side driver: deploy the payload to the phone and run pytest on it."""

from __future__ import annotations

from pathlib import Path

import android_deploy

PAYLOAD = Path(__file__).parents[1] / 'payload'

PYTEST_ARGS = [
    '.',
    '-v',
    '-s',
    '--tb=short',
    '-m',
    'llama_cpp or qairt',
]

# A DSP abort here drops the phone off adb for the rest of the QDC job, so run it last.
LAST = 'vlm and GGUF and npu'


def _collect(expr: str) -> list[str]:
    out = android_deploy.adb_shell(
        f'cd {android_deploy.DEV_TESTS} && {android_deploy.device_env()} '
        f"{android_deploy.DEV_PY}/bin/python3 -m pytest --collect-only -q -m 'llama_cpp or qairt' -k '{expr}'",
        check=False,
    ).stdout
    return [ln.strip() for ln in out.splitlines() if '.py::' in ln]


def test_run_device():
    android_deploy.deploy(PAYLOAD)
    last = _collect(LAST)
    rc = android_deploy.run_pytest([*PYTEST_ARGS, '-k', f'not ({LAST})'])
    rc_last = android_deploy.run_pytest([*PYTEST_ARGS, '-k', LAST])
    if rc_last != 0:
        # The CI summary parses stdout; the first status seen per case wins, so this only books unfinished cases.
        for nodeid in last:
            print(f'{nodeid} FAILED (device pytest aborted)')
    assert rc == 0 and rc_last == 0, f'on-device pytest exited {rc}, {rc_last}'
