# serve.py
import requests

BASE_URL = "http://127.0.0.1:18181"  
# BASE_URL = "https://api.hyperlinkos.com"  
CHAT_URL = f"{BASE_URL}/v1/chat/completions"
ASR_URL  = f"{BASE_URL}/v1/audio/transcriptions"
ASR_REPO_ID = "NexaAI/parakeet-tdt-0.6b-v2-MLX"
INFER_REPO_ID = "NexaAI/Qwen3-4B-GGUF"

class LLMService:

    @staticmethod
    def speech_to_text(audio):
        files = {
            "file": (audio, open(audio, "rb"), "audio/wav")
        }
        
        data = {
            "model": ASR_REPO_ID,
            "language": "en"
        }
        resp = requests.post(ASR_URL, data=data, files=files)
        return resp.json().get("text", "")

    @staticmethod
    def chat(messages, tools=None):
        body = {
            "model": INFER_REPO_ID,
            "messages": messages,
            "tools": tools if tools else [],
            "enable_think": False
        }
        resp = requests.post(CHAT_URL, json=body)
        return resp.json()