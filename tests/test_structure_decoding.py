import json
from nexa.gguf import NexaTextInference
from pathlib import Path

def test_structural_output():
    nexa = NexaTextInference(model_path="Meta-Llama-3.1-8B-Instruct:q4_0")
    prompt = "Emily Carter, a 32-year-old owner, drives a 2023 Audi Q5 available in White, Black, and Gray, equipped with Bang & Olufsen audio (19 speakers, Bluetooth), 10 airbags, 8 parking sensors, lane assist, and a turbocharged inline-4 engine delivering 261 horsepower and a top speed of 130 mph."
    schema_abs_path = (Path(__file__).parent / "structure_decoding_resources/schema.json").resolve()
    response = nexa.structure_output(prompt=prompt, json_schema_path=schema_abs_path)
    print(f"response: {json.dumps(response, indent=4)}")

def echo_integer(number):
    return number

def test_function_calling():
    tools = [
        {
            "type": "function",
            "function": {
                "name": "add_integer",
                "description": "Returns the addition of input integers.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "number1": {"type": "integer", "description": "An integer to add."},
                        "number2": {"type": "integer", "description": "An integer to add."} 
                    },
                    "required": ["number"],
                    "additionalProperties": False
                },
                "strict": True
            }
        }
    ]
    messages = [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Please echo the number 142."}
    ]
    
    nexa = NexaTextInference(model_path="Meta-Llama-3.1-8B-Instruct:q4_0")
    response = nexa.function_calling(messages=messages, tools=tools)
    print(response)


if __name__ == "__main__":
    # test_structural_output()
    test_function_calling()
