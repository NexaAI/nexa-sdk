#!/usr/bin/env python3
"""NexaAI VLM Function Call Demo with Google Calendar MCP"""

import asyncio
import json
import os
import argparse
import re

from nexaai import GenerationConfig, ModelConfig, VlmChatMessage, VlmContent, setup_logging
from nexaai.vlm import VLM
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

from mcp_utils import get_mcp_tools, execute_mcp_tool


def extract_function_call(text: str):
    """Extract function call JSON from LLM response."""
    text = re.sub(r"<\|[^|]+\|>", "", text.strip())
    match = re.search(r'\{[^{}]*"name"\s*:\s*"([^"]+)"[^{}]*"arguments"\s*:\s*(\{[^{}]*\})[^{}]*\}', text, re.DOTALL)
    if match:
        try:
            return match.group(1), json.loads(match.group(2))
        except:
            pass
    try:
        parsed = json.loads(text)
        if isinstance(parsed, dict) and "name" in parsed:
            return parsed.get("name"), parsed.get("arguments", {})
    except:
        pass
    return None


def build_system_prompt(tools: list) -> str:
    """Build system prompt from tool schemas."""
    tools_descriptions = []
    for t in tools:
        func = t.get('function', {})
        name = func.get('name', '')
        desc = func.get('description', '')
        params = func.get('parameters', {})
        props = params.get('properties', {})
        required = params.get('required', [])
        
        param_list = []
        for param_name, param_info in props.items():
            param_type = param_info.get('type', 'string')
            param_desc = param_info.get('description', '')
            is_required = param_name in required
            req_mark = " (required)" if is_required else " (optional)"
            param_list.append(f"  - {param_name} ({param_type}){req_mark}: {param_desc}")
        
        params_str = "\n".join(param_list) if param_list else "  (no parameters)"
        tools_descriptions.append(f"{name}: {desc}\nParameters:\n{params_str}")
    
    tools_list = "\n\n".join([f"{i+1}. {td}" for i, td in enumerate(tools_descriptions)])
    system_prompt = f"""You are a helpful AI assistant with access to Google Calendar through function calling.

Available functions (use EXACT names and parameter names):
{tools_list}

Output JSON for function calls: {{"name": "function_name", "arguments": {{"parameter_name": "value"}}}}

CRITICAL RULES:
- Use EXACT function names and parameter names as shown above
- All required parameters MUST be included
- Parameter names are case-sensitive
- Use the exact parameter types (string, number, boolean, array, object)
- ALWAYS use function calls when user asks about:
  * Dates, times, schedules, events (use list-events)
  * What to do on a specific date/month (use list-events)
  * Creating/adding calendar events (use create-event)
  * Searching for events (use list-events)
  * Current time (use get-current-time)
- When user asks "what should I do on [date/month]", call list-events to check their calendar"""
    
    return system_prompt


