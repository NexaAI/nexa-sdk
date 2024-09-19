import argparse
import logging
import multiprocessing
import time
import requests
from nexa.eval.eval_runner import evaluate_model
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
            "num_fewshot": None,
            "hf_hub_log_args": "",
            "batch_size": 8,
            "max_batch_size": None,
            "device": "cuda",
            "output_path": str(output_path),
            "limit": None,
            "use_cache": None,
            "cache_requests": None,
            "check_integrity": False,
            "write_out": False,
            "log_samples": False,
            "system_instruction": None,
            "apply_chat_template": False,
            "fewshot_as_multiturn": False,
            "include_path": None,
            "gen_kwargs": None,
            "verbosity": "INFO",
            "predict_only": False,
            "wandb_args": None,
            "show_config": False,
            "seed": [0, 1234, 1234, 1234],
            "trust_remote_code": False,
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

    def run_evaluation(self):
        try:
            self.start_server()
            if self.wait_for_server():
                logging.info(f"Starting evaluation for tasks: {self.tasks}")
                args = argparse.Namespace(**self.eval_args)
                evaluate_model(args)
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