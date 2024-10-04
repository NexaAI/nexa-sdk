import abc
import ast
import logging
import random
import re
from functools import partial
from collections.abc import Callable
from copy import deepcopy
from dataclasses import asdict, dataclass
from inspect import getsource
from typing import (
    Any,
    Dict,
    Iterable,
    Iterator,
    List,
    Mapping,
    Optional,
    Tuple,
    Union,
)

import datasets
import numpy as np
from tqdm import tqdm

from nexa.eval import utils
from nexa.eval.nexa_task import samplers
from nexa.eval.nexa_task.instance import Instance, OutputType
from nexa.eval.nexa_task.metrics import mean
from nexa.eval.nexa_task.registry import (
    AGGREGATION_REGISTRY,
    DEFAULT_METRIC_REGISTRY,
    get_aggregation,
    get_metric,
    get_metric_aggregation,
    is_higher_better,
)
from nexa.eval.nexa_task.filter import build_filter_ensemble
from nexa.eval.prompts import get_prompt


ALL_OUTPUT_TYPES = [
    "multiple_choice",
    "generate_until",
]

eval_logger = logging.getLogger("nexa-eval")


@dataclass
class TaskConfig(dict):
    # task naming/registry
    task: Optional[str] = None
    task_alias: Optional[str] = None
    tag: Optional[Union[str, list]] = None
    group: Optional[Union[str, list]] = None
    # HF dataset options.
    # which dataset to use,
    # and what splits for what purpose
    dataset_path: Optional[str] = None
    dataset_name: Optional[str] = None
    dataset_kwargs: Optional[dict] = None
    training_split: Optional[str] = None
    validation_split: Optional[str] = None
    test_split: Optional[str] = None
    fewshot_split: Optional[str] = (
        None  # TODO: assert that this not None if num_fewshot > 0. (?) assert if this is same split as one evaling (?)
    )
    # formatting / prompting options.
    # see docs/advanced_task_guide.md for more info
    process_docs: Optional[Callable] = None
    doc_to_text: Optional[Union[Callable, str]] = None
    doc_to_target: Optional[Union[Callable, str]] = None
    doc_to_choice: Optional[Union[Callable, str, dict, list]] = None
    process_results: Optional[Union[Callable, str]] = None
    use_prompt: Optional[str] = None
    description: str = ""
    target_delimiter: str = " "
    fewshot_delimiter: str = "\n\n"
    fewshot_config: Optional[dict] = None
    # runtime configuration options
    num_fewshot: Optional[int] = None
    # scoring options
    metric_list: Optional[list] = None
    output_type: OutputType = "generate_until"
    generation_kwargs: Optional[dict] = None
    repeats: int = 1
    filter_list: Optional[Union[str, list]] = None
    should_decontaminate: bool = False
    doc_to_decontamination_query: Optional[str] = None
    metadata: Optional[dict] = (
        None  # by default, not used in the code. allows for users to pass arbitrary info to tasks
    )

    def __post_init__(self) -> None:
        if self.group is not None:
            # eval_logger.warning(
            #     "A task YAML file was found to contain a `group` key. Groups which provide aggregate scores over several subtasks now require a separate config file--if not aggregating, you may want to use the `tag` config option instead within your config. "
            # )

            if self.tag is None:
                self.tag = self.group
            else:
                raise ValueError(
                    "Got both a `group` and `tag` entry within a TaskConfig. Please use one or the other--`group` values will be deprecated in v0.4.4."
                )

        if self.generation_kwargs is not None:
            if self.output_type != "generate_until":
                eval_logger.warning(
                    f"[{self.task}] passed `generation_kwargs`, but not using `output_type: generate_until`!"
                )

            if "temperature" in self.generation_kwargs:
                self.generation_kwargs["temperature"] = float(
                    self.generation_kwargs["temperature"]
                )

            if "until" not in self.generation_kwargs:
                self.generation_kwargs["until"] = [self.fewshot_delimiter]
        else:
            if self.output_type == "generate_until":
                # ensure that we greedily generate in absence of explicit arguments otherwise
                self.generation_kwargs = {
                    "until": (
                        None
                        if self.fewshot_delimiter is None
                        else [self.fewshot_delimiter]
                    ),
                    "do_sample": False,
                }

    def __getitem__(self, item):
        return getattr(self, item)

    def __setitem__(self, item, value):
        return setattr(self, item, value)

    def to_dict(self, keep_callable: bool = False) -> dict:
        """dumps the current config as a dictionary object, as a printable format.
        null fields will not be printed.
        Used for dumping results alongside full task configuration

        :return: dict
            A printable dictionary version of the TaskConfig object.

        # TODO: should any default value in the TaskConfig not be printed?
        """
        cfg_dict = asdict(self)
        # remove values that are `None`
        for k, v in list(cfg_dict.items()):
            if v is None:
                cfg_dict.pop(k)
            elif k == "metric_list":
                for metric_dict in v:
                    for metric_key, metric_value in metric_dict.items():
                        if callable(metric_value):
                            metric_dict[metric_key] = self.serialize_function(
                                metric_value, keep_callable=keep_callable
                            )
                cfg_dict[k] = v
            elif callable(v):
                cfg_dict[k] = self.serialize_function(v, keep_callable=keep_callable)
        return cfg_dict

    def serialize_function(
        self, value: Union[Callable, str], keep_callable=False
    ) -> Union[Callable, str]:
        """Serializes a given function or string.

        If 'keep_callable' is True, the original callable is returned.
        Otherwise, attempts to return the source code of the callable using 'getsource'.
        """
        if keep_callable:
            return value
        else:
            try:
                return getsource(value)
            except (TypeError, OSError):
                return str(value)


