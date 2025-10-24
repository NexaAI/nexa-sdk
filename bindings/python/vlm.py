#!/usr/bin/env python3

"""
NexaAI VLM Example - Llama Model Testing

This example demonstrates how to use the NexaAI SDK to work with Llama models.
It includes basic model initialization, text generation, streaming, and chat template functionality.
"""

import os
import argparse
import io
import shlex
from pathlib import Path
from typing import Optional, List, Tuple

from nexaai.vlm import VLM, GenerationConfig
from nexaai.common import ModelConfig, MultiModalMessage, MultiModalMessageContent


def parse_media_from_input(user_input: str) -> Tuple[str, Optional[List[str]], Optional[List[str]]]:
    tokens = shlex.split(user_input, posix=False)
    image_exts = {'.png', '.jpg', '.jpeg', '.gif', '.bmp', '.tiff', '.webp'}
    audio_exts = {'.mp3', '.wav', '.flac', '.aac', '.ogg', '.m4a'}

    image_paths, audio_paths, prompt_parts = [], [], []

    for token in tokens:
        path = Path(os.path.expanduser(token))
        if path.exists():
            ext = path.suffix.lower()
            if ext in image_exts:
                image_paths.append(str(path))
            elif ext in audio_exts:
                audio_paths.append(str(path))
            else:
                prompt_parts.append(token)
        else:
            prompt_parts.append(token)

    prompt = ' '.join(prompt_parts).strip()
    return prompt, image_paths or None, audio_paths or None


def main():
    parser = argparse.ArgumentParser(description="NexaAI VLM Example")
    parser.add_argument("--model",
                        default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/gemma-3n-E4B-it-4bit-MLX/model-00001-of-00002.safetensors",
                        help="Path to the VLM model")
    parser.add_argument("--device", default="", help="Device to run on")
    parser.add_argument("--max-tokens", type=int, default=100,
                        help="Maximum tokens to generate")
    parser.add_argument("--system", default="You are a helpful assistant.",
                        help="System message")
    parser.add_argument("--plugin-id", default="cpu_gpu",
                        help="Plugin ID to use")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    m_cfg = ModelConfig()

    instance = VLM.from_(name_or_path=model_path, m_cfg=m_cfg,
                         plugin_id=args.plugin_id, device_id=args.device)

    conversation: List[MultiModalMessage] = [MultiModalMessage(role="system",
                                                               content=[MultiModalMessageContent(type="text", text=args.system)])]
    strbuff = io.StringIO()

    print("Multi-round conversation started. Type '/quit' or '/exit' to end.")
    print("=" * 50)

    while True:
        user_input = input("\nUser: ").strip()
        if not user_input:
            print("Please provide an input or type '/quit' to exit.")
            continue

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

        prompt, images, audios = parse_media_from_input(user_input)

        contents = []
        if prompt:
            contents.append(MultiModalMessageContent(type="text", text=prompt))
        if images:
            for image in images:
                contents.append(MultiModalMessageContent(
                    type="image", path=image))
        if audios:
            for audio in audios:
                contents.append(MultiModalMessageContent(
                    type="audio", path=audio))
        conversation.append(MultiModalMessage(role="user", content=contents))

        # Apply the chat template
        prompt = instance.apply_chat_template(conversation)
        strbuff.truncate(0)
        strbuff.seek(0)

        print("Assistant: ", end="", flush=True)
        for token in instance.generate_stream(prompt, g_cfg=GenerationConfig(max_tokens=args.max_tokens, image_paths=images, audio_paths=audios)):
            print(token, end="", flush=True)
            strbuff.write(token)

        # Get profiling data
        profiling_data = instance.get_profiling_data()
        if profiling_data is not None:
            print(profiling_data)

        conversation.append(MultiModalMessage(role="assistant", content=[
                            MultiModalMessageContent(type="text", text=strbuff.getvalue())]))


if __name__ == "__main__":
    main()
