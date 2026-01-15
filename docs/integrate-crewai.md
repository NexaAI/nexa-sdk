# CrewAI + Nexa SDK — Minimal Integration

Use Nexa SDK’s OpenAI-compatible server with CrewAI.

## Quick Start

1) Start Nexa server
```bash
nexa pull NexaAI/Qwen3-4B-GGUF
nexa serve
```

2) Install dependencies
```bash
pip install crewai
```

3) Run the example
```bash
python cookbook/crewai/example.py
```

## Minimal CrewAI Setup

```python
from crewai import Agent, Task, Crew, LLM

# Point CrewAI to Nexa's OpenAI-compatible endpoint
llm = LLM(
    model="huggingface/NexaAI/Qwen3-4B-GGUF",
    base_url="http://localhost:18181/v1",
)

agent = Agent(
    role="Python Developer",
    goal="Write clean and efficient Python code",
    backstory="Expert Python developer",
    llm=llm,
)

task = Task(
    description="Write a Python function to calculate the sum of a list",
    expected_output="A working function with docstring",
    agent=agent,
)

crew = Crew(agents=[agent], tasks=[task])
result = crew.kickoff()
print(result)
```

Notes:
- Ensure the Nexa server is running at http://localhost:18181/v1.
- The model string should match what you pulled with `nexa pull`.
