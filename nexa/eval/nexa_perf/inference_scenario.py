from dataclasses import dataclass, field
from logging import getLogger
from typing import Any, Dict, Optional, ClassVar
import time
from contextlib import ExitStack

from nexa.eval.nexa_perf.utils.system_utils import is_rocm_system
from nexa.eval.nexa_perf.nexa_backend import NexaBackend
from nexa.eval.nexa_perf.perf_benchmark import BenchmarkReport
from nexa.eval.nexa_perf.input_generator import InputGenerator
from nexa.eval.nexa_perf.energy_tracker import Efficiency, EnergyTracker
from nexa.eval.nexa_perf.latency_tracker import LatencyTracker, Throughput
from nexa.eval.nexa_perf.memory_tracker import MemoryTracker

LOGGER = getLogger("inference")

INPUT_SHAPES = {"batch_size": 2, "num_choices": 2, "sequence_length": 16}


@dataclass
class InferenceConfig:
    name: str = "inference"
    _target_: str = "nexa.eval.nexa_perf.inference_scenario.InferenceScenario"

    # benchmark options
    iterations: int = field(
        default=10,
        metadata={
            "help": "Minimum number of iterations to run the benchmark. "
            "The number of tracked inferences will be at least this value."
            "Set to 0 to disable this constraint (benchmark will run for `duration` seconds)."
        },
    )
    duration: int = field(
        default=10,
        metadata={
            "help": "Minimum duration of the benchmark in seconds. "
            "The sum of tracked inferences will be at least this value."
            "Set to 0 to disable this constraint (benchmark will run for `iterations` iterations)."
        },
    )
    warmup_runs: int = field(
        default=10,
        metadata={"help": "Number of warmup runs to perform before benchmarking."},
    )

    # input/output config
    input_shapes: Dict[str, Any] = field(
        default_factory=dict,
        metadata={"help": "Input shapes for the model. Missing keys will be filled with default values."},
    )
    new_tokens: Optional[int] = field(
        default=None,
        metadata={"help": "If set, `max_new_tokens` and `min_new_tokens` will be set to this value."},
    )

    # tracking options
    memory: bool = field(default=False, metadata={"help": "Measure max memory usage"})
    latency: bool = field(default=True, metadata={"help": "Measure latencies and throughputs"})
    energy: bool = field(default=False, metadata={"help": "Measure energy usage and efficiency"})

    # methods kwargs
    forward_kwargs: Dict[str, Any] = field(
        default_factory=dict, metadata={"help": "Keyword arguments to pass to the forward method of the backend."}
    )
    generate_kwargs: Dict[str, Any] = field(
        default_factory=dict, metadata={"help": "Keyword arguments to pass to the generate method of the backend."}
    )
    call_kwargs: Dict[str, Any] = field(
        default_factory=dict, metadata={"help": "Keyword arguments to pass to the call method of the backend."}
    )

    def __post_init__(self):
        self.input_shapes = {**INPUT_SHAPES, **self.input_shapes}

        if self.new_tokens is not None:
            LOGGER.warning(
                "`new_tokens` is deprecated. Use `max_new_tokens` and `min_new_tokens` instead. "
                "Setting `max_new_tokens` and `min_new_tokens` to `new_tokens`."
            )
            self.generate_kwargs["max_new_tokens"] = self.new_tokens
            self.generate_kwargs["min_new_tokens"] = self.new_tokens

        if (
            "max_new_tokens" in self.generate_kwargs
            and "min_new_tokens" in self.generate_kwargs
            and self.generate_kwargs["max_new_tokens"] != self.generate_kwargs["min_new_tokens"]
        ):
            raise ValueError(
                "Setting `min_new_tokens` and `max_new_tokens` to different values results in non-deterministic behavior."
            )

        elif "max_new_tokens" in self.generate_kwargs and "min_new_tokens" not in self.generate_kwargs:
            LOGGER.warning(
                "Setting `max_new_tokens` without `min_new_tokens` results in non-deterministic behavior. "
                "Setting `min_new_tokens` to `max_new_tokens`."
            )
            self.generate_kwargs["min_new_tokens"] = self.generate_kwargs["max_new_tokens"]

        elif "min_new_tokens" in self.generate_kwargs and "max_new_tokens" not in self.generate_kwargs:
            LOGGER.warning(
                "Setting `min_new_tokens` without `max_new_tokens` results in non-deterministic behavior. "
                "Setting `max_new_tokens` to `min_new_tokens`."
            )
            self.generate_kwargs["max_new_tokens"] = self.generate_kwargs["min_new_tokens"]

        if self.energy and is_rocm_system():
            raise ValueError("Energy measurement through codecarbon is not yet available on ROCm-powered devices.")


