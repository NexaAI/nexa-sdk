from dataclasses import asdict, dataclass, field, make_dataclass
from logging import getLogger
from pathlib import Path
from typing import TYPE_CHECKING, Type, Union, Dict, Any, List, Optional

import pandas as pd
from flatten_dict import flatten
from hydra.utils import get_class
from rich.console import Console
from rich.markdown import Markdown

from nexa.eval.nexa_perf.utils.system_utils import get_system_info
from nexa.eval.nexa_perf.energy_tracker import Efficiency, Energy
from nexa.eval.nexa_perf.latency_tracker import Latency, Throughput
from nexa.eval.nexa_perf.memory_tracker import Memory

if TYPE_CHECKING:
    from nexa.eval.nexa_perf.nexa_backend import NexaBackend
    from nexa.eval.nexa_perf.process_launcher import ProcessLauncher
    from nexa.eval.nexa_perf.inference_scenario import InferenceScenario

LOGGER = getLogger("benchmark")
CONSOLE = Console()

# BenchmarkConfig
@dataclass
class BenchmarkConfig:
    name: str
    backend: Any
    scenario: Any
    launcher: Any
    environment: Dict[str, Any] = field(default_factory=lambda: {**get_system_info()})
    print_report: bool = False
    log_report: bool = True

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "BenchmarkConfig":
        return cls(**data)

    @property
    def default_filename(cls) -> str:
        return "benchmark_config.json"

# TargetMeasurements
@dataclass
class TargetMeasurements:
    memory: Optional[Memory] = None
    latency: Optional[Latency] = None
    throughput: Optional[Throughput] = None
    energy: Optional[Energy] = None
    efficiency: Optional[Efficiency] = None

    def __post_init__(self):
        if self.memory is not None and isinstance(self.memory, dict):
            self.memory = Memory(**self.memory)
        if self.latency is not None and isinstance(self.latency, dict):
            self.latency = Latency(**self.latency)
        if self.throughput is not None and isinstance(self.throughput, dict):
            self.throughput = Throughput(**self.throughput)
        if self.energy is not None and isinstance(self.energy, dict):
            self.energy = Energy(**self.energy)
        if self.efficiency is not None and isinstance(self.efficiency, dict):
            self.efficiency = Efficiency(**self.efficiency)

    @staticmethod
    def aggregate(measurements: List["TargetMeasurements"]) -> "TargetMeasurements":
        assert len(measurements) > 0, "No measurements to aggregate"

        m0 = measurements[0]

        memory = Memory.aggregate([m.memory for m in measurements]) if m0.memory is not None else None
        latency = Latency.aggregate([m.latency for m in measurements]) if m0.latency is not None else None
        throughput = Throughput.aggregate([m.throughput for m in measurements]) if m0.throughput is not None else None
        energy = Energy.aggregate([m.energy for m in measurements]) if m0.energy is not None else None
        efficiency = Efficiency.aggregate([m.efficiency for m in measurements]) if m0.efficiency is not None else None

        return TargetMeasurements(
            memory=memory, latency=latency, throughput=throughput, energy=energy, efficiency=efficiency
        )

    def to_plain_text(self) -> str:
        plain_text = ""

        for key in ["memory", "latency", "throughput", "energy", "efficiency"]:
            measurement = getattr(self, key)
            if measurement is not None:
                plain_text += f"\t+ {key}:\n"
                plain_text += measurement.to_plain_text()

        return plain_text

    def log(self):
        for line in self.to_plain_text().split("\n"):
            if line:
                LOGGER.info(line)

    def to_markdown_text(self) -> str:
        markdown_text = ""

        for key in ["memory", "latency", "throughput", "energy", "efficiency"]:
            measurement = getattr(self, key)
            if measurement is not None:
                markdown_text += f"## {key}:\n\n"
                markdown_text += measurement.to_markdown_text()

        return markdown_text

    def print(self):
        CONSOLE.print(Markdown(self.to_markdown_text()))

