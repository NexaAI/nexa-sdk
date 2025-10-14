import json
import os
import shutil
import subprocess
import sys
from io import TextIOWrapper
from pathlib import Path
from typing import Any

nexa_path = None


def _search_nexa() -> str:
    search_dirs = [
        '../build',
        './build',
        './runner/build',
    ]
    for d in search_dirs:
        path = Path(d) / 'nexa'
        if path.exists() and os.access(path, os.X_OK):
            return str(path.resolve())

    global_nexa = shutil.which('nexa')
    if global_nexa is not None:
        return global_nexa

    raise FileNotFoundError("nexa command not found")


def load_param(case: str) -> list[str]:
    path = Path(__file__).parent.parent / 'cases' / case / 'param.txt'
    if not path.exists():
        raise FileNotFoundError(f"{path} not found")
    with open(path, 'r') as f:
        return [l.rstrip() for l in f.readlines()]


def execute_nexa(
        args: list[str],
        **kwargs: Any  # pyright: ignore[reportAny, reportExplicitAny]
) -> subprocess.CompletedProcess[str]:
    global nexa_path

    if nexa_path is None:
        nexa_path = _search_nexa()

    stdout: TextIOWrapper = kwargs.pop('stdout', sys.stdout)  # pyright: ignore[reportAny]
    stderr: TextIOWrapper = kwargs.pop('stderr', sys.stderr)  # pyright: ignore[reportAny]

    res = subprocess.run([nexa_path] + args, capture_output=True, text=True, **kwargs)  # pyright: ignore[reportAny]
    stdout.write(res.stdout)
    compat_out = False
    for line in res.stderr.splitlines():
        try:
            json.loads(line)
            stderr.write(line + '\n')
        except Exception:
            if not compat_out:
                stdout.write('\n====== Compatibility Logs ======\n')
                compat_out = True
            stdout.write(line + '\n')

    return res
