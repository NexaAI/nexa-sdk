# Copyright 2024-2026 Nexa AI, Inc.
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
EDSL Integration Demo with Nexa SDK

This demo shows how to use EDSL (Expected Parrot Domain-Specific Language) with
Nexa SDK's OpenAI-compatible API for AI-powered research simulations and surveys.

Prerequisites:
1. Install Nexa CLI and start the server:
   nexa serve

2. Download the model (if not already cached):
   nexa pull NexaAI/Qwen3-4B-GGUF

3. Install dependencies:
   pip install -r requirements.txt

Usage:
    python example.py
"""

from edsl import (
    QuestionMultipleChoice,
    QuestionFreeText,
    QuestionLinearScale,
    Model,
    Agent,
    AgentList,
    ScenarioList,
    Survey,
)
import sys


def configure_nexa_model():
    """Configure EDSL to use Nexa SDK's OpenAI-compatible API."""

    # Configuration for Nexa SDK OpenAI-compatible API
    base_url = 'http://localhost:18181/v1'
    model_name = 'NexaAI/Qwen3-4B-GGUF'

    # Create custom Model class that uses Nexa SDK's endpoint
    # EDSL's OpenAI v2 service supports custom base_url
    model = Model(
        model=f'openai/{model_name}',
        parameters={
            'base_url': base_url,
            'api_key': 'not-needed',  # Nexa SDK doesn't require authentication
        },
    )

    return model, model_name, base_url


def example_simple_survey():
    """Example 1: Simple single-question survey."""
    print('=' * 60)
    print('Example 1: Simple Survey with Nexa SDK')
    print('=' * 60)

    try:
        model, model_name, base_url = configure_nexa_model()

        print(f'Model: {model_name}')
        print(f'API Endpoint: {base_url}')
        print()

        # Create a multiple choice question
        q = QuestionMultipleChoice(
            question_name='ai_feeling',
            question_text='How do you feel about artificial intelligence?',
            question_options=['Excited', 'Concerned', 'Neutral', 'Curious'],
        )

        print('Question: How do you feel about artificial intelligence?')
        print('Options: Excited, Concerned, Neutral, Curious')
        print('\nRunning survey...')

        # Run the survey with Nexa SDK model
        results = q.by(model).run()

        # Display results
        answer = results.select('ai_feeling').first()
        print(f'\nAnswer: {answer}')
        print()

    except Exception as e:
        print(f'✗ Error: {e}')
        print('\nPlease ensure:')
        print('1. Nexa server is running: nexa serve')
        print('2. Model is downloaded: nexa pull NexaAI/Qwen3-4B-GGUF')
        sys.exit(1)


def example_agent_personas():
    """Example 2: Survey with different AI agent personas."""
    print('=' * 60)
    print('Example 2: Agent Personas with Different Perspectives')
    print('=' * 60)

    try:
        model, model_name, base_url = configure_nexa_model()

        print(f'Model: {model_name}')
        print(f'API Endpoint: {base_url}')
        print()

        # Create agents with different personas
        agents = AgentList(
            [
                Agent(traits={'persona': 'data scientist'}),
                Agent(traits={'persona': 'software engineer'}),
                Agent(traits={'persona': 'product manager'}),
            ]
        )

        # Create a free text question
        q = QuestionFreeText(
            question_name='favorite_tools', question_text='What are your favorite professional tools and why?'
        )

        print('Question: What are your favorite professional tools and why?')
        print('Agent Personas: data scientist, software engineer, product manager')
        print('\nRunning survey with multiple agents...')

        # Run survey with multiple agents using Nexa SDK
        results = q.by(agents).by(model).run()

        # Display results
        print('\nResults:')
        print('-' * 60)
        for row in results.select('persona', 'favorite_tools'):
            print(f'\nPersona: {row["agent"]["persona"]}')
            print(f'Answer: {row["answer"]["favorite_tools"]}')
        print()

    except Exception as e:
        print(f'✗ Error: {e}')
        sys.exit(1)


def example_parameterized_prompts():
    """Example 3: Survey with parameterized prompts using scenarios."""
    print('=' * 60)
    print('Example 3: Parameterized Prompts with Scenarios')
    print('=' * 60)

    try:
        model, model_name, base_url = configure_nexa_model()

        print(f'Model: {model_name}')
        print(f'API Endpoint: {base_url}')
        print()

        # Create scenarios for parameterized questions
        scenarios = ScenarioList.from_list('technology', ['machine learning', 'cloud computing', 'blockchain'])

        # Create a linear scale question with parameter
        q = QuestionLinearScale(
            question_name='interest_level',
            question_text='How interested are you in {{ technology }}?',
            question_options=[1, 2, 3, 4, 5],
            option_labels={1: 'Not interested', 5: 'Very interested'},
        )

        print('Question: How interested are you in {{ technology }}?')
        print('Technologies: machine learning, cloud computing, blockchain')
        print('Scale: 1 (Not interested) to 5 (Very interested)')
        print('\nRunning parameterized survey...')

        # Run survey with scenarios using Nexa SDK
        results = q.by(scenarios).by(model).run()

        # Display results
        print('\nResults:')
        print('-' * 60)
        for row in results.select('technology', 'interest_level'):
            print(
                f'Technology: {row["scenario"]["technology"]:20} -> Interest Level: {row["answer"]["interest_level"]}'
            )
        print()

    except Exception as e:
        print(f'✗ Error: {e}')
        sys.exit(1)


def example_multi_question_survey():
    """Example 4: Multi-question survey with skip logic."""
    print('=' * 60)
    print('Example 4: Multi-Question Survey')
    print('=' * 60)

    try:
        model, model_name, base_url = configure_nexa_model()

        print(f'Model: {model_name}')
        print(f'API Endpoint: {base_url}')
        print()

        # Create multiple questions
        q1 = QuestionMultipleChoice(
            question_name='programming_language',
            question_text='What is your preferred programming language?',
            question_options=['Python', 'JavaScript', 'Go', 'Rust'],
        )

        q2 = QuestionFreeText(
            question_name='language_reason', question_text='Why do you prefer {{ programming_language.answer }}?'
        )

        # Create survey with both questions
        survey = Survey(questions=[q1, q2])

        print('Survey Questions:')
        print('1. What is your preferred programming language?')
        print('   Options: Python, JavaScript, Go, Rust')
        print('2. Why do you prefer [your answer]?')
        print('\nRunning multi-question survey...')

        # Run survey using Nexa SDK
        results = survey.by(model).run()

        # Display results
        print('\nResults:')
        print('-' * 60)
        for row in results.select('programming_language', 'language_reason'):
            print(f'Preferred Language: {row["answer"]["programming_language"]}')
            print(f'Reason: {row["answer"]["language_reason"]}')
        print()

    except Exception as e:
        print(f'✗ Error: {e}')
        sys.exit(1)


def main():
    """Main demo function demonstrating EDSL integration with Nexa SDK."""

    print('\n' + '=' * 60)
    print('EDSL + Nexa SDK Integration Demo')
    print('AI-Powered Research Simulations with Local Models')
    print('=' * 60 + '\n')

    # Run all examples
    example_simple_survey()

    print('\n')
    example_agent_personas()

    print('\n')
    example_parameterized_prompts()

    print('\n')
    example_multi_question_survey()

    print('=' * 60)
    print('Demo completed successfully!')
    print('=' * 60)


if __name__ == '__main__':
    main()
