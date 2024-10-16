import os
import platform
import re
import subprocess
from typing import List, Optional

import psutil

from .import_utils import is_amdsmi_available, is_pynvml_available, is_pyrsmi_available


# Network related stuff
def get_socket_ifname() -> Optional[str]:
    for interface in psutil.net_if_addrs():
        if interface.startswith("e"):
            return interface

    raise None


## CPU related stuff
def get_cpu() -> Optional[str]:
    if platform.system() == "Windows":
        return platform.processor()

    elif platform.system() == "Darwin":
        command = "sysctl -n machdep.cpu.brand_string"
        return str(subprocess.check_output(command, shell=True).decode().strip())

    elif platform.system() == "Linux":
        command = "cat /proc/cpuinfo"
        all_info = subprocess.check_output(command, shell=True).decode().strip()
        for line in all_info.split("\n"):
            if "model name" in line:
                return re.sub(".*model name.*:", "", line, 1)
        return "Could not find device name"

    else:
        raise ValueError(f"Unknown system '{platform.system()}'")


def get_cpu_ram_mb():
    return psutil.virtual_memory().total / 1e6


## GPU related stuff
try:
    subprocess.check_output("nvidia-smi")
    _nvidia_system = True
except Exception:
    _nvidia_system = False

try:
    subprocess.check_output("rocm-smi")
    _rocm_system = True
except Exception:
    _rocm_system = False


def is_nvidia_system():
    return _nvidia_system


def is_rocm_system():
    return _rocm_system


if is_nvidia_system() and is_pynvml_available():
    import pynvml

if is_rocm_system() and is_amdsmi_available():
    import amdsmi  # type: ignore

if is_rocm_system() and is_pyrsmi_available():
    from pyrsmi import rocml


def get_rocm_version():
    for folder in os.listdir("/opt/"):
        if "rocm" in folder and "rocm" != folder:
            return folder.split("-")[-1]
    raise ValueError("Could not find ROCm version.")


def get_gpus():
    if is_nvidia_system():
        if not is_pynvml_available():
            raise ValueError(
                "The library PyNVML is required to get available GPUs, but is not installed. "
                "Please install the official and NVIDIA maintained PyNVML library through `pip install nvidia-ml-py`."
            )

        gpus = []

        pynvml.nvmlInit()
        for i in range(pynvml.nvmlDeviceGetCount()):
            handle = pynvml.nvmlDeviceGetHandleByIndex(i)
            gpu = pynvml.nvmlDeviceGetName(handle)
            # Older pynvml versions may return bytes
            gpu = gpu.decode("utf-8") if isinstance(gpu, bytes) else gpu
            gpus.append(gpu)
        pynvml.nvmlShutdown()

    elif is_rocm_system():
        if not is_amdsmi_available() and not is_pyrsmi_available():
            raise ValueError(
                "Either the library AMD SMI or PyRSMI is required to get available GPUs, but neither is installed. "
                "Please install the official and AMD maintained AMD SMI library from https://github.com/ROCm/amdsmi "
                "or PyRSMI library from https://github.com/ROCm/pyrsmi."
            )

        gpus = []

        if is_amdsmi_available():
            amdsmi.amdsmi_init()
            for processor_handles in amdsmi.amdsmi_get_processor_handles():
                gpus.append(amdsmi.amdsmi_get_gpu_vendor_name(processor_handles))
            amdsmi.amdsmi_shut_down()

        elif is_pyrsmi_available():
            rocml.smi_initialize()
            for i in range(rocml.smi_get_device_count()):
                gpus.append(rocml.smi_get_device_name(i))
            rocml.smi_shutdown()

    else:
        raise ValueError("No NVIDIA or ROCm GPUs found.")

    return gpus


def get_gpu_vram_mb() -> List[int]:
    if is_nvidia_system():
        if not is_pynvml_available():
            raise ValueError(
                "The library PyNVML is required to get GPU VRAM, but is not installed. "
                "Please install the official and NVIDIA maintained PyNVML library through `pip install nvidia-ml-py`."
            )

        pynvml.nvmlInit()
        vrams = [
            pynvml.nvmlDeviceGetMemoryInfo(pynvml.nvmlDeviceGetHandleByIndex(i)).total
            for i in range(pynvml.nvmlDeviceGetCount())
        ]
        pynvml.nvmlShutdown()

    elif is_rocm_system():
        if not is_amdsmi_available() and not is_pyrsmi_available():
            raise ValueError(
                "Either the library AMD SMI or PyRSMI is required to get GPU VRAM, but neither is installed. "
                "Please install the official and AMD maintained AMD SMI library from https://github.com/ROCm/amdsmi "
                "or PyRSMI library from https://github.com/ROCm/pyrsmi."
            )

        if is_amdsmi_available():
            amdsmi.amdsmi_init()
            vrams = [
                amdsmi.amdsmi_get_gpu_memory_total(processor_handles, mem_type=amdsmi.AmdSmiMemoryType.VRAM)
                for processor_handles in amdsmi.amdsmi_get_processor_handles()
            ]
            amdsmi.amdsmi_shut_down()

        elif is_pyrsmi_available():
            rocml.smi_initialize()
            vrams = [rocml.smi_get_device_memory_total(i) for i in range(rocml.smi_get_device_count())]
            rocml.smi_shutdown()

    else:
        raise ValueError("No NVIDIA or ROCm GPUs found.")

    return sum(vrams)


def get_gpu_device_ids() -> str:
    if is_nvidia_system():
        if os.environ.get("NVIDIA_VISIBLE_DEVICES", None) is not None:
            device_ids = os.environ["NVIDIA_VISIBLE_DEVICES"]
        elif os.environ.get("CUDA_VISIBLE_DEVICES", None) is not None:
            device_ids = os.environ["CUDA_VISIBLE_DEVICES"]
        else:
            if not is_pynvml_available():
                raise ValueError(
                    "The library PyNVML is required to get GPU device ids, but is not installed. "
                    "Please install the official and NVIDIA maintained PyNVML library through `pip install nvidia-ml-py`."
                )

            pynvml.nvmlInit()
            device_ids = list(range(pynvml.nvmlDeviceGetCount()))
            device_ids = ",".join(str(i) for i in device_ids)
            pynvml.nvmlShutdown()
    elif is_rocm_system():
        if os.environ.get("ROCR_VISIBLE_DEVICES", None) is not None:
            device_ids = os.environ["ROCR_VISIBLE_DEVICES"]
        elif os.environ.get("HIP_VISIBLE_DEVICES", None) is not None:
            device_ids = os.environ["HIP_VISIBLE_DEVICES"]
        elif os.environ.get("CUDA_VISIBLE_DEVICES", None) is not None:
            device_ids = os.environ["CUDA_VISIBLE_DEVICES"]
        else:
            if not is_amdsmi_available() or not is_pyrsmi_available():
                raise ValueError(
                    "Either the library AMD SMI or PyRSMI is required to get GPU device ids, but neither is installed. "
                    "Please install the official and AMD maintained AMD SMI library from https://github.com/ROCm/amdsmi "
                    "or PyRSMI library from https://github.com/ROCm/pyrsmi."
                )

            if is_pyrsmi_available():
                rocml.smi_initialize()
                device_ids = list(range(rocml.smi_get_device_count()))
                device_ids = ",".join(str(i) for i in device_ids)
                rocml.smi_shutdown()

            elif is_amdsmi_available():
                amdsmi.amdsmi_init()
                device_ids = list(range(len(amdsmi.amdsmi_get_processor_handles())))
                device_ids = ",".join(str(i) for i in device_ids)
                amdsmi.amdsmi_shut_down()

    else:
        raise ValueError("Couldn't infer GPU device ids.")

    return device_ids


## System related stuff
def get_system_info() -> dict:
    system_dict = {
        "cpu": get_cpu(),
        "cpu_count": os.cpu_count(),
        "cpu_ram_mb": get_cpu_ram_mb(),
        "system": platform.system(),
        "machine": platform.machine(),
        "platform": platform.platform(),
        "processor": platform.processor(),
        "python_version": platform.python_version(),
    }

    if is_nvidia_system() or is_rocm_system():
        system_dict["gpu"] = get_gpus()
        system_dict["gpu_count"] = len(get_gpus())
        system_dict["gpu_vram_mb"] = get_gpu_vram_mb()

    return system_dict
