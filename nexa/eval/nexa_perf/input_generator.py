import logging
import random
import string
from abc import ABC
from typing import List, Tuple, Any, Dict

import numpy as np

LOGGER = logging.getLogger("generator")

DEFAULT_VOCAB_SIZE = 2
DEFAULT_TYPE_VOCAB_SIZE = 2

class TaskGenerator(ABC):
    def __init__(self, shapes, with_labels: bool):
        self.shapes = shapes
        self.with_labels = with_labels

    @staticmethod
    def generate_random_integers(min_value: int, max_value: int, shape: Tuple[int]):
        return np.random.randint(min_value, max_value, size=shape)

    @staticmethod
    def generate_random_floats(min_value: float, max_value: float, shape: Tuple[int]):
        return np.random.rand(*shape) * (max_value - min_value) + min_value

    @staticmethod
    def generate_ranges(start: int, stop: int, shape: Tuple[int]):
        return np.tile(np.arange(start, stop), (shape[0], 1))

    @staticmethod
    def generate_random_strings(num_seq: int) -> List[str]:
        return [
            "".join(random.choice(string.ascii_letters + string.digits) for _ in range(random.randint(10, 100)))
            for _ in range(num_seq)
        ]

    def __call__(self):
        raise NotImplementedError("Generator must implement __call__ method")

class TextGenerator(TaskGenerator):
    def input_ids(self):
        return self.generate_random_integers(
            min_value=0,
            max_value=self.shapes["vocab_size"] or DEFAULT_VOCAB_SIZE,
            shape=(self.shapes["batch_size"], self.shapes["sequence_length"]),
        )

    def attention_mask(self):
        return self.generate_random_integers(
            min_value=1,  # avoid sparse attention
            max_value=2,
            shape=(self.shapes["batch_size"], self.shapes["sequence_length"]),
        )

    def token_type_ids(self):
        return self.generate_random_integers(
            min_value=0,
            max_value=self.shapes["type_vocab_size"] or DEFAULT_TYPE_VOCAB_SIZE,
            shape=(self.shapes["batch_size"], self.shapes["sequence_length"]),
        )

    def position_ids(self):
        return self.generate_ranges(
            start=0,
            stop=self.shapes["sequence_length"],
            shape=(self.shapes["batch_size"], self.shapes["sequence_length"]),
        )

    def requires_token_type_ids(self):
        return self.shapes["type_vocab_size"] is not None and self.shapes["type_vocab_size"] > 1

    def requires_position_ids(self):
        return self.shapes["max_position_embeddings"] is not None

    def __call__(self):
        dummy = {
            "input_ids": self.input_ids(),
            "attention_mask": self.attention_mask()
        }

        if self.with_labels:
            dummy["labels"] = self.input_ids()

        return dummy

class InputGenerator:
    def __init__(self, task: str, input_shapes: Dict[str, int]) -> None:
        shapes = {**input_shapes}
        self.task_generator = TextGenerator(shapes=shapes, with_labels=False)

    def __call__(self) -> Dict[str, Any]:
        return self.task_generator()