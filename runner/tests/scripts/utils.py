import os
import platform
import shutil
import subprocess
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


def start_nexa(args: list[str],
               debug_log: bool = False,
               stdout: Any = subprocess.PIPE,
               stderr: Any = subprocess.PIPE,
               **kwargs: Any) -> subprocess.Popen[str]:
    global nexa_path

    if nexa_path is None:
        nexa_path = _search_nexa()

    env = os.environ.copy()
    env['NEXA_LOG'] = 'trace' if debug_log else ''
    env['NO_COLOR'] = '1'

    return subprocess.Popen([nexa_path, '--test-mode', '--skip-migrate', '--skip-update'] + args,
                            text=True,
                            encoding='utf-8',
                            cwd=Path(__file__).parent.parent,
                            env=env,
                            stdout=stdout,
                            stderr=stderr,
                            **kwargs)


def execute_nexa(args: list[str],
                 debug_log: bool = False,
                 timeout: int | None = None,
                 **kwargs: Any) -> subprocess.CompletedProcess[str]:
    proc = start_nexa(args, debug_log=debug_log, **kwargs)
    stdout, stderr = proc.communicate(timeout=timeout)
    return subprocess.CompletedProcess(proc.args, proc.returncode, stdout, stderr)
