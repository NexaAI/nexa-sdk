import json
from nexa.gguf import NexaTextInference
from pathlib import Path


def test_structural_output():
    nexa = NexaTextInference(model_path="Meta-Llama-3.1-8B-Instruct:q4_0")
    prompt = "Emily Carter, a 32-year-old owner, drives a 2023 Audi Q5 available in White, Black, and Gray, equipped with Bang & Olufsen audio (19 speakers, Bluetooth), 10 airbags, 8 parking sensors, lane assist, and a turbocharged inline-4 engine delivering 261 horsepower and a top speed of 130 mph."
    schema_abs_path = (Path(__file__).parent /
                       "structure_decoding_resources/schema.json").resolve()
    response = nexa.structure_output(
        prompt=prompt, json_schema_path=schema_abs_path)
    print(f"response: {json.dumps(response, indent=4)}")


def add_integer(num1, num2):
    return num1, num2


def test_function_calling():

    system_prompt = (
        "You are an AI assistant that generates structured function calling responses. "
        "Identify the correct function from the available tools and return a JSON object "
        "containing the function name and all required parameters. Ensure the parameters "
        "are accurately derived from the user's input and formatted correctly."
    )

    # tools = [
    #     {
    #         "type": "function",
    #         "function": {
    #             "name": "add_integer",
    #             "description": "Returns the addition of input integers.",
    #             "parameters": {
    #                 "type": "object",
    #                 "properties": {
    #                     "num1": {"type": "integer", "description": "An integer to add."},
    #                     "num2": {"type": "integer", "description": "An integer to add."}
    #                 },
    #                 "required": ["number"],
    #                 "additionalProperties": False
    #             },
    #             "strict": True
    #         }
    #     }
    # ]
    # messages = [
    #     {"role": "system", "content": system_prompt},
    #     {"role": "user", "content": "Please calculate the sum of 42 and 100."}
    # ]

    # tools = [
    #     {
    #         "type": "function",
    #         "function": {
    #             "name": "max_integer",
    #             "description": "Returns the maximum of input integers.",
    #             "parameters": {
    #                 "type": "object",
    #                 "properties": {
    #                     "num1": {"type": "integer", "description": "An input integer."},
    #                     "num2": {"type": "integer", "description": "An input integer."}
    #                 },
    #                 "required": ["num1", "num2"],
    #                 "additionalProperties": False
    #             },
    #             "strict": True
    #         }
    #     }
    # ]
    # messages = [
    #     {"role": "system", "content": system_prompt},
    #     {"role": "user", "content": "Please calculate the maximum of 42 and 100."}
    # ]

    tools = [{
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Get current temperature for provided coordinates in celsius.",
            "parameters": {
                "type": "object",
                "properties": {
                    "latitude": {"type": "number"},
                    "longitude": {"type": "number"}
                },
                "required": ["latitude", "longitude"],
                "additionalProperties": False
            },
            "strict": True
        }
    }]

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "What's the weather like in Paris today?"}
    ]

    nexa = NexaTextInference(
        model_path="Meta-Llama-3.1-8B-Instruct:q4_0", function_calling=True)
    response = nexa.function_calling(messages=messages, tools=tools)
    print(response)


if __name__ == "__main__":
    test_structural_output()
    test_function_calling()
