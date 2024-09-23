import os
import numpy as np
import subprocess
import tempfile
from typing import Dict, List, Any
from collections import defaultdict
# For each problem, we generate up to 200 code samples using temperature 0.8 and nucleus sampling with top_p = 0.95.

def process_results(doc: Dict[str, Any], results: List[str]) -> Dict[str, Any]:
    completion = results[0]
    result = check_correctness(doc, completion)
    return {
        "pass@1": int(result["passed"]),
    }

def check_correctness(problem: Dict[str, Any], completion: str, timeout: float = 3.0) -> Dict[str, Any]:
    full_code = problem["prompt"] + completion + "\n" + problem["test"]
    
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as temp_file:
        temp_file.write(full_code)
        temp_file_path = temp_file.name

    try:
        result = subprocess.run(['python', temp_file_path], capture_output=True, text=True, timeout=timeout)
        passed = result.returncode == 0 and "test cases" not in result.stderr.lower()
        return {"passed": passed, "result": result.stdout if passed else result.stderr}
    except subprocess.TimeoutExpired:
        return {"passed": False, "result": "Timeout"}
    finally:
        os.remove(temp_file_path)

def estimate_pass_at_k(n: int, c: int, k: int) -> float:
    if n - c < k:
        return 1.0
    return 1.0 - np.prod(1.0 - k / np.arange(n - c + 1, n + 1))

def pass_at_k(samples: List[int], k: int) -> float:
    n_samples = len(samples)
    n_correct = sum(samples)
    if n_samples < k:
        return 1.0
    return estimate_pass_at_k(n_samples, n_correct, k)

def pass_at_1(samples: List[int]) -> float:
    return pass_at_k(samples, 1)