async def main():
    setup_logging()
    
    parser = argparse.ArgumentParser()
    parser.add_argument("--credentials", default="gcp-oauth.keys.json")
    parser.add_argument("--text", help="Text input")
    parser.add_argument("--image", help="Image file path")
    parser.add_argument("--audio", help="Audio file path")
    args = parser.parse_args()
    
    if not args.text and not args.image and not args.audio:
        parser.print_help()
        return

    # Initialize MCP first to get tools
    print("[info] Connecting to Google Calendar MCP server...")
    if not os.path.exists(args.credentials):
        print(f"[error] Credentials file not found: {os.path.abspath(args.credentials)}")
        return

    server = StdioServerParameters(
        command="npx",
        args=["-y", "@cocal/google-calendar-mcp"],
        env={"GOOGLE_OAUTH_CREDENTIALS": os.path.abspath(args.credentials)},
    )

    try:
        async with stdio_client(server) as (read, write):
            async with ClientSession(read, write) as session:
                await session.initialize()
                print("[info] MCP connection established")

                # Get tools
                tools = await get_mcp_tools(session)
                print(f"[info] Found {len(tools)} available tools")

                # only select tools in the list
                tools = [t for t in tools if t.get('function', {}).get('name', '') in ['create-event', 'list-events', 'get-current-time']]
                
                # Build system prompt
                system_prompt = build_system_prompt(tools)
                print(f"[info] System prompt: {system_prompt}")
                vlm:VLM = VLM.from_("NexaAI/OmniNeural-4B", config=ModelConfig(
                    system_prompt=system_prompt,
                    n_ctx=4096, n_threads=0, n_threads_batch=0, n_batch=0, n_ubatch=0, n_seq_max=0, n_gpu_layers=999
                ))
                print("[info] Model loaded successfully")

                conversation = []
                # Build message content from command line arguments
                contents = []
                image_paths = []
                audio_paths = []
                
                if args.text:
                    contents.append(VlmContent(type="text", text=args.text))
                
                if args.image:
                    image_path = os.path.abspath(args.image)
                    if not os.path.exists(image_path):
                        print(f"[error] Image file not found: {image_path}")
                        return
                    image_paths.append(image_path)
                    contents.append(VlmContent(type="image", text=image_path))
                
                if args.audio:
                    audio_path = os.path.abspath(args.audio)
                    if not os.path.exists(audio_path):
                        print(f"[error] Audio file not found: {audio_path}")
                        return
                    audio_paths.append(audio_path)
                    contents.append(VlmContent(type="audio", text=audio_path))

                if not contents:
                    print("[error] No input provided")
                    return

                conversation.append(VlmChatMessage(role="user", contents=contents))

                # Generate
                prompt = vlm.apply_chat_template(conversation, enable_thinking=False)
                print(f"[info] Prompt: {prompt}")
                print("Assistant: ", end="", flush=True)
                
                response_text = ""
                for token in vlm.generate_stream(prompt, config=GenerationConfig(
                    max_tokens=512, image_paths=image_paths or None, audio_paths=audio_paths or None, image_max_length=512
                )):
                    print(token, end="", flush=True)
                    response_text += token
                print()

                # Check function call
                func_call = extract_function_call(response_text)
                print(f"[info] Function call: {func_call}")
                if func_call:
                    func_name, func_args = func_call
                    if func_name and isinstance(func_name, str):
                        max_retries = 3
                        retry_count = 0
                        func_result = None
                        
                        while retry_count <= max_retries:
                            print(f"\n[Function call: {func_name}]")
                            func_result = await execute_mcp_tool(session, func_name, func_args, tools)
                            print(f"[Function result: {func_result}]")
                            
                            # Check if result contains errors
                            try:
                                result_data = json.loads(func_result) if isinstance(func_result, str) else func_result
                                is_error = result_data.get('isError', False)
                                
                                if is_error:
                                    error_text = ""
                                    if isinstance(result_data.get('content'), list):
                                        for item in result_data['content']:
                                            if item.get('type') == 'text':
                                                error_text = item.get('text', '')
                                                break
                                    
                                    # Auto-fix account errors
                                    if "Account" in error_text and "not found" in error_text and "Available accounts:" in error_text:
                                        match = re.search(r'Available accounts:\s*(\w+)', error_text)
                                        if match:
                                            correct_account = match.group(1)
                                            print(f"[info] Auto-fixing account: using '{correct_account}'")
                                            func_args['account'] = correct_account
                                            retry_count += 1
                                            continue
                                        else:
                                            # If account is optional, remove it
                                            if 'account' in func_args:
                                                print(f"[info] Removing invalid account parameter (optional)")
                                                del func_args['account']
                                                retry_count += 1
                                                continue
                                    
                                    # Auto-fix eventId errors (eventId should not be provided for create-event)
                                    if "Invalid event ID" in error_text or ("event ID" in error_text.lower() and "invalid" in error_text.lower()):
                                        if 'eventId' in func_args:
                                            print(f"[info] Removing invalid eventId parameter (not needed for create-event)")
                                            del func_args['eventId']
                                            retry_count += 1
                                            continue
                                    
                                    # Auto-remove optional parameters that cause errors
                                    if ("validation error" in error_text.lower() or "invalid" in error_text.lower()) and retry_count < max_retries:
                                        # Try removing optional parameters that might be invalid
                                        optional_params_to_remove = ['account', 'eventId', 'timeZone', 'fields']
                                        removed = False
                                        for param in optional_params_to_remove:
                                            if param in func_args:
                                                print(f"[info] Removing optional parameter '{param}' due to validation error")
                                                del func_args[param]
                                                removed = True
                                                break
                                        
                                        if removed:
                                            retry_count += 1
                                            continue
                                        else:
                                            # No more optional params to remove, break
                                            break
                                
                                # Success or non-retryable error
                                break
                            except Exception:
                                break

                        # Follow-up response
                        followup = conversation + [
                            VlmChatMessage(role="assistant", contents=[VlmContent(type="text", text=response_text)]),
                            VlmChatMessage(role="user", contents=[VlmContent(type="text", text=f"You called {func_name} with {func_args}. Result: {func_result}. Provide a natural language response. Do NOT call any function again.")])
                        ]
                        print("\nAssistant: ", end="", flush=True)
                        followup_response = ""
                        for token in vlm.generate_stream(
                            vlm.apply_chat_template(followup, enable_thinking=False),
                            config=GenerationConfig(max_tokens=512)
                        ):
                            print(token, end="", flush=True)
                            followup_response += token
                        print()
                else:
                    print(response_text)

    except Exception as e:
        print(f"[error] Failed to initialize MCP connection: {e}")


if __name__ == "__main__":
    asyncio.run(main())
