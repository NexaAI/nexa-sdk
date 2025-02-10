import json

from nexa.gguf.nexa_inference_text import NexaTextInference
from utils import suppress_stdout_stderr
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
    # Include add_integer in the available tools to check if the model could selects the correct function.
    tools = [tool_get_weather, tool_add_integer]
    with suppress_stdout_stderr():
        model = NexaTextInference(
            model_path="Meta-Llama-3.1-8B-Instruct:q4_0", function_calling=True)
    print('-' * 20)
    print('Successfully loaded model.')
    print('-' * 20)
    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "What's the weather in New York today?"}
    ]

    responses = model.function_calling(messages=messages, tools=tools)
    print('Received function calling arguments.')
    print('-' * 20)
    response = responses[0]
    func_info = response['function']
    func_name = func_info['name']
    func_args = json.loads(func_info['arguments'])

    recv = call_function(func_name, **func_args)

    print('Received weather data from wttr.in api.')
    print('-' * 20)
    print(recv[:recv.find("Location")].strip())

    # OpenAI-style fucntion calling

    # messages.append({"role": "assistant", "content": None, "function_call": response['function']})
    # messages.append({"role": "function", "name": func_name, "content": str(res)})
    # output = model.create_chat_completion(messages=messages)
    # print(f'Model Output: {output['choices'][0]['message']['content']}')
    # print('-' * 20)
