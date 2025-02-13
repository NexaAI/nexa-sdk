import json

from nexa.gguf.nexa_inference_text import NexaTextInference
from utils import suppress_stdout_stderr
from utils import call_function, add_integer, get_weather, get_factorial
from utils import system_prompt, tool_get_weather, tool_add_integer, tool_get_factorial

MODEL_PATH = "Meta-Llama-3.1-8B-Instruct:q4_0"

if __name__ == "__main__":
    # Include add_integer and get_factorial in the available tools as well to check if the model could selects the correct function.
    tools = [tool_get_weather, tool_add_integer, tool_get_factorial]
    with suppress_stdout_stderr():
        model = NexaTextInference(
            model_path=MODEL_PATH, function_calling=True)
    print('-' * 20)
    print(f'Successfully loaded model: {MODEL_PATH}.')
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
