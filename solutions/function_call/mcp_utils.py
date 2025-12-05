"""MCP utility functions."""

import json
import os
from typing import List, Dict, Any, Optional
from mcp import ClientSession, StdioServerParameters


def mcp_tool_to_openai_format(tool) -> Dict[str, Any]:
    """Convert MCP tool to OpenAI function calling format."""
    properties = {}
    required = []
    
    if tool.inputSchema and "properties" in tool.inputSchema:
        for prop_name, prop_schema in tool.inputSchema["properties"].items():
            properties[prop_name] = {
                "type": prop_schema.get("type", "string"),
                "description": prop_schema.get("description", ""),
            }
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
    tools = []
    for tool in result.tools:
        tools.append(mcp_tool_to_openai_format(tool))
    return tools


def normalize_tool_name(tool_name: str, available_tools: List[Dict[str, Any]]) -> str:
    """Normalize tool name to match available tools."""
    # Common name mappings
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
    
    # First try direct mapping
    if tool_name in name_mappings:
        normalized = name_mappings[tool_name]
        # Verify it exists in available tools
        tool_names = [t.get("function", {}).get("name", "") for t in available_tools]
        if normalized in tool_names:
            return normalized
    
    # If exact match exists, use it
    tool_names = [t.get("function", {}).get("name", "") for t in available_tools]
    if tool_name in tool_names:
        return tool_name
    
    # Try fuzzy matching (replace underscores with hyphens)
    normalized = tool_name.replace("_", "-")
    if normalized in tool_names:
        return normalized
    
    # Return original if no match found
    return tool_name


async def execute_mcp_tool(session: ClientSession, tool_name: str, arguments: Dict[str, Any], available_tools: Optional[List[Dict[str, Any]]] = None) -> str:
    """Execute a tool call via MCP server."""
    try:
        # Normalize tool name if available tools provided
        if available_tools:
            tool_name = normalize_tool_name(tool_name, available_tools)
        
        result = await session.call_tool(tool_name, arguments=arguments)
        return result.model_dump_json(indent=2)
    except Exception as e:
        return f"Error: {str(e)}"


def create_calendar_server(credentials: str) -> StdioServerParameters:
    """Create Google Calendar MCP server parameters."""
    if not credentials:
        credentials = "gcp-oauth.keys.json"
    
    # Check if credentials file exists
    if not os.path.exists(credentials):
        raise FileNotFoundError(
            f"Credentials file not found: {credentials}\n"
            f"Please create the OAuth credentials file at: {os.path.abspath(credentials)}"
        )
    
    # Use absolute path for credentials
    abs_credentials = os.path.abspath(credentials)
    
    return StdioServerParameters(
        command="npx",
        args=["-y", "@cocal/google-calendar-mcp"],
        env={"GOOGLE_OAUTH_CREDENTIALS": abs_credentials},
    )