class Task:
    VERSION: Optional[Union[int, str]] = "Yaml"
    OUTPUT_TYPE: Optional[str] = None
    DATASET_PATH: Optional[str] = None
    DATASET_NAME: Optional[str] = None
    CONFIG = None

    def __init__(
        self,
        data_dir=None,
        cache_dir=None,
        download_mode=None,
        config: Optional[dict] = None,
    ) -> None:  # TODO no super() call here
        
        self.config = self.CONFIG if self.CONFIG else (TaskConfig(**config) if config else TaskConfig())

        if config is not None:
            self.config.__dict__.update(config)

        if self.config is None:
            raise ValueError("Must pass a config to Task, either in cls.CONFIG or `config` kwarg")

        if isinstance(self.config.metadata, dict):
            if "version" in self.config.metadata:
                self.VERSION = self.config.metadata["version"]

        if self.config.output_type is not None:
            if self.config.output_type not in ALL_OUTPUT_TYPES:
                raise ValueError(
                    f"Got invalid output_type '{self.config.output_type}', must be in '{','.join(ALL_OUTPUT_TYPES)}'"
                )
            self.OUTPUT_TYPE = self.config.output_type

        if self.config.dataset_path is not None:
            self.DATASET_PATH = self.config.dataset_path

        if self.config.dataset_name is not None:
            self.DATASET_NAME = self.config.dataset_name

        self._metric_fn_list = {}
        self._metric_fn_kwargs = {}
        self._aggregation_list = {}
        self._higher_is_better = {}

        if self.config.metric_list is None:
            # TODO: handle this in TaskConfig.__post_init__ ?
            _metric_list = DEFAULT_METRIC_REGISTRY[self.config.output_type]

            for metric_name in _metric_list:
                self._metric_fn_list[metric_name] = get_metric(metric_name)
                self._metric_fn_kwargs[metric_name] = {}
                self._aggregation_list[metric_name] = get_metric_aggregation(
                    metric_name
                )
                self._higher_is_better[metric_name] = is_higher_better(metric_name)
        else:
            for metric_config in self.config.metric_list:
                if "metric" not in metric_config:
                    raise ValueError(
                        "'metric' key not provided for an entry in 'metric_list', must be specified!"
                    )
                metric_name = metric_config["metric"]
                kwargs = {
                    key: metric_config[key]
                    for key in metric_config
                    if key
                    not in ["metric", "aggregation", "higher_is_better", "hf_evaluate"]
                }
                hf_evaluate_metric = (
                    "hf_evaluate" in metric_config
                    and metric_config["hf_evaluate"] is True
                )

                if self.config.process_results is not None:
                    self._metric_fn_list[metric_name] = None
                    self._metric_fn_kwargs[metric_name] = {}
                elif callable(metric_name):
                    metric_fn = metric_name.__call__
                    metric_name = metric_name.__name__
                    self._metric_fn_list[metric_name] = metric_fn
                    self._metric_fn_kwargs[metric_name] = kwargs
                else:
                    self._metric_fn_list[metric_name] = get_metric(
                        metric_name, hf_evaluate_metric
                    )
                    self._metric_fn_kwargs[metric_name] = kwargs

                if "aggregation" in metric_config:
                    agg_name = metric_config["aggregation"]
                    if isinstance(agg_name, str):
                        self._aggregation_list[metric_name] = get_aggregation(agg_name)
                    elif callable(agg_name):  # noqa: E721
                        self._aggregation_list[metric_name] = metric_config[
                            "aggregation"
                        ]
                else:
                    INV_AGG_REGISTRY = {v: k for k, v in AGGREGATION_REGISTRY.items()}
                    metric_agg = get_metric_aggregation(metric_name)
                    eval_logger.warning(
                        f"[Task: {self.config.task}] metric {metric_name} is defined, but aggregation is not. "
                        f"using default "
                        f"aggregation={INV_AGG_REGISTRY[metric_agg]}"
                    )
                    self._aggregation_list[metric_name] = metric_agg

                if "higher_is_better" in metric_config:
                    self._higher_is_better[metric_name] = metric_config[
                        "higher_is_better"
                    ]
                else:
                    eval_logger.warning(
                        f"[Task: {self.config.task}] metric {metric_name} is defined, but higher_is_better is not. "
                        f"using default "
                        f"higher_is_better={is_higher_better(metric_name)}"
                    )
                    self._higher_is_better[metric_name] = is_higher_better(metric_name)

        self.download(self.config.dataset_kwargs)
        self._training_docs = None
        self._fewshot_docs = None

        if self.config.filter_list is not None:
            self._filters = []
            for filter_config in self.config.filter_list:
                filter_name = filter_config["name"]
                filter_functions = filter_config["filter"]
                components = []
                for function in filter_functions:
                    kwargs = {
                        key: function[key] for key in function if key != "function"
                    }
                    components.append([function["function"], kwargs])
                filter_pipeline = build_filter_ensemble(filter_name, components)
                self._filters.append(filter_pipeline)
        else:
            self._filters = [build_filter_ensemble("none", [["take_first", None]])]

        if self.config.use_prompt is not None:
            eval_logger.info(f"loading prompt {self.config.use_prompt}")
            self.prompt = get_prompt(
                self.config.use_prompt, self.DATASET_PATH, self.DATASET_NAME
            )
        else:
            self.prompt = None

        if self.fewshot_docs() is not None:
            self.fewshot_rnd = (
                random.Random()
            )  # setting with no seed, to be overridden at a later time
            config_sampler: Union[str, Callable] = (
                self.config.fewshot_config.get("sampler", "default")
                if self.config.fewshot_config
                else "default"
            )
            if isinstance(config_sampler, str):
                self.sampler = samplers.get_sampler(config_sampler)(
                    list(self.fewshot_docs()), self, rnd=self.fewshot_rnd
                )
            elif callable(config_sampler) and issubclass(
                config_sampler, samplers.ContextSampler
            ):
                self.sampler = config_sampler(
                    docs=list(self.fewshot_docs()), task=self, rnd=self.fewshot_rnd
                )
            else:
                raise TypeError(
                    f"fewshot_config.sampler should be a string or callable of ContextSampler type, "
                    f"not {type(config_sampler)}"
                )

        self.task_docs = self.eval_docs

        # Test One Doc
        self.features = list(self.task_docs.features.keys())
        self.multiple_input = 0
        self.multiple_target = 0
        test_doc = self.task_docs[0]
        test_text = self.doc_to_text(test_doc)
        test_target = self.doc_to_target(test_doc)

        if self.config.doc_to_choice is not None:
            test_choice = self.doc_to_choice(test_doc)
            if not isinstance(test_choice, list):
                eval_logger.error("doc_to_choice must return list")
            else:
                num_choice = len(test_choice)

            if isinstance(test_text, int):
                self.multiple_input = num_choice
        else:
            test_choice = None

        if isinstance(test_target, list):
            self.multiple_target = len(test_target)
        else:
            if (isinstance(test_target, int)) and (test_choice is not None):
                test_target = test_choice[test_target]
            else:
                test_target = str(test_target)

        if test_choice is not None:
            check_choices = test_choice
        else:
            check_choices = [test_target]
        if self.config.doc_to_choice is not None:
            for choice in check_choices:
                choice_has_whitespace = True if choice[0].isspace() else False
                delimiter_has_whitespace = (
                    True
                    if self.config.target_delimiter.rstrip()
                    != self.config.target_delimiter
                    else False
                )

                if delimiter_has_whitespace and choice_has_whitespace:
                    eval_logger.debug(
                        f'Both target_delimiter "{self.config.target_delimiter}" and target choice: "{choice}" have whitespace'
                    )
                elif (not delimiter_has_whitespace) and (not choice_has_whitespace):
                    eval_logger.debug(
                        f'Both target_delimiter "{self.config.target_delimiter}" and target choice: "{choice}" do not have whitespace, ignore if the language you are evaluating on does not require/use whitespace'
                    )

    def download(self, dataset_kwargs: Optional[Dict[str, Any]] = None) -> None:
        self.dataset = datasets.load_dataset(
            path=self.DATASET_PATH,
            name=self.DATASET_NAME,
            **dataset_kwargs if dataset_kwargs is not None else {},
        )


    def has_training_docs(self) -> bool:
        if self.config.training_split is not None:
            return True
        else:
            return False

    def has_validation_docs(self) -> bool:
        if self.config.validation_split is not None:
            return True
        else:
            return False

    def has_test_docs(self) -> bool:
        if self.config.test_split is not None:
            return True
        else:
            return False

    def training_docs(self) -> datasets.Dataset:
        if self.has_training_docs():
            if self.config.process_docs is not None:
                return self.config.process_docs(
                    self.dataset[self.config.training_split]
                )
            return self.dataset[self.config.training_split]

    def validation_docs(self) -> datasets.Dataset:
        if self.has_validation_docs():
            if self.config.process_docs is not None:
                return self.config.process_docs(
                    self.dataset[self.config.validation_split]
                )
            return self.dataset[self.config.validation_split]

    def test_docs(self) -> datasets.Dataset:
        if self.has_test_docs():
            if self.config.process_docs is not None:
                return self.config.process_docs(self.dataset[self.config.test_split])
            return self.dataset[self.config.test_split]

    def fewshot_docs(self):
        if self.config.fewshot_split is not None:
            if self.config.process_docs is not None:
                return self.config.process_docs(self.dataset[self.config.fewshot_split])
            return self.dataset[self.config.fewshot_split]
        elif (
            self.config.fewshot_config is not None
            and self.config.fewshot_config.get("samples", None) is not None
        ):
            if isinstance(self.config.fewshot_config["samples"], list):
                return self.config.fewshot_config["samples"]
            elif callable(self.config.fewshot_config["samples"]):
                return self.config.fewshot_config["samples"]()
            else:
                raise Exception(
                    "`fewshot_config['samples']` was incorrectly defined in the configuration. It should be either a list of samples as a dict, or function returning this list."
                )
        else:
            if (self.config.num_fewshot is not None) and (self.config.num_fewshot > 0):
                eval_logger.warning(
                    f"[Task: {self.config.task}] "
                    "num_fewshot > 0 but fewshot_split is None. "
                    "using preconfigured rule."
                )
            return super().fewshot_docs()

    def fewshot_context(
        self,
        doc: str,
        num_fewshot: int,
    ) -> str:
        """Returns a fewshot context string that is made up of a prepended description
        (if provided), the `num_fewshot` number of examples, and an appended prompt example.

        :param doc: str
            The document as returned from training_docs, validation_docs, or test_docs.
        :param num_fewshot: int
            The number of fewshot examples to provide in the returned context string.
        """

        labeled_examples = ""

        # get task description
        if description := self.config.description:
            description = utils.apply_template(self.config.description, doc)

        # create system prompt based on the provided system instruction and description
        if description:
            system_prompt = description
        else:
            system_prompt = ""

        # add system prompt if specified
        if system_prompt:
            labeled_examples = system_prompt

        # if few-shot - append examples after the system prompt
        if num_fewshot > 0:
            labeled_examples += self.sampler.get_context(doc, num_fewshot)

        example = self.doc_to_text(doc)

        if self.multiple_input:
            return labeled_examples
        if isinstance(example, str):
            return labeled_examples + example
        elif isinstance(example, list):
            return [labeled_examples + ex for ex in example]
        elif isinstance(example, int):
            if self.config.doc_to_choice is not None:
                choices = self.doc_to_choice(doc)
                return labeled_examples + choices[example]
            else:
                return labeled_examples + str(example)

    def apply_filters(self):
        """Iterates over FilterEnsembles and applies them to instances"""
        if hasattr(self, "_filters"):
            for f in self._filters:
                f.apply(self._instances)
        else:
            eval_logger.warning("No filter defined, passing through instances")
            return self._instances

    def should_decontaminate(self):
        return self.config.should_decontaminate

    def doc_to_decontamination_query(self, doc):
        if self.config.should_decontaminate:
            if self.config.doc_to_decontamination_query is None:
                return self.doc_to_text(doc)
            else:
                doc_to_decontamination_query = self.config.doc_to_decontamination_query
                if doc_to_decontamination_query in self.features:
                    return doc[doc_to_decontamination_query]
                elif callable(doc_to_decontamination_query):
                    return doc_to_decontamination_query(doc)
                else:
                    return ast.literal_eval(
                        utils.apply_template(
                            self.config.doc_to_decontamination_query, doc
                        )
                    )

    def _process_doc(self, doc: dict) -> dict:
        """
        Override this to process (detokenize, strip, replace, etc.) individual
        documents. This can be used in a map over documents of a data split.
        E.g. `map(self._process_doc, self.dataset["validation"])`

        :return: dict
            The processed version of the specified `doc`.
        """
        return doc

    def doc_to_text(self, doc, doc_to_text=None):
        if self.prompt is not None:
            doc_to_text = self.prompt
        elif doc_to_text is not None:
            doc_to_text = doc_to_text
        else:
            doc_to_text = self.config.doc_to_text

        if isinstance(doc_to_text, int):
            return doc_to_text
        elif isinstance(doc_to_text, str):
            if doc_to_text in self.features:
                # if self.config.doc_to_choice is not None:
                #     return self.doc_to_choice(doc)[doc[doc_to_text]]
                # else:
                return doc[doc_to_text]
            else:
                text_string = utils.apply_template(doc_to_text, doc)
                if text_string.isdigit() and self.config.doc_to_choice is not None:
                    return ast.literal_eval(text_string)
                else:
                    return text_string
        elif callable(doc_to_text):
            return doc_to_text(doc)
        # Used when applying a Promptsource template
        elif hasattr(doc_to_text, "apply"):
            applied_prompt = doc_to_text.apply(doc)
            if len(applied_prompt) == 2:
                return applied_prompt[0]
            else:
                eval_logger.warning("Applied prompt returns empty string")
                return self.config.fewshot_delimiter
        else:
            print(type(doc_to_text))
            raise TypeError

    def doc_to_target(self, doc: Mapping, doc_to_target=None) -> Union[int, str, list]:
        if self.prompt is not None:
            doc_to_target = self.prompt
        elif doc_to_target is not None:
            doc_to_target = doc_to_target
        else:
            doc_to_target = self.config.doc_to_target

        if isinstance(doc_to_target, int):
            return doc_to_target
        elif isinstance(doc_to_target, str):
            if doc_to_target in self.features:
                # if self.config.doc_to_choice is not None:
                #     return self.doc_to_choice(doc)[doc[doc_to_target]]
                # else:
                return doc[doc_to_target]
            else:
                target_string = utils.apply_template(doc_to_target, doc)
                if target_string.isdigit() and self.config.doc_to_choice is not None:
                    return ast.literal_eval(target_string)
                elif (
                    len(target_string) >= 2
                    and (target_string[0] == "[")
                    and (target_string[-1] == "]")
                ):
                    try:
                        return ast.literal_eval(target_string)
                    except (SyntaxError, ValueError):
                        return target_string
                else:
                    return target_string
        elif isinstance(doc_to_target, list):
            return doc_to_target
        elif callable(doc_to_target):
            return doc_to_target(doc)
        # Used when applying a Promptsource template
        elif hasattr(doc_to_target, "apply"):
            applied_prompt = doc_to_target.apply(doc)
            if len(applied_prompt) == 2:
                return applied_prompt[1]
            else:
                eval_logger.warning("Applied prompt returns empty string")
                return self.config.fewshot_delimiter
        else:
            raise TypeError

    def doc_to_choice(self, doc: Any, doc_to_choice=None) -> List[str]:
        if self.prompt is not None:
            doc_to_choice = self.prompt
        elif doc_to_choice is not None:
            doc_to_choice = doc_to_choice
        elif self.config.doc_to_choice is None:
            eval_logger.error("doc_to_choice was called but not set in config")
        else:
            doc_to_choice = self.config.doc_to_choice

        if isinstance(doc_to_choice, str):
            if doc_to_choice in self.features:
                return doc[doc_to_choice]
            else:
                return ast.literal_eval(utils.apply_template(doc_to_choice, doc))
        elif isinstance(doc_to_choice, list):
            return doc_to_choice
        elif isinstance(doc_to_choice, dict):
            return list(doc_to_choice.values())
        elif callable(doc_to_choice):
            return doc_to_choice(doc)
        elif hasattr(doc_to_choice, "get_answer_choices_list"):
            return doc_to_choice.get_answer_choices_list(doc)
        else:
            raise TypeError

    def construct_requests(
        self, doc: dict, ctx: str, **kwargs
    ) -> Union[List[Instance], Instance]:

        if self.OUTPUT_TYPE == "multiple_choice":
            choices = self.doc_to_choice(doc)
            target_delimiter = self.config.target_delimiter
            if self.multiple_input:
                # If there are multiple inputs, choices are placed in the ctx
                cont = self.doc_to_target(doc)
                arguments = [
                    (ctx + choice, f"{target_delimiter}{cont}") for choice in choices
                ]
            else:
                # Otherwise they are placed in the continuation
                arguments = [(ctx, f"{target_delimiter}{cont}") for cont in choices]

            request_list = [
                Instance(
                    request_type="loglikelihood",
                    doc=doc,
                    arguments=arg,
                    idx=i,
                    **kwargs,
                )
                for i, arg in enumerate(arguments)
            ]
            # TODO: we should raise a warning telling users this will at most ~2x runtime.
            if "acc_mutual_info" in self._metric_fn_list.keys():
                # if we are calculating multiple choice accuracy
                # using mutual information instead of raw loglikelihood as metric, need unconditional lls.

                # here mutual info refers to calculating
                # log(P(choice|ctx) / P(choice)) = log(P(choice|ctx)) - log(P(choice))
                # in other words normalizing by subtracting the unconditional logprob of each choice.
                request_list.extend(
                    [
                        Instance(
                            request_type="loglikelihood",
                            doc=doc,
                            arguments=("", "{}".format(choice)),
                            idx=i,
                            **kwargs,
                        )
                        for i, choice in enumerate(choices)
                    ]
                )
            return request_list

        elif self.OUTPUT_TYPE == "generate_until":
            arguments = (ctx, deepcopy(self.config.generation_kwargs))

        return Instance(
            request_type=self.OUTPUT_TYPE, doc=doc, arguments=arguments, idx=0, **kwargs
        )

    def process_results(self, doc, results):
        if callable(self.config.process_results):
            return self.config.process_results(doc, results)

        result_dict = {}
        use_metric = list(self._metric_fn_list.keys())

        if self.OUTPUT_TYPE == "multiple_choice":
            lls, is_greedy = zip(*results)

            # retrieve choices in List[str] form, to compute choice lengths, etc.
            choices = self.doc_to_choice(doc)
            completion_len = np.array([float(len(i)) for i in choices])

            if (
                2 * len(choices) == len(lls)
                and "acc_mutual_info" in self._metric_fn_list.keys()
            ):
                # then we are doing mutual info.
                # this stores the "dryrun" / unconditional answer loglikelihoods
                lls_unconditional = lls[1::2]
                if len(lls_unconditional) != len(choices):
                    raise ValueError
                # and this stores our "regular" conditional loglikelihoods
                lls = lls[::2]

            pred = np.argmax(lls)
            pred_norm = np.argmax(lls / completion_len)

            if self.multiple_input:
                gold = self.doc_to_text(doc)
            else:
                gold = self.doc_to_target(doc)

            gold_index_error = False
            if isinstance(gold, list):
                gold = [i if i < len(choices) else -100 for i in gold]
                if -100 in gold:
                    gold_index_error = True
            else:
                if isinstance(gold, int):
                    gold = gold if gold < len(choices) else -100
                elif isinstance(gold, str):
                    gold = choices.index(gold) if gold in choices else -100

                if gold == -100:
                    gold_index_error = True

            if gold_index_error:
                eval_logger.warning(
                    f"Label index was not in within range of available choices,"
                    f"Sample:\n\n{doc}\n\n"
                )

            if self.multiple_target:
                acc = 1.0 if pred in gold else 0.0
                acc_norm = 1.0 if pred_norm in gold else 0.0
                exact_match = int(any([is_greedy[i] if i != -100 else 0 for i in gold]))
            else:
                acc = 1.0 if pred == gold else 0.0
                acc_norm = 1.0 if pred_norm == gold else 0.0
                # TODO: this gets score of 0 on arc_challenge for pythia-70m. need to test that this works properly
                exact_match = int(is_greedy[gold]) if gold != -100 else 0

            prob_norm = utils.softmax(lls)

            # TODO use keyword arguments to the metric?
            # gold, pred, norm stuff, the original lls,
            result_dict = {
                **({"acc": acc} if "acc" in use_metric else {}),
                **({"f1": (gold, pred)} if "f1" in use_metric else {}),
                **({"mcc": (gold, pred)} if "mcc" in use_metric else {}),
                **({"acc_norm": acc_norm} if "acc_norm" in use_metric else {}),
                **({"exact_match": exact_match} if "exact_match" in use_metric else {}),
                **(
                    {"brier_score": (gold, prob_norm)}
                    if "brier_score" in use_metric
                    else {}
                ),
            }

            if "acc_mutual_info" in use_metric:
                lls_mutual_info = [
                    ll_c - ll_u for ll_c, ll_u in zip(lls, lls_unconditional)
                ]
                acc_mutual_info = 1.0 if np.argmax(lls_mutual_info) == gold else 0.0
                result_dict["acc_mutual_info"] = acc_mutual_info

        elif self.OUTPUT_TYPE == "generate_until":
            gold = self.doc_to_target(doc)
            result = results[0]
            if self.config.doc_to_choice is not None:
                # If you set doc_to_choice,
                # it assumes that doc_to_target returns a number.
                choices = self.doc_to_choice(doc)
                gold = choices[gold]
            # we expect multiple_targets to be a list.
            elif self.multiple_target:
                gold = list(gold)
            elif type(gold) != type(result):
                # cast gold to the same type as result
                gold = type(result)(gold)

            for metric in self._metric_fn_list.keys():
                if self.multiple_target:
                    # in the case where we have multiple targets,
                    # return true if any are true
                    # TODO: this may break for multipLe_target, non zero-or-1 metrics
                    scores = []
                    if not isinstance(gold, list):
                        # sometimes, a multiple_target dataset has exceptions where one doc has only one string answer
                        # print(gold)
                        gold = [gold]
                    if metric == "exact_match":
                        result = [result for _ in range(len(gold))]
                        scores = self._metric_fn_list[metric](
                            references=gold,
                            predictions=result,
                            **self._metric_fn_kwargs[metric],
                        )[metric]
                        result_score = 1.0 if scores > 0.0 else 0.0
                    else:
                        for gold_option in gold:
                            try:
                                result_score = self._metric_fn_list[metric](
                                    references=[gold_option],
                                    predictions=[result],
                                    **self._metric_fn_kwargs[metric],
                                )
                            except (
                                TypeError
                            ):  # TODO: this is hacky and I don't want to do it
                                result_score = self._metric_fn_list[metric](
                                    [gold_option, result]
                                )
                            if isinstance(result_score, dict):
                                # TODO: this handles the case where HF evaluate returns a dict.
                                result_score = result_score[metric]
                            scores.append(result_score)
                        if any(scores):
                            result_score = 1.0
                        else:
                            result_score = 0.0
                else:
                    try:
                        result_score = self._metric_fn_list[metric](
                            references=[gold],
                            predictions=[result],
                            **self._metric_fn_kwargs[metric],
                        )
                    except TypeError:  # needed for now in order to use a different interface between our own metrics and HF Evaluate metrics
                        result_score = self._metric_fn_list[metric]([gold, result])
                    if isinstance(result_score, dict):
                        # TODO: this handles the case where HF evaluate returns a dict.
                        result_score = result_score[metric]
                result_dict[metric] = result_score
        else:
            raise ValueError(
                f"Passed invalid output_type '{self.OUTPUT_TYPE}' ! Please use one of ",
                "'generate_until' or 'multiple_choice'",
            )

        return result_dict

    def aggregation(self) -> dict:
        return self._aggregation_list

    def higher_is_better(self) -> dict:
        return self._higher_is_better

    def get_config(self, key: str) -> Any:
        return getattr(self.config, key, None)

    @property
    def task_name(self) -> Any:
        return getattr(self.config, "task", None)

    def __repr__(self):
        return (
            f"Task(task_name={getattr(self.config, 'task', None)},"
            f"output_type={self.OUTPUT_TYPE},"
            f"num_fewshot={getattr(self.config, 'num_fewshot', None)},"
            f"num_samples={len(self.eval_docs)})"
        )

    def fewshot_docs(self) -> Iterable:
        """
        :return: Iterable[obj]
            A iterable of any object, that doc_to_text can handle
        """
        if self.has_training_docs():
            return self.training_docs()
        elif self.has_validation_docs():
            return self.validation_docs()
        else:
            # eval_logger.warning(
            #     f"[Task: {self.config.task}] has_training_docs and has_validation_docs are False"
            #     ", using test_docs as fewshot_docs but this is not recommended."
            # )
            return self.test_docs()

    @property
    def eval_docs(self) -> Union[datasets.Dataset, List[dict]]:
        if self.has_test_docs():
            return self.test_docs()
        elif self.has_validation_docs():
            return self.validation_docs()
        else:
            raise ValueError(
                f"Task dataset (path={self.DATASET_PATH}, name={self.DATASET_NAME}) must have valid or test docs!"
            )

    def set_fewshot_seed(self, seed: Optional[int] = None) -> None:
        self.fewshot_rnd = random.Random(seed)
        if hasattr(self, "sampler"):
            self.sampler.rnd = self.fewshot_rnd

    def dump_config(self) -> dict:
        """Returns the config as a dictionary."""
        # TODO: this should only return the overrides applied to a non-YAML task's configuration.
        # (num_fewshot)
        return self.config.to_dict()

    def build_all_requests(
        self,
        *,
        limit: Union[int, None] = None,
        rank: int = 0,
        world_size: int = 1,
    ) -> None:
        """Build a set of Instances for a task, and store them in task.instances"""

        # used with caching
        og_limit = limit

        cache_key = f"requests-{self.config.task}-{self.config.num_fewshot}shot-rank{rank}-world_size{world_size}"

        eval_logger.info(f"Building contexts for {self.config.task} on rank {rank}...")

        instances = []

        doc_id_docs = list(
            self.doc_iterator(rank=rank, limit=limit, world_size=world_size)
        )

        num_docs = len(doc_id_docs)

        for doc_id, doc in tqdm(
            doc_id_docs,
            total=num_docs,
        ):
            # sample fewshot context #TODO: need to offset doc_id by rank now!
            fewshot_ctx = self.fewshot_context(
                doc,
                0 if self.config.num_fewshot is None else self.config.num_fewshot,
            )

            # TODO: we should override self.config.repeats if doing greedy gen so users don't waste time+compute
            inst = self.construct_requests(
                doc=doc,
                ctx=fewshot_ctx,
                metadata=(self.config["task"], doc_id, self.config.repeats),
            )

            if not isinstance(inst, list):
                inst = [inst]

            instances.append(inst)

        # now flatten, this is to allow slicing to work with pickles

        sliced_instances = instances[:og_limit]

        flattened_instances = [
            instance
            for instance_group in sliced_instances
            for instance in instance_group
        ]

        self._instances = flattened_instances

        if len(self._instances) == 0:
            raise ValueError("task.build_requests() did not find any docs!")
        
    def doc_iterator(
        self, *, rank: int = 0, limit: Union[int, None] = None, world_size: int = 1
    ) -> Iterator[Tuple[int, Any]]:
        limit = int(limit) if limit else None
        doc_iterator = utils.create_iterator(
            enumerate(self.eval_docs),
            rank=int(rank),
            limit=limit,
            world_size=int(world_size),
        )
        return doc_iterator

    @property
    def instances(self) -> List[Instance]:
        """After calling `task.build_all_requests()`, tasks
        maintain a list of the dataset instances which will be evaluated.
        """
        return self._instances

    def set_config(self, key: str, value: Any, update: bool = False) -> None:
        """Set or update the configuration for a given key."""
        if key is None:
            raise ValueError("Key must be provided.")

        if update:
            current_value = getattr(self.config, key, {})
            if not isinstance(current_value, dict):
                raise TypeError(
                    f"Expected a dict for key '{key}', got {type(current_value).__name__} instead."
                )
            current_value.update(value)
        else:
            setattr(self.config, key, value)