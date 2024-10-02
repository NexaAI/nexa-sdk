import argparse
import multiprocessing
import time
import requests
import sys
from nexa.eval import evaluator
from nexa.eval.evaluation_tracker import EvaluationTracker
from nexa.eval.tasks import TaskManager
from nexa.eval.utils import make_table, simple_parse_args_string
from nexa.gguf.server.nexa_service import run_nexa_ai_service as NexaServer
from nexa.constants import NEXA_MODEL_EVAL_RESULTS_PATH, NEXA_RUN_MODEL_MAP
from pathlib import Path
from datetime import datetime

def print_message(level, message):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"{timestamp} - {level} - {message}")

class NexaEval:
    def __init__(self, model_path: str, tasks: str):
        model_path = NEXA_RUN_MODEL_MAP.get(model_path, model_path)
        self.model_path = model_path
        
        self.model_name = model_path.split(":")[0].lower()
        self.model_tag = model_path.split(":")[1].lower()
        self.tasks = tasks
        self.server_process = None
        self.server_url = "http://0.0.0.0:8300"
        output_path = Path(NEXA_MODEL_EVAL_RESULTS_PATH) / self.model_name / self.model_tag / tasks.replace(',', '_')
        self.eval_args = {
            "model": model_path,
            "tasks": tasks,
            "model_args": f"base_url={self.server_url}/v1/completions",
            "hf_hub_log_args": "",
            "batch_size": 8,
            "output_path": str(output_path),
            "include_path": None,
            "verbosity": "INFO",
        }


    def start_server(self):
        self.server_process = multiprocessing.Process(
            target=NexaServer,
            args=(self.model_path,),
            kwargs={"host": "0.0.0.0", "port": 8300, "nctx": 4096},
        )
        self.server_process.start()
        print("INFO", f"Started server process for model: {self.model_path}")

    def wait_for_server(self, timeout: int = 60) -> bool:
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                response = requests.get(f"{self.server_url}/")
                if response.status_code == 200:
                    print_message("INFO", "Server is ready")
                    return True
            except requests.exceptions.ConnectionError:
                pass
            time.sleep(1)
        raise Exception("Server did not become ready within the specified timeout.")
    

    def evaluate_model(self, args):
        
        evaluation_tracker = EvaluationTracker(output_path=args.output_path, model_name=args.model)
        task_manager = TaskManager(args.verbosity, include_path=args.include_path)
        
        if args.tasks is None:
            print("ERROR", "Need to specify task to evaluate.")
            sys.exit()
        else:
            task_list = args.tasks.split(",")
            task_names = task_manager.match_tasks(task_list)
        
        print_message("INFO", f"Selected Tasks: {task_names}")

        from datasets.exceptions import DatasetNotFoundError
        try:
            results = evaluator.nexa_evaluate(
                model=args.model,
                model_args=args.model_args,
                tasks=task_names,
                batch_size=args.batch_size,
                task_manager=task_manager
            )
        except ValueError as e:
            if "No tasks specified, or no tasks found" in str(e):
                print_message("ERROR", f"Error: No valid tasks were found for evaluation. Specified tasks: {args.tasks}. Please verify the task names and try again.")
            else:
                print_message("ERROR", f"An unexpected ValueError occurred: {e}")
            return
        except DatasetNotFoundError as e:
            print_message("ERROR", f"Error: {e}")
            print_message("ERROR", "Run 'huggingface-cli login' to authenticate with the Hugging Face Hub.")
            return
        except RuntimeError as e:
            if "TensorFlow 2.0 or PyTorch should be installed" in str(e):
                print_message("ERROR", "This task requires either TensorFlow or PyTorch, but neither is installed.")
                print_message("ERROR", "To run this task, please install one of the following:")
                print_message("ERROR", "- PyTorch: Visit https://pytorch.org/ for installation instructions.")
                print_message("ERROR", "- TensorFlow: Visit https://www.tensorflow.org/install/ for installation instructions.")
            else:
                print_message("ERROR", f"An unexpected error occurred: {e}")
            return
        
        if results is not None:
            evaluation_tracker.save_results_aggregated(results=results)
            print(make_table(results))
            if "groups" in results:
                print(make_table(results, "groups"))
    

    def run_evaluation(self):
        try:
            self.start_server()
            if self.wait_for_server():
                print_message("INFO", f"Starting evaluation for tasks: {self.tasks}")
                args = argparse.Namespace(**self.eval_args)
                self.evaluate_model(args)
                print_message("INFO", "Evaluation completed")
        finally:
            if self.server_process:
                self.server_process.terminate()
                self.server_process.join()
                print_message("INFO", "Server process terminated")

def run_eval_inference(model_path: str, tasks: str):
    evaluator = NexaEval(model_path, tasks)
    evaluator.run_evaluation()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Nexa Model Evaluation")
    parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    parser.add_argument("--tasks", type=str, help="Tasks to evaluate, comma-separated")
    
    args = parser.parse_args()
    run_eval_inference(args.model_path, args.tasks)