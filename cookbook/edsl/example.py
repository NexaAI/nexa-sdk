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

"""
Basic survey example with Nexa SDK and EDSL.

Demonstrates:
- Configuring EDSL Model for Nexa (OpenAI-compatible API)
- Running a simple survey question
"""

import os
from edsl import QuestionMultipleChoice, Model


def create_model():
    """Create EDSL Model pointing at Nexa SDK API."""
    base_url = "http://localhost:18181/v1"
    model_name = "unsloth/Qwen3-4B-Instruct-2507-GGUF"
    os.environ["OPENAI_BASE_URL"] = base_url
    os.environ.setdefault("OPENAI_API_KEY", "not-needed")
    model = Model(
        model=f"openai/{model_name}",
        parameters={"base_url": base_url, "api_key": "not-needed"},
    )
    model.model = model_name
    return model


def simple_survey_example():
    """Run a single multiple-choice question with Nexa."""
    print("=" * 50)
    print("Simple Survey Example")
    print("=" * 50)

    model = create_model()
    q = QuestionMultipleChoice(
        question_name="ai_feeling",
        question_text="How do you feel about artificial intelligence?",
        question_options=["Excited", "Concerned", "Neutral", "Curious"],
    )

    print("\nRunning survey...")
    results = q.by(model).run()
    answer = results.select("ai_feeling").first()
    print(f"Answer: {answer}")


def main():
    """Run EDSL + Nexa SDK example."""
    print("\n" + "=" * 50)
    print("EDSL + Nexa SDK Example")
    print("=" * 50 + "\n")

    simple_survey_example()

    print("\n" + "=" * 50)
    print("Done.")
    print("=" * 50)


if __name__ == "__main__":
    main()
