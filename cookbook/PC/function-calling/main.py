# Copyright 2024-2026 Nexa AI, Inc.
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

#!/usr/bin/env python3
"""NexaAI VLM Function Call Demo with Google Calendar MCP"""

import asyncio
import json
import os
import sys
import argparse
import re
from dataclasses import dataclass
from typing import List, Dict, Any, Optional

from nexaai import GenerationConfig, ModelConfig, VlmChatMessage, VlmContent, setup_logging
from nexaai.vlm import VLM
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client


def _convert_schema_property(prop_schema: Dict[str, Any]) -> Dict[str, Any]:
    """Recursively convert a schema property, handling nested objects."""
    result = {
        "type": prop_schema.get("type", "string"),
    }
    
    if "description" in prop_schema:
        result["description"] = prop_schema["description"]
    
    # Handle nested objects
    if prop_schema.get("type") == "object" and "properties" in prop_schema:
        nested_props = {}
        nested_required = []
        
        for nested_name, nested_schema in prop_schema["properties"].items():
            nested_props[nested_name] = _convert_schema_property(nested_schema)
            if nested_name in prop_schema.get("required", []):
                nested_required.append(nested_name)
        
        result["properties"] = nested_props
        if nested_required:
            result["required"] = nested_required
    
    # Handle arrays of objects
    if prop_schema.get("type") == "array" and "items" in prop_schema:
        result["items"] = _convert_schema_property(prop_schema["items"])
    
    return result


def mcp_tool_to_openai_format(tool) -> Dict[str, Any]:
    """Convert MCP tool to OpenAI function calling format."""
    properties = {}
    required = []
    
    if tool.inputSchema and "properties" in tool.inputSchema:
        for prop_name, prop_schema in tool.inputSchema["properties"].items():
            properties[prop_name] = _convert_schema_property(prop_schema)
            if tool.inputSchema.get("required") and prop_name in tool.inputSchema["required"]:
                required.append(prop_name)
    
    return {
        "type": "function",
        "function": {
            "name": tool.name,
            "description": tool.description or "",
            "parameters": {
                "type": "object",
                "properties": properties,
                "required": required,
            }
        }
    }


async def get_mcp_tools(session: ClientSession) -> List[Dict[str, Any]]:
    """Get tools from MCP server and convert to OpenAI format."""
    result = await session.list_tools()
    return [mcp_tool_to_openai_format(tool) for tool in result.tools]


def normalize_tool_name(tool_name: str, available_tools: List[Dict[str, Any]]) -> str:
    """Normalize tool name to match available tools."""
    name_mappings = {
        "create_calendar_event": "create-event",
        "create-event": "create-event",
        "list_calendar_events": "list-events",
        "list-events": "list-events",
        "update_calendar_event": "update-event",
        "update-event": "update-event",
        "delete_calendar_event": "delete-event",
        "delete-event": "delete-event",
        "get_current_time": "get-current-time",
        "get-current-time": "get-current-time",
    }
    
    tool_names = [t.get("function", {}).get("name", "") for t in available_tools]
    
    if tool_name in name_mappings:
        normalized = name_mappings[tool_name]
        if normalized in tool_names:
            return normalized
    
    if tool_name in tool_names:
        return tool_name
    
    normalized = tool_name.replace("_", "-")
    return normalized if normalized in tool_names else tool_name


async def execute_mcp_tool(session: ClientSession, tool_name: str, arguments: Dict[str, Any], 
                           available_tools: Optional[List[Dict[str, Any]]] = None) -> str:
    """Execute a tool call via MCP server."""
    try:
        if available_tools:
            tool_name = normalize_tool_name(tool_name, available_tools)
        result = await session.call_tool(tool_name, arguments=arguments)
        return result.model_dump_json(indent=2)
    except Exception as e:
        return f"Error: {str(e)}"


