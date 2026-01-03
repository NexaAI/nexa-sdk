# Copyright 2024-2025 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/env python3
"""
LangChain Integration Demo with Nexa SDK VLM Chat

This demo shows how to use LangChain's ChatOpenAI with Nexa SDK's OpenAI-compatible API
for VLM (Vision-Language Model) chat using the Qwen3-VL-4B-Instruct-GGUF model.

Prerequisites:
1. Install Nexa CLI and start the server:
   nexa serve

2. Download the model (if not already cached):
   nexa pull NexaAI/Qwen3-VL-4B-Instruct-GGUF

3. Install dependencies:
   pip install -r requirements.txt

Usage:
    python example.py
"""

from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, SystemMessage
import sys


def main():
    """Main demo function demonstrating LangChain integration with Nexa SDK."""

    # Configuration for Nexa SDK OpenAI-compatible API
    base_url = "http://localhost:18181/v1"
    api_key = "not-needed"  # Nexa SDK doesn't require authentication
    model_name = "NexaAI/Qwen3-VL-4B-Instruct-GGUF"

    print("=" * 60)
    print("LangChain + Nexa SDK VLM Chat Demo")
    print("=" * 60)
    print(f"Model: {model_name}")
    print(f"API Endpoint: {base_url}")
    print()

    # Initialize ChatOpenAI with Nexa SDK configuration
    try:
        llm = ChatOpenAI(
            model=model_name,
            base_url=base_url,
            api_key=api_key,
            temperature=0.7,
            max_tokens=512,
        )
        print("✓ Successfully initialized ChatOpenAI with Nexa SDK")
        print()
    except Exception as e:
        print(f"✗ Failed to initialize ChatOpenAI: {e}")
        print("\nPlease ensure:")
        print("1. Nexa server is running: nexa serve")
        print("2. Model is downloaded: nexa pull NexaAI/Qwen3-VL-4B-Instruct-GGUF")
        sys.exit(1)

    # Example 1: Simple text chat
    print("-" * 60)
    print("Example 1: Simple Text Chat")
    print("-" * 60)

    try:
        messages = [
            HumanMessage(
                content="What is artificial intelligence? Please explain in 2-3 sentences."
            )
        ]

        print("User: What is artificial intelligence? Please explain in 2-3 sentences.")
        print("\nAssistant: ", end="", flush=True)

        response = llm.invoke(messages)
        print(response.content)
        print()

    except Exception as e:
        print(f"✗ Error during chat: {e}")
        print("\nPlease check:")
        print("1. Nexa server is running and accessible")
        print("2. Model is loaded correctly")
        sys.exit(1)

    # Example 2: Chat with system prompt
    print("-" * 60)
    print("Example 2: Chat with System Prompt")
    print("-" * 60)

    try:
        messages = [
            SystemMessage(
                content="You are a helpful AI assistant that explains complex topics in simple terms."
            ),
            HumanMessage(content="Explain quantum computing in simple terms."),
        ]

        print(
            "System: You are a helpful AI assistant that explains complex topics in simple terms."
        )
        print("User: Explain quantum computing in simple terms.")
        print("\nAssistant: ", end="", flush=True)

        response = llm.invoke(messages)
        print(response.content)
        print()

    except Exception as e:
        print(f"✗ Error during chat: {e}")
        sys.exit(1)

    # Example 3: Multi-turn conversation
    print("-" * 60)
    print("Example 3: Multi-turn Conversation")
    print("-" * 60)

    try:
        conversation = [
            HumanMessage(content="My name is Alice. I'm learning Python programming.")
        ]

        print("User: My name is Alice. I'm learning Python programming.")
        print("\nAssistant: ", end="", flush=True)

        response = llm.invoke(conversation)
        print(response.content)
        print()

        # Continue the conversation
        conversation.append(response)
        conversation.append(
            HumanMessage(content="Can you recommend a good beginner project?")
        )

        print("User: Can you recommend a good beginner project?")
        print("\nAssistant: ", end="", flush=True)

        response = llm.invoke(conversation)
        print(response.content)
        print()

    except Exception as e:
        print(f"✗ Error during conversation: {e}")
        sys.exit(1)

    print("=" * 60)
    print("Demo completed successfully!")
    print("=" * 60)


if __name__ == "__main__":
    main()

