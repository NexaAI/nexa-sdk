#!/usr/bin/env python3
"""
Streaming example with Nexa SDK and LlamaIndex.

Demonstrates:
- Streaming text completions
- Streaming chat responses
- Handling streaming output in real-time
"""

from llama_index.llms.openai_like import OpenAILike
from llama_index.core.base.llms.types import ChatMessage
from config import LLM_CONFIG


def streaming_completion_example():
    """Example of streaming text completion."""
    print("=" * 50)
    print("Streaming Text Completion Example")
    print("=" * 50)

    llm = OpenAILike(**LLM_CONFIG)

    prompt = "Write a creative story about a robot learning to paint"
    print(f"\nPrompt: {prompt}")
    print("\nStreaming response:")
    print("-" * 50)

    response_stream = llm.stream_complete(prompt)
    full_response = ""

    for chunk in response_stream:
        print(chunk.delta, end="", flush=True)
        full_response += chunk.delta

    print("\n" + "-" * 50)
    print(f"\nTotal length: {len(full_response)} characters")


def streaming_chat_example():
    """Example of streaming chat responses."""
    print("\n" + "=" * 50)
    print("Streaming Chat Example")
    print("=" * 50)

    llm = OpenAILike(**LLM_CONFIG)

    messages = [
        ChatMessage(role="system", content="You are a creative writing assistant."),
        ChatMessage(role="user", content="Tell me a short story about time travel"),
    ]

    print("\nChat messages:")
    for msg in messages:
        print(f"  {msg.role}: {msg.content}")

    print("\nStreaming response:")
    print("-" * 50)

    response_stream = llm.stream_chat(messages)
    full_response = ""

    for chunk in response_stream:
        print(chunk.delta, end="", flush=True)
        full_response += chunk.delta

    print("\n" + "-" * 50)
    print(f"\nTotal length: {len(full_response)} characters")


def streaming_multi_turn_with_context():
    """Example of streaming in a multi-turn conversation."""
    print("\n" + "=" * 50)
    print("Streaming Multi-turn Conversation Example")
    print("=" * 50)

    llm = OpenAILike(**LLM_CONFIG)

    messages = [
        ChatMessage(
            role="system",
            content="You are a helpful assistant answering questions about space exploration.",
        ),
    ]

    # First turn
    print("\nUser: Tell me about the Apollo 11 mission")
    messages.append(
        ChatMessage(role="user", content="Tell me about the Apollo 11 mission")
    )

    print("Assistant: ", end="")
    response_stream = llm.stream_chat(messages)
    response_text = ""

    for chunk in response_stream:
        print(chunk.delta, end="", flush=True)
        response_text += chunk.delta

    print("\n")
    messages.append(ChatMessage(role="assistant", content=response_text))

    # Second turn
    print("\nUser: How long did the mission last?")
    messages.append(ChatMessage(role="user", content="How long did the mission last?"))

    print("Assistant: ", end="")
    response_stream = llm.stream_chat(messages)
    response_text = ""

    for chunk in response_stream:
        print(chunk.delta, end="", flush=True)
        response_text += chunk.delta

    print("\n")


def main():
    """Run all streaming examples."""
    print("\n" + "=" * 50)
    print("Nexa SDK + LlamaIndex Streaming Examples")
    print("=" * 50)

    streaming_completion_example()
    streaming_chat_example()
    streaming_multi_turn_with_context()

    print("\n" + "=" * 50)
    print("All streaming examples completed!")
    print("=" * 50)


if __name__ == "__main__":
    main()
