from __future__ import annotations
import os
import re
import json
import argparse
import requests
from typing import Any, Tuple, Optional, List
from serpapi import GoogleSearch

from nexaai.llm import LLM, GenerationConfig
from nexaai.common import ModelConfig, ChatMessage

# Nexa config
DEFAULT_MODEL = "NexaAI/Granite-4-Micro-NPU"
# DEFAULT_MODEL = "NexaAI/granite-4-Nano-NPU"

# You can get a free API key from https://serpapi.com/
SEARCH_API_KEY = "7467f292f9d4ce3324da285ca111ea11477ba7fc84ee7e9fa5f867a9d1b35856"

# ... existing code ...

SYSTEM_PROMPT = """You are Granite Agent with function calling.

Your goals:
- Understand the user's request.
- Decide whether a function call is needed.
- If yes, output a structured JSON function call (no explanations).
- If no, directly respond to the user in natural language.

Functions:
1. search_web(query: string) - Web search
2. write_to_file(file_path: string) - Save text to file

Output JSON for function calls:
{"name": "function_name", "arguments": {"key": "value"}}

Rules:
- JSON only for functions (no explanations)
- Summarize results in 2-3 sentences
- Never invent function names

Examples:
User: Latest AI news?
Assistant: {"name": "search_web", "arguments": {"query": "latest AI news"}}

User: Save that.
Assistant: {"name": "write_to_file", "arguments": {"file_path": "notes.txt"}}

User: Hello
Assistant: How can I assist you today?
"""

FUNCTION_TOOLS = [
    {
        "type":"function",
        "function": {
            "name": "search_web",
            "description": "Searches the web for a given query and returns the latest information.",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {"type": "string", "description": "User search query"}
                },
                "required": ["query"],
             }
        },
    },
    {
        "type":"function",
        "function": {
            "name": "write_to_file",
            "description": "Writes text content into a file on the local filesystem.",
            "parameters": {
                "type": "object",
                "properties": {"path": {"type": "string"}},
                "required": ["path"],
            }
        },
    }
]

def search_web(query: str):
    """Search the web using SerpAPI"""
    params = {
        "engine": "google",
        "q": query,
        "google_domain": "google.com",
        "num": "2",
        "start": "1",
        "safe": "active",
        "api_key": SEARCH_API_KEY,
    }
    search = GoogleSearch(params)
    results = search.get_dict()
    organic_results = results.get("organic_results", [])

    # Print search results for debugging
    print(f"\n[Web Search Results for '{query}']:")
    for idx, result in enumerate(organic_results, 1):
        print(f"{idx}. {result.get('title', 'N/A')}")
        print(f"   URL: {result.get('link', 'N/A')}")
        print(f"   Snippet: {result.get('snippet', 'N/A')[:200]}...")
        print()

    return organic_results

def write_to_file(file_path: str, content: str):
    """Write content to a file"""
    with open(file_path, "a", encoding="utf-8") as f:
        f.write(content + "\n")

FUNCTION_REGISTRY = {"search_web": search_web, "write_to_file": write_to_file}

def handle_function_call(func_name: str, func_args: dict, model: str):
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

    # Customize prompt based on function type
    if func_name == "search_web":
        user_followup_prompt = f"""
        You called search_web with query: {func_args.get('query')}
        
        Here are the search results:
        {tool_result}
        
        Please provide a short summary of these web search results in 2-3 bullet points. For each result:
        - Include the title
        - Mention the source/website
        - Summarize the key information from the snippet
        - Make it informative and comprehensive
        
        Be verbose and helpful. Do NOT call any function again.
        """
    else:
        user_followup_prompt = f"""
        You previously decided to call the function `{func_name}` with arguments {func_args}.
        Here is the result returned by that function:

        {tool_result}

        Now, based on this result, please summarize and respond naturally to the user.
        Do NOT call any function again.
        """

    try:
        for piece in nexa_chat_stream(model, user_followup_prompt):
            yield piece
    except Exception as e:
        print(f"[error] failed to call nexa for followup: {e}")
        yield f"\n[Error: {e}]"
        return

def nexa_chat_stream(model: str, prompt: str):
    """
    Generate streaming conversation with local LLM 
    """
    m_cfg = ModelConfig()
    llm = LLM.from_(model, plugin_id="npu", device_id="npu", m_cfg=m_cfg)
    
    messages: List[ChatMessage] = [ChatMessage(role="user", content=prompt)]
    prompt = llm.apply_chat_template(messages)
    
    for token in llm.generate_stream(
        prompt, g_cfg=GenerationConfig(max_tokens=512)
    ):
        yield token

def nexa_chat_completion(model: str, messages: list):
    """
    Non-streaming conversation with local LLM
    """
    m_cfg = ModelConfig()
    llm = LLM.from_(model, plugin_id="npu", device_id="npu", m_cfg=m_cfg)
    prompt = llm.apply_chat_template(messages, tools=FUNCTION_TOOLS)
    
    return llm.generate(prompt, g_cfg=GenerationConfig(max_tokens=512))

