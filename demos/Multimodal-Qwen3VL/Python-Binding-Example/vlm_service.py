#!/usr/bin/env python3
"""
NexaAI VLM Service
"""

import os
import argparse
import io
import re

from typing import List, Optional

from nexaai.vlm import VLM, GenerationConfig
from nexaai.common import ModelConfig, MultiModalMessage, MultiModalMessageContent, SamplerConfig

default_system_prompt = """
You are a witty, sarcastic, and sassy AI who comments on images with humor and attitude.
You always respond in JSON format according to the grammar below.
Your humor should be clever and lighthearted, never mean or offensive.

Grammar:
{
  "response": string,  // Your main sarcastic or witty comment about the image
  "comeback"?: string  // Optional humorous follow-up or playful comeback
}

Constraints:
- Always produce valid JSON.
- The response should reflect a strong personality (sarcastic, witty, or sassy).
- The "comeback" is optional — include it only if it naturally fits.
"""

class VLMService:
    """
    NexaAI vlm service
    """

    def __init__(
        self, 
        model_name: str, 
        mmproj_name: str,
        plugin_id: str = "nexaml", 
        device: str = "gpu",
        system_prompt: str = ""
    ):
        self.model_name = model_name
        self.mmproj_name = mmproj_name
        self.plugin_id = plugin_id
        self.device = device
        if len(system_prompt.strip()) > 0:
            self.system_prompt = system_prompt
        else:
            self.system_prompt = default_system_prompt
            
        self._model = None
        self._conversation: List[MultiModalMessage] = []
        self._init_model()

    def _init_model(self):
        """Initialize the VLM model and conversation context."""
        m_cfg = ModelConfig()
        
        print("[debug: ]", self.model_name, self.plugin_id, self.device, self.mmproj_name)
        self._model = VLM.from_(
            name_or_path=self.model_name,
            m_cfg=m_cfg,
            plugin_id=self.plugin_id,
            device_id=self.device,
            mmproj_path=self.mmproj_name
        )
        
        # initialize conversation with system prompt
        self._conversation = [
            MultiModalMessage(
                role="system",
                content=[MultiModalMessageContent(type="text", text=self.system_prompt)]
            )
        ]

    # -------------------------------------------------------------
    # Public API
    # -------------------------------------------------------------
    def stream_response(
        self,
        prompt: str,
        images: Optional[List[str]] = None,
        audios: Optional[List[str]] = None,
        max_tokens: int = 200,
        temperature: float = 0.7,
        top_p: float = 0.9
    ):
        """
        Stream generate

        Args:
            prompt: User input text
            images: List of image paths (optional)
            audios: List of audio paths (optional)
            max_tokens: Maximum generation length

        Returns:
            dict: { "text": Model output text, "images": Image paths}
        """
        contents = []
        if prompt:
            contents.append(MultiModalMessageContent(type="text", text=prompt))
        if images:
            for img in images:
                contents.append(MultiModalMessageContent(type="image", path=img))
        if audios:
            for audio in audios:
                contents.append(MultiModalMessageContent(type="audio", path=audio))

        user_msg = MultiModalMessage(role="user", content=contents)
        self._conversation.append(user_msg)

        prompt = self._model.apply_chat_template(self._conversation)
     
        sampler_config = SamplerConfig()
        sampler_config.temperature = temperature
        sampler_config.top_p = top_p
        sampler_config.grammar_string=r"""
        char ::= [^"\\\x7F\x00-\x1F] | [\\] (["\\bfnrt] | "u" [0-9a-fA-F]{4})
        comeback-kv ::= "\"comeback\"" space ":" space string
        response-kv ::= "\"response\"" space ":" space string
        root ::= "{" space response-kv ( "," space ( comeback-kv ) )? "}" space
        space ::= | " " | "\n"{1,2} [ \t]{0,20}
        string ::= "\"" char* "\"" space
        """
    
        gen_cfg = GenerationConfig(
            max_tokens=max_tokens, 
            sampler_config=sampler_config,
            image_paths=images,
            audio_paths=audios
        )
        
        
        strbuff = io.StringIO()
        strbuff.truncate(0)
        strbuff.seek(0)
        for token in self._model.generate_stream(prompt, gen_cfg):
            strbuff.write(token)
            yield token
        
        self._conversation.append(MultiModalMessage(role="assistant", content=[MultiModalMessageContent(type="text", text=strbuff.getvalue())]))

    def update_system_prompt(self, system_prompt: str):
        """
        Update system prompt and reset conversation
        """
        self.system_prompt = system_prompt
        self.reset()
        print(f"[Info] System prompt updated to: {system_prompt}")
    
    def reset(self):
        """
        Reset conversation context
        """
        self._conversation = [
            MultiModalMessage(
                role="system",
                content=[MultiModalMessageContent(type="text", text=self.system_prompt)]
            )
        ]
        self._model.reset()

    def save_cache(self, path: str):
        """Save KV cache"""
        self._model.save_kv_cache(path)

    def load_cache(self, path: str):
        """Load KV cache"""
        self._model.load_kv_cache(path)
            

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
                       default="NexaAI/qwen2.5vl",
                       help="Path to the VLM model")
    parser.add_argument("--device", default="", help="Device to run on")
    parser.add_argument("--max-tokens", type=int, default=100, help="Maximum tokens to generate")
    parser.add_argument("--system", default="You are a helpful assistant.", 
                       help="System message")
    parser.add_argument("--plugin-id", default="cpu_gpu", help="Plugin ID to use")
    
    args = parser.parse_args()
    # Create VLM service via the viewmodel so UI can share the same instance
    vlm_service = vlm_viewmodel.create(
        model_name=args.model,
        plugin_id=args.plugin_id,
        device=args.device,
        system_prompt=args.system,
    )
    
    print("NexaAI VLM Service is ready. Type 'exit' to quit.")
    while True:
        user_input = input("User: ")
        if user_input.lower() == "exit":
            break

        prompt, image_paths, audio_paths = parse_media_from_input(user_input)
        print("[Debug]:", prompt, image_paths, audio_paths) 
        
        flag = False
        for token in vlm_service.stream_response(
            prompt=prompt,
            images=image_paths,
            audios=audio_paths,
            max_tokens=args.max_tokens
        ):
            if not flag:
                print("Assistant: ", end="", flush=True)
                flag = True
            print(token, end="", flush=True)
        print()
    
if __name__ == "__main__":
    main()