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

import time

# mock tools

def get_weather(location):
    # get weather
    return f"location: {location} weather: 18℃"

def send_email(content):
    # send email
    return f"email send: content: {content}"

def timenow(unuse):
    # get time
    return f"{time.strftime("%a %b %d %H:%M:%S %Y", time.localtime())}"

def finished():
    # finished
    return "done"


TOOL_FUNCTION=[
    {
        "type": "function",
        "function": {
            "name": "timenow",
            "description": "Return the current date and time.",
            "parameters": {
                "type": "object",
                "properties": {},
                "required": []
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Return the weather for a certain city.",
            "parameters": {
                "type": "object",
                "properties": {
                    "city": {
                        "type": "string",
                        "description": "The city to query weather for."
                    }
                },
                "required": [
                    "city"
                ]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "send_email",
            "description": "Send an email to a recipient containing a message.",
            "parameters": {
                "type": "object",
                "properties": {
                    "to": {
                        "type": "string",
                        "description": "The receiver's name or email address."
                    },
                    "email_message": {
                        "type": "string",
                        "description": "The content of the email message."
                    }
                },
                "required": [
                    "to",
                    "email_message"
                ]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "finished",
            "description": "Call this when the user's goal is complete. No parameters.",
            "parameters": {
                "type": "object",
                "properties": {},
                "required": []
            }
        }
    }
]
