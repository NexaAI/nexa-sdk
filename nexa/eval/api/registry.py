import logging
from typing import Callable, Dict

import evaluate as hf_evaluate


eval_logger = logging.getLogger("nexa-eval")

MODEL_REGISTRY = {}


TASK_REGISTRY = {}
GROUP_REGISTRY = {}
ALL_TASKS = set()
func2task_index = {}


OUTPUT_TYPE_REGISTRY = {}
METRIC_REGISTRY = {}
METRIC_AGGREGATION_REGISTRY = {}
AGGREGATION_REGISTRY: Dict[str, Callable[[], Dict[str, Callable]]] = {}
HIGHER_IS_BETTER_REGISTRY = {}
FILTER_REGISTRY = {}

DEFAULT_METRIC_REGISTRY = {
    "multiple_choice": ["acc", "acc_norm"],
    "generate_until": ["exact_match"],
}


def register_metric(**args):
    # TODO: do we want to enforce a certain interface to registered metrics?
    def decorate(fn):
        assert "metric" in args
        name = args["metric"]

        for key, registry in [
            ("metric", METRIC_REGISTRY),
            ("higher_is_better", HIGHER_IS_BETTER_REGISTRY),
            ("aggregation", METRIC_AGGREGATION_REGISTRY),
        ]:
            if key in args:
                value = args[key]
                assert (
                    value not in registry
                ), f"{key} named '{value}' conflicts with existing registered {key}!"

                if key == "metric":
                    registry[name] = fn
                elif key == "aggregation":
                    registry[name] = AGGREGATION_REGISTRY[value]
                else:
                    registry[name] = value

        return fn

    return decorate


def get_metric(name: str, hf_evaluate_metric=False) -> Callable:
    if not hf_evaluate_metric:
        if name in METRIC_REGISTRY:
            return METRIC_REGISTRY[name]
        else:
            eval_logger.warning(
                f"Could not find registered metric '{name}' in nexa-eval, searching in HF Evaluate library..."
            )

    try:
        metric_object = hf_evaluate.load(name)
        return metric_object.compute
    except Exception:
        eval_logger.error(
            f"{name} not found in the evaluate library! Please check https://huggingface.co/evaluate-metric",
        )


def register_aggregation(name: str):
    def decorate(fn):
        assert (
            name not in AGGREGATION_REGISTRY
        ), f"aggregation named '{name}' conflicts with existing registered aggregation!"

        AGGREGATION_REGISTRY[name] = fn
        return fn

    return decorate


def get_aggregation(name: str) -> Callable[[], Dict[str, Callable]]:
    try:
        return AGGREGATION_REGISTRY[name]
    except KeyError:
        eval_logger.warning(f"{name} not a registered aggregation metric!")


def get_metric_aggregation(name: str) -> Callable[[], Dict[str, Callable]]:
    try:
        return METRIC_AGGREGATION_REGISTRY[name]
    except KeyError:
        eval_logger.warning(f"{name} metric is not assigned a default aggregation!")


def is_higher_better(metric_name) -> bool:
    try:
        return HIGHER_IS_BETTER_REGISTRY[metric_name]
    except KeyError:
        eval_logger.warning(
            f"higher_is_better not specified for metric '{metric_name}'!"
        )


def register_filter(name):
    def decorate(cls):
        if name in FILTER_REGISTRY:
            eval_logger.info(
                f"Registering filter `{name}` that is already in Registry {FILTER_REGISTRY}"
            )
        FILTER_REGISTRY[name] = cls
        return cls

    return decorate


def get_filter(filter_name: str) -> type:
    try:
        return FILTER_REGISTRY[filter_name]
    except KeyError:
        eval_logger.warning(f"filter `{filter_name}` is not registered!")
