#!/usr/bin/env python3
"""
Function Call Demo with Google Calendar

This demo showcases the function calling capabilities of NexaAI/OmniNeural-4B model,
integrated with Google Calendar MCP application.
"""

import sys
import io
import argparse
import logging
from typing import List

from nexaai import LLM, GenerationConfig, ModelConfig, LlmChatMessage, setup_logging

from mcp import FunctionRegistry, GoogleCalendarMCP
from models.config import AppModelConfig, SYSTEM_PROMPT, get_tools
from utils.function_parser import extract_function_call, FunctionCall


def handle_function_call(
    func_call: FunctionCall,
    registry: FunctionRegistry,
    model: LLM,
    model_config: AppModelConfig,
    system_prompt: str,
) -> str:
    """
    Execute a function call and generate a natural language response.

    Args:
        func_call: The parsed function call
        registry: Function registry
        model: LLM instance
        model_config: Model configuration
        system_prompt: System prompt

    Returns:
        Natural language response about the function execution result
    """
    try:
        # Execute the function
        result = registry.execute(func_call.name, func_call.arguments)

        # Generate follow-up response
        followup_prompt = f"""You previously decided to call the function `{func_call.name}` with arguments {func_call.arguments}.
Here is the result returned by that function:

{result}

Now, based on this result, please provide a natural language response to the user.
Be helpful and conversational. Do NOT call any function again."""

        # Create model config with system prompt
        config = ModelConfig()
        if hasattr(config, 'system_prompt'):
            config.system_prompt = system_prompt

        # Create LLM instance for follow-up
        llm = LLM.from_(
            model=model_config.model_path,
            plugin_id=model_config.plugin_id,
            config=config,
        )

        messages: List[LlmChatMessage] = [
            LlmChatMessage(role="user", content=followup_prompt)
        ]
        prompt = llm.apply_chat_template(messages)

        response = llm.generate(prompt, GenerationConfig(max_tokens=model_config.max_tokens))
        return response
    except KeyError as e:
        return f"Error: Function '{func_call.name}' is not available. {str(e)}"
    except Exception as e:
        return f"Error executing function '{func_call.name}': {str(e)}"


def main():
    """Main entry point."""
    setup_logging(level=logging.INFO)

    parser = argparse.ArgumentParser(
        description="Function Call Demo with Google Calendar"
    )
    parser.add_argument(
        "-m",
        "--model",
        default=None,
        help="Path to the model (overrides environment variable)",
    )
    parser.add_argument(
        "--plugin-id", default=None, help="Plugin ID (overrides environment variable)"
    )
    parser.add_argument(
        "--device-id", default=None, help="Device ID (overrides environment variable)"
    )
    parser.add_argument(
        "--max-tokens", type=int, default=2048, help="Maximum tokens to generate"
    )
    parser.add_argument(
        "--credentials",
        default="gcp-oauth.keys.json",
        help="Path to Google OAuth credentials file (for MCP server)",
    )
    args = parser.parse_args()

    # Load model configuration
    model_config = AppModelConfig.from_env()
    if args.model:
        model_config.model_path = args.model
    if args.plugin_id:
        model_config.plugin_id = args.plugin_id
    if args.device_id:
        model_config.device_id = args.device_id
    if args.max_tokens:
        model_config.max_tokens = args.max_tokens

    print(f"[info] Initializing model: {model_config.model_path}")
    print(f"[info] Plugin: {model_config.plugin_id}, Device: {model_config.device_id}")

    # Initialize function registry and MCP applications
    registry = FunctionRegistry()

    try:
        GoogleCalendarMCP(
            registry=registry, credentials_file=args.credentials
        )
        print("[info] Google Calendar MCP initialized successfully")
    except FileNotFoundError as e:
        print(f"[warning] {e}")
        print("[warning] Google Calendar functions will not be available")
    except Exception as e:
        print(f"[warning] Failed to initialize Google Calendar MCP: {e}")
        print("[warning] Google Calendar functions will not be available")

    # Get tools for function calling
    tools = get_tools(registry)
    if not tools:
        print("[warning] No tools available. Function calling will not work.")

    # Initialize LLM
    try:
        config = ModelConfig()
        if hasattr(config, 'system_prompt'):
            config.system_prompt = SYSTEM_PROMPT
        
        llm = LLM.from_(
            model=model_config.model_path,
            plugin_id=model_config.plugin_id,
            config=config,
        )
        print("[info] Model loaded successfully")
    except Exception as e:
        print(f"[error] Failed to load model: {e}")
        sys.exit(1)

    # Initialize conversation
    conversation: List[LlmChatMessage] = []
    if hasattr(config, 'system_prompt') and config.system_prompt:
        conversation.append(LlmChatMessage(role="system", content=SYSTEM_PROMPT))
    strbuff = io.StringIO()

    print("\n" + "=" * 60)
    print("Function Call Demo with Google Calendar")
    print("=" * 60)
    print("Type '/quit' or '/exit' to end the conversation")
    print("Type '/reset' to reset the conversation")
    print("=" * 60 + "\n")

    while True:
        try:
            user_input = input("User: ").strip()

            if not user_input:
                print("Please provide an input or type '/quit' to exit.")
                continue

            # Handle commands
            if user_input.startswith("/"):
                cmds = user_input.split()
                if cmds[0] in {"/quit", "/exit", "/q"}:
                    print("Goodbye!")
                    break
                elif cmds[0] in {"/reset", "/r"}:
                    llm.reset()
                    conversation = []
                    print("Conversation reset")
                    continue
                else:
                    print("Unknown command. Available commands: /quit, /exit, /reset")
                    continue

            # Add user message to conversation
            conversation.append(LlmChatMessage(role="user", content=user_input))

            # Apply chat template with tools
            prompt = llm.apply_chat_template(conversation, tools=tools)

            # Generate response
            strbuff.truncate(0)
            strbuff.seek(0)

            print("Assistant: ", end="", flush=True)
            gen = llm.generate_stream(prompt, GenerationConfig(max_tokens=model_config.max_tokens))

            try:
                while True:
                    token = next(gen)
                    print(token, end="", flush=True)
                    strbuff.write(token)
            except StopIteration:
                pass

            response_text = strbuff.getvalue()

            # Check for function call
            func_call = extract_function_call(response_text)

            if func_call:
                print(f"\n[Function call: {func_call.name}]")
                print(f"[Arguments: {func_call.arguments}]")

                # Execute function and get follow-up response
                followup_response = handle_function_call(
                    func_call, registry, llm, model_config, SYSTEM_PROMPT
                )

                print(f"\n{followup_response}")

                # Add assistant message with function call result
                conversation.append(
                    LlmChatMessage(
                        role="assistant",
                        content=f"Function call: {func_call.name}\nResult: {followup_response}",
                    )
                )
            else:
                # No function call, just add the response
                conversation.append(
                    LlmChatMessage(role="assistant", content=response_text)
                )

            print()  # New line after response

        except KeyboardInterrupt:
            print("\n\n[info] Interrupted by user. Goodbye!")
            break
        except Exception as e:
            print(f"\n[error] An error occurred: {e}")
            logging.exception("Error in main loop")
            continue


if __name__ == "__main__":
    main()
