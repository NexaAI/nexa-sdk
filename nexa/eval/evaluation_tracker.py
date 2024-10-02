import json
import time
from datetime import datetime
from pathlib import Path

from nexa.eval.utils import (
    eval_logger,
    handle_non_serializable,
)


class EvaluationTracker:
    """
    Keeps track and saves relevant information of the evaluation process.
    Compiles the data from trackers and writes it to files, which can be published to the Hugging Face hub if requested.
    """

    def __init__(
        self,
        model_name: str = None,
        output_path: str = None,
    ) -> None:
        """
        Creates all the necessary loggers for evaluation tracking.

        Args:
            output_path (str): Path to save the results. If not provided, the results won't be saved.
        """
        # Initialize tracking variables
        self.model_name: str = model_name
        self.start_time: float = time.perf_counter()
        self.end_time: float = None
        self.total_evaluation_time_seconds: str = None
        self.output_path = output_path

    def log_end_time(self) -> None:
        """Logs the end time of the evaluation and calculates the total evaluation time."""
        self.end_time = time.perf_counter()
        self.total_evaluation_time_seconds = str(self.end_time - self.start_time)

    def save_results_aggregated(
        self,
        results: dict,
    ) -> None:
        """
        Saves the aggregated results to the output path and pushes them to the Hugging Face hub if requested.

        Args:
            results (dict): The aggregated results to save.
        """
        self.log_end_time()

        if self.output_path:
            try:
                eval_logger.info("Saving results aggregated")

                # Collect configuration attributes
                config_attrs = {
                    "model_name": self.model_name,
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