def create_calendar_server(credentials: str) -> StdioServerParameters:
    """Create Google Calendar MCP server parameters."""
    if not os.path.exists(credentials):
        raise FileNotFoundError(
            f"Credentials file not found: {credentials}\n"
            f"Please create the OAuth credentials file at: {os.path.abspath(credentials)}"
        )
    return StdioServerParameters(
        command="npx",
        args=["-y", "@cocal/google-calendar-mcp"],
        env={"GOOGLE_OAUTH_CREDENTIALS": os.path.abspath(credentials)},
    )


def extract_function_call(text: str):
    """Extract function call JSON from LLM response."""
    if not text:
        return None
    
    text = re.sub(r"<\|[^|]+\|>", "", text.strip())
    
    try:
        parsed = json.loads(text)
        if isinstance(parsed, dict) and "name" in parsed:
            return parsed.get("name"), parsed.get("arguments", {})
    except json.JSONDecodeError:
        json_start = text.find('{')
        if json_start == -1:
            return None
        
        brace_count = 0
        for i in range(json_start, len(text)):
            if text[i] == '{':
                brace_count += 1
            elif text[i] == '}':
                brace_count -= 1
                if brace_count == 0:
                    json_str = text[json_start:i + 1]
                    try:
                        parsed = json.loads(json_str)
                        if isinstance(parsed, dict) and "name" in parsed:
                            return parsed.get("name"), parsed.get("arguments", {})
                    except json.JSONDecodeError:
                        pass
                    break
    
    return None


def _format_nested_properties(props: Dict[str, Any], required: List[str], prefix: str = "", indent: int = 2) -> List[str]:
    """Recursively format nested object properties with required field indicators."""
    param_list = []
    indent_str = " " * indent
    
    for param_name, param_info in props.items():
        param_type = param_info.get('type', 'string')
        param_desc = param_info.get('description', '')
        is_required = param_name in required
        full_param_name = f"{prefix}.{param_name}" if prefix else param_name
        
        # Handle nested objects
        if param_type == 'object' and 'properties' in param_info:
            nested_props = param_info.get('properties', {})
            nested_required = param_info.get('required', [])
            req_mark = " (REQUIRED)" if is_required else ""
            
            if param_desc:
                param_list.append(f"{indent_str}{param_name} (object){req_mark}: {param_desc}")
            else:
                param_list.append(f"{indent_str}{param_name} (object){req_mark}")
            
            # Recursively format nested properties
            nested_params = _format_nested_properties(
                nested_props, nested_required, full_param_name, indent + 2
            )
            param_list.extend(nested_params)
        else:
            req_mark = " (REQUIRED)" if is_required else ""
            if param_desc:
                param_list.append(f"{indent_str}{param_name} ({param_type}){req_mark}: {param_desc}")
            else:
                param_list.append(f"{indent_str}{param_name} ({param_type}){req_mark}")
    
    return param_list


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
        
        # Format parameters with nested object support
        param_list = _format_nested_properties(props, required)
        
        params_str = "\n".join(param_list) if param_list else "  (no parameters)"
        
        # Highlight required parameters at the top
        required_params = [p for p in required]
        required_str = f"\n  REQUIRED parameters: {', '.join(required_params)}" if required_params else ""
        
        tools_descriptions.append(f"{name}: {desc}{required_str}\n{params_str}")
    
    tools_list = "\n\n".join([f"{i+1}. {td}" for i, td in enumerate(tools_descriptions)])
    
    # Add example for create-event
    example_json = """{
  "name": "create-event",
  "arguments": {
    "calendarId": "primary",
    "summary": "Meeting",
    "start": "2025-01-01T10:00:00",
    "end": "2025-01-01T11:00:00"
  }
}"""
    
    return f"""You are a calendar assistant. When the user requests calendar actions, respond with ONLY a JSON object in this format:

{{"name": "function_name", "arguments": {{"param": "value"}}}}

CRITICAL RULES:
- You MUST include ALL required parameters (marked as REQUIRED)
- For nested objects, ALL required fields within the object must be included
- If a parameter is marked as REQUIRED, it cannot be omitted
- Output ONLY valid JSON, no other text before or after
- Use exact function and parameter names (case-sensitive)

Example for create-event:
{example_json}

Available functions:
{tools_list}

IMPORTANT: Before creating events, you may need to call get-current-time first to get accurate date/time context. Always include calendarId="primary" for create-event unless specified otherwise.
"""


