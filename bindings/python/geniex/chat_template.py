# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

"""Python-owned results and metadata for model chat-template application."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class GrammarTrigger:
    """A token, word, or pattern that activates a lazy chat grammar."""

    trigger_type: int
    value: str | None
    token: int


@dataclass(frozen=True)
class ChatTemplateGrammar:
    """Owned copy of grammar metadata returned by chat-template application."""

    grammar_type: int = 0
    grammar: str | None = None
    lazy: bool = False
    generation_prompt: str | None = None
    triggers: tuple[GrammarTrigger, ...] = ()
    preserved_tokens: tuple[int, ...] = ()


@dataclass(frozen=True)
class ChatTemplateOutput:
    """Formatted prompt plus optional sampler grammar metadata."""

    formatted_text: str
    chat_template_grammar: ChatTemplateGrammar
