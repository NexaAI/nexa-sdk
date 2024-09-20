import argparse
import logging
import multiprocessing
import time
import requests
import os
import sys
from nexa.eval import evaluator, utils
from nexa.eval.loggers import EvaluationTracker
from nexa.eval.tasks import TaskManager
from nexa.eval.utils import handle_non_serializable, make_table, simple_parse_args_string
from nexa.gguf.server.nexa_service import run_nexa_ai_service as NexaServer
from nexa.constants import NEXA_MODEL_EVAL_RESULTS_PATH, NEXA_RUN_MODEL_MAP
from pathlib import Path

logging.basicConfig(level=logging.INFO)

class NexaEval:
    def __init__(self, model_path: str, tasks: str):
        model_path = NEXA_RUN_MODEL_MAP.get(model_path, model_path)
        self.model_path = model_path
        
        self.model_name = model_path.split(":")[0].lower()
        self.model_tag = model_path.split(":")[1].lower()
        self.tasks = tasks
        self.server_process = None
        self.server_url = "http://0.0.0.0:8000"
        output_path = Path(NEXA_MODEL_EVAL_RESULTS_PATH) / self.model_name / self.model_tag / tasks.replace(',', '_')
        self.eval_args = {
            "model": "nexa-gguf",
            "tasks": tasks,
            "model_args": f"base_url={self.server_url}/v1/completions",
            "hf_hub_log_args": "",
            "batch_size": 8,
            "device": "cuda",
            "output_path": str(output_path),
            "cache_requests": None,
            "log_samples": False,
            "include_path": None,
            "verbosity": "INFO",
            "seed": [0, 1234, 1234, 1234],
        }


    def start_server(self):
        self.server_process = multiprocessing.Process(
            target=NexaServer,
            args=(self.model_path,),
            kwargs={"host": "0.0.0.0", "port": 8000}
        )
        self.server_process.start()
        logging.info(f"Started server process for model: {self.model_path}")

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
        raise Exception("Server did not become ready within the specified timeout.")
    

    def evaluate_model(self, args):

        if args.output_path:
            args.hf_hub_log_args += f",output_path={args.output_path}"
        if os.environ.get("HF_TOKEN", None):
            args.hf_hub_log_args += f",token={os.environ.get('HF_TOKEN')}"
        
        evaluation_tracker_args = simple_parse_args_string(args.hf_hub_log_args)
        evaluation_tracker = EvaluationTracker(**evaluation_tracker_args)
        task_manager = TaskManager(args.verbosity, include_path=args.include_path)
        
        if args.tasks is None:
            logging.error("Need to specify task to evaluate.")
            sys.exit()
        else:
            task_list = args.tasks.split(",")
            task_names = task_manager.match_tasks(task_list)
        
        logging.info(f"Selected Tasks: {task_names}")

        request_caching_args = evaluator.request_caching_arg_to_dict(cache_requests=args.cache_requests)
        results = evaluator.simple_evaluate(
            model=args.model,
            model_args=args.model_args,
            tasks=task_names,
            batch_size=args.batch_size,
            device=args.device,
            evaluation_tracker=evaluation_tracker,
            task_manager=task_manager,
            random_seed=args.seed[0],
            numpy_random_seed=args.seed[1],
            torch_random_seed=args.seed[2],
            fewshot_random_seed=args.seed[3],
            **request_caching_args,
        )

        if results is not None:
            if args.log_samples:
                samples = results.pop("samples")
            evaluation_tracker.save_results_aggregated(results=results, samples=samples if args.log_samples else None)

            if args.log_samples:
                for task_name, config in results["configs"].items():
                    evaluation_tracker.save_results_samples(task_name=task_name, samples=results["samples"][task_name])
            print(make_table(results))
            if "groups" in results:
                print(make_table(results, "groups"))
    

    def run_evaluation(self):
        try:
            self.start_server()
            if self.wait_for_server():
                logging.info(f"Starting evaluation for tasks: {self.tasks}")
                args = argparse.Namespace(**self.eval_args)
                self.evaluate_model(args)
                logging.info("Evaluation completed")
        finally:
            if self.server_process:
                self.server_process.terminate()
                self.server_process.join()
                logging.info("Server process terminated")

def run_eval_inference(model_path: str, tasks: str):
    evaluator = NexaEval(model_path, tasks)
    evaluator.run_evaluation()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Nexa Model Evaluation")
    parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    parser.add_argument("--tasks", type=str, help="Tasks to evaluate, comma-separated")
    
    args = parser.parse_args()
    run_eval_inference(args.model_path, args.tasks)