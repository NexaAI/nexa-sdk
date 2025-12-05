"""Function call parser for extracting function calls from model output."""

import json
import re
from typing import Optional, Tuple, Dict, Any


class FunctionCall:
    """Represents a parsed function call."""
    
    def __init__(self, name: str, arguments: Dict[str, Any]):
        self.name = name
        self.arguments = arguments
    
    def __repr__(self) -> str:
        return f"FunctionCall(name='{self.name}', arguments={self.arguments})"


def extract_function_call(text: str) -> Optional[FunctionCall]:
    """
    Extract function call JSON from LLM response using regex.
    Handles cases where JSON is embedded in text or has extra tokens.

    Args:
        text: The model output text to parse

    Returns:
        FunctionCall object if found, None otherwise
    """
    # Clean up common LLM artifacts
    text = text.strip()

    # Remove special tokens like <|end_of_text|>, <|im_end|>, etc.
    text = re.sub(r"<\|[^|]+\|>", "", text)
    text = text.strip()

    # Try to find JSON-like structure with "name" and "arguments" fields
    # Pattern matches: { "name": "...", "arguments": { ... } }
    json_pattern = (
        r'\{[^{}]*"name"\s*:\s*"([^"]+)"[^{}]*"arguments"\s*:\s*(\{[^}]*\})[^{}]*\}'
    )

    match = re.search(json_pattern, text, re.DOTALL)
    if match:
        func_name = match.group(1)
        try:
            # Try to parse the arguments
            args_str = match.group(2)
            func_args = json.loads(args_str)
            return FunctionCall(func_name, func_args)
        except json.JSONDecodeError:
            # If arguments parsing fails, try to extract manually
            # Pattern for simple key-value like { "query": "..." }
            args_match = re.search(r'"(\w+)"\s*:\s*"([^"]*)"', match.group(2))
            if args_match:
                return FunctionCall(func_name, {args_match.group(1): args_match.group(2)})

    # Fallback: try to parse the entire cleaned text as JSON
    try:
        parsed = json.loads(text)
        if isinstance(parsed, dict) and "name" in parsed:
            func_name = parsed.get("name")
            func_args = parsed.get("arguments", {})
            return FunctionCall(func_name, func_args)
    except json.JSONDecodeError:
        pass

    # Try to find JSON block in code blocks
    code_block_pattern = r'```(?:json)?\s*(\{.*?\})\s*```'
    match = re.search(code_block_pattern, text, re.DOTALL)
    if match:
        try:
            parsed = json.loads(match.group(1))
            if isinstance(parsed, dict) and "name" in parsed:
                func_name = parsed.get("name")
                func_args = parsed.get("arguments", {})
                return FunctionCall(func_name, func_args)
        except json.JSONDecodeError:
            pass

    return None

