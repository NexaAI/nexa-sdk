"""
NexaAI LLM Example

This example demonstrates how to use the NexaAI SDK to work with LLM models.
"""

import argparse
import io
import os
from typing import List

from nexaai.llm import LLM, GenerationConfig
from nexaai.common import ModelConfig, ChatMessage


def main():
    parser = argparse.ArgumentParser(description="NexaAI LLM Example")
    parser.add_argument("--model",
                        default="~/.cache/nexa.ai/nexa_sdk/models/Qwen/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf",
                        help="Path to the LLM model")
    parser.add_argument("--device", default="cpu", help="Device to run on")
    parser.add_argument("--max-tokens", type=int, default=100, help="Maximum tokens to generate")
    parser.add_argument("--system", default="You are a helpful assistant.", 
                       help="System message")
    parser.add_argument("--plugin-id", default="cpu_gpu", help="Plugin ID to use")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    m_cfg = ModelConfig()

    instance = LLM.from_(model_path, plugin_id=args.plugin_id, device_id=args.device, m_cfg=m_cfg)

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
        for token in instance.generate_stream(prompt, g_cfg=GenerationConfig(max_tokens=args.max_tokens)):
            print(token, end="", flush=True)
            strbuff.write(token)

        profiling_data = instance.get_profiling_data()
        if profiling_data is not None:
            print(profiling_data)

        conversation.append(ChatMessage(role="assistant", content=strbuff.getvalue()))


if __name__ == "__main__":
    main()
