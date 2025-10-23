import builtins
import os
import platform
import sys
from datetime import datetime
from io import TextIOWrapper
from pathlib import Path

from . import utils

log_dir = Path('bench-' + datetime.now().strftime('%Y%m%d-%H%M%S'))
log_file: TextIOWrapper


def init():
    global log_file

    os.makedirs(log_dir, exist_ok=True)

    log_file = open(os.path.join(log_dir, 'environment'), 'w', encoding='utf-8')
    print(f"========== Environment ===========")
    print(f"OS: {platform.system()}")
    print(f"Arch: {platform.machine()}")
    print(f"Python version: {sys.version}")
    utils.execute_nexa(['version'], stdout=log_file, stderr=log_file)
    utils.execute_nexa(['version'])
    log_file.close()

    log_file = open(log_dir / 'bench.log', 'w', encoding='utf-8')


def print(msg: str):
    global log_file

    ts = datetime.now().strftime('%Y-%m-%d %H:%M:%S.%f')
    data = f'[{ts}] {msg}'
    builtins.print(data)
    log_file.write(data)
    log_file.write('\n')
    log_file.flush()
