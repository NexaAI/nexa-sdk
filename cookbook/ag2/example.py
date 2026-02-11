#!/usr/bin/env python3
"""
Basic agent example with Nexa SDK and AG2.

Demonstrates:
- Configuring AG2 with Nexa SDK's OpenAI-compatible endpoint
- Creating a single conversable agent
- Running a two-agent conversation
"""

from autogen import ConversableAgent, LLMConfig


def create_llm_config():
    """Create LLM config pointing to Nexa's OpenAI-compatible endpoint."""
    return LLMConfig({
        "model": "NexaAI/Qwen3-4B-GGUF",
        "base_url": "http://localhost:18181/v1",
        "api_type": "openai",
        "api_key": "not-needed",
    })


def single_agent_example():
    """Example of a single conversable agent."""
    print("=" * 50)
    print("Single Agent Example")
    print("=" * 50)

    llm_config = create_llm_config()

    agent = ConversableAgent(
        name="helpful_agent",
        system_message="You are a helpful AI assistant that gives concise answers.",
        llm_config=llm_config,
    )

    print("\nRunning agent...")
    response = agent.run(
        message="Write a Python function to calculate the factorial of a number.",
        max_turns=3,
    )
    response.process()

    for msg in response.messages:
        print(f"\n[{msg.get('role', 'unknown')}]: {msg.get('content', '')}")


def two_agent_example():
    """Example of a two-agent conversation."""
    print("\n" + "=" * 50)
    print("Two-Agent Conversation Example")
    print("=" * 50)

    llm_config = create_llm_config()

    # Create a coding assistant
    coder = ConversableAgent(
        name="coder",
        system_message=(
            "You are a Python developer. Write clean, well-documented code. "
            "Reply TERMINATE when the task is complete."
        ),
        llm_config=llm_config,
    )

    # Create a code reviewer
    reviewer = ConversableAgent(
        name="reviewer",
        system_message=(
            "You are a code reviewer. Review the code for correctness, style, "
            "and best practices. Reply TERMINATE when the review is done."
        ),
        llm_config=llm_config,
    )

    print("\nStarting two-agent conversation...")
    result = coder.initiate_chat(
        recipient=reviewer,
        message="Write a Python function to check if a string is a palindrome.",
        max_turns=3,
    )

    print(f"\nConversation summary:\n{result.summary}")


def main():
    """Run all agent examples."""
    print("\n" + "=" * 50)
    print("AG2 + Nexa SDK Basic Agent Examples")
    print("=" * 50 + "\n")

    single_agent_example()
    two_agent_example()

    print("\n" + "=" * 50)
    print("All examples completed!")
    print("=" * 50)


if __name__ == "__main__":
    main()
