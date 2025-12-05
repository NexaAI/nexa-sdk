"""Function call parser."""

import json
import re
from typing import Optional, Tuple, Dict, Any


def extract_function_call(text: str) -> Optional[Tuple[str, Dict[str, Any]]]:
    """Extract function call JSON from LLM response."""
    text = text.strip()
    text = re.sub(r"<\|[^|]+\|>", "", text)
    text = text.strip()

    # Pattern matches: { "name": "...", "arguments": { ... } }
    json_pattern = r'\{[^{}]*"name"\s*:\s*"([^"]+)"[^{}]*"arguments"\s*:\s*(\{[^{}]*\})[^{}]*\}'
    match = re.search(json_pattern, text, re.DOTALL)
    if match:
        func_name = match.group(1)
        try:
            args_str = match.group(2)
            func_args = json.loads(args_str)
            return func_name, func_args
        except json.JSONDecodeError:
            pass

    # Fallback: try to parse the entire cleaned text as JSON
    try:
        parsed = json.loads(text)
        if isinstance(parsed, dict) and "name" in parsed:
            func_name = parsed.get("name")
            func_args = parsed.get("arguments", {})
            if func_name and isinstance(func_name, str):
                return func_name, func_args
    except json.JSONDecodeError:
        pass

    return None


