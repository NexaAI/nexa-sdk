import json
import os
import platform
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
        exe = 'nexa' if platform.system() != 'Windows' else 'nexa.exe'
        path = Path(d) / exe
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
    with open(path, 'r', encoding='utf-8') as f:
        return [l.rstrip() for l in f.readlines()]


def execute_nexa(
        args: list[str],
        debug_log: bool = False,
        **kwargs: Any  # pyright: ignore[reportAny, reportExplicitAny]
) -> subprocess.CompletedProcess[str]:
    global nexa_path

    if nexa_path is None:
        nexa_path = _search_nexa()

    stdout: TextIOWrapper = kwargs.pop('stdout', sys.stdout)  # pyright: ignore[reportAny]
    stderr: TextIOWrapper = kwargs.pop('stderr', sys.stderr)  # pyright: ignore[reportAny]

    env = os.environ.copy()
    env['NEXA_LOG'] = 'trace' if debug_log else ''
    env['NO_COLOR'] = '1'

    res = subprocess.run([nexa_path, '--test-mode', '--skip-migrate', '--skip-update'] + args,
                         capture_output=True,
                         text=True,
                         encoding='utf-8',
                         cwd=Path(__file__).parent.parent,
                         env=env,
                         **kwargs)  # pyright: ignore[reportAny]

    stdout.write('========== Output Log ==========\n')
    stdout.write(res.stdout)
    compat_out = False
    for line in res.stderr.splitlines():
        try:
            json.loads(line)
            stderr.write(line + '\n')
        except Exception:
            if not compat_out:
                stdout.write('\n========== Debug Log ===========\n')
                compat_out = True
            stdout.write(line + '\n')

    return res
