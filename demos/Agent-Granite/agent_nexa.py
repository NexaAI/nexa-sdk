from __future__ import annotations

import os
import re
import json
import argparse
import requests
import faiss
import numpy as np
from typing import List, Dict, Any, Iterable, Tuple
from serpapi import GoogleSearch

from langchain_core.language_models.llms import LLM



# Nexa config
DEFAULT_MODEL = "NexaAI/granite-4.0-micro-GGUF"
DEFAULT_ENDPOINT = "http://127.0.0.1:18181"
# You can get a free API key from https://serpapi.com/
SEARCH_API_KEY = "7467f292f9d4ce3324da285ca111ea11477ba7fc84ee7e9fa5f867a9d1b35856"

SYSTEM_PROMPT = """
You are Granite Agent, a lightweight on-device AI assistant that can take actions via function calling.

Your goals:
- Understand the user’s request.
- Decide whether a function call is needed.
- If yes, output a structured JSON function call (no explanations).
- If no, directly respond to the user in natural language.

Available Functions:
1. search_web(query: string)

Purpose: Fetch recent or real-time information from the web.
Use when: The user asks about latest, recent, current, trending, or time-sensitive topics.
Examples:
User: What’s the latest AI news this week?
Assistant:
{
  "name": "search_web",
  "arguments": { "query": "latest AI news this week" }
}

2. write_to_file(file_path: string, content: string)

Purpose: Save or append text to a local file.
Use when: The user asks to “save,” “record,” “note down,” or “write” something.

Special rule:
If the user says “save that,” “save this,” or “save the previous answer,”
Use the previous assistant message as the content.

Example:
User: Save that answer to my ‘AI news’ notes.
Assistant:
{
  "name": "write_to_file",
  "arguments": { 
    "file_path": "AI_news.txt"
  }
}

Rules:
- JSON format when calling a function:
  {
    "name": "function_name",
    "arguments": { "..." }
  }

- When you receive the function result, summarize it in 2-3 sentences.
- Keep responses concise, factual, and readable.
- Never hallucinate function names.
- The user interface is real-time and visual, so keep your tone direct and agentic.
- You are a fast and capable assistant running on-device (300M LLM).

Behavior Examples:

Example 1 — Web Search
User: What's the latest news on AI?
Assistant:
{
  "name": "search_web",
  "arguments": { "query": "latest news on AI" } 
}

Example 2 — File Write
User: Save that to my AI chip notes.
Assistant:
{
  "name": "write_to_file",
  "arguments": { 
    "file_path": "AI_chip_notes.txt"
  }
}

Example 3 — No Function Needed
User: Hello!
Assistant: Hi! How can I assist you today?

"""

FUNCTION_TOOLS = [
    {
        "name": "search_web",
        "description": "Searches the web for a given query and returns the latest information.",
        "parameters": {
            "type": "object",
            "properties": {
                "query": {"type": "string", "description": "User search query"}
            },
            "required": ["query"]
        }
    },
    {
        "name": "write_file",
        "description": "Writes text content into a file on the local filesystem.",
        "parameters": {
            "type": "object",
            "properties": {
                "path": {"type": "string"}
            },
            "required": ["path"]
        }
    }
]

def search_web(query: str):
    params = {
        "engine": "google",
        "q": query,
        "google_domain": "google.com",
        "num": "3",
        "start": "10",
        "safe": "active",
        "api_key": SEARCH_API_KEY
    }
    search = GoogleSearch(params)
    results = search.get_dict()
    organic_results = results["organic_results"]
    return organic_results

def write_to_file(file_path: str, content: str):
    with open(file_path, "a", encoding="utf-8") as f:
        f.write(content + "\n")

FUNCTION_REGISTRY = {
    "search_web": search_web,
    "write_to_file": write_to_file
}

def handle_function_call(func_name: str, func_args: dict, model: str, endpoint: str):
    """
    Execute the registered function, print the tool result, then call Nexa to produce
    a natural language summary based on the tool output.
    """
    if isinstance(func_args, str):
        try:
            func_args = json.loads(func_args)
        except json.JSONDecodeError:
            func_args = {}
                
    tool_result = FUNCTION_REGISTRY[func_name](**func_args)
    
    user_followup_prompt = f"""
    You previously decided to call the function `{func_name}` with arguments {func_args}.
    Here is the result returned by that function:

    {tool_result}

    Now, based on this result, please summarize and respond naturally to the user.
    Do NOT call any function again.
    """

    try:
        for piece in stream_call_nexa_chat(model, user_followup_prompt, endpoint):
            yield piece
    except Exception as e:
        print(f"[error] failed to call nexa for followup: {e}")
        return


# Nexa low-level call
def _post_json(url: str, payload: dict, timeout: int = 300) -> dict:
    headers = {"Content-Type": "application/json"}
    resp = requests.post(url, headers=headers, data=json.dumps(payload), timeout=timeout)
    if resp.status_code >= 400:
        raise requests.HTTPError(f"{resp.status_code} {url}\n{resp.text}", response=resp)
    return resp.json()

def call_nexa_chat(model: str, prompt: str, base: str) -> str:
    url = base.rstrip("/") + "/v1/chat/completions"
    data = _post_json(url, {
        "model": model,
        "messages": [{"role": "user", "content": prompt}],
        "stream": False,
        "max_tokens": 512
    })
    
    try:
        return data["choices"][0]["message"]["content"]
    except Exception:
        # tolerate slight variants
        return data.get("text", "") or data.get("response", "")