def extract_function_call(text: str) -> Optional[Tuple[str, dict]]:
    """
    Extract function call JSON from LLM response using regex.
    Handles cases where JSON is embedded in text or has extra tokens.

    Returns:
        Tuple of (function_name, arguments) if found, None otherwise
    """
    # Clean up common LLM artifacts
    text = text.strip()

    # Remove special tokens like <|end_of_text|>, <|im_end|>, etc.
    text = re.sub(r"<\|[^|]+\|>", "", text)
    text = text.strip()

    # Try to find JSON-like structure with "name" and "arguments" fields
    # Pattern matches: { "name": "...", "arguments": { ... } }
    json_pattern = (
        r'\{[^{}]*"name"\s*:\s*"([^"]+)"[^{}]*"arguments"\s*:\s*(\{[^}]*\})[^{}]*\}'
    )

    match = re.search(json_pattern, text, re.DOTALL)
    if match:
        func_name = match.group(1)
        try:
            # Try to parse the arguments
            args_str = match.group(2)
            func_args = json.loads(args_str)
            return func_name, func_args
        except json.JSONDecodeError:
            # If arguments parsing fails, try to extract manually
            # Pattern for simple key-value like { "query": "..." }
            args_match = re.search(r'"(\w+)"\s*:\s*"([^"]*)"', match.group(2))
            if args_match:
                return func_name, {args_match.group(1): args_match.group(2)}

    # Fallback: try to parse the entire cleaned text as JSON
    try:
        parsed = json.loads(text)
        if isinstance(parsed, dict) and "name" in parsed:
            func_name = parsed.get("name")
            func_args = parsed.get("arguments", {})
            return func_name, func_args
    except json.JSONDecodeError:
        pass

    return None


def nexa_start_search_stream(
    query: str,
    last_message: str = "",
    model: str = DEFAULT_MODEL,
):
    """
    Main agent function that handles user query and function calling.
    Yields JSON-formatted status messages.
    """
    messages = [
        {"role": "system", "content": SYSTEM_PROMPT},
        {"role": "user", "content": query},
    ]

    try:
        yield json.dumps({"status": "proccess", "message": "Starting analysis..."})
        result = nexa_chat_completion(model, messages)

        # Try to extract function call using helper function
        function_call = extract_function_call(result)

        if function_call:
            func_name, func_args = function_call

            # Validate function exists in registry
            if func_name in FUNCTION_REGISTRY:
                yield json.dumps(
                    {
                        "status": "function",
                        "message": json.dumps(
                            {"name": func_name, "arguments": func_args}
                        ),
                    }
                )

                if func_name == "write_to_file":
                    yield json.dumps({"status": "proccess", "message": "Function calling..."})
                    
                    file_path = func_args.get("file_path") or func_args.get("path")
                    write_to_file(file_path, last_message)
                    
                    yield json.dumps(
                        {
                            "status": "proccess",
                            "message": "Function call finished.",
                        }
                    )
                    
                    message = f"Successfully saved the previous answer to **{file_path}**. You can check it anytime!"
                    yield json.dumps({"status": "stream", "message": message})
                else:
                    # Execute the function and stream results
                    try:
                        yield json.dumps(
                            {"status": "proccess", "message": "Function calling..."}
                        )
                        flag = False
                        for piece in handle_function_call(
                            func_name, func_args, model
                        ):
                            if not flag:
                                yield json.dumps(
                                    {
                                        "status": "proccess",
                                        "message": "Function call finished.",
                                    }
                                )
                                flag = True
                            
                            yield json.dumps({"status": "stream", "message": piece})
                    except Exception as e:
                        yield json.dumps(
                            {"status": "function_call_error", "message": f"{e}"}
                        )
                        # try again
                        try:
                            for piece in handle_function_call(
                                func_name, func_args, model
                            ):
                                yield json.dumps({"status": "stream", "message": piece})
                        except Exception as retry_error:
                            yield json.dumps(
                                {
                                    "status": "error",
                                    "message": f"Retry failed: {retry_error}",
                                }
                            )
            else:
                # Function not in registry, treat as regular response
                yield json.dumps({"status": "stream", "message": result})
        else:
            # No function call detected, treat as regular response
            yield json.dumps({"status": "stream", "message": result})

    except (ConnectionError, TimeoutError) as e:
        yield json.dumps({"status": "error", "message": f"Connection error: {e}"})
    except requests.HTTPError as e:
        yield json.dumps({"status": "error", "message": f"HTTP error: {e}"})
    except Exception as e:
        yield json.dumps({"status": "error", "message": f"Unexpected error: {e}"})


# CLI
def main():
    """Main CLI entry point"""
    ap = argparse.ArgumentParser(description="Function Tool with Nexa SDK server")
    ap.add_argument("--model", default=DEFAULT_MODEL, help="Nexa model path.")
    args = ap.parse_args()

    print(f"[info] Ready. Using model={args.model}")
    print("Type your question (or just press Enter to quit):")

    last_message = ""
    while True:
        try:
            q = input("[user] ").strip()
            if not q:
                break

            print("[assistant] ", end="", flush=True)
            response_content = ""

            # Parse JSON responses and display nicely
            for piece in nexa_start_search_stream(
                q, last_message, args.model
            ):
                try:
                    parsed = json.loads(piece)
                    status = parsed.get("status")
                    message = parsed.get("message", "")

                    if status == "proccess":
                        print(f"\n[{message}]", end=" ", flush=True)
                    elif status == "function":
                        print(f"\n[Function call: {message}]", flush=True)
                    elif status == "stream":
                        print(message, end="", flush=True)
                        response_content += message
                    elif status == "error":
                        print(f"\n[ERROR: {message}]", flush=True)
                    elif status == "function_call_error":
                        print(f"\n[Function call error: {message}]", flush=True)
                    else:
                        print(message, end="", flush=True)
                        response_content += message
                except json.JSONDecodeError:
                    # If not JSON, just print it
                    print(piece, end="", flush=True)
                    response_content += piece

            last_message = response_content
            print()  # newline after response

        except KeyboardInterrupt:
            print("\n[info] Bye.")
            break
        except Exception as e:
            print(f"\n[error] {e}")
            continue


if __name__ == "__main__":
    main()
