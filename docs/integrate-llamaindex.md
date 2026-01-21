# LlamaIndex Integration Guide

Quickly use Nexa SDK with LlamaIndex using OpenAI-compatible endpoint.

## Installation

```bash
pip install llama-index-core llama-index-llms-openai-like
```

## Basic Setup

```python
from llama_index.llms.openai_like import OpenAILike

llm = OpenAILike(
    model="your-model",
    api_base="http://localhost:18181/v1",  # Nexa SDK endpoint
    api_key="fake",
    context_window=8000,
    is_chat_model=True,
)

# Text completion
response = llm.complete("What is AI?")
print(response)

# Chat
from llama_index.core.base.llms.types import ChatMessage

messages = [
    ChatMessage(role="system", content="You are helpful."),
    ChatMessage(role="user", content="What is AI?"),
]
chat_response = llm.chat(messages)
print(chat_response)
```

## Streaming

```python
# Stream completion
response_stream = llm.stream_complete("Tell me a story")
for chunk in response_stream:
    print(chunk.delta, end="")
```

## Key Methods

| Method | Description |
|--------|-------------|
| `complete(prompt)` | Text completion |
| `stream_complete(prompt)` | Streaming completion |
| `acomplete(prompt)` | Async completion |
| `chat(messages)` | Chat with messages |
| `achat(messages)` | Async chat |

## Next Steps

- See cookbook examples: [../../cookbook/llama-index/](../../cookbook/llama-index/)
- [LlamaIndex documentation](https://docs.llamaindex.ai/)
- [Nexa SDK README](README.md)