def call_nexa(prompt: str, model: str, endpoint_base: str) -> str:
    """
    Use /v1/chat/completions endpoint.
    """
    return call_nexa_chat(model, prompt, endpoint_base)

def stream_call_nexa_chat(model: str, prompt: str, base: str):
    """
    Stream /v1/chat/completions.
    Yields incremental text pieces as they arrive.
    """
    url = base.rstrip("/") + "/v1/chat/completions"
    headers = {"Content-Type": "application/json"}
    payload = {
        "model": model, 
        "messages": [{"role": "user", "content": prompt}],
        "stream": True,
        "max_tokens": 512
    }

    with requests.post(url, headers=headers, data=json.dumps(payload), stream=True, timeout=300) as resp:
        resp.raise_for_status()
        for raw in resp.iter_lines(decode_unicode=True):
            if not raw:
                continue
            # typical line: "data: {json}" or "data: [DONE]"
            if raw.startswith("data:"):
                data = raw[len("data:"):].strip()
                if data == "[DONE]":
                    break
                try:
                    obj = json.loads(data)
                except Exception:
                    continue
                # chat stream usually in choices[0].delta.content
                choices = obj.get("choices", [])
                if choices:
                    delta = choices[0].get("delta") or {}
                    piece = delta.get("content", "")
                    if piece:
                        yield piece


def nexa_chat_messages(model: str, messages: list, base: str):
    """
    Use /v1/chat/completions.
    By default this function returns the full response string.
    """
    url = base.rstrip("/") + "/v1/chat/completions"
    data = _post_json(url, {
        "model": model,
        "messages": messages,
        "stream": False,
        "max_tokens": 512
    })
    try:
        return data["choices"][0]["message"]["content"]
    except Exception:
        # tolerate slight variants
        return data.get("text", "") or data.get("response", "")

def nexa_start_search_stream(
        query: str, 
        last_message: str = "",
        model: str = DEFAULT_MODEL, 
        endpoint: str = DEFAULT_ENDPOINT
    ):
    
    messages = [ 
        { "role": "system", "content": SYSTEM_PROMPT },
        { "role": "user", "content": query },
    ]

    try:
        yield json.dumps({"status": "proccess", "message": "Starting analysis..."})
        result = nexa_chat_messages(model, messages, endpoint)
        try:
            parsed = json.loads(result)
            if isinstance(parsed, dict):
                func_name = parsed.get("name")
                func_args = parsed.get("arguments", {})
                if func_name and func_name in FUNCTION_REGISTRY:
                    yield json.dumps({"status": "function", "message": result})
                    if func_name == "write_to_file":
                        file_path = func_args.get("file_path") or func_args.get("path")
                        write_to_file(file_path, last_message)
                        message = f"Successfully saved the previous answer to **{file_path}**. You can check it anytime!"
                        yield json.dumps({"status": "stream", "message": message})
                    else:
                        try:
                            yield json.dumps({"status": "proccess", "message": "Function calling..."})
                            flag = False
                            for piece in handle_function_call(func_name, func_args, model, endpoint):
                                if not flag:
                                    yield json.dumps({"status": "proccess", "message": "Function call finished."})
                                    flag = True
                                else:
                                    yield json.dumps({"status": "stream", "message": piece})
                        except Exception as e:
                            yield json.dumps({"status": "function_call_error", "message": f"{e}"})
                            # try again
                            for piece in handle_function_call(func_name, func_args, model, endpoint):
                                yield json.dumps({"status": "stream", "message": piece})
                else:
                    yield json.dumps({"status": "stream", "message": result})
            else:
                yield json.dumps({"status": "stream", "message": result})
        except json.JSONDecodeError:
            # Not JSON: yield the raw result
            yield json.dumps({"status": "stream", "message": result})

    except requests.HTTPError as e:
        yield json.dumps({"status": "stream", "message": f"{e}"})


# LangChain LLM adapter
class NexaLLM(LLM):
    """A minimal LangChain LLM adapter that calls Nexa's OpenAI-style endpoints."""
    model: str = DEFAULT_MODEL
    endpoint: str = DEFAULT_ENDPOINT

    def _call(self, prompt: str, **kwargs: Any) -> str:
        return call_nexa(prompt, self.model, self.endpoint)

    @property
    def _llm_type(self) -> str:
        return f"nexa:{self.model}"

# CLI
def main():
    ap = argparse.ArgumentParser(description="Function Tool with Nexa SDK server")
    ap.add_argument("--model", default=DEFAULT_MODEL, help="Nexa model name or alias.")
    ap.add_argument("--endpoint", default=DEFAULT_ENDPOINT, help="Nexa base endpoint, e.g. http://127.0.0.1:18181")
    args = ap.parse_args()

    print(f"[info] Ready. Using model={args.model} endpoint={args.endpoint}")
    print("Type your question (or just press Enter to quit):")

    last_message = ""
    while True:
        try:
            q = input("[user] ").strip()
            if not q:
                break
            response = ""
            for piece in nexa_start_search_stream(q, last_message, args.model, args.endpoint):
                print(piece, end="", flush=True)
                response += piece
            last_message = response
            print()
        except KeyboardInterrupt:
            print("\n[info] Bye.")
            break

if __name__ == "__main__":
    main()
