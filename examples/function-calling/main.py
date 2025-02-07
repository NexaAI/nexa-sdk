from nexa.gguf.nexa_inference_text import NexaTextInference
from utils import call_function, add_integer, get_weather
from utils import system_prompt

tool_get_weather = {
    "type": "function",
    "function": {
            "name": "get_weather",
            "description": "Get current temperature for provided city.",
            "parameters": {
                "type": "object",
                "properties": {
                    "city": {"type": "string"},
                },
                "required": ["city"],
                "additionalProperties": False
            },
        "strict": True
    }
}

tool_add_integer = {
    "type": "function",
    "function": {
            "name": "add_integer",
            "description": "Returns the addition of input integers.",
            "parameters": {
                "type": "object",
                "properties": {
                    "num1": {"type": "integer", "description": "An integer to add."},
                    "num2": {"type": "integer", "description": "An integer to add."}
                },
                "required": ["number"],
                "additionalProperties": False
            },
        "strict": True
    }
}


if __name__ == "__main__":

    tools = [tool_get_weather, tool_add_integer]

    model = NexaTextInference(
        model_path="Meta-Llama-3.1-8B-Instruct:q4_0", function_calling=True)

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "What's the weather in Paris today?"}
    ]

    responses = model.function_calling(messages=messages, tools=tools)
    for response in responses:
        func_info = response['function']
        func_name = func_info['name']
        func_args = func_info['arguments']

        print('-' * 20)
        print(func_name)
        print(func_args)
        print('-' * 20)
