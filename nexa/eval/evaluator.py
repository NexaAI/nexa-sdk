import json
import logging
import random
import time
from collections import defaultdict
from typing import TYPE_CHECKING, List, Optional, Union

import numpy as np

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

def nexa_evaluate(
    model,
    model_args: Optional[str] = None,
    tasks: Optional[List[str]] = None,
    num_fewshot: Optional[int] = None,
    batch_size: Optional[Union[int, str]] = None,
    limit: Optional[Union[int, float]] = None,
    bootstrap_iters: int = 100000,
    task_manager: Optional[TaskManager] = None,
    verbosity: str = "INFO",
    random_seed: int = 0,
    numpy_random_seed: int = 1234,
    fewshot_random_seed: int = 1234,
):
    """Instantiate and evaluate a model on a list of tasks.

    :param model: Union[str, LM]
        Name of model or LM object, see nexa.eval.models.get_model
    :param model_args: Optional[str]
        String for model class, see LM.create_from_arg_string.
        Ignored if `model` argument is a LM object.
    :param tasks: list[str]
        List of task names or Task objects.
    :param num_fewshot: int
        Number of examples in few-shot context
    :param batch_size: int
        Batch size for model
    :param limit: int or float, optional
        Limit the number of examples per task (only use this for testing). If <1, limit is a percentage of the total number of examples.
    :param bootstrap_iters:
        Number of iterations for bootstrap statistics, used when calculating stderrs. Set to 0 for no stderr calculations to be performed.
    :param random_seed: int
        Random seed for Python's random module. If set to None, the seed will not be set.
    :param numpy_random_seed: int
        Random seed for NumPy. If set to None, the seed will not be set.
    :param fewshot_random_seed: int
        Random seed for few-shot sampler random generator. If set to None, the seed of generator will be set to None.

    :return
        Dictionary of results
    """
    eval_logger.setLevel(getattr(logging, f"{verbosity}"))
    start_date = time.time()

    # Set random seeds
    random.seed(random_seed)
    np.random.seed(numpy_random_seed)

    if not tasks:
        raise ValueError("No tasks specified, or no tasks found. Please verify the task names.")

    lm = GGUFLM.create_from_arg_string(
        model_args,
        {
            "batch_size": batch_size,
        },
    )

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
                # eval_logger.info(f"Setting few-shot random generator seed to {fewshot_random_seed}")
                adjusted_task_dict[task_name] = task_obj
        return adjusted_task_dict

    task_dict = _adjust_config(task_dict)

    # Begin evaluation logic
    requests = defaultdict(list)
    padding_requests = defaultdict(int)
    eval_tasks = get_task_list(task_dict)

    for task_output in eval_tasks:
        task: Task = task_output.task
        task_limit = get_sample_size(task, limit)
        task.build_all_requests(
            limit=task_limit,
            rank=lm.rank,
            world_size=lm.world_size,
        )
        for instance in task.instances:
            reqtype = instance.request_type
            requests[reqtype].append(instance)

    # Run LM on inputs, get all outputs
    for reqtype, reqs in requests.items(): # TODO: probably change to multiprocessing
        eval_logger.info(f"Running {reqtype} requests")
        cloned_reqs = []
        for req in reqs:
            cloned_reqs.extend([req] * req.repeats)

        if (lm.world_size > 1) and (padding_requests[reqtype] > 0):
            for _ in range(padding_requests[reqtype]):
                cloned_reqs.extend([req] * req.repeats)

        resps = getattr(lm, reqtype)(cloned_reqs)

        for x, req in zip(resps, cloned_reqs):
            req.resps.append(x)

        if lm.world_size > 1:
            lm.accelerator.wait_for_everyone()

    # Postprocess outputs
    for task_output in eval_tasks:
        task = task_output.task
        task.apply_filters()
        instances_by_doc_id = defaultdict(list)
        for instance in task.instances:
            instances_by_doc_id[instance.doc_id].append(instance)
        for instances in instances_by_doc_id.values():
            instances.sort(key=lambda x: x.idx)
        for filter_key in task.instances[0].filtered_resps.keys():
            doc_iterator = task.doc_iterator(
                rank=lm.rank, limit=task_limit, world_size=lm.world_size
            )
            for doc_id, doc in doc_iterator:
                requests = instances_by_doc_id[doc_id]
                metrics = task.process_results(
                    doc, [req.filtered_resps[filter_key] for req in requests]
                )
                target = task.doc_to_target(doc)
                example = {
                    "doc_id": doc_id,
                    "doc": doc,
                    "target": target,
                    "arguments": [req.args for req in requests],
                    "resps": [req.resps for req in requests],
                    "filtered_resps": [
                        req.filtered_resps[filter_key] for req in requests
                    ],
                    "doc_hash": hash_string(
                        json.dumps(
                            requests[0].doc,
                            indent=2,
                            default=handle_non_serializable,
                            ensure_ascii=False,
                        )
                    ),
                    "prompt_hash": hash_string(requests[0].arguments[0]),
                    "target_hash": hash_string(str(target)),
                }
                example.update(metrics)
                task_output.logged_samples.append(example)
                for metric, value in metrics.items():
                    task_output.sample_metrics[(metric, filter_key)].append(value)

    if lm.rank == 0:
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

        # Add model info to results
        if isinstance(model, str):
            model_name = model
        elif hasattr(model, "config") and hasattr(model.config, "_name_or_path"):
            model_name = model.config._name_or_path
        else:
            model_name = type(model).__name__

        results_dict["config"] = {
            "model": model_name,
            "model_args": model_args,
            "batch_size": batch_size,
            "limit": limit,
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
    else:
        return None
