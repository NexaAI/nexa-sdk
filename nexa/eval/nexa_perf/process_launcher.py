import os
import traceback
import shutil
import sys
import tempfile
from contextlib import ExitStack, contextmanager
from dataclasses import dataclass, field
from logging import getLogger, Logger
from multiprocessing import Pipe, Process, set_executable, get_start_method, set_start_method
from multiprocessing.connection import Connection
from typing import Any, Callable, Dict, List, Optional, ClassVar

import psutil

from nexa.eval.nexa_perf import BenchmarkReport
from nexa.eval.nexa_perf.utils.logging_utils import setup_logging
from nexa.eval.nexa_perf.utils.process_utils import sync_with_child, sync_with_parent
from nexa.eval.nexa_perf.utils.system_utils import is_nvidia_system, is_rocm_system
from nexa.eval.nexa_perf.utils.device_isolation_utils import assert_device_isolation


LOGGER = getLogger("launcher")

@dataclass
class ProcessConfig:
    name: str = "process"
    _target_: str = "nexa.eval.nexa_perf.process_launcher.ProcessLauncher"

    device_isolation: bool = False
    device_isolation_action: Optional[str] = None

    numactl: bool = False
    numactl_kwargs: Dict[str, Any] = field(default_factory=dict)

    start_method: str = "spawn"

    def __post_init__(self):
        if self.device_isolation and not is_nvidia_system() and not is_rocm_system():
            raise ValueError(
                "Device isolation is only supported on NVIDIA and ROCm systems. "
                "Please set `device_isolation` to False or make sure your drivers "
                "are correctly installed by running `nvidia-smi` or `rocm-smi`."
            )

        if self.device_isolation and self.device_isolation_action is None:
            LOGGER.warning(
                "Device isolation is enabled but no action is specified. "
                "Please set `device_isolation_action` to either `error`, `warn`, or `kill`. "
                "Defaulting to `warn`."
            )
            self.device_isolation_action = "warn"

        elif self.device_isolation and self.device_isolation_action not in {"error", "warn", "kill"}:
            raise ValueError(
                f"Unsupported device isolation action {self.device_isolation_action}. "
                "Please set `device_isolation_action` to either `error`, `warn`, or `kill`."
            )

        if self.start_method not in ["spawn", "fork"]:
            raise ValueError(f"start_method must be one of ['spawn', 'fork'], got {self.start_method}")


NUMA_EXECUTABLE_CONTENT = """#!/bin/bash
echo "Running with numactl wrapper"
echo "numactl path: {numactl_path}"
echo "numactl args: {numactl_args}"
echo "python path: {python_path}"
echo "python args: $@"
{numactl_path} {numactl_args} {python_path} "$@"
"""

