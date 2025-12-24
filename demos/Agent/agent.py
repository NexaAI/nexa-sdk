# Copyright 2024-2025 Nexa AI, Inc.
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

import json
from serve import LLMService
import tools

SYSTEM_PROMPT = """
You are an expert at breaking down a complex user request into a sequence of function calls. Respect the chronological order of actions described by the user.  

Based on the user's request and the history of previously executed functions, decide on the next function to call to achieve the user's goal.

If the goal is complete and you have the result that you need call the finished function.
If the input does not match any supported function call the finished function.
If the input sounds like a conversation or the user just says thanks for the previous request call the finished function.

Here is the list of supported functions:

- timenow(): return the current date and time
- get_weather(city): return the weather for a certain city.
- send_email(to, email_message): send an email to a recipient containing a message.
- finished: call this function with NO parameters when the user's goal is complete.

You must return exactly one JSON object representing a function call per response.

Respond only with a valid JSON. Do not include comments, explanations, tabs, or extra spaces.
{"function":"function_name","describe":"describe your intent in three words","parameter":"parameter_value or Leave empty string '' if no parameters"}`
"""


class AgentRunner:
    def __init__(self):
        self.history = [
            {"role": "system", "content": SYSTEM_PROMPT}
        ]

    def run(self, base_url, task, model):
        self.history.append({"role": "user", "content": task})

        yield json.dumps({"status": "proccess", "message": "Starting analysis task..."})
        
        while True:
            
            max_retries = 3
            for attempt in range(1, max_retries + 1):
                try:
                    response = LLMService.chat(
                        base_url=base_url, 
                        messages=self.history,
                        model=model
                    )
                    message = response["choices"][0]["message"]["content"]
                    data = json.loads(message)
                    break
                except Exception as e:
                    if attempt < max_retries:
                        continue
                    yield json.dumps({"status": "error", "message": f"{e}"})
                    return

            func = data.get("function")
            param = data.get("parameter")
            describe = data.get("describe")
            yield json.dumps({"status": "function", "message": f"{data}"})
            
            if func == "finished":
                yield json.dumps({"status": "finished", "message": f"{describe}!"})
                return

            if hasattr(tools, func):
                yield json.dumps({"status": "task", "message": f"{describe}..."})
                result = getattr(tools, func)(param)
                self.history.append({
                    "role": "assistant",
                    "content": f"running `{func}`, result: {result}"
                })
            else:
                yield json.dumps({"status": "error", "message": f"unknow func: {func}"})
                break
