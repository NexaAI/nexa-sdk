#!/usr/bin/env python3

"""
NexaAI VLM Example - Vision Language Model Testing

This example demonstrates how to use the NexaAI SDK to work with Vision Language Models.
It includes basic model initialization, text generation, image understanding, and chat functionality.
"""

import argparse
from typing import List

from nexaai.vlm import VLM, GenerationConfig

from nexaai.common import MultiModalMessage


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--model", type=str, default="ggml-org/Qwen2.5-Omni-3B-GGUF", help="Path to the model")
    parser.add_argument("--mmproj", type=str, default="",
                        help="Path to the mmproj")
    parser.add_argument("--plugin_id", type=str,
                        default="llama_cpp", help="Plugin ID(llama_cpp, mlx)")
    parser.add_argument("--device_id", type=str, default="cpu",
                        help="Device ID(cpu, cuda, metal)")
    args = parser.parse_args()

    # Load model
    try:
        instance: VLM = VLM.from_(
            args.model, mmproj_path=args.mmproj, plugin_id=args.plugin_id, device_id=args.device_id)
    except Exception as e:
        print(f"Error loading model: {e}")
        return

    conversation: List[MultiModalMessage] = []

    print("Multi-round conversation started. Type 'quit' or 'exit' to end.")
    print("=" * 50)

    while True:
        user_input = input("\nUser: ").strip()

        if user_input.lower() in ["quit", "exit", "q"]:
            print("Goodbye!")
            break

        if not user_input:
            print("Please provide an input or type 'quit' to exit.")
            continue

        conversation.append(MultiModalMessage(role="user", content=user_input))

        # Apply the chat template
        try:
            prompt = instance.apply_chat_template(conversation)
        except Exception as e:
            print(f"Error applying chat template: {e}")
            continue

        print("Assistant: ", end="", flush=True)

        try:
            strbuff = ""
            # Generate the model response
            for token in instance.generate_stream(prompt, g_cfg=GenerationConfig(max_tokens=100)):
                print(token, end="", flush=True)
                strbuff += token
            conversation.append(MultiModalMessage(
                role="assistant", content=strbuff))
        except Exception as e:
            print(f"Error generating response: {e}")


if __name__ == "__main__":
    main()
