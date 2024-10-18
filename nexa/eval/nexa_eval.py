import argparse
import multiprocessing
import time
import requests
import sys
import json
import socket
import logging
from datetime import datetime
from pathlib import Path
from contextlib import ExitStack
from nexa.eval import evaluator
from nexa.eval.nexa_task.task_manager import TaskManager
from nexa.eval.utils import make_table, handle_non_serializable
from nexa.gguf.server.nexa_service import run_nexa_ai_service as NexaServer
from nexa.constants import NEXA_MODEL_EVAL_RESULTS_PATH, NEXA_RUN_MODEL_MAP
from nexa.eval.nexa_perf import (
    Benchmark,
    BenchmarkConfig,
    InferenceConfig,
    ProcessConfig,
    NexaConfig,
)

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

class NexaEval:
    def __init__(self, model_path: str, tasks: str = None, limit: float = None, port: int = None, nctx: int = None):
        model_path = NEXA_RUN_MODEL_MAP.get(model_path, model_path)
        self.model_path = model_path
        
        self.model_name = model_path.split(":")[0].lower()
        self.model_tag = model_path.split(":")[1].lower()
        self.limit = limit
        self.tasks = tasks
        self.server_process = None
        self.nctx = nctx if nctx is not None else 4096
        self.initial_port = port if port is not None else 8300
        self.port = self.initial_port
        self.server_url = f"http://localhost:{self.port}"
        output_path = Path(NEXA_MODEL_EVAL_RESULTS_PATH) / self.model_name / self.model_tag
        if self.tasks:
            output_path = output_path / self.tasks.replace(',', '_')
        self.eval_args = {
            "model": self.model_path,
            "tasks": self.tasks,
            "limit": self.limit,
            "model_args": f"base_url={self.server_url}/v1/completions",
            "batch_size": 8,
            "output_path": str(output_path),
            "include_path": None,
            "verbosity": "INFO",
        }

    def find_available_port(self):
        while True:
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                    s.bind(('localhost', self.port))
                    return self.port
            except socket.error:
                logging.info(f"Port {self.port} is in use, trying {self.port + 1}")
                self.port += 1

    def update_urls(self):
        self.server_url = f"http://localhost:{self.port}"
        self.eval_args["model_args"] = f"base_url={self.server_url}/v1/completions"

    def start_server(self):
        self.port = self.find_available_port()
        self.update_urls()
        self.server_process = multiprocessing.Process(
            target=NexaServer,
            args=(self.model_path,),
            kwargs={"host": "localhost", "port": self.port, "nctx": self.nctx},
        )
        self.server_process.start()
        logging.info(f"Started server process for model: {self.model_path} on port {self.port} with context window {self.nctx}")

    def wait_for_server(self, timeout: int = 60) -> bool:
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                response = requests.get(f"{self.server_url}/")
                if response.status_code == 200:
                    logging.info("Server is ready")
                    return True
            except requests.exceptions.ConnectionError:
                pass
            time.sleep(1)
        raise RuntimeError(f"Server did not become ready within the specified timeout of {timeout} seconds")

    def evaluate_model(self, args):

        start_time = time.perf_counter()
        task_manager = TaskManager(args.verbosity, include_path=args.include_path)

        if args.tasks is None:
            logging.error("Need to specify task to evaluate.")
            sys.exit()
        else:
            task_list = args.tasks.split(",")
            task_names = task_manager.match_tasks(task_list)
        
        logging.info(f"Selected Tasks: {task_names}")

        from datasets.exceptions import DatasetNotFoundError
        try:
            results = evaluator.nexa_evaluate(
                model=args.model,
                model_args=args.model_args,
                limit=args.limit,
                tasks=task_names,
                batch_size=args.batch_size,
                task_manager=task_manager
            )
        except ValueError as e:
            if "No tasks specified, or no tasks found" in str(e):
                logging.error(f"Error: No valid tasks were found for evaluation. Specified tasks: {args.tasks}. Please verify the task names and try again.")
            else:
                logging.error(f"An unexpected ValueError occurred: {e}")
            return
        except DatasetNotFoundError as e:
            logging.error(f"Error: {e}")
            logging.error("Run 'huggingface-cli login' to authenticate with the Hugging Face Hub.")
            return
        except RuntimeError as e:
            if "TensorFlow 2.0 or PyTorch should be installed" in str(e):
                logging.error("This task requires either TensorFlow or PyTorch, but neither is installed.")
                logging.error("To run this task, please install one of the following:")
                logging.error("- PyTorch: Visit https://pytorch.org/ for installation instructions.")
                logging.error("- TensorFlow: Visit https://www.tensorflow.org/install/ for installation instructions.")
            else:
                logging.error(f"An unexpected error occurred: {e}")
            return
        
        if results is not None:
            end_time = time.perf_counter()
            total_evaluation_time_seconds = str(end_time - start_time)

            config_attrs = {
                "model_name": args.model,
                "start_time": start_time,
                "end_time": end_time,
                "total_evaluation_time_seconds": total_evaluation_time_seconds,
            }
            results.update(config_attrs)

            if args.output_path:
                try:
                    logging.info("Saving aggregated results")

                    dumped = json.dumps(
                        results,
                        indent=2,
                        default=handle_non_serializable,
                        ensure_ascii=False,
                    )

                    path = Path(args.output_path)
                    path.mkdir(parents=True, exist_ok=True)

                    date_id = datetime.now().isoformat().replace(":", "-")
                    file_results_aggregated = path.joinpath(f"results_{date_id}.json")
                    with file_results_aggregated.open("w", encoding="utf-8") as f:
                        f.write(dumped)

                except Exception as e:
                    logging.warning("Could not save aggregated results")
                    logging.info(repr(e))
            else:
                logging.info("Output path not provided, skipping saving aggregated results")

            print(make_table(results))
            if "groups" in results:
                print(make_table(results, "groups"))
    

    def run_evaluation(self):
        with ExitStack() as stack:
            try:
                self.start_server()
                stack.callback(self.stop_server)
                if self.wait_for_server():
                    logging.info(f"Starting evaluation for tasks: {self.tasks}")
                    args = argparse.Namespace(**self.eval_args)
                    self.evaluate_model(args)
                    logging.info("Evaluation completed")
                    logging.info(f"Output file has been saved to {self.eval_args['output_path']}")
            except Exception as e:
                logging.error(f"An error occurred during evaluation: {e}")

    def run_perf_eval(self, device: str, new_tokens: int):
        BENCHMARK_NAME = f"nexa_sdk_{self.model_path}"
        launcher_config = ProcessConfig()
        backend_config = NexaConfig(
            device=device,
            model=self.model_path,
            task="text-generation",
        )
        scenario_config = InferenceConfig(
            latency=True,
            memory=True,  
            energy=True,
            input_shapes={
                "batch_size": 1,        # TODO: make it dynamic, hardcoded to 1 for now
                "sequence_length": 256,
                "vocab_size": 32000,
            },
            generate_kwargs={
                "max_new_tokens": new_tokens,
                "min_new_tokens": new_tokens,
            },
        )
        benchmark_config = BenchmarkConfig(
            name=BENCHMARK_NAME,
            launcher=launcher_config,
            backend=backend_config,
            scenario=scenario_config,
        )

        # Launch the benchmark with the specified configuration
        benchmark_report = Benchmark.launch(benchmark_config)
        benchmark_report.save_csv(f"benchmark_report_{self.model_path}.csv")


    def stop_server(self):
        if self.server_process:
            self.server_process.terminate()
            self.server_process.join()
            logging.info("Server process terminated")

def run_eval_inference(model_path: str, tasks: str = None, limit: float = None, port: int = None, nctx: int = None, device: str = "cpu", new_tokens: int = 100):
    evaluator = NexaEval(model_path, tasks, limit, port, nctx)
    if not tasks:
        evaluator.run_perf_eval(device, new_tokens)
    else:
        evaluator.run_evaluation()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Nexa Model Evaluation")
    parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    parser.add_argument("--tasks", type=str, help="Tasks to evaluate, comma-separated", default=None)
    parser.add_argument("--limit", type=float, help="Limit the number of examples per task. If <1, limit is a percentage of the total number of examples.", default=None)
    parser.add_argument("--port", type=int, help="Initial port to bind the server to", default=8300)
    parser.add_argument("--nctx", type=int, help="Length of context window", default=4096)
    parser.add_argument("--device", type=str, help="Device to run the inference on, choose from 'cpu', 'cuda', 'mps'", default="cpu")
    parser.add_argument("--new_tokens", type=int, help="Number of new tokens to evaluate", default=100)

    args = parser.parse_args()
    run_eval_inference(args.model_path, args.tasks, args.limit, args.port, args.nctx, args.device, args.new_tokens)