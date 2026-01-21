#!/usr/bin/env python3
"""
Basic agent example with Nexa SDK and CrewAI.

Demonstrates:
- Creating a simple agent
- Defining tasks
- Executing a crew
"""

from crewai import Agent, Task, Crew, LLM


def create_llm():
    """Create LLM instance from config."""
    return LLM(
        model="huggingface/NexaAI/Qwen3-4B-GGUF",
        base_url="http://localhost:18181/v1",
    )


def basic_agent_example():
    """Example of a basic agent with simple task."""
    print("=" * 50)
    print("Basic Agent Example")
    print("=" * 50)

    # Initialize LLM with Nexa SDK
    llm = create_llm()

    # Create agent
    agent = Agent(
        role="Python Developer",
        goal="Write clean and efficient Python code",
        backstory="Expert Python developer with 10 years of experience",
        llm=llm,
    )

    # Create task
    task = Task(
        description="Write a Python function to calculate the sum of a list",
        expected_output="Working Python function with docstring and examples",
        agent=agent,
    )

    # Create and execute crew
    crew = Crew(agents=[agent], tasks=[task])

    print("\nExecuting crew...")
    result = crew.kickoff()
    print(f"\nResult:\n{result}")


def writer_example():
    """Example of a writer agent."""
    print("\n" + "=" * 50)
    print("Writer Agent Example")
    print("=" * 50)

    llm = create_llm()

    writer = Agent(
        role="Technical Writer",
        goal="Write clear and concise technical documentation",
        backstory="Professional technical writer with expertise in API documentation",
        llm=llm,
    )

    task = Task(
        description="Write documentation for a REST API endpoint",
        expected_output="Complete API documentation with examples",
        agent=writer,
    )

    crew = Crew(
        agents=[writer],
        tasks=[task],
    )

    print("\nExecuting crew...")
    result = crew.kickoff()
    print(f"\nResult:\n{result}")


def main():
    """Run all agent examples."""
    print("\n" + "=" * 50)
    print("CrewAI + Nexa SDK Basic Agent Examples")
    print("=" * 50 + "\n")

    basic_agent_example()

    print("\n" + "=" * 50)
    print("All examples completed!")
    print("=" * 50)


if __name__ == "__main__":
    main()
