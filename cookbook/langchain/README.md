# LangChain Integration with Nexa SDK

This guide demonstrates how to integrate [LangChain](https://www.langchain.com/) with Nexa SDK's OpenAI-compatible API for VLM (Vision-Language Model) chat capabilities.

## Overview

Nexa SDK provides an OpenAI-compatible REST API that allows you to use LangChain's `ChatOpenAI` class with your local VLM models. This integration enables you to leverage LangChain's powerful features (chains, agents, memory, etc.) while running models locally with Nexa SDK.

## Prerequisites

1. **Nexa CLI installed** - Download from [Nexa SDK Documentation](https://docs.nexa.ai)
2. **Model downloaded** - The example uses `NexaAI/Qwen3-VL-4B-Instruct-GGUF`
3. **Python 3.8+** with pip

## Setup

### 1. Install Nexa CLI

Follow the installation instructions in the [main README](../../README.md) to install Nexa CLI for your platform.

### 2. Download the Model

```bash
nexa pull NexaAI/Qwen3-VL-4B-Instruct-GGUF
```

### 3. Start Nexa Server

Start the Nexa server with the OpenAI-compatible API:

```bash
nexa serve
```

The server will be available at `http://localhost:18181/v1` (note the `/v1` suffix).

### 4. Install Python Dependencies

```bash
pip install -r requirements.txt
```

## Usage

### Basic Configuration

LangChain's `ChatOpenAI` class can be configured to use Nexa SDK's API by setting the `base_url` parameter:

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="NexaAI/Qwen3-VL-4B-Instruct-GGUF",
    base_url="http://localhost:18181/v1",
    api_key="not-needed",  # Nexa SDK doesn't require authentication
    temperature=0.7,
    max_tokens=512,
)
```

### Key Configuration Parameters

- **`base_url`**: `"http://localhost:18181/v1"` - Nexa SDK's OpenAI-compatible API endpoint
- **`api_key`**: `"not-needed"` - Nexa SDK doesn't require authentication
- **`model`**: `"NexaAI/Qwen3-VL-4B-Instruct-GGUF"` - The model identifier (must match the model name used with `nexa pull`)

### Example: Simple Chat

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

llm = ChatOpenAI(
    model="NexaAI/Qwen3-VL-4B-Instruct-GGUF",
    base_url="http://localhost:18181/v1",
    api_key="not-needed",
)

messages = [HumanMessage(content="What is artificial intelligence?")]
response = llm.invoke(messages)
print(response.content)
```

### Example: With System Prompt

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, SystemMessage

llm = ChatOpenAI(
    model="NexaAI/Qwen3-VL-4B-Instruct-GGUF",
    base_url="http://localhost:18181/v1",
    api_key="not-needed",
)

messages = [
    SystemMessage(content="You are a helpful AI assistant."),
    HumanMessage(content="Explain quantum computing in simple terms.")
]
response = llm.invoke(messages)
print(response.content)
```

### Example: Multi-turn Conversation

```python
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage

llm = ChatOpenAI(
    model="NexaAI/Qwen3-VL-4B-Instruct-GGUF",
    base_url="http://localhost:18181/v1",
    api_key="not-needed",
)

conversation = [
    HumanMessage(content="My name is Alice.")
]
response = llm.invoke(conversation)
print(response.content)

# Continue conversation
conversation.append(response)
conversation.append(HumanMessage(content="What's my name?"))
response = llm.invoke(conversation)
print(response.content)
```

## Running the Demo

Run the included example script:

```bash
python example.py
```

The demo showcases:
1. Simple text chat
2. Chat with system prompt
3. Multi-turn conversation

## Integration with LangChain Features

Once configured, you can use the `ChatOpenAI` instance with all LangChain features:

- **Chains**: Build complex workflows
- **Agents**: Create AI agents with tools
- **Memory**: Add conversation memory
- **Streaming**: Enable streaming responses
- **Callbacks**: Add custom callbacks

### Example: Using with LangChain Chains

```python
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate
from langchain.chains import LLMChain

llm = ChatOpenAI(
    model="NexaAI/Qwen3-VL-4B-Instruct-GGUF",
    base_url="http://localhost:18181/v1",
    api_key="not-needed",
)

prompt = ChatPromptTemplate.from_messages([
    ("system", "You are a helpful assistant."),
    ("human", "{question}")
])

chain = LLMChain(llm=llm, prompt=prompt)
result = chain.run(question="What is Python?")
print(result)
```

## Troubleshooting

### Server Connection Error

If you see connection errors, ensure:
1. Nexa server is running: `nexa serve --host 127.0.0.1:18181`
2. The server is accessible at `http://localhost:18181`
3. Check the server logs for any errors

### Model Not Found

If you get a "model not found" error:
1. Ensure the model is downloaded: `nexa pull NexaAI/Qwen3-VL-4B-Instruct-GGUF`
2. Verify the model name matches exactly (case-sensitive)
3. Check available models: `nexa list`

### API Compatibility Issues

Nexa SDK's API is compatible with OpenAI's API, but some advanced features may differ. If you encounter issues:
1. Check the [Nexa SDK documentation](https://docs.nexa.ai)
2. Review the API compatibility in the main README
3. Test with the OpenAI Python client first to verify the endpoint

## Additional Resources

- [Nexa SDK Documentation](https://docs.nexa.ai)
- [LangChain Documentation](https://python.langchain.com/)
- [Nexa SDK GitHub Repository](https://github.com/NexaAI/nexa-sdk)

## License

This integration example follows the same license as Nexa SDK. See the [LICENSE](../../LICENSE) file for details.

