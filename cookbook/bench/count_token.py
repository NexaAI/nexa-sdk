#!/usr/bin/env python3
"""Token counter for benchmark system prompts.

This utility accurately counts tokens in text files using multiple tokenizer backends:
- tiktoken: OpenAI's tokenizer (default, fast, no model download)
- transformers: Hugging Face tokenizers (requires model download)

Usage:
    python count_words.py prompt500.txt
    python count_words.py prompt500.txt --backend transformers --model meta-llama/Llama-2-7b-hf
    python count_words.py prompt500.txt --backend tiktoken --model gpt-4

Examples:
    # Count tokens using OpenAI's gpt-3.5-turbo tokenizer (default)
    python count_words.py prompt500.txt

    # Count tokens for GPT-4
    python count_words.py prompt500.txt --model gpt-4

    # Count tokens using Hugging Face transformers
    python count_words.py prompt500.txt --backend transformers

    # Count tokens using specific LLaMA model
    python count_words.py prompt500.txt --backend transformers --model meta-llama/Llama-2-7b-hf
"""

from __future__ import annotations

import argparse
import pathlib
import sys
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    import tiktoken


def count_tokens_tiktoken(text: str, model: str = "gpt-3.5-turbo") -> int:
    """Count tokens using tiktoken (OpenAI tokenizer)."""
    try:
        import tiktoken
    except ImportError:
        print(
            "Error: tiktoken not installed. Install with: pip install tiktoken",
            file=sys.stderr,
        )
        raise

    encoding = tiktoken.encoding_for_model(model)
    return len(encoding.encode(text))


def count_tokens_transformers(
    text: str, model: str = "meta-llama/Llama-2-7b-hf"
) -> int:
    """Count tokens using transformers (Hugging Face)."""
    try:
        from transformers import AutoTokenizer
    except ImportError:
        print(
            "Error: transformers not installed. Install with: pip install transformers",
            file=sys.stderr,
        )
        raise

    tokenizer = AutoTokenizer.from_pretrained(model)
    tokens = tokenizer.encode(text, add_special_tokens=False)
    return len(tokens)


def main() -> int:
    parser = argparse.ArgumentParser(description="Count tokens in a text file.")
    parser.add_argument("path", type=pathlib.Path, help="Path to the input text file")
    parser.add_argument(
        "--encoding",
        default="utf-8",
        help="File encoding (default: utf-8)",
    )
    parser.add_argument(
        "--backend",
        default="tiktoken",
        choices=["tiktoken", "transformers"],
        help="Tokenizer backend to use (default: tiktoken)",
    )
    parser.add_argument(
        "--model",
        default=None,
        help="Model for tokenizer. "
        "For tiktoken: gpt-3.5-turbo, gpt-4, etc. (default: gpt-3.5-turbo). "
        "For transformers: HF model ID (default: meta-llama/Llama-2-7b-hf)",
    )
    args = parser.parse_args()

    if not args.path.is_file():
        parser.error(f"File not found: {args.path}")

    try:
        text = args.path.read_text(encoding=args.encoding)
    except Exception as exc:  # noqa: BLE001
        print(f"Failed to read file: {exc}", file=sys.stderr)
        return 1

    try:
        if args.backend == "tiktoken":
            model = args.model or "gpt-3.5-turbo"
            total = count_tokens_tiktoken(text, model=model)
        else:  # transformers
            model = args.model or "meta-llama/Llama-2-7b-hf"
            total = count_tokens_transformers(text, model=model)
    except Exception as exc:  # noqa: BLE001
        print(f"Failed to count tokens: {exc}", file=sys.stderr)
        return 1

    print(total)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
