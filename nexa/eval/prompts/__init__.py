import ast
import os
from typing import Dict

from nexa.eval import utils
from nexa.eval.utils import eval_logger

PROMPT_REGISTRY: Dict[str, Dict[str, str]] = {
    "qa-basic": {
        "question-newline-answer": "Question: {{question}}\nAnswer:",
        "q-newline-a": "Q: {{question}}\nA:",
    },
}


def get_prompt(prompt_id: str, dataset_name: str = None, subset_name: str = None):
    # unpack prompt name
    category_name, prompt_name = prompt_id.split(":")
    if subset_name is None:
        dataset_full_name = dataset_name
    else:
        dataset_full_name = f"{dataset_name}-{subset_name}"
    eval_logger.info(f"Loading prompt from {category_name} for {dataset_full_name}")
    if ".yaml" in category_name:
        import yaml

        with open(category_name, "rb") as file:
            prompt_yaml_file = yaml.full_load(file)

        prompt_string = prompt_yaml_file["prompts"][prompt_name]
        return PromptString(prompt_string)
    else:
        try:
            return PROMPT_REGISTRY[category_name][prompt_name]
        except Exception:
            raise ValueError(
                f"expected only a single `:` as separator between \
                prompt category and name, but got `{prompt_id}` instead"
            )


class PromptString:
    def __init__(self, prompt_string):
        self.prompt_string = prompt_string

    def apply(self, doc):
        doc_to_text = self.prompt_string["doc_to_text"]
        doc_to_target = self.prompt_string["doc_to_target"]

        # TODO need a way to process doc_to_choice
        if "doc_to_choice" in self.prompt_string:
            raise Exception("Not yet implemented to accept doc_to_choice")

        text_string = utils.apply_template(doc_to_text, doc)
        target_string = utils.apply_template(doc_to_target, doc)

        return [text_string, target_string]