def _handle_function_call_error(error_text: str, func_args: Dict[str, Any]) -> bool:
    """Handle function call errors and auto-fix parameters. Returns True if should retry."""
    # Auto-fix account errors
    if "Account" in error_text and "not found" in error_text and "Available accounts:" in error_text:
        match = re.search(r'Available accounts:\s*(\w+)', error_text)
        if match:
            func_args['account'] = match.group(1)
            return True
        elif 'account' in func_args:
            del func_args['account']
            return True
    
    # Auto-fix eventId errors
    if "Invalid event ID" in error_text or ("event ID" in error_text.lower() and "invalid" in error_text.lower()):
        if 'eventId' in func_args:
            del func_args['eventId']
            return True
    
    # Auto-remove optional parameters that cause errors
    if "validation error" in error_text.lower() or "invalid" in error_text.lower():
        optional_params = ['account', 'eventId', 'timeZone', 'fields']
        for param in optional_params:
            if param in func_args:
                del func_args[param]
                return True
    
    return False


async def _execute_with_retry(session: ClientSession, func_name: str, func_args: Dict[str, Any], 
                               tools: List[Dict[str, Any]], max_retries: int = 3) -> str:
    """Execute function call with automatic error handling and retry."""
    retry_count = 0
    func_result = ""
    
    while retry_count <= max_retries:
        func_result = await execute_mcp_tool(session, func_name, func_args, tools)
        
        try:
            result_data = json.loads(func_result) if isinstance(func_result, str) else func_result
            if result_data.get('isError', False):
                error_text = ""
                if isinstance(result_data.get('content'), list):
                    for item in result_data['content']:
                        if item.get('type') == 'text':
                            error_text = item.get('text', '')
                            break
                
                if retry_count < max_retries and _handle_function_call_error(error_text, func_args):
                    retry_count += 1
                    continue
            break
        except Exception:
            break
    
    return func_result or ""


def init_vlm(tools: List[Dict[str, Any]]) -> VLM:
    """Initialize VLM with tools."""
    system_prompt = build_system_prompt(tools)
    print('[debug] system_prompt:', system_prompt)
    return VLM.from_("NexaAI/OmniNeural-4B", config=ModelConfig(
        system_prompt=system_prompt,
        n_ctx=4096, n_threads=0, n_threads_batch=0, n_batch=0, 
        n_ubatch=0, n_seq_max=0, n_gpu_layers=999
    ))



@dataclass
class FunctionCallAgentResult:
    """Result of function call agent execution."""
    func_name: Optional[str]
    func_result: Optional[str]
    response_text: str

