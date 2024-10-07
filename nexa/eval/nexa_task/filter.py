import re
import sys
import unicodedata
from collections import Counter
from functools import partial
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Callable, Iterable, List, Union

from nexa.eval.nexa_task.instance import Instance
from nexa.eval.nexa_task.registry import register_filter, get_filter


class Filter(ABC):
    """
    Filter classes operate on a per-task level.
    They take all model outputs (`instance.resps` for all `task.instances`)
    across all instances of a task, and perform operations.
    In a single run, one can configure any number of separate filters or lists of filters.

    """

    def __init__(self, **kwargs) -> None:
        """
        Can define custom behavior here, if an individual instantiation of a Filter class should have state.
        """

    @abstractmethod
    def apply(self, resps: Union[List, Iterable], docs: List[dict]) -> Iterable:
        """
        Defines the operation to perform on a list of the `inst.resps` properties of `Instance` objects.
        Should return the list of (filtered) response lists *in the same order as they were input*, e.g.
        if pass in [<inst.resps for instance 0>, <inst.resps for instance 1>] should return
        [<filtered resps for instance 0>, <filtered resps for instance 1>]
        """
        return resps


@register_filter("regex")
class RegexFilter(Filter):
    """ """

    def __init__(
        self,
        regex_pattern: str = r"#### (\-?[0-9\.\,]+)",
        group_select=0,
        fallback: str = "[invalid]",
    ) -> None:
        """
        pass a string `regex` to run `re.compile(r"regex")` on.
        `fallback` defines the output returned if no matches for the regex are located.
        """
        self.regex_pattern = regex_pattern
        self.regex = re.compile(regex_pattern)
        self.group_select = group_select
        self.fallback = fallback

    def apply(self, resps, docs):
        # here, we assume we have a list, in which each element is
        # a list of model responses for some particular input/target pair.
        # so we process each of these (same input/target response sets)
        # independently (and keep them a list.)
        def filter_set(inst):
            filtered = []
            for resp in inst:
                match = self.regex.findall(resp)
                if match:
                    match = match[self.group_select]
                    if isinstance(match, tuple):
                        match = [m for m in match if m][0]
                    match = match.strip()
                else:
                    match = self.fallback
                filtered.append(match)
            return filtered

        # print(resps)
        filtered_resps = list(map(lambda x: filter_set(x), resps))
        # print(filtered_resps)

        return filtered_resps


@register_filter("remove_whitespace")
class WhitespaceFilter(Filter):
    """ """

    def __init__(self) -> None:
        pass

    def apply(self, resps, docs):
        def filter_set(inst):
            filtered_resp = []
            for resp in inst:
                resp = resp.lstrip()
                filtered_resp.append(resp)
            return filtered_resp

        filtered_resps = [filter_set(resp) for resp in resps]

        return filtered_resps


@register_filter("take_first")
class TakeFirstFilter(Filter):
    def __init__(self) -> None:
        """
        Can define custom behavior here, if an individual instantiation of a Filter class should have state.
        """

    def apply(self, resps, docs):
        """
        Assuming each entry of `resps` is a list of model responses, we discard all but the first response.
        """
        return map(lambda r: r[0], resps)


@dataclass
class FilterEnsemble:
    """
    FilterEnsemble creates a pipeline applying multiple filters.
    Its intended usage is to stack multiple post-processing steps in order.
    `task.apply_filters` should use a list of FilterEnsemble classes that it stores, to apply each
    pipeline separately.
    """

    name: str
    filters: List[Callable[[], Filter]]

    def apply(self, instances: List[Instance]) -> None:
        resps, docs = zip(*((inst.resps, inst.doc) for inst in instances))
        resps, docs = list(resps), list(docs)

        for f in self.filters:
            # apply filters in sequence
            resps = f().apply(resps, docs)

        # add the end results after filtering to filtered_requests of their respective source instances.
        # has key `self.name`: each FilterEnsemble applied in a given run should use a different name.
        for inst, resp in zip(instances, resps):
            inst.filtered_resps[self.name] = resp


def build_filter_ensemble(
    filter_name: str, components: List[List[str]]
) -> FilterEnsemble:
    """
    Create a filtering pipeline.
    """
    filters = []
    for function, kwargs in components:
        if kwargs is None:
            kwargs = {}
        # create a filter given its name in the registry
        f = partial(get_filter(function), **kwargs)
        # add the filter as a pipeline step
        filters.append(f)

    return FilterEnsemble(name=filter_name, filters=filters)
