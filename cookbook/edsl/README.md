# EDSL Integration with Nexa SDK

This guide demonstrates how to integrate [EDSL (Expected Parrot Domain-Specific Language)](https://github.com/expectedparrot/edsl) with Nexa SDK's OpenAI-compatible API for AI-powered research simulations and surveys.

## Overview

EDSL is a powerful Python package designed for conducting computational social science and market research with AI. It allows you to:

- Design and run surveys with AI agents
- Create experiments with multiple language models
- Generate diverse responses using agent personas
- Use parameterized prompts with scenarios
- Build complex data labeling flows

By integrating EDSL with Nexa SDK, you can run all these research simulations locally with your own models, ensuring data privacy and full control over your research infrastructure.

## Prerequisites

1. **Nexa CLI installed** - Download from [Nexa SDK Documentation](https://docs.nexa.ai)
2. **Model downloaded** - The example uses `NexaAI/Qwen3-4B-GGUF`
3. **Python 3.9+** with pip

## Setup

### 1. Install Nexa CLI

Follow the installation instructions in the [main README](../../README.md) to install Nexa CLI for your platform.

### 2. Download the Model

```bash
nexa pull NexaAI/Qwen3-4B-GGUF
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

EDSL's Model class can be configured to use Nexa SDK's API by setting the `base_url` parameter:

```python
from edsl import Model

model = Model(
    model="openai/NexaAI/Qwen3-4B-GGUF",
    parameters={
        "base_url": "http://localhost:18181/v1",
        "api_key": "not-needed"  # Nexa SDK doesn't require authentication
    }
)
```

### Key Configuration Parameters

- **`model`**: `"openai/NexaAI/Qwen3-4B-GGUF"` - Use the "openai/" prefix to specify OpenAI-compatible API
- **`base_url`**: `"http://localhost:18181/v1"` - Nexa SDK's OpenAI-compatible API endpoint
- **`api_key`**: `"not-needed"` - Nexa SDK doesn't require authentication

## Examples

### Example 1: Simple Survey

Run a basic single-question survey with Nexa SDK:

```python
from edsl import QuestionMultipleChoice, Model

# Configure Nexa SDK model
model = Model(
    model="openai/NexaAI/Qwen3-4B-GGUF",
    parameters={
        "base_url": "http://localhost:18181/v1",
        "api_key": "not-needed"
    }
)

# Create a question
q = QuestionMultipleChoice(
    question_name="ai_feeling",
    question_text="How do you feel about artificial intelligence?",
    question_options=["Excited", "Concerned", "Neutral", "Curious"]
)

# Run the survey
results = q.by(model).run()

# Display results
print(results.select("ai_feeling"))
```

### Example 2: Agent Personas

Use different AI agent personas to get diverse perspectives:

```python
from edsl import QuestionFreeText, Model, Agent, AgentList

# Configure model
model = Model(
    model="openai/NexaAI/Qwen3-4B-GGUF",
    parameters={
        "base_url": "http://localhost:18181/v1",
        "api_key": "not-needed"
    }
)

# Create agents with different personas
agents = AgentList([
    Agent(traits={"persona": "data scientist"}),
    Agent(traits={"persona": "software engineer"}),
    Agent(traits={"persona": "product manager"})
])

# Create a question
q = QuestionFreeText(
    question_name="favorite_tools",
    question_text="What are your favorite professional tools and why?"
)

# Run survey with multiple agents
results = q.by(agents).by(model).run()

# Display results by persona
for row in results.select("persona", "favorite_tools"):
    print(f"Persona: {row['agent']['persona']}")
    print(f"Answer: {row['answer']['favorite_tools']}\n")
```

### Example 3: Parameterized Prompts with Scenarios

Use scenarios to parameterize questions across different contexts:

```python
from edsl import QuestionLinearScale, Model, ScenarioList

# Configure model
model = Model(
    model="openai/NexaAI/Qwen3-4B-GGUF",
    parameters={
        "base_url": "http://localhost:18181/v1",
        "api_key": "not-needed"
    }
)

# Create scenarios
scenarios = ScenarioList.from_list(
    "technology",
    ["machine learning", "cloud computing", "blockchain"]
)

# Create parameterized question
q = QuestionLinearScale(
    question_name="interest_level",
    question_text="How interested are you in {{ technology }}?",
    question_options=[1, 2, 3, 4, 5],
    option_labels={1: "Not interested", 5: "Very interested"}
)

# Run survey with scenarios
results = q.by(scenarios).by(model).run()

# Display results
for row in results.select("technology", "interest_level"):
    print(f"{row['scenario']['technology']}: {row['answer']['interest_level']}")
```

### Example 4: Multi-Question Survey

Create surveys with multiple questions and skip logic:

```python
from edsl import QuestionMultipleChoice, QuestionFreeText, Survey, Model

# Configure model
model = Model(
    model="openai/NexaAI/Qwen3-4B-GGUF",
    parameters={
        "base_url": "http://localhost:18181/v1",
        "api_key": "not-needed"
    }
)

# Create multiple questions
q1 = QuestionMultipleChoice(
    question_name="programming_language",
    question_text="What is your preferred programming language?",
    question_options=["Python", "JavaScript", "Go", "Rust"]
)

q2 = QuestionFreeText(
    question_name="language_reason",
    question_text="Why do you prefer {{ programming_language.answer }}?"
)

# Create survey
survey = Survey(questions=[q1, q2])

# Run survey
results = survey.by(model).run()

# Display results
for row in results.select("programming_language", "language_reason"):
    print(f"Language: {row['answer']['programming_language']}")
    print(f"Reason: {row['answer']['language_reason']}\n")
```

## Running the Demo

Run the included example script to see all features in action:

```bash
python example.py
```

The demo showcases:
1. Simple single-question survey
2. Agent personas with different perspectives
3. Parameterized prompts with scenarios
4. Multi-question surveys with skip logic

## Key Features of EDSL Integration

### Research Simulations
- Design surveys with declarative question types
- Get consistent, structured responses
- No need for complex JSON schemas

### Agent Personas
- Create AI agents with specific traits
- Get diverse perspectives from a single model
- Simulate different demographic groups or expertise levels

### Parameterization
- Use scenarios to parameterize questions
- Test hypotheses across different contexts
- Import data from CSV, PDF, images, etc.

### Survey Logic
- Build complex flows with skip and stop rules
- Pipe answers between questions
- Chain multiple questions together

### Data Analysis
- Results come as structured datasets
- Built-in methods for analysis and visualization
- Easy to export and share findings

## Troubleshooting

### Server Connection Error

If you see connection errors, ensure:
1. Nexa server is running: `nexa serve`
2. The server is accessible at `http://localhost:18181`
3. Check the server logs for any errors

### Model Not Found

If you get a "model not found" error:
1. Ensure the model is downloaded: `nexa pull NexaAI/Qwen3-4B-GGUF`
2. Verify the model name matches exactly (case-sensitive)
3. Check available models: `nexa list`

### API Compatibility Issues

Nexa SDK's API is compatible with OpenAI's API, but some advanced features may differ. If you encounter issues:
1. Check the [Nexa SDK documentation](https://docs.nexa.ai)
2. Review the API compatibility in the main README
3. Try with a different model if needed

### EDSL-Specific Issues

For EDSL-specific questions:
1. Check the [EDSL documentation](https://docs.expectedparrot.com)
2. Join the [EDSL Discord](https://discord.com/invite/mxAYkjfy9m)
3. Review [EDSL examples on Coop](https://www.expectedparrot.com/content/explore)

## Benefits of Local Deployment

Running EDSL with Nexa SDK locally offers several advantages:

- **Privacy**: All data stays on your machine
- **Cost**: No API usage fees
- **Control**: Full control over model selection and parameters
- **Offline**: Works without internet connection
- **Speed**: No network latency for requests
- **Customization**: Use any GGUF model you prefer

## Additional Resources

- [Nexa SDK Documentation](https://docs.nexa.ai)
- [EDSL Documentation](https://docs.expectedparrot.com)
- [EDSL GitHub Repository](https://github.com/expectedparrot/edsl)
- [Nexa SDK GitHub Repository](https://github.com/NexaAI/nexa-sdk)
- [EDSL Examples on Coop](https://www.expectedparrot.com/content/explore)

## License

This integration example follows the same license as Nexa SDK. See the [LICENSE](../../LICENSE) file for details.
