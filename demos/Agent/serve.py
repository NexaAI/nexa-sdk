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