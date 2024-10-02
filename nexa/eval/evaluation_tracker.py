import json
import time
from datetime import datetime
from pathlib import Path

from nexa.eval.utils import (
    eval_logger,
    handle_non_serializable,
    hash_string,
    sanitize_model_name,
)


class EvaluationTracker:
    """
    Keeps track and saves relevant information of the evaluation process.
    Compiles the data from trackers and writes it to files, which can be published to the Hugging Face hub if requested.
    """

    def __init__(
        self,
        output_path: str = None,
    ) -> None:
        """
        Creates all the necessary loggers for evaluation tracking.

        Args:
            output_path (str): Path to save the results. If not provided, the results won't be saved.
        """
        # Initialize tracking variables
        self.model_source: str = None
        self.model_name: str = None
        self.model_name_sanitized: str = None
        self.start_time: float = time.perf_counter()
        self.end_time: float = None
        self.total_evaluation_time_seconds: str = None

        self.output_path = output_path

    @staticmethod
    def _get_model_name(model_args: str) -> str:
        """Extracts the model name from the model arguments."""

        def extract_model_name(model_args: str, key: str) -> str:
            """Extracts the model name from the model arguments using a key."""
            args_after_key = model_args.split(key)[1]
            return args_after_key.split(",")[0]

        # Order does matter; e.g., 'peft' and 'delta' are provided together with 'pretrained'
        prefixes = ["peft=", "delta=", "pretrained=", "model=", "path=", "engine="]
        for prefix in prefixes:
            if prefix in model_args:
                return extract_model_name(model_args, prefix)
        return ""

    def log_experiment_args(
        self,
        model_source: str,
        model_args: str,
    ) -> None:
        """Logs model parameters and job ID."""
        self.model_source = model_source
        self.model_name = self._get_model_name(model_args)
        self.model_name_sanitized = sanitize_model_name(self.model_name)

    def log_end_time(self) -> None:
        """Logs the end time of the evaluation and calculates the total evaluation time."""
        self.end_time = time.perf_counter()
        self.total_evaluation_time_seconds = str(self.end_time - self.start_time)

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
        self.log_end_time()

        if self.output_path:
            try:
                eval_logger.info("Saving results aggregated")

                # Calculate cumulative hash for each task - only if samples are provided
                task_hashes = {}
                if samples:
                    for task_name, task_samples in samples.items():
                        sample_hashes = [
                            s["doc_hash"] + s["prompt_hash"] + s["target_hash"]
                            for s in task_samples
                        ]
                        task_hashes[task_name] = hash_string("".join(sample_hashes))

                # Update initial results dict
                results.update({"task_hashes": task_hashes})

                # Collect configuration attributes
                config_attrs = {
                    "model_source": self.model_source,
                    "model_name": self.model_name,
                    "model_name_sanitized": self.model_name_sanitized,
                    "start_time": self.start_time,
                    "end_time": self.end_time,
                    "total_evaluation_time_seconds": self.total_evaluation_time_seconds,
                }
                results.update(config_attrs)

                dumped = json.dumps(
                    results,
                    indent=2,
                    default=handle_non_serializable,
                    ensure_ascii=False,
                )

                path = Path(self.output_path if self.output_path else Path.cwd())
                path = path.joinpath(self.model_name_sanitized)
                path.mkdir(parents=True, exist_ok=True)

                date_id = datetime.now().isoformat().replace(":", "-")
                file_results_aggregated = path.joinpath(f"results_{date_id}.json")
                file_results_aggregated.open("w", encoding="utf-8").write(dumped)

            except Exception as e:
                eval_logger.warning("Could not save results aggregated")
                eval_logger.info(repr(e))
        else:
            eval_logger.info(
                "Output path not provided, skipping saving results aggregated"
            )