# BenchmarkReport
@dataclass
class BenchmarkReport:
    @classmethod
    def from_list(cls, targets: List[str]) -> "BenchmarkReport":
        return cls.from_dict({target: None for target in targets})

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "BenchmarkReport":
        return make_dataclass(cls_name=cls.__name__, fields=data.keys(), bases=(cls,))(**data)
    
    def to_dict(self, flat=False) -> Dict[str, Any]:
        data = asdict(self)

        if flat:
            data = flatten(data, reducer="dot")

        return data
    
    def __post_init__(self):
        for target in self.to_dict().keys():
            if getattr(self, target) is None:
                setattr(self, target, TargetMeasurements())
            elif isinstance(getattr(self, target), dict):
                setattr(self, target, TargetMeasurements(**getattr(self, target)))

    @classmethod
    def aggregate(cls, reports: List["BenchmarkReport"]) -> "BenchmarkReport":
        aggregated_measurements = {}
        for target in reports[0].to_dict().keys():
            measurements = [getattr(report, target) for report in reports]
            aggregated_measurements[target] = TargetMeasurements.aggregate(measurements)

        return cls.from_dict(aggregated_measurements)

    def to_plain_text(self) -> str:
        plain_text = ""

        for target in self.to_dict().keys():
            plain_text += f"+ {target}:\n"
            plain_text += getattr(self, target).to_plain_text()

        return plain_text

    def to_markdown_text(self) -> str:
        markdown_text = ""

        for target in self.to_dict().keys():
            markdown_text += f"# {target}:\n\n"
            markdown_text += getattr(self, target).to_markdown_text()

        return markdown_text

    def save_text(self, filename: str):
        with open(filename, mode="w") as f:
            f.write(self.to_plain_text())

    def save_markdown(self, filename: str):
        with open(filename, mode="w") as f:
            f.write(self.to_markdown_text())

    def to_dict(self, flat=False) -> Dict[str, Any]:
        data = asdict(self)

        if flat:
            data = flatten(data, reducer="dot")

        return data

    def to_dataframe(self) -> pd.DataFrame:
        flat_dict_data = self.to_dict(flat=True)
        return pd.DataFrame.from_dict(flat_dict_data, orient="index").T    

    def save_csv(self, path: Union[str, Path]) -> None:
        self.to_dataframe().to_csv(path, index=False)

    def log(self):
        for line in self.to_plain_text().split("\n"):
            if line:
                LOGGER.info(line)

    def print(self):
        CONSOLE.print(Markdown(self.to_markdown_text()))

    @property
    def default_filename(self) -> str:
        return "benchmark_report.json"

# Benchmark
@dataclass
class Benchmark:
    config: BenchmarkConfig
    report: BenchmarkReport

    def __post_init__(self):
        if isinstance(self.config, dict):
            self.config = BenchmarkConfig.from_dict(self.config)
        elif not isinstance(self.config, BenchmarkConfig):
            raise ValueError("config must be either a dict or a BenchmarkConfig instance")

        if isinstance(self.report, dict):
            self.report = BenchmarkReport.from_dict(self.report)
        elif not isinstance(self.report, BenchmarkReport):
            raise ValueError("report must be either a dict or a BenchmarkReport instance")

    @staticmethod
    def launch(config: BenchmarkConfig):
        """
        Runs an benchmark using specified launcher configuration/logic
        """

        # Allocate requested launcher
        launcher_config: ProcessConfig = config.launcher
        launcher_factory: Type[ProcessLauncher] = get_class(launcher_config._target_)
        launcher: ProcessLauncher = launcher_factory(launcher_config)

        # Launch the benchmark using the launcher
        report = launcher.launch(worker=Benchmark.run, worker_args=[config])

        if config.log_report:
            report.log()

        if config.print_report:
            report.print()

        return report

    @staticmethod
    def run(config: BenchmarkConfig):
        """
        Runs a scenario using specified backend configuration/logic
        """

        # Allocate requested backend
        backend_config: NexaConfig = config.backend
        backend_factory: Type[NexaBackend] = get_class(backend_config._target_)
        backend: NexaBackend = backend_factory(backend_config)

        # Allocate requested scenario
        scenario_config: InferenceConfig = config.scenario
        scenario_factory: Type[InferenceScenario] = get_class(scenario_config._target_)
        scenario: InferenceScenario = scenario_factory(scenario_config)

        # Run the scenario using the backend
        report = scenario.run(backend)

        return report

    def to_dict(self, flat=False) -> Dict[str, Any]:
        data = asdict(self)

        if flat:
            data = flatten(data, reducer="dot")

        return data

    def to_dataframe(self) -> pd.DataFrame:
        flat_dict_data = self.to_dict(flat=True)
        return pd.DataFrame.from_dict(flat_dict_data, orient="index").T    

    def save_csv(self, path: Union[str, Path]) -> None:
        self.to_dataframe().to_csv(path, index=False)

    @property
    def default_filename(self) -> str:
        return "benchmark.json"
