"""Configuration for Nexa SDK + LlamaIndex examples."""

import os

NEXA_API_BASE = os.getenv("NEXA_API_BASE", "http://localhost:18181/v1")
NEXA_MODEL = os.getenv("NEXA_MODEL", "NexaAI/Qwen3-0.6B-GGUF")
NEXA_API_KEY = os.getenv("NEXA_API_KEY", "")

# LLM Configuration
LLM_CONFIG = {
    "model": NEXA_MODEL,
    "api_base": NEXA_API_BASE,
    "api_key": NEXA_API_KEY,
    "context_window": 8000,
    "is_chat_model": True,
    "temperature": 0.1,
    "timeout": 60.0,
}
