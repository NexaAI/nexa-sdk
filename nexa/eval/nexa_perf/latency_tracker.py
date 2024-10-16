import time
from contextlib import contextmanager
from dataclasses import asdict, dataclass
from logging import getLogger
from typing import List, Literal, Optional

import numpy as np
from rich.console import Console
from rich.markdown import Markdown

CONSOLE = Console()
LOGGER = getLogger("latency")

LATENCY_UNIT = "s"

Latency_Unit_Literal = Literal["s"]
Throughput_Unit_Literal = Literal["samples/s", "tokens/s", "images/s", "steps/s"]


@dataclass
class Latency:
    unit: Latency_Unit_Literal

    values: List[float]

    count: int
    total: float
    mean: float
    p50: float
    p90: float
    p95: float
    p99: float
    stdev: float
    stdev_: float

    def __getitem__(self, index) -> float:
        if isinstance(index, slice):
            return Latency.from_values(values=self.values[index], unit=self.unit)
        elif isinstance(index, int):
            return Latency.from_values(values=[self.values[index]], unit=self.unit)
        else:
            raise ValueError(f"Invalid index type: {type(index)}, expected int or slice")

    def __sub__(self, latency: "Latency") -> "Latency":
        latencies = [lat - latency.mean for lat in self.values]

        assert not any(
            latency < 0 for latency in latencies
        ), "Negative latency detected. Please increase the dimensions of your benchmark (inputs/warmup/iterations)."

        return Latency.from_values(values=latencies, unit=self.unit)

    @staticmethod
    def aggregate(latencies: List["Latency"]) -> "Latency":
        if len(latencies) == 0 or all(latency is None for latency in latencies):
            return None
        elif any(latency is None for latency in latencies):
            raise ValueError("Some latency measurements are missing")

        unit = latencies[0].unit
        values = sum((lat.values for lat in latencies), [])
        return Latency.from_values(values=values, unit=unit)

    @staticmethod
    def from_values(values: List[float], unit: str) -> "Latency":
        return Latency(
            unit=unit,
            values=values,
            count=len(values),
            total=sum(values),
            mean=np.mean(values),
            p50=np.percentile(values, 50),
            p90=np.percentile(values, 90),
            p95=np.percentile(values, 95),
            p99=np.percentile(values, 99),
            stdev=np.std(values) if len(values) > 1 else 0,
            stdev_=(np.std(values) / np.abs(np.mean(values))) * 100 if len(values) > 1 else 0,
        )

    def to_plain_text(self) -> str:
        plain_text = ""
        plain_text += "\t\t+ count: {count}\n"
        plain_text += "\t\t+ total: {total:.6f} ({unit})\n"
        plain_text += "\t\t+ mean: {mean:.6f} ({unit})\n"
        plain_text += "\t\t+ p50: {p50:.6f} ({unit})\n"
        plain_text += "\t\t+ p90: {p90:.6f} ({unit})\n"
        plain_text += "\t\t+ p95: {p95:.6f} ({unit})\n"
        plain_text += "\t\t+ p99: {p99:.6f} ({unit})\n"
        plain_text += "\t\t+ stdev: {stdev:.6f} ({unit})\n"
        plain_text += "\t\t+ stdev_: {stdev_:.2f} (%)\n"
        return plain_text.format(**asdict(self))

    def log(self):
        for line in self.to_plain_text().split("\n"):
            if line:
                LOGGER.info(line)

    def to_markdown_text(self) -> str:
        markdown_text = ""
        markdown_text += "| metric | value        | unit   |\n"
        markdown_text += "| :----- | -----------: |------: |\n"
        markdown_text += "| count  |      {count} |      - |\n"
        markdown_text += "| total  |    {total:f} | {unit} |\n"
        markdown_text += "| mean   |     {mean:f} | {unit} |\n"
        markdown_text += "| p50    |      {p50:f} | {unit} |\n"
        markdown_text += "| p90    |      {p90:f} | {unit} |\n"
        markdown_text += "| p95    |      {p95:f} | {unit} |\n"
        markdown_text += "| p99    |      {p99:f} | {unit} |\n"
        markdown_text += "| stdev  |    {stdev:f} | {unit} |\n"
        markdown_text += "| stdev_ | {stdev_:.2f} |      % |\n"
        return markdown_text.format(**asdict(self))

    def print(self):
        CONSOLE.print(Markdown(self.to_markdown_text()))


@dataclass
class Throughput:
    unit: Throughput_Unit_Literal

    value: float

    @staticmethod
    def aggregate(throughputs: List["Throughput"]) -> "Throughput":
        if len(throughputs) == 0:
            raise ValueError("No throughput measurements to aggregate")
        elif any(throughput is None for throughput in throughputs):
            raise ValueError("Some throughput measurements are missing")

        unit = throughputs[0].unit
        value = sum(throughput.value for throughput in throughputs)

        return Throughput(value=value, unit=unit)

    @staticmethod
    def from_latency(latency: Latency, volume: int, unit: str) -> "Throughput":
        value = volume / latency.mean if latency.mean > 0 else 0
        return Throughput(value=value, unit=unit)

    def to_plain_text(self) -> str:
        plain_text = ""
        plain_text += "\t\t+ throughput: {value:.2f} ({unit})\n"
        return plain_text.format(**asdict(self))

    def log(self):
        for line in self.to_plain_text().split("\n"):
            if line:
                LOGGER.info(line)

    def to_markdown_text(self) -> str:
        markdown_text = ""
        markdown_text += "| metric     |     value   |   unit |\n"
        markdown_text += "| :--------- | --------:   | -----: |\n"
        markdown_text += "| throughput | {value:.2f} | {unit} |\n"
        return markdown_text.format(**asdict(self))

    def print(self):
        CONSOLE.print(Markdown(self.to_markdown_text()))


class LatencyTracker:
    def __init__(self, device: str, backend: str):
        self.device = device
        self.backend = backend

        LOGGER.info("\t\t+ Tracking latency using CPU performance counter")

        self.start_time: Optional[float] = None
        self.start_events: List[float] = []
        self.end_events: List[float] = []

    def reset(self):
        self.start_time = None
        self.start_events = []
        self.end_events = []

    @contextmanager
    def track(self):
        yield from self._cpu_latency()

    def _cpu_latency(self):
        self.start_events.append(time.perf_counter())

        yield

        self.end_events.append(time.perf_counter())

    def get_latency(self) -> Latency:
        latencies_list = [(self.end_events[i] - self.start_events[i]) for i in range(len(self.start_events))]

        assert not any(
            latency < 0 for latency in latencies_list
        ), "Negative latency detected. Please increase the dimensions of your benchmark (inputs/warmup/iterations)."

        return Latency.from_values(latencies_list, unit=LATENCY_UNIT)

    def count(self):
        assert len(self.start_events) == len(
            self.end_events
        ), "Mismatched number of start and end events, count() should only be called outside of track() context"

        return len(self.start_events)

    def elapsed(self):
        if self.start_time is None:
            assert (
                len(self.start_events) == 0 and len(self.end_events) == 0
            ), "Number of recorded events is not zero, make sure to reset() the tracker properly"

            self.start_time = time.perf_counter()

        return time.perf_counter() - self.start_time