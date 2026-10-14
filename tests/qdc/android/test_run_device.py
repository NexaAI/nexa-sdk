# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

"""Host-side driver: deploy the payload to the phone and run pytest on it."""

from __future__ import annotations

import json
from pathlib import Path

import android_deploy

PAYLOAD = Path(__file__).parents[1] / 'payload'
DEVICE_REPORT_HOST = 'device-report.log'
MAX_RUNS = 10

PYTEST_ARGS = [
    '.',
    '-v',
    '-s',
    '--tb=short',
    '-m',
    'llama_cpp or qairt',
]


def _nodeids(report_log: bytes) -> list[str]:
    ids: dict[str, None] = {}
    for line in report_log.splitlines():
        try:
            rec = json.loads(line)
        except json.JSONDecodeError:
            continue
        if rec.get('$report_type') == 'TestReport' and rec.get('nodeid'):
            ids[rec['nodeid']] = None
    return list(ids)


def _run_once(index: int, deselect: list[str]) -> tuple[int, bytes]:
    name = f'device-report.{index}.log'
    on_phone = f'{android_deploy.DEV_TESTS}/{name}'
    rc = android_deploy.run_pytest([*PYTEST_ARGS, f'--report-log={name}', *[f'--deselect={n}' for n in deselect]])
    android_deploy.adb('pull', on_phone, name, check=False)
    path = Path(name)
    return rc, path.read_bytes() if path.exists() else b''


def test_run_device():
    android_deploy.deploy(PAYLOAD)
    # A native abort kills the device shell and every later case with it, so relaunch
    # pytest without the cases that already started until the whole suite has run.
    merged = b''
    seen: list[str] = []
    rcs: list[int] = []
    for index in range(MAX_RUNS):
        rc, data = _run_once(index, seen)
        merged += data if data.endswith(b'\n') or not data else data + b'\n'
        rcs.append(rc)
        done = _nodeids(merged)
        if rc in (0, 1) or len(done) == len(seen):
            break
        seen = done
    Path(DEVICE_REPORT_HOST).write_bytes(merged)
    assert all(rc == 0 for rc in rcs), f'on-device pytest exited {rcs}'
