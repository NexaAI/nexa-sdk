# Copyright 2024-2025 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# serve.py
import requests

BASE_URL = "http://127.0.0.1:18181" 
# BASE_URL = "https://api.hyperlinkos.com" 

ALL_ASR_MODELS = ["NexaAI/parakeet-tdt-0.6b-v2-MLX"]
ALL_INFER_MODELS = ["NexaAI/Qwen3-4B-GGUF"]

class LLMService:

    @staticmethod
    def speech_to_text(base_url, audio, model):
        files = {
            "file": (audio, open(audio, "rb"), "audio/wav")
        }
        
        data = {
            "model": model,
            "language": "en"
        }
        resp = requests.post(f"{base_url}/v1/audio/transcriptions", data=data, files=files)
        return resp.json().get("text", "")

    @staticmethod
    def chat(base_url, messages, model, tools=None):
        body = {
            "model": model,
            "messages": messages,
            "tools": tools if tools else [],
            "enable_think": False
        }
        resp = requests.post(f"{base_url}/v1/chat/completions", json=body)
        return resp.json()
