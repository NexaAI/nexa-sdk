import json
import logging
import random
import time
from collections import defaultdict
from typing import TYPE_CHECKING, List, Optional, Union
import multiprocessing
import queue
import numpy as np
from tqdm import tqdm

from nexa import __version__
import nexa.eval.nexa_task.metrics
import nexa.eval.nexa_task.registry
from nexa.eval.nexa_task.task import Task
from nexa.eval.nexa_models import GGUFLM
from nexa.eval.evaluator_utils import (
    consolidate_group_results,
    consolidate_results,
    get_sample_size,
    get_subtask_list,
    get_task_list,
    prepare_print_tasks,
)
from nexa.eval.nexa_task.task_manager import (
    TaskManager,
    get_task_dict,
)
from nexa.eval.utils import (
    eval_logger,
    handle_non_serializable,
    hash_string,
)

if TYPE_CHECKING:
    from nexa.eval.nexa_task.task import Task

# Define the worker function at the global scope
def worker(task_queue, result_queue, stop_event, model_path):
    # Disable tqdm in worker processes
    import sys
    import os
    import tqdm
    tqdm.tqdm = lambda *args, **kwargs: iter(args[0]) if args else iter([])
    # Redirect stdout and stderr to null to suppress any prints
    sys.stdout = open(os.devnull, 'w')
    sys.stderr = open(os.devnull, 'w')

    # Initialize the model in each process
    lm_worker = GGUFLM(model_path)
    while not stop_event.is_set():
        try:
            item = task_queue.get(timeout=0.1)
            if item is None:
                task_queue.task_done()
                break  # Received sentinel value, exit loop
            idx, req = item
            # Process the request
            reqtype = req.request_type
            resps = getattr(lm_worker, reqtype)([req])
            # Put the response in the result queue
            result_queue.put((idx, resps[0]))
            task_queue.task_done()
        except queue.Empty:
            continue
        except Exception as e:
            task_queue.task_done()

