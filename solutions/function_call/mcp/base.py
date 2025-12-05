"""Base classes and interfaces for MCP applications."""

from abc import ABC, abstractmethod
from typing import Dict, Callable, Any, List, Optional
import json


class FunctionRegistry:
    """Registry for function implementations."""
    
    def __init__(self):
        self._functions: Dict[str, Callable] = {}
        self._tools: List[Dict[str, Any]] = []
    
    def register(self, name: str, func: Callable, tool_def: Dict[str, Any]):
        """
        Register a function with its tool definition.
        
        Args:
            name: Function name
            func: Function implementation
            tool_def: Tool definition in OpenAI function calling format
        """
        self._functions[name] = func
        self._tools.append(tool_def)
    
    def get_function(self, name: str) -> Optional[Callable]:
        """Get a registered function by name."""
        return self._functions.get(name)
    
    def get_tools(self) -> List[Dict[str, Any]]:
        """Get all registered tool definitions."""
        return self._tools.copy()
    
    def execute(self, name: str, arguments: Dict[str, Any]) -> Any:
        """
        Execute a registered function.
        
        Args:
            name: Function name
            arguments: Function arguments
            
        Returns:
            Function result
            
        Raises:
            KeyError: If function is not registered
        """
        if name not in self._functions:
            raise KeyError(f"Function '{name}' is not registered")
        return self._functions[name](**arguments)


class MCPApp(ABC):
    """Base class for MCP applications."""
    
    def __init__(self, registry: FunctionRegistry):
        """
        Initialize MCP application.
        
        Args:
            registry: Function registry to register tools
        """
        self.registry = registry
        self._register_tools()
    
    @abstractmethod
    def _register_tools(self):
        """Register all tools provided by this MCP application."""
        pass
    
    @abstractmethod
    def get_name(self) -> str:
        """Get the name of this MCP application."""
        pass

