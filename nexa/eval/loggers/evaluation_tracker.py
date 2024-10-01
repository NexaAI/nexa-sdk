import json
import os
import re
import time
from dataclasses import asdict, dataclass
from datetime import datetime
from pathlib import Path

from nexa.eval.utils import (
    eval_logger,
    handle_non_serializable,
    hash_string,
    sanitize_list,
    sanitize_model_name,
)


@dataclass(init=False)
class GeneralConfigTracker:
    """
    Tracker for the evaluation parameters.

    Attributes:
        model_source (str): Source of the model (e.g. Hugging Face, GGUF, etc.)
        model_name (str): Name of the model.
        model_name_sanitized (str): Sanitized model name for directory creation.
        start_time (float): Start time of the experiment. Logged at class init.
        end_time (float): Start time of the experiment. Logged when calling [`GeneralConfigTracker.log_end_time`]
        total_evaluation_time_seconds (str): Inferred total evaluation time in seconds (from the start and end times).
    """

    model_source: str = None
    model_name: str = None
    model_name_sanitized: str = None
    system_instruction: str = None
    system_instruction_sha: str = None
    start_time: float = None
    end_time: float = None
    total_evaluation_time_seconds: str = None

    def __init__(self) -> None:
        """Starts the evaluation timer."""
        self.start_time = time.perf_counter()

    @staticmethod
    def _get_model_name(model_args: str) -> str:
        """Extracts the model name from the model arguments."""

        def extract_model_name(model_args: str, key: str) -> str:
            """Extracts the model name from the model arguments using a key."""
            args_after_key = model_args.split(key)[1]
            return args_after_key.split(",")[0]

        # order does matter, e.g. peft and delta are provided together with pretrained
        prefixes = ["peft=", "delta=", "pretrained=", "model=", "path=", "engine="]
        for prefix in prefixes:
            if prefix in model_args:
                return extract_model_name(model_args, prefix)
        return ""

    def log_experiment_args(
        self,
        model_source: str,
        model_args: str,
        system_instruction: str,
    ) -> None:
        """Logs model parameters and job ID."""
        self.model_source = model_source
        self.model_name = GeneralConfigTracker._get_model_name(model_args)
        self.model_name_sanitized = sanitize_model_name(self.model_name)
        self.system_instruction = system_instruction
        self.system_instruction_sha = (
            hash_string(system_instruction) if system_instruction else None
        )

    def log_end_time(self) -> None:
        """Logs the end time of the evaluation and calculates the total evaluation time."""
        self.end_time = time.perf_counter()
        self.total_evaluation_time_seconds = str(self.end_time - self.start_time)


class EvaluationTracker:
    """
    Keeps track and saves relevant information of the evaluation process.
    Compiles the data from trackers and writes it to files, which can be published to the Hugging Face hub if requested.
    """

    def __init__(
        self,
        output_path: str = None,
        public_repo: bool = False,
        leaderboard_url: str = "",
        point_of_contact: str = "",
    ) -> None:
        """
        Creates all the necessary loggers for evaluation tracking.

        Args:
            output_path (str): Path to save the results. If not provided, the results won't be saved.
            public_repo (bool): Whether to push the results to a public or private repository.
            leaderboard_url (str): URL to the leaderboard on the Hugging Face hub on the dataset card.
            point_of_contact (str): Contact information on the Hugging Face hub dataset card.
        """
        self.general_config_tracker = GeneralConfigTracker()

        self.output_path = output_path
        self.public_repo = public_repo
        self.leaderboard_url = leaderboard_url
        self.point_of_contact = point_of_contact


    def save_results_aggregated(
        self,
        results: dict,
        samples: dict,
    ) -> None:
        """
        Saves the aggregated results and samples to the output path and pushes them to the Hugging Face hub if requested.

        Args:
            results (dict): The aggregated results to save.
            samples (dict): The samples results to save.
        """
        self.general_config_tracker.log_end_time()

        if self.output_path:
            try:
                eval_logger.info("Saving results aggregated")

                # calculate cumulative hash for each task - only if samples are provided
                task_hashes = {}
                if samples:
                    for task_name, task_samples in samples.items():
                        sample_hashes = [
                            s["doc_hash"] + s["prompt_hash"] + s["target_hash"]
                            for s in task_samples
                        ]
                        task_hashes[task_name] = hash_string("".join(sample_hashes))

                # update initial results dict
                results.update({"task_hashes": task_hashes})
                results.update(asdict(self.general_config_tracker))
                dumped = json.dumps(
                    results,
                    indent=2,
                    default=handle_non_serializable,
                    ensure_ascii=False,
                )

                path = Path(self.output_path if self.output_path else Path.cwd())
                path = path.joinpath(self.general_config_tracker.model_name_sanitized)
                path.mkdir(parents=True, exist_ok=True)

                self.date_id = datetime.now().isoformat().replace(":", "-")
                file_results_aggregated = path.joinpath(f"results_{self.date_id}.json")
                file_results_aggregated.open("w", encoding="utf-8").write(dumped)

            except Exception as e:
                eval_logger.warning("Could not save results aggregated")
                eval_logger.info(repr(e))
        else:
            eval_logger.info(
                "Output path not provided, skipping saving results aggregated"
            )

    def save_results_samples(
        self,
        task_name: str,
        samples: dict,
    ) -> None:
        """
        Saves the samples results to the output path and pushes them to the Hugging Face hub if requested.

        Args:
            task_name (str): The task name to save the samples for.
            samples (dict): The samples results to save.
        """
        if self.output_path:
            try:
                eval_logger.info(f"Saving per-sample results for: {task_name}")

                path = Path(self.output_path if self.output_path else Path.cwd())
                path = path.joinpath(self.general_config_tracker.model_name_sanitized)
                path.mkdir(parents=True, exist_ok=True)

                file_results_samples = path.joinpath(
                    f"samples_{task_name}_{self.date_id}.jsonl"
                )

                for sample in samples:
                    arguments = {}
                    for i, arg in enumerate(sample["arguments"]):
                        arguments[f"gen_args_{i}"] = {}
                        for j, tmp in enumerate(arg):
                            arguments[f"gen_args_{i}"][f"arg_{j}"] = tmp

                    sample["resps"] = sanitize_list(sample["resps"])
                    sample["filtered_resps"] = sanitize_list(sample["filtered_resps"])
                    sample["arguments"] = arguments
                    sample["target"] = str(sample["target"])

                    sample_dump = (
                        json.dumps(
                            sample,
                            default=handle_non_serializable,
                            ensure_ascii=False,
                        )
                        + "\n"
                    )

                    with open(file_results_samples, "a", encoding="utf-8") as f:
                        f.write(sample_dump)

            except Exception as e:
                eval_logger.warning("Could not save sample results")
                eval_logger.info(repr(e))
        else:
            eval_logger.info("Output path not provided, skipping saving sample results")