from dataclasses import dataclass, field
from typing import Any, Dict, Optional, ClassVar
from logging import getLogger
import os
import random
import numpy as np
from tempfile import TemporaryDirectory
from psutil import cpu_count

from nexa.eval.nexa_perf.utils.import_utils import nexa_sdk_version
from nexa.eval.nexa_perf.utils.system_utils import get_gpu_device_ids, is_nvidia_system, is_rocm_system
from nexa.gguf import NexaTextInference

LOGGER = getLogger("backend")

@dataclass
class NexaConfig:
    name: str = "nexa_backend"
    version: Optional[str] = nexa_sdk_version()
    _target_: str = "nexa.eval.nexa_perf.nexa_backend.NexaBackend"

    task: Optional[str] = None
    library: Optional[str] = "nexa_backend"
    model_type: Optional[str] = "nexa_backend"

    model: Optional[str] = None
    processor: Optional[str] = None

    device: Optional[str] = None
    device_ids: Optional[str] = None

    seed: int = 42
    inter_op_num_threads: Optional[int] = None
    intra_op_num_threads: Optional[int] = None

    model_kwargs: Dict[str, Any] = field(default_factory=dict)
    processor_kwargs: Dict[str, Any] = field(default_factory=dict)

    filename: Optional[str] = None

    def __post_init__(self):
        if self.model is None:
            raise ValueError("`model` must be specified.")

        if self.processor is None:
            self.processor = self.model

        if self.device is None:
            if is_nvidia_system() or is_rocm_system():
                self.device = "cuda"
            else:
                self.device = "cpu"

        if self.device not in ["cuda", "cpu", "mps"]:
            raise ValueError(f"`device` must be either `cuda`, `cpu`, or `mps`, but got {self.device}")

        if self.device == "cuda":
            if self.device_ids is None:
                LOGGER.warning("`device_ids` was not specified, using all available GPUs.")
                self.device_ids = get_gpu_device_ids()
                LOGGER.warning(f"`device_ids` is now set to `{self.device_ids}` based on system configuration.")

            if is_nvidia_system():
                os.environ["CUDA_DEVICE_ORDER"] = "PCI_BUS_ID"
                os.environ["CUDA_VISIBLE_DEVICES"] = self.device_ids
                LOGGER.info(f"CUDA_VISIBLE_DEVICES was set to {os.environ['CUDA_VISIBLE_DEVICES']}.")
            elif is_rocm_system():
                os.environ["ROCR_VISIBLE_DEVICES"] = self.device_ids
                LOGGER.info(f"ROCR_VISIBLE_DEVICES was set to {os.environ['ROCR_VISIBLE_DEVICES']}.")
            else:
                raise RuntimeError("CUDA device is only supported on systems with NVIDIA or ROCm drivers.")
            
        if self.inter_op_num_threads is not None:
            if self.inter_op_num_threads == -1:
                self.inter_op_num_threads = cpu_count()

        if self.intra_op_num_threads is not None:
            if self.intra_op_num_threads == -1:
                self.intra_op_num_threads = cpu_count()


class NexaBackend:
    NAME: ClassVar[str] = "nexa_backend"

    def __init__(self, config: NexaConfig) -> None:
        self.config = config

        self.logger = getLogger(self.NAME)
        self.logger.info(f"Allocating {self.NAME}")

        self.logger.info(f"\t+ Seeding backend with {self.config.seed}")
        self.seed()

        self.logger.info("\t+ Benchmarking a nexa model")
        self.pretrained_processor = None
        self.generation_config = None
        self.pretrained_config = None
        self.automodel_loader = None
        # TODO: need a custom method to extract shapes from gguf

    def seed(self) -> None:
        random.seed(self.config.seed)
        np.random.seed(self.config.seed)

    def prepare_input_shapes(self, input_shapes: Dict[str, Any]) -> Dict[str, Any]:
        if self.config.task == "text-generation":
            if input_shapes["batch_size"] != 1:
                raise ValueError("Batch size must be 1 for text generation")
        return input_shapes

    def prepare_inputs(self, inputs: Dict[str, Any]) -> Dict[str, Any]:
        return {"tokens": inputs["input_ids"].squeeze(0).tolist()}

    def load(self) -> None:
        self.logger.info("\t+ Creating backend temporary directory")
        self.tmpdir = TemporaryDirectory()
        self.logger.info("\t+ Loading pretrained model")
        self.load_model()
        self.tmpdir.cleanup()

    def load_model(self) -> None:
        """
        Load the model from the given model path (normally GGUF, GGML)
        """
        # TODO: add mps (apple metal) support, currently cant benchmark mps device accurately for energy
        if self.config.device == "cuda" or self.config.device == "mps":
            nexa_model = NexaTextInference(model_path=self.config.model, device="gpu", **self.config.model_kwargs)
        elif self.config.device == "cpu":
            nexa_model = NexaTextInference(model_path=self.config.model, device="cpu", **self.config.model_kwargs)
        else:
            raise ValueError(f"Invalid device: {self.config.device}")
        
        self.pretrained_model = nexa_model.model

    def prefill(self, inputs: Dict[str, Any], kwargs: Dict[str, Any]) -> list[int]:
        next(self.pretrained_model.generate(**inputs))

    def generate(self, inputs: Dict[str, Any], kwargs: Dict[str, Any]) -> list[int]:
        generator = self.pretrained_model.generate(**inputs)
        for _ in range(kwargs["max_new_tokens"]):
            next(generator)