def nexa_evaluate(
    model_path,
    tasks: Optional[List[str]] = None,
    num_fewshot: Optional[int] = None,
    limit: Optional[Union[int, float]] = None,
    bootstrap_iters: int = 100000,
    task_manager: Optional[TaskManager] = None,
    verbosity: str = "INFO",
    random_seed: int = 0,
    numpy_random_seed: int = 1234,
    fewshot_random_seed: int = 1234,
    num_workers: int = 1,
):
    eval_logger.setLevel(getattr(logging, f"{verbosity}"))
    start_date = time.time()

    # Set random seeds
    random.seed(random_seed)
    np.random.seed(numpy_random_seed)

    if not tasks:
        raise ValueError("No tasks specified, or no tasks found. Please verify the task names.")

    task_dict = get_task_dict(tasks, task_manager)

    # Helper function to recursively apply config overrides to leaf subtasks
    def _adjust_config(task_dict):
        adjusted_task_dict = {}
        for task_name, task_obj in task_dict.items():
            if isinstance(task_obj, dict):
                adjusted_task_dict[task_name] = _adjust_config(task_obj)
            else:
                if num_fewshot is not None:
                    default_num_fewshot = task_obj.get_config("num_fewshot")
                    if default_num_fewshot == 0:
                        eval_logger.info(
                            f"num_fewshot has been set to 0 for {task_name} in its config. Manual configuration will be ignored."
                        )
                    else:
                        eval_logger.warning(
                            f"Overwriting default num_fewshot of {task_name} from {default_num_fewshot} to {num_fewshot}"
                        )
                        task_obj.set_config(key="num_fewshot", value=num_fewshot)
                else:
                    default_num_fewshot = task_obj.get_config("num_fewshot")
                    if default_num_fewshot is None:
                        task_obj.set_config(key="num_fewshot", value=0)
                task_obj.set_fewshot_seed(seed=fewshot_random_seed)
                adjusted_task_dict[task_name] = task_obj
        return adjusted_task_dict

    task_dict = _adjust_config(task_dict)

    # Begin evaluation logic
    requests = []
    eval_tasks = get_task_list(task_dict)

    for task_output in eval_tasks:
        task: Task = task_output.task
        task_limit = get_sample_size(task, limit)
        task.build_all_requests(limit=task_limit)
        for instance in task.instances:
            requests.extend([instance] * instance.repeats)

    # Assign an index to each request
    idx_to_req = {}
    indexed_requests = []
    for idx, req in enumerate(requests):
        idx_to_req[idx] = req
        indexed_requests.append((idx, req))

    # Run LM on inputs, get all outputs
    if num_workers == 1:
        # Without multiprocessing
        lm = GGUFLM(model_path)
        eval_logger.info(f"Running requests with a single worker")
        reqs_by_type = defaultdict(list)
        for idx, req in indexed_requests:
            reqs_by_type[req.request_type].append((idx, req))
        for reqtype, reqs in reqs_by_type.items():
            idxs, reqs_only = zip(*reqs)
            resps = getattr(lm, reqtype)(reqs_only)
            for idx, x in zip(idxs, resps):
                req = idx_to_req[idx]
                req.resps.append(x)
    else:
        # Multiprocessing logic
        # Define the task queue, result queue, and stop event
        task_queue = multiprocessing.JoinableQueue()
        result_queue = multiprocessing.Queue()
        stop_event = multiprocessing.Event()

        # Add all requests to the task queue
        for item in indexed_requests:
            task_queue.put(item)

        # Add sentinel values to stop workers
        for _ in range(num_workers):
            task_queue.put(None)

        # Start worker processes
        processes = []
        for _ in range(num_workers):
            p = multiprocessing.Process(
                target=worker,
                args=(task_queue, result_queue, stop_event, model_path),
            )
            p.start()
            processes.append(p)

        # Create progress bar in the main process
        pbar = tqdm(total=len(requests))

        # Collect results and update progress bar
        results_received = 0
        total_results = len(requests)
        while results_received < total_results:
            try:
                # Get result from result queue
                idx, resp = result_queue.get(timeout=1)
                req = idx_to_req[idx]
                req.resps.append(resp)
                results_received += 1
                pbar.update(1)
            except queue.Empty:
                continue
            except Exception:
                continue
            
        pbar.close()

        # Ensure all processes have finished
        stop_event.set()
        for p in processes:
            p.join()

    # Postprocess outputs
    for task_output in eval_tasks:
        task = task_output.task
        task.apply_filters()
        instances_by_doc_id = defaultdict(list)
        for instance in task.instances:
            instances_by_doc_id[instance.doc_id].append(instance)
        for instances in instances_by_doc_id.values():
            instances.sort(key=lambda x: x.idx)
        if not task.instances:
            continue  # Skip if no instances are present
        if not task.instances[0].filtered_resps:
            continue  # Skip if filtered_resps is empty
        for filter_key in task.instances[0].filtered_resps.keys():
            doc_iterator = task.doc_iterator(limit=task_limit)
            for doc_id, doc in doc_iterator:
                requests_list = instances_by_doc_id.get(doc_id, [])
                if not requests_list:
                    continue  # No requests for this doc_id
                filtered_resps = [req.filtered_resps.get(filter_key, None) for req in requests_list]
                if None in filtered_resps:
                    continue  # Skip if any filtered_resps is None
                metrics = task.process_results(
                    doc, filtered_resps
                )
                target = task.doc_to_target(doc)
                example = {
                    "doc_id": doc_id,
                    "doc": doc,
                    "target": target,
                    "arguments": [req.args for req in requests_list],
                    "resps": [req.resps for req in requests_list],
                    "filtered_resps": [
                        req.filtered_resps[filter_key] for req in requests_list
                    ],
                    "doc_hash": hash_string(
                        json.dumps(
                            requests_list[0].doc,
                            indent=2,
                            default=handle_non_serializable,
                            ensure_ascii=False,
                        )
                    ),
                    "prompt_hash": hash_string(requests_list[0].arguments[0]),
                    "target_hash": hash_string(str(target)),
                }
                example.update(metrics)
                task_output.logged_samples.append(example)
                for metric, value in metrics.items():
                    task_output.sample_metrics[(metric, filter_key)].append(value)

    # Aggregate results over all datapoints
    for task_output in eval_tasks:
        task_output.calculate_aggregate_metric(bootstrap_iters=bootstrap_iters)
    (
        results,
        samples,
        configs,
        versions,
        num_fewshot,
        higher_is_better,
    ) = consolidate_results(eval_tasks)

    # Calculate group metrics
    if bool(results):
        results, versions, show_group_table, *_ = consolidate_group_results(
            results, versions, task_dict
        )

    results_agg, group_agg = prepare_print_tasks(task_dict, results)
    subtask_list = get_subtask_list(task_dict)

    # Collect higher_is_better values
    for group, task_list in subtask_list.items():
        _higher_is_better = {}
        if task_list:
            for task_name in task_list:
                task_higher_is_better = higher_is_better[task_name]
                for m, h in task_higher_is_better.items():
                    if m not in _higher_is_better:
                        _higher_is_better[m] = h
                    elif _higher_is_better[m] != h:
                        eval_logger.warning(
                            f"Higher_is_better values for metric {m} in group {group} are not consistent. Defaulting to None."
                        )
                        _higher_is_better[m] = None
            higher_is_better[group] = _higher_is_better

    results_dict = {
        "results": dict(results_agg.items()),
        **(
            {"groups": dict(group_agg.items())}
            if (bool(group_agg) & show_group_table)
            else {}
        ),
        "group_subtasks": dict(reversed(subtask_list.items())),
        "configs": dict(sorted(configs.items())),
        "versions": dict(sorted(versions.items())),
        "n-shot": dict(sorted(num_fewshot.items())),
        "higher_is_better": dict(sorted(higher_is_better.items())),
        "n-samples": {
            task_output.task_name: {
                "original": len(task_output.task.eval_docs),
                "effective": min(
                    task_limit if task_limit else len(task_output.task.eval_docs),
                    len(task_output.task.eval_docs),
                ),
            }
            for task_output in eval_tasks
        },
        "samples": dict(samples),
    }

    results_dict["config"] = {
        "model": model_path,
        "limit": limit,
        "num_workers": num_workers,
        "bootstrap_iters": bootstrap_iters,
        "random_seed": random_seed,
        "numpy_seed": numpy_random_seed,
        "fewshot_seed": fewshot_random_seed,
    }
    results_dict.update({
        "date": start_date,
        "nexa_sdk_version":  __version__,
    })
    return results_dict
