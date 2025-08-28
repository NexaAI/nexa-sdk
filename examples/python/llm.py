"""
NexaAI LLM Example

This example demonstrates how to use the NexaAI SDK to work with LLM models.
"""

import io
import os
from typing import List

from nexaai.llm import LLM, GenerationConfig
from nexaai.common import ModelConfig, ChatMessage


def main():
    # Your model path
    model = os.path.expanduser(
        "~/.cache/nexa.ai/nexa_sdk/models/Qwen/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf")

    # Model configuration
    m_cfg = ModelConfig()

    # Load model
    instance: LLM = LLM.from_(
        model, plugin_id="llama_cpp", device_id="cpu", m_cfg=m_cfg)

    conversation: List[ChatMessage] = [ChatMessage(
        role="system", content="You are a helpful assistant.")]
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

        # Apply the chat template
        prompt = instance.apply_chat_template(conversation)

        strbuff.truncate(0)
        strbuff.seek(0)

        print("Assistant: ", end="", flush=True)
        # Generate the model response
        for token in instance.generate_stream(prompt, g_cfg=GenerationConfig(max_tokens=100)):
            print(token, end="", flush=True)
            strbuff.write(token)

        conversation.append(ChatMessage(
            role="assistant", content=strbuff.getvalue()))


if __name__ == "__main__":
    main()
