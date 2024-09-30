from collections import Counter

from nexa.eval.api.filter import Filter
from nexa.eval.api.registry import register_filter


# TODO: implement "arg_max" filter. either it should take in an arbitrary "scoring"/reward function
# that takes an input and returns a scalar and then should select the max reward,
# or should implement different filters for different ways of handling a reward model's inference.


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