async def call_agent(
    vlm: VLM,
    session: ClientSession,
    tools: List[Dict[str, Any]],
    text: Optional[str] = None,
    image: Optional[str] = None,
    audio: Optional[str] = None
) -> FunctionCallAgentResult:
    if not text and not image and not audio:
        raise ValueError("At least one of text, image, or audio must be provided")
    
    contents = []
    image_paths = []
    audio_paths = []
    
    if image:
        image_path = os.path.abspath(image)
        if not os.path.exists(image_path):
            raise FileNotFoundError(f"Image file not found: {image_path}")
        image_paths.append(image_path)
        contents.append(VlmContent(type="image", text=image_path))
    
    if audio:
        audio_path = os.path.abspath(audio)
        if not os.path.exists(audio_path):
            raise FileNotFoundError(f"Audio file not found: {audio_path}")
        audio_paths.append(audio_path)
        contents.append(VlmContent(type="audio", text=audio_path))
    
    if text:
        contents.append(VlmContent(type="text", text=text))

    conversation = [VlmChatMessage(role="user", contents=contents)]
    
    # Generate initial response
    prompt = vlm.apply_chat_template(conversation)
    print('[debug] prompt:', prompt)
    print('[debug] generate_stream...')
    response_text = ""
    for token in vlm.generate_stream(prompt, config=GenerationConfig(
        max_tokens=2048, image_paths=image_paths or None, 
        audio_paths=audio_paths or None, image_max_length=512
    )):
        print(token, end="", flush=True)
        response_text += token
    print()
    print('[debug] response_text:', response_text)
    func_call = extract_function_call(response_text)
    if not func_call:
        print(f"[error] Failed to extract function call from response")
        return FunctionCallAgentResult(
            func_name=None,
            func_result=None,
            response_text=response_text
        )
    
    func_name, func_args = func_call
    if func_name and isinstance(func_name, str):
        print('[debug] calling function:', func_name)
        func_result = await _execute_with_retry(session, func_name, func_args, tools)
        print('[debug] func_result:', func_result)
        
        # Parse function result to extract success/error message
        result_message = ""
        try:
            result_data = json.loads(func_result) if isinstance(func_result, str) else func_result
            if result_data.get('isError', False):
                # Extract error message
                if isinstance(result_data.get('content'), list):
                    for item in result_data['content']:
                        if item.get('type') == 'text':
                            result_message = item.get('text', '')
                            break
            else:
                # Extract success message or summary
                if isinstance(result_data.get('content'), list):
                    for item in result_data['content']:
                        if item.get('type') == 'text':
                            result_message = item.get('text', '')
                            break
        except Exception:
            result_message = str(func_result)
        
        followup = conversation + [
            VlmChatMessage(role="assistant", contents=[VlmContent(type="text", text=response_text)]),
            VlmChatMessage(role="user", contents=[VlmContent(type="text", 
                text=f"Function execution completed. Result: {result_message}\n\n"
                     f"Now respond to the user in natural language. You are in RESPONSE MODE, not function calling mode.\n"
                     f"- DO NOT output any JSON format\n"
                     f"- DO NOT use {{}} brackets\n"
                     f"- DO NOT call any function\n"
                     f"- Just speak naturally like a helpful assistant\n"
                     f"- Tell the user what happened with the calendar event in a friendly way")])
        ]
        followup_response = ""
        for token in vlm.generate_stream(
            vlm.apply_chat_template(followup, enable_thinking=False),
            config=GenerationConfig(max_tokens=2048)
        ):
            followup_response += token
        return FunctionCallAgentResult(
            func_name=func_name,
            func_result=func_result,
            response_text=followup_response
        )
    
    return FunctionCallAgentResult(
        func_name=None,
        func_result=None,
        response_text=response_text
    )


async def call_agent_wrapper(
    text: Optional[str] = None,
    image: Optional[str] = None,
    audio: Optional[str] = None,
    credentials: str = "gcp-oauth.keys.json"
) -> FunctionCallAgentResult:
    setup_logging()

    if not text and not image and not audio:
        raise ValueError("At least one of text, image, or audio must be provided") 
    
    server = create_calendar_server(credentials)
    async with stdio_client(server) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            tools = await get_mcp_tools(session)
            tools = [t for t in tools if t.get('function', {}).get('name', '') in 
                    ['create-event', 'get-current-time']]
            
            vlm = init_vlm(tools)
            result = await call_agent(vlm, session, tools, text, image, audio)
            return result
    
    
async def main():
    """Command-line interface for the agent."""
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
    
    server = create_calendar_server(args.credentials)
    async with stdio_client(server) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            tools = await get_mcp_tools(session)
            tools = [t for t in tools if t.get('function', {}).get('name', '') in 
                    ['create-event', 'get-current-time']]
            
            vlm = init_vlm(tools)
            result = await call_agent(vlm, session, tools, args.text, args.image, args.audio)

    if result.response_text:
        print(result.response_text)
    

if __name__ == "__main__":
    asyncio.run(main())