TEXT_GENERATION_DEFAULT_KWARGS = {
    "num_return_sequences": 1,
    "max_new_tokens": 100,
    "min_new_tokens": 100,
    "do_sample": False,
    "use_cache": True,
    "pad_token_id": 0,
    "eos_token_id": 0,
    "num_beams": 1,
}
TEXT_GENERATION_PREFILL_OVERRIDES = {
    "max_new_tokens": 1,
    "min_new_tokens": 1,
}
TEXT_GENERATION_WARMUP_OVERRIDES = {
    "max_new_tokens": 2,
    "min_new_tokens": 2,
}

TEXT_GENERATION_THROUGHPUT_UNIT = "tokens/s"
TEXT_GENERATION_EFFICIENCY_UNIT = "tokens/kWh"


class InferenceScenario:
    NAME: ClassVar[str] = "inference"

    def __init__(self, config: InferenceConfig) -> None:
        self.config = config
        self.logger = getLogger(self.NAME)
        self.logger.info(f"Allocating {self.NAME} scenario")

    def run(self, backend: NexaBackend) -> BenchmarkReport:
        self.logger.info("\t+ Creating input generator")
        self.input_generator = InputGenerator(
            task=backend.config.task, input_shapes=self.config.input_shapes
        )

        self.logger.info("\t+ Generating Text Generation inputs")
        self.inputs = self.input_generator()
        self.logger.info("\t+ Updating Text Generation kwargs with default values")
        self.config.generate_kwargs = {**TEXT_GENERATION_DEFAULT_KWARGS, **self.config.generate_kwargs}
        self.logger.info("\t+ Initializing Text Generation report")
        self.report = BenchmarkReport.from_list(targets=["load", "prefill", "decode", "per_token"])

        self.logger.info("\t+ Preparing input shapes for Inference")
        self.config.input_shapes = backend.prepare_input_shapes(input_shapes=self.config.input_shapes)

        self.run_model_loading_tracking(backend)

        self.logger.info("\t+ Preparing inputs for Inference")
        self.inputs = backend.prepare_inputs(inputs=self.inputs)

        if self.config.latency or self.config.energy:
            # latency and energy are metrics that require some warmup
            if self.config.warmup_runs > 0:
                self.warmup_text_generation(backend)

        if self.config.latency:
            self.run_text_generation_latency_tracking(backend)

        if self.config.memory:
            self.run_text_generation_memory_tracking(backend)


        if self.config.energy:
            self.run_text_generation_energy_tracking(backend)

        return self.report

    # Warmup
    def warmup_text_generation(self, backend: NexaBackend):
        self.logger.info("\t+ Warming up backend for Text Generation")
        _ = backend.generate(self.inputs, self.config.generate_kwargs)
        for _ in range(self.config.warmup_runs):
            _ = backend.generate(self.inputs, {**self.config.generate_kwargs, **TEXT_GENERATION_WARMUP_OVERRIDES})

    # Loading tracking
    def run_model_loading_tracking(self, backend: NexaBackend):
        self.logger.info("\t+ Running model loading tracking")

        if self.config.memory:
            memory_tracker = MemoryTracker(
                backend=backend.config.name, device=backend.config.device, device_ids=backend.config.device_ids
            )
        if self.config.latency:
            latency_tracker = LatencyTracker(backend=backend.config.name, device=backend.config.device)

        with ExitStack() as context_stack:
            if self.config.memory:
                context_stack.enter_context(memory_tracker.track())
            if self.config.latency:
                context_stack.enter_context(latency_tracker.track())

            backend.load()

        if self.config.latency:
            self.report.load.latency = latency_tracker.get_latency()
        if self.config.memory:
            self.report.load.memory = memory_tracker.get_max_memory()

    ## Memory tracking
    def run_text_generation_memory_tracking(self, backend: NexaBackend):
        self.logger.info("\t+ Running Text Generation memory tracking")
        self.memory_tracker = MemoryTracker(
            backend=backend.config.name, device=backend.config.device, device_ids=backend.config.device_ids
        )
        prefill_kwargs = {**self.config.generate_kwargs, **TEXT_GENERATION_PREFILL_OVERRIDES}

        with self.memory_tracker.track():
            _ = backend.prefill(self.inputs, prefill_kwargs)

        self.report.prefill.memory = self.memory_tracker.get_max_memory()

        with self.memory_tracker.track():
            _ = backend.generate(self.inputs, self.config.generate_kwargs)

        self.report.decode.memory = self.memory_tracker.get_max_memory()

    def run_inference_memory_tracking(self, backend: NexaBackend):
        self.logger.info("\t+ Running Inference memory tracking")
        self.memory_tracker = MemoryTracker(
            backend=backend.config.name, device=backend.config.device, device_ids=backend.config.device_ids
        )

        with self.memory_tracker.track():
            _ = backend.forward(self.inputs, self.config.forward_kwargs)

        self.report.forward.memory = self.memory_tracker.get_max_memory()

    ## Latency tracking
    def run_text_generation_latency_tracking(self, backend: NexaBackend):
        self.logger.info("\t+ Running Text Generation latency tracking")
        latency_tracker = LatencyTracker(backend=backend.config.name, device=backend.config.device)
        prefill_kwargs = {**self.config.generate_kwargs, **TEXT_GENERATION_PREFILL_OVERRIDES}

        while latency_tracker.elapsed() < self.config.duration or latency_tracker.count() < self.config.iterations:
            with latency_tracker.track():
                _ = backend.prefill(self.inputs, prefill_kwargs)

        prefill_latency = latency_tracker.get_latency()
        prefill_volume = self.atomic_prefill_volume

        self.report.prefill.latency = prefill_latency
        self.report.prefill.throughput = Throughput.from_latency(
            prefill_latency, prefill_volume, unit=TEXT_GENERATION_THROUGHPUT_UNIT
        )

        latency_tracker.reset()
        while latency_tracker.elapsed() < self.config.duration or latency_tracker.count() < self.config.iterations:
            with latency_tracker.track():
                _ = backend.generate(self.inputs, self.config.generate_kwargs)

        generate_latency = latency_tracker.get_latency()
        decode_latency = generate_latency - prefill_latency
        decode_volume = self.atomic_decode_volume

        self.report.decode.latency = decode_latency
        self.report.decode.throughput = Throughput.from_latency(
            decode_latency, decode_volume, unit=TEXT_GENERATION_THROUGHPUT_UNIT
        )

    ## Energy tracking
    def run_text_generation_energy_tracking(self, backend: NexaBackend):
        self.logger.info("\t+ Running Text Generation energy tracking")
        energy_tracker = EnergyTracker(
            backend=backend.config.name, device=backend.config.device, device_ids=backend.config.device_ids
        )
        prefill_kwargs = {**self.config.generate_kwargs, **TEXT_GENERATION_PREFILL_OVERRIDES}

        count = 0
        elapsed = 0
        start_time = time.perf_counter()

        with energy_tracker.track(file_prefix="prefill"):
            while elapsed < self.config.duration or count < self.config.iterations:
                _ = backend.prefill(self.inputs, prefill_kwargs)
                elapsed = time.perf_counter() - start_time
                count += 1

        prefill_energy = energy_tracker.get_energy() / count
        prefill_volume = self.atomic_prefill_volume

        self.report.prefill.energy = prefill_energy
        self.report.prefill.efficiency = Efficiency.from_energy(
            prefill_energy, prefill_volume, unit=TEXT_GENERATION_EFFICIENCY_UNIT
        )

        count = 0
        elapsed = 0
        start_time = time.perf_counter()

        with energy_tracker.track(file_prefix="generate"):
            while elapsed < self.config.duration or count < self.config.iterations:
                _ = backend.generate(self.inputs, self.config.generate_kwargs)
                elapsed = time.perf_counter() - start_time
                count += 1

        generate_energy = energy_tracker.get_energy() / count
        decode_energy = generate_energy - prefill_energy
        decode_volume = self.atomic_decode_volume

        self.report.decode.energy = decode_energy
        self.report.decode.efficiency = Efficiency.from_energy(
            decode_energy, decode_volume, unit=TEXT_GENERATION_EFFICIENCY_UNIT
        )

    @property
    def atomic_prefill_volume(self) -> int:  # in tokens
        if {"input_ids", "prompt", "prompts"} & set(self.inputs.keys()):
            # text conditioned generation (1 bos token or sequence_length tokens)
            return self.config.input_shapes["batch_size"] * max(self.config.input_shapes["sequence_length"], 1)
        else:
            # image/audio conditioned generation (1 bos token)
            return self.config.input_shapes["batch_size"]

    @property
    def atomic_decode_volume(self) -> int:  # in tokens
        return (
            self.config.input_shapes["batch_size"]
            * self.config.generate_kwargs["num_beams"]  # at each beam stage there are num_beams tokens generated
            * (self.config.generate_kwargs["max_new_tokens"] - 1)  # 1 token is generated during prefill
        )
