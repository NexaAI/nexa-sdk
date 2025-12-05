"""Model configuration and tool definitions."""

import os
from dataclasses import dataclass
from typing import List, Dict, Any, Optional
from dotenv import load_dotenv

from ..mcp import FunctionRegistry, GoogleCalendarMCP

# Load environment variables
load_dotenv()


@dataclass
class AppModelConfig:
    """Application model configuration."""
    model_path: str
    plugin_id: str
    device_id: str
    max_tokens: int = 2048
    
    @classmethod
    def from_env(cls) -> "AppModelConfig":
        """Create AppModelConfig from environment variables."""
        return cls(
            model_path=os.getenv("MODEL_PATH", "NexaAI/OmniNeural-4B"),
            plugin_id=os.getenv("PLUGIN_ID", "npu"),
            device_id=os.getenv("DEVICE_ID", "npu"),
            max_tokens=int(os.getenv("MAX_TOKENS", "2048"))
        )


SYSTEM_PROMPT = """You are a helpful AI assistant with access to Google Calendar through function calling.

Your goals:
- Understand the user's request about calendar events
- Decide whether a function call is needed to interact with Google Calendar
- If yes, output a structured JSON function call (no explanations before the JSON)
- If no, directly respond to the user in natural language

Available functions:
1. create_calendar_event - Create a new calendar event
2. list_calendar_events - List calendar events
3. update_calendar_event - Update an existing calendar event
4. delete_calendar_event - Delete a calendar event

Output JSON for function calls:
{"name": "function_name", "arguments": {"key": "value"}}

Rules:
- Use JSON format only when calling functions (no explanations before JSON)
- After function execution, provide a natural language summary of the result
- Be helpful and conversational
- Never invent function names that don't exist
- When creating events, parse relative times like "tomorrow at 2pm" or "next Monday at 10am"

Examples:
User: Add a meeting tomorrow at 2pm called "Team Standup"
Assistant: {"name": "create_calendar_event", "arguments": {"summary": "Team Standup", "start_time": "tomorrow at 2pm"}}

User: What events do I have this week?
Assistant: {"name": "list_calendar_events", "arguments": {"max_results": 20}}

User: Hello
Assistant: Hello! How can I help you with your calendar today?
"""


def get_tools(registry: Optional[FunctionRegistry] = None) -> List[Dict[str, Any]]:
    """
    Get tool definitions for function calling.
    
    Args:
        registry: Optional function registry. If provided, will register tools.
        
    Returns:
        List of tool definitions in OpenAI function calling format
    """
    if registry is None:
        registry = FunctionRegistry()
    
    # Initialize Google Calendar MCP
    credentials_file = os.getenv("GOOGLE_CREDENTIALS_FILE", "gcp-oauth.keys.json")
    
    try:
        google_calendar = GoogleCalendarMCP(
            registry=registry,
            credentials_file=credentials_file
        )
    except FileNotFoundError as e:
        print(f"Warning: {e}")
        print("Google Calendar MCP will not be available. Please configure credentials.json")
    except Exception as e:
        print(f"Warning: Failed to initialize Google Calendar MCP: {e}")
        print("Google Calendar MCP will not be available.")
    
    return registry.get_tools()

