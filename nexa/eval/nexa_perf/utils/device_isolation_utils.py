import os
import signal
import sys
import time
from logging import getLogger
from typing import Set

from .import_utils import is_amdsmi_available, is_psutil_available, is_pynvml_available
from .logging_utils import setup_logging
from .system_utils import is_nvidia_system, is_rocm_system

if is_psutil_available():
    import psutil

if is_pynvml_available():
    import pynvml

if is_amdsmi_available():
    import amdsmi  # type: ignore


LOGGER = getLogger("device-isolation")


class DeviceIsolationError(Exception):
    pass


def isolation_error_signal_handler(signum, frame):
    raise DeviceIsolationError("Received an error signal from the device isolation process")


if sys.platform == "linux":
    signal.signal(signal.SIGUSR1, isolation_error_signal_handler)


def get_nvidia_devices_pids(device_ids: str) -> Set[int]:
    if not is_pynvml_available():
        raise ValueError(
            "The library pynvml is required to get the pids running on NVIDIA GPUs, but is not installed. "
            "Please install the official and NVIDIA maintained PyNVML library through `pip install nvidia-ml-py`."
        )

    pynvml.nvmlInit()

    devices_pids = set()
    devices_ids = list(map(int, device_ids.split(",")))

    for device_id in devices_ids:
        device_handle = pynvml.nvmlDeviceGetHandleByIndex(device_id)
        device_processes = pynvml.nvmlDeviceGetComputeRunningProcesses(device_handle)
        for device_process in device_processes:
            devices_pids.add(device_process.pid)

    pynvml.nvmlShutdown()

    return devices_pids


def get_amd_devices_pids(device_ids: str) -> Set[int]:
    if not is_amdsmi_available():
        raise ValueError(
            "The library amdsmi is required to get the pids running on AMD GPUs, but is not installed. "
            "Please install the official and AMD maintained amdsmi library from https://github.com/ROCm/amdsmi."
        )

    amdsmi.amdsmi_init()
    permission_denied = False

    devices_pids = set()
    devices_ids = list(map(int, device_ids.split(",")))

    processor_handles = amdsmi.amdsmi_get_processor_handles()
    for device_id in devices_ids:
        processor_handle = processor_handles[device_id]

        if permission_denied:
            continue

        try:
            # these functions fail a lot for no apparent reason
            processes_handles = amdsmi.amdsmi_get_gpu_process_list(processor_handle)
        except Exception as e:
            permission_denied = "Permission denied" in str(e)
            continue

        for process_handle in processes_handles:
            try:
                # these functions fail a lot for no apparent reason
                info = amdsmi.amdsmi_get_gpu_process_info(processor_handle, process_handle)
            except Exception as e:
                permission_denied = "Permission denied" in str(e)
                continue

            if info["memory_usage"]["vram_mem"] == 4096:
                # not sure why these processes are always present
                continue

            devices_pids.add(info["pid"])

    amdsmi.amdsmi_shut_down()

    return devices_pids


def get_pids_running_on_system_devices(device_ids: str) -> Set[int]:
    """Returns the set of pids running on the system device(s)."""
    if is_nvidia_system():
        devices_pids = get_nvidia_devices_pids(device_ids)
    elif is_rocm_system():
        devices_pids = get_amd_devices_pids(device_ids)
    else:
        raise ValueError("get_pids_running_on_system_device is only supported on NVIDIA and AMD GPUs")

    return devices_pids


def get_children_pids(pid: int) -> Set[int]:
    """Returns the set of pids of the children of the given process."""
    if not is_psutil_available():
        raise ValueError(
            "The library psutil is required to get the children pids of a process, but is not installed. "
            "Please install the official and cross-platform psutil library through `pip install psutil`."
        )

    if not psutil.pid_exists(pid):
        LOGGER.warn(f"Process with pid [{pid}] does not exist.")
        return set()

    process = psutil.Process(pid)
    children = process.children(recursive=True)
    children_pids = {child.pid for child in children}

    return children_pids


def assert_device_isolation(pid: int, device_ids: str, action: str):
    log_level = os.environ.get("LOG_LEVEL", "INFO")
    log_to_file = os.environ.get("LOG_TO_FILE", "1") == "1"
    setup_logging(log_level, to_file=log_to_file, prefix="DEVICE-ISOLATION-PROCESS")

    device_isolation_pid = os.getpid()
    permitted_parent_pids = {pid, device_isolation_pid}

    while any(psutil.pid_exists(p) for p in permitted_parent_pids):
        device_pids = get_pids_running_on_system_devices(device_ids=device_ids)
        device_pids = {p for p in device_pids if psutil.pid_exists(p)}

        permitted_children_pids = set()
        for pid in permitted_parent_pids:
            permitted_children_pids |= get_children_pids(pid)

        permitted_pids = permitted_parent_pids | permitted_children_pids
        permitted_pids = {p for p in permitted_pids if psutil.pid_exists(p)}

        non_permitted_pids = device_pids - permitted_pids

        if len(non_permitted_pids) > 0:
            LOGGER.warn(
                f"Found process(es) [{non_permitted_pids}] running on device(s) [{device_ids}], "
                f"other than the isolated process [{pid}], the device isolation process [{device_isolation_pid}] "
                f"and their children [{permitted_children_pids}]."
            )

            if action == "warn":
                LOGGER.warn("Make sure no other process is running on the device(s) while benchmarking.")
            elif action == "error":
                LOGGER.error("Signaling the isolated process to error out.")
                if sys.platform == "linux":
                    os.kill(pid, signal.SIGUSR1)
                else:
                    LOGGER.error("Sending an error signal is only supported on Linux. Killing the isolated process.")
                    os.kill(pid, signal.SIGKILL)
            elif action == "kill":
                LOGGER.error("Killing the isolated process.")
                os.kill(pid, signal.SIGKILL)

            LOGGER.warn("Exiting device isolation process.")
            exit(0)

        time.sleep(1)