class ProcessLauncher:
    NAME: ClassVar[str] = "process"

    def __init__(self, config: ProcessConfig):
        self.config = config
        self.logger = getLogger(self.NAME)
        self.logger.info(f"Allocated {self.NAME} launcher")

        if get_start_method(allow_none=True) != self.config.start_method:
            self.logger.info(f"\t+ Setting multiprocessing start method to {self.config.start_method}")
            set_start_method(self.config.start_method, force=True)

    def launch(self, worker: Callable[..., BenchmarkReport], worker_args: List[Any]) -> BenchmarkReport:
        child_connection, parent_connection = Pipe()
        main_process_pid = os.getpid()
        isolated_process = Process(
            target=target, args=(worker, worker_args, child_connection, main_process_pid, self.logger), daemon=False
        )

        with ExitStack() as stack:
            if self.config.numactl:
                stack.enter_context(self.numactl_executable())

            isolated_process.start()

            if isolated_process.is_alive():
                sync_with_child(parent_connection)
            else:
                raise RuntimeError("Could not synchronize with isolated process")

            if self.config.device_isolation:
                stack.enter_context(self.device_isolation(isolated_process.pid))

            if isolated_process.is_alive():
                sync_with_child(parent_connection)
            else:
                raise RuntimeError("Could not synchronize with isolated process")

            isolated_process.join()

        if isolated_process.exitcode != 0:
            raise RuntimeError(f"Isolated process exited with non-zero code {isolated_process.exitcode}")

        if parent_connection.poll():
            response = parent_connection.recv()
        else:
            raise RuntimeError("Received no response from isolated process")

        if "traceback" in response:
            self.logger.error("\t+ Received traceback from isolated process")
            raise ChildProcessError(response["traceback"])
        elif "exception" in response:
            self.logger.error("\t+ Received exception from isolated process")
            raise ChildProcessError(response["exception"])
        elif "report" in response:
            self.logger.info("\t+ Received report from isolated process")
            report = BenchmarkReport.from_dict(response["report"])
        else:
            raise RuntimeError(f"Received an unexpected response from isolated process: {response}")

        return report

    @contextmanager
    def device_isolation(self, pid: int, device_ids: Optional[str] = None):
        if device_ids is None:
            if is_rocm_system():
                device_ids = os.environ.get("ROCR_VISIBLE_DEVICES", None)
            elif is_nvidia_system():
                device_ids = os.environ.get("CUDA_VISIBLE_DEVICES", None)

        self.device_isolation_process = Process(
            target=assert_device_isolation,
            kwargs={"action": self.config.device_isolation_action, "device_ids": device_ids, "pid": pid},
            daemon=True,
        )
        self.device_isolation_process.start()
        self.logger.info(f"\t+ Isolating device(s) [{device_ids}] for process [{pid}] and its children")
        self.logger.info(f"\t+ Executing action [{self.config.device_isolation_action}] in case of violation")

        yield

        self.logger.info("\t+ Stopping device isolation process")
        self.device_isolation_process.terminate()
        self.device_isolation_process.join()
        self.device_isolation_process.close()

    @contextmanager
    def numactl_executable(self):
        self.logger.info("\t+ Warming up multiprocessing context")
        dummy_process = Process(target=dummy_target, daemon=False)
        dummy_process.start()
        dummy_process.join()
        dummy_process.close()

        self.logger.info("\t+ Creating numactl wrapper executable for multiprocessing")
        python_path = sys.executable
        numactl_path = shutil.which("numactl")
        if numactl_path is None:
            raise RuntimeError("Could not find numactl executable. Please install numactl and try again.")
        numactl_args = " ".join([f"--{key}={value}" for key, value in self.config.numactl_kwargs.items()])
        numa_executable = tempfile.NamedTemporaryFile(delete=False, prefix="numa_executable_", suffix=".sh")
        numa_executable_content = NUMA_EXECUTABLE_CONTENT.format(
            numactl_path=numactl_path, numactl_args=numactl_args, python_path=python_path
        )
        numa_executable.write(numa_executable_content.encode())
        os.chmod(numa_executable.name, 0o777)
        numa_executable.close()

        self.logger.info("\t+ Setting multiprocessing executable to numactl wrapper")
        set_executable(numa_executable.name)

        yield

        self.logger.info("\t+ Resetting default multiprocessing executable")
        os.unlink(numa_executable.name)
        set_executable(sys.executable)


def dummy_target() -> None:
    exit(0)

def target(
    worker: Callable[..., BenchmarkReport],
    worker_args: List[Any],
    child_connection: Connection,
    main_process_pid: int,
    logger: Logger,
) -> None:
    main_process = psutil.Process(main_process_pid)

    if main_process.is_running():
        sync_with_parent(child_connection)
    else:
        raise RuntimeError("Could not synchronize with main process")

    log_level = os.environ.get("LOG_LEVEL", "INFO")
    log_to_file = os.environ.get("LOG_TO_FILE", "1") == "1"
    setup_logging(level=log_level, to_file=log_to_file, prefix="ISOLATED-PROCESS")

    if main_process.is_running():
        sync_with_parent(child_connection)
    else:
        raise RuntimeError("Could not synchronize with main process")

    try:
        report = worker(*worker_args)
    except Exception:
        logger.error("\t+ Sending traceback to main process")
        child_connection.send({"traceback": traceback.format_exc()})
    else:
        logger.info("\t+ Sending report to main process")
        child_connection.send({"report": report.to_dict()})
    finally:
        logger.info("\t+ Exiting isolated process")
        child_connection.close()
        exit(0)