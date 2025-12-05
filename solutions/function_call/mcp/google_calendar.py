"""Google Calendar MCP implementation using existing MCP server."""

import asyncio
from typing import Dict, Any, List, Optional

try:
    from mcp import ClientSession, StdioServerParameters  # type: ignore
    from mcp.client.stdio import stdio_client  # type: ignore
    MCP_AVAILABLE = True
except ImportError:
    # Fallback if mcp library is not available
    MCP_AVAILABLE = False
    ClientSession = None  # type: ignore
    StdioServerParameters = None  # type: ignore
    stdio_client = None  # type: ignore

from .base import MCPApp, FunctionRegistry


class GoogleCalendarMCP(MCPApp):
    """Google Calendar MCP application using existing MCP server."""
    
    def __init__(
        self,
        registry: FunctionRegistry,
        credentials_file: str = "gcp-oauth.keys.json",
    ):
        """
        Initialize Google Calendar MCP.
        
        Args:
            registry: Function registry
            credentials_file: Path to Google OAuth credentials JSON file
        """
        if not MCP_AVAILABLE:
            raise ImportError(
                "MCP library is not installed. Please install it with: pip install mcp"
            )
        
        self.credentials_file = credentials_file
        # StdioServerParameters is guaranteed to be available if MCP_AVAILABLE is True
        self.server_params = StdioServerParameters(  # type: ignore
            command="npx",
            args=["-y", "@cocal/google-calendar-mcp"],
            env={
                "GOOGLE_OAUTH_CREDENTIALS": credentials_file,
            },
        )
        self.session: Optional[Any] = None  # type: ignore
        self._tools_cache: List[Dict[str, Any]] = []
        super().__init__(registry)
        self._initialize_async()
    
    def _initialize_async(self):
        """Initialize MCP connection asynchronously."""
        try:
            # Run async initialization in a new event loop
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            loop.run_until_complete(self._connect())
        except Exception as e:
            print(f"Warning: Failed to initialize MCP connection: {e}")
            print("MCP functions may not be available.")
    
    async def _connect(self):
        """Connect to MCP server and cache tools."""
        try:
            async with stdio_client(self.server_params) as (read, write):  # type: ignore
                async with ClientSession(read, write) as session:  # type: ignore
                    await session.initialize()
                    self.session = session
                    # Cache available tools
                    result = await session.list_tools()
                    self._tools_cache = [
                        {
                            "type": "function",
                            "function": {
                                "name": tool.name,
                                "description": tool.description or "",
                                "parameters": tool.inputSchema if hasattr(tool, 'inputSchema') else {},
                            }
                        }
                        for tool in result.tools
                    ]
        except Exception as e:
            print(f"Error connecting to MCP server: {e}")
            self.session = None
    
    async def _call_tool_async(self, tool_name: str, arguments: Dict[str, Any]) -> Any:
        """Call MCP tool asynchronously."""
        if not self.session:
            # Reconnect if needed
            await self._connect()
            if not self.session:
                raise RuntimeError("MCP session not available")
        
        async with stdio_client(self.server_params) as (read, write):  # type: ignore
            async with ClientSession(read, write) as session:  # type: ignore
                await session.initialize()
                result = await session.call_tool(tool_name, arguments=arguments)
                return result
    
    def _call_tool_sync(self, tool_name: str, arguments: Dict[str, Any]) -> Any:
        """Call MCP tool synchronously."""
        try:
            loop = asyncio.get_event_loop()
        except RuntimeError:
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
        
        return loop.run_until_complete(self._call_tool_async(tool_name, arguments))
    
    def get_name(self) -> str:
        """Get the name of this MCP application."""
        return "google_calendar"
    
    def _register_tools(self):
        """Register Google Calendar tools from MCP server."""
        # First, try to get tools from MCP server
        if not self._tools_cache:
            # If no cached tools, register common Google Calendar tools
            self._register_default_tools()
        else:
            # Register tools from MCP server
            for tool_def in self._tools_cache:
                tool_name = tool_def["function"]["name"]
                # Create a wrapper function for each tool
                def make_tool_func(name):
                    def tool_func(**kwargs):
                        result = self._call_tool_sync(name, kwargs)
                        # Extract content from MCP result
                        if hasattr(result, 'content'):
                            return result.content
                        elif hasattr(result, 'text'):
                            return result.text
                        elif isinstance(result, dict):
                            return result
                        else:
                            return str(result)
                    return tool_func
                
                self.registry.register(
                    tool_name,
                    make_tool_func(tool_name),
                    tool_def
                )
    
    def _register_default_tools(self):
        """Register default Google Calendar tools if MCP server is not available."""
        # Common Google Calendar MCP tools
        default_tools = [
            {
                "type": "function",
                "function": {
                    "name": "create-event",
                    "description": "Create a new calendar event",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "summary": {"type": "string", "description": "Event title"},
                            "start": {"type": "string", "description": "Start time (ISO 8601)"},
                            "end": {"type": "string", "description": "End time (ISO 8601)"},
                            "description": {"type": "string", "description": "Event description"},
                            "location": {"type": "string", "description": "Event location"},
                        },
                        "required": ["summary", "start"]
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "list-events",
                    "description": "List calendar events",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "timeMin": {"type": "string", "description": "Start time (ISO 8601)"},
                            "timeMax": {"type": "string", "description": "End time (ISO 8601)"},
                            "maxResults": {"type": "integer", "description": "Maximum results"},
                        }
                    }
                }
            },
            {
                "type": "function",
                "function": {
                    "name": "get-current-time",
                    "description": "Get current time in specified timezone",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "timeZone": {"type": "string", "description": "Timezone (e.g., 'Asia/Shanghai')"},
                        }
                    }
                }
            }
        ]
        
        for tool_def in default_tools:
            tool_name = tool_def["function"]["name"]
            def make_tool_func(name):
                def tool_func(**kwargs):
                    try:
                        result = self._call_tool_sync(name, kwargs)
                        if hasattr(result, 'content'):
                            return result.content
                        elif hasattr(result, 'text'):
                            return result.text
                        elif isinstance(result, dict):
                            return result
                        else:
                            return str(result)
                    except Exception as e:
                        return {"error": str(e), "message": f"Failed to call {name}"}
                return tool_func
            
            self.registry.register(
                tool_name,
                make_tool_func(tool_name),
                tool_def
            )
