import os
from typing import Dict

from nexa.eval.utils import eval_logger

PROMPT_REGISTRY: Dict[str, Dict[str, str]] = {
    "qa-basic": {
        "question-newline-answer": "Question: {{question}}\nAnswer:",
        "q-newline-a": "Q: {{question}}\nA:",
    },
}

def get_prompt(prompt_id: str, dataset_name: str = None, subset_name: str = None):
    # Unpack prompt name
    try:
        category_name, prompt_name = prompt_id.split(":")
    except ValueError:
        raise ValueError(
            f"Expected prompt_id in the format 'category_name:prompt_name', but got '{prompt_id}'"
        )

    # Construct the dataset full name
    dataset_full_name = (
        f"{dataset_name}-{subset_name}" if subset_name else dataset_name
    )
    eval_logger.info(f"Loading prompt from '{category_name}' for '{dataset_full_name}'")

    # Check if the category is a YAML file
    if category_name.endswith(".yaml"):
        import yaml

        if not os.path.exists(category_name):
            raise FileNotFoundError(f"YAML file '{category_name}' not found.")

        with open(category_name, "r") as file:
            prompt_yaml_file = yaml.safe_load(file)

        prompt_string = prompt_yaml_file.get("prompts", {}).get(prompt_name)
        if prompt_string is None:
            raise ValueError(
                f"Prompt '{prompt_name}' not found in YAML file '{category_name}'."
            )
        return prompt_string
    else:
        # Retrieve the prompt from the registry
        try:
            prompt_string = PROMPT_REGISTRY[category_name][prompt_name]
            return prompt_string
        except KeyError:
            raise ValueError(
                f"Prompt '{prompt_name}' not found in category '{category_name}'."
            )
