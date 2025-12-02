"""
NexaAI LLM Example

This example demonstrates how to use the NexaAI SDK to work with LLM models.
"""

import argparse
import io
import os
from typing import List

from nexaai import LLM, GenerationConfig, ModelConfig, LlmChatMessage as ChatMessage


def main():
    parser = argparse.ArgumentParser(description="NexaAI LLM Example")
    parser.add_argument(
        "-m",
        "--model",
        default="~/.cache/nexa.ai/nexa_sdk/models/Qwen/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf",
        help="Path to the LLM model",
    )
    parser.add_argument("--device", default=None, help="Device to run on")
    parser.add_argument(
        "--max-tokens", type=int, default=128, help="Maximum tokens to generate"
    )
    parser.add_argument(
        "--system", default="You are a helpful assistant.", help="System message"
    )
    parser.add_argument("--plugin-id", default=None, help="Plugin ID to use")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    config = ModelConfig()

    instance: LLM = LLM.from_(model=model_path, plugin_id=args.plugin_id, config=config)

    conversation: List[ChatMessage] = [ChatMessage(role="system", content=args.system)]
    strbuff = io.StringIO()
    print("Multi-round conversation started. Type '/quit' or '/exit' to end.")
    print("=" * 50)
    while True:
        user_input = input("\nUser: ").strip()

        if user_input.startswith("/"):
            cmds = user_input.split()
            if cmds[0] in {"/quit", "/exit", "/q"}:
                print("Goodbye!")
                break
            elif cmds[0] in {"/save", "/s"}:
                instance.save_kv_cache(cmds[1])
                print("KV cache saved to", cmds[1])
                continue
            elif cmds[0] in {"/load", "/l"}:
                instance.load_kv_cache(cmds[1])
                print("KV cache loaded from", cmds[1])
                continue
            elif cmds[0] in {"/reset", "/r"}:
                instance.reset()
                print("Conversation reset")
                continue
            else:
                print("Unknown command")
                continue

        if not user_input:
            print("Please provide an input or type '/quit' to exit.")
            continue

        conversation.append(ChatMessage(role="user", content=user_input))
        prompt = instance.apply_chat_template(conversation)

        strbuff.truncate(0)
        strbuff.seek(0)

        print("Assistant: ", end="", flush=True)
        gen = instance.generate_stream(
            prompt, GenerationConfig(max_tokens=args.max_tokens)
        )
        result = None
        try:
            while True:
                token = next(gen)
                print(token, end="", flush=True)
                strbuff.write(token)
        except StopIteration as e:
            result = e.value

        if result and hasattr(result, "profile_data") and result.profile_data:
            print(f"\n{result.profile_data}")

        conversation.append(ChatMessage(role="assistant", content=strbuff.getvalue()))


if __name__ == "__main__":
    main()
