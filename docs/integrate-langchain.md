# LangChain Integration Guide

Quickly connect LangChain to Nexa Serve (OpenAI-compatible endpoint).

## Prerequisites
- Python 3.10+
- Nexa SDK installed: see [Installation Guide](../README.md)
- Model downloaded, e.g. `nexa pull Qwen/Qwen3-7B-Instruct-GGUF`
- Nexa Serve running: `nexa serve` (default `http://127.0.0.1:18181/v1`)

## Install
```bash
pip install langchain langchain-openai openai
```

## Quick Start
```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    base_url="http://127.0.0.1:18181/v1",
    api_key="not-needed",
    model="Qwen/Qwen3-7B-Instruct-GGUF"
)

print(llm.invoke("Hello from Nexa!").content)
```

## Common Recipes

### Simple chat (with system prompt)
```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import SystemMessage, HumanMessage

llm = ChatOpenAI(base_url="http://127.0.0.1:18181/v1", api_key="not-needed", model="Qwen/Qwen3-7B-Instruct-GGUF")

messages = [
    SystemMessage(content="You are a helpful assistant."),
    HumanMessage(content="Explain decorators in Python")
]

print(llm.invoke(messages).content)
```
## More Examples
- Full runnable sample: [cookbook/langchain/example.py](../../cookbook/langchain/example.py)
- LangChain docs: https://python.langchain.com
