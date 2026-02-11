# AG2 + Nexa SDK — Minimal Integration

Use Nexa SDK's OpenAI-compatible server with AG2 (formerly AutoGen).

## Quick Start

1) Start Nexa server
```bash
nexa pull NexaAI/Qwen3-4B-GGUF
nexa serve
```

2) Install dependencies
```bash
pip install "ag2[openai]"
```

3) Run the example
```bash
python cookbook/ag2/example.py
```

## Minimal AG2 Setup

```python
from autogen import ConversableAgent, LLMConfig

# Point AG2 to Nexa's OpenAI-compatible endpoint
llm_config = LLMConfig({
    "model": "NexaAI/Qwen3-4B-GGUF",
    "base_url": "http://localhost:18181/v1",
    "api_type": "openai",
    "api_key": "not-needed",
})

agent = ConversableAgent(
    name="helpful_agent",
    system_message="You are a helpful AI assistant.",
    llm_config=llm_config,
)

response = agent.run(
    message="Write a Python function to calculate the sum of a list.",
    max_turns=3,
)
response.process()
print(response.messages)
```

## Two-Agent Conversation

```python
from autogen import ConversableAgent, LLMConfig

llm_config = LLMConfig({
    "model": "NexaAI/Qwen3-4B-GGUF",
    "base_url": "http://localhost:18181/v1",
    "api_type": "openai",
    "api_key": "not-needed",
})

coder = ConversableAgent(
    name="coder",
    system_message="You are a Python developer. Reply TERMINATE when done.",
    llm_config=llm_config,
)

reviewer = ConversableAgent(
    name="reviewer",
    system_message="You are a code reviewer. Reply TERMINATE when done.",
    llm_config=llm_config,
)

result = coder.initiate_chat(
    recipient=reviewer,
    message="Write a function to check if a string is a palindrome.",
    max_turns=3,
)
print(result.summary)
```

Notes:
- Ensure the Nexa server is running at http://localhost:18181/v1.
- The model string should match what you pulled with `nexa pull`.
- AG2 requires Python 3.10+.

## More Examples
- Full runnable sample: [cookbook/ag2/example.py](../cookbook/ag2/example.py)
- AG2 docs: https://docs.ag2.ai
