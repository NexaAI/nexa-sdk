#!/usr/bin/env python3

"""
NexaAI VLM Example - Llama Model Testing

This example demonstrates how to use the NexaAI SDK to work with Llama models.
It includes basic model initialization, text generation, streaming, and chat template functionality.
"""

import argparse
import io
import os
import re
from typing import List, Optional

from nexaai.vlm import VLM, GenerationConfig
from nexaai.common import ModelConfig, MultiModalMessage, MultiModalMessageContent


def parse_media_from_input(user_input: str) -> tuple[str, Optional[List[str]], Optional[List[str]]]:
    quoted_pattern = r'["\']([^"\']*)["\']'
    quoted_matches = re.findall(quoted_pattern, user_input)

    prompt = re.sub(quoted_pattern, '', user_input).strip()

    image_extensions = {'.png', '.jpg', '.jpeg', '.gif', '.bmp', '.tiff', '.webp'}
    audio_extensions = {'.mp3', '.wav', '.flac', '.aac', '.ogg', '.m4a'}

    image_paths = []
    audio_paths = []

    for quoted_file in quoted_matches:
        if quoted_file:
            if quoted_file.startswith('~'):
                quoted_file = os.path.expanduser(quoted_file)

            if not os.path.exists(quoted_file):
                print(f"Warning: File '{quoted_file}' not found")
                continue

            file_ext = os.path.splitext(quoted_file.lower())[1]
            if file_ext in image_extensions:
                image_paths.append(quoted_file)
            elif file_ext in audio_extensions:
                audio_paths.append(quoted_file)

    return prompt, image_paths if image_paths else None, audio_paths if audio_paths else None


def main():
    parser = argparse.ArgumentParser(description="NexaAI VLM Example")
    parser.add_argument("--model", 
                       default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/gemma-3n-E4B-it-4bit-MLX/model-00001-of-00002.safetensors",
                       help="Path to the VLM model")
    parser.add_argument("--mmproj", default="", help="Path to multimodal projection model")
    parser.add_argument("--device", default="", help="Device to run on")
    parser.add_argument("--max-tokens", type=int, default=100, help="Maximum tokens to generate")
    parser.add_argument("--system", default="You are a helpful assistant.", 
                       help="System message")
    parser.add_argument("--plugin-id", default="cpu_gpu", help="Plugin ID to use")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    m_cfg = ModelConfig()

    print(f"Loading model from {model_path} with plugin {args.plugin_id} and device {args.device}")
    print(f"MMProj path: {args.mmproj}")
    print(f"System message: {args.system}")
    print(f"Max tokens: {args.max_tokens}")
    print(f"Plugin ID: {args.plugin_id}")
    print(f"Device: {args.device}")
    print(f"Model path: {model_path}")
    print(f"MMProj path: {args.mmproj}")
    print(f"System message: {args.system}")
    print(f"Max tokens: {args.max_tokens}")
    print(f"Plugin ID: {args.plugin_id}")
    print(f"Device: {args.device}")
    instance = VLM.from_(name_or_path=model_path, mmproj_path=args.mmproj, 
                        m_cfg=m_cfg, plugin_id=args.plugin_id, device_id=args.device)

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
                contents.append(MultiModalMessageContent(type="image", text=image))
        if audios:
            for audio in audios:
                contents.append(MultiModalMessageContent(type="audio", text=audio))
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

        conversation.append(MultiModalMessage(role="assistant", content=[MultiModalMessageContent(type="text", text=strbuff.getvalue())]))


if __name__ == "__main__":
    main()
