#!/usr/bin/env python3

"""
NexaAI LLM Example - Llama Model Testing

This example demonstrates how to use the NexaAI SDK to work with Llama models.
It includes basic model initialization, text generation, streaming, and chat template functionality.
"""

import argparse
from typing import List
from nexaai.llm import LLM, GenerationConfig
from nexaai.common import ModelConfig, ChatMessage


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", type=str, default="Qwen/Qwen3-0.6B-GGUF", help="Path to the model")
    parser.add_argument("--plugin_id", type=str, default="llama_cpp", help="Plugin ID(llama_cpp, mlx)")
    parser.add_argument("--device_id", type=str, default="cpu", help="Device ID(cpu, cuda, metal)")
    args = parser.parse_args()

    # Model configuration
    cfg = ModelConfig(
        n_ctx=4096,
        n_threads_batch=12,
        n_batch=512,
        n_ubatch=512,
        n_seq_max=1
    )

    # Load model
    try:
        llm_instance: LLM = LLM.from_(args.model, plugin_id=args.plugin_id, device_id=args.device_id)
    except Exception as e:
        print(f"Error loading model: {e}")
        return

    conversation: List[ChatMessage] = []
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

        conversation.append(ChatMessage(role="user", content=user_input))

        # Apply the chat template
        try:
            prompt = llm_instance.apply_chat_template(conversation)
        except Exception as e:
            print(f"Error applying chat template: {e}")
            continue

        print("Assistant: ", end="", flush=True)

        try:
            strbuff = ""
            # Generate the model response
            for token in llm_instance.generate_stream(prompt, g_cfg=GenerationConfig(max_tokens=100)):
                print(token, end="", flush=True)
                strbuff += token
            conversation.append(ChatMessage(role="assistant", content=strbuff))
        except Exception as e:
            print(f"Error generating response: {e}")


if __name__ == "__main__":
    main()
