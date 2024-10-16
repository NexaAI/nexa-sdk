import os
from pathlib import Path
from enum import Enum

# Paths for caching, model hub, and tokens
NEXA_CACHE_ROOT = Path(os.getenv("NEXA_CACHE_ROOT") or "~/.cache/nexa").expanduser()
NEXA_TOKEN_PATH = NEXA_CACHE_ROOT / "token"
NEXA_MODELS_HUB_DIR = NEXA_CACHE_ROOT / "hub"
NEXA_MODEL_EVAL_RESULTS_PATH = NEXA_CACHE_ROOT / "eval"
NEXA_MODELS_HUB_OFFICIAL_DIR = NEXA_MODELS_HUB_DIR / "official"
NEXA_MODELS_HUB_HF_DIR = NEXA_MODELS_HUB_DIR / "huggingface"
NEXA_MODEL_LIST_PATH = NEXA_MODELS_HUB_DIR / "model_list.json"

# URLs and buckets
NEXA_API_URL = "https://model-hub-backend.nexa4ai.com"
NEXA_OFFICIAL_BUCKET = "https://public-storage.nexa4ai.com/"

# Nexa logo
NEXA_LOGO = """
      _|    _|  _|_|_|  _|    _|    _|_|      _|_|    _|_|_|_|
      _|_|  _|  _|       _|  _|   _|    _|  _|    _|     _|
      _|_|_|_|  _|_|_|     _|     _|_|_|_|  _|_|_|_|     _|
      _|  _|_|  _|        _| _|   _|    _|  _|    _|     _|
      _|    _|  _|_|_|  _|    _|  _|    _|  _|    _|  _|_|_|_|
"""

# Model producer info
PRODUCER_INFO = {
    "producer_version": "0.0.0",
    "doc_string": "Model exported by Nexa.ai",
}


class ModelType(Enum):
    NLP = "NLP"
    COMPUTER_VISION = "Computer Vision"
    AUDIO = "Audio"
    MULTIMODAL = "Multimodal"
    TEXT_EMBEDDING = "Text Embedding"


NEXA_RUN_MODEL_MAP_TEXT = {
    "octopus-v2": "Octopus-v2:q4_0",
    "octopus-v4": "Octopus-v4:q4_0",
    "gpt2": "gpt2:q4_0",
    "tinyllama": "TinyLlama-1.1B-Chat-v1.0:fp16",
    "llama2": "Llama-2-7b-chat:q4_0",
    "llama3": "Meta-Llama-3-8B-Instruct:q4_0",
    "llama3.1": "Meta-Llama-3.1-8B-Instruct:q4_0",
    "llama3.2": "Llama3.2-3B-Instruct:q4_0",
    "gemma": "gemma-1.1-2b-instruct:q4_0",
    "gemma2": "gemma-2-2b-instruct:q4_0",
    "qwen1.5": "Qwen1.5-7B-Instruct:q4_0",
    "qwen2": "Qwen2-1.5B-Instruct:q4_0",
    "qwen2.5": "Qwen2.5-1.5B-Instruct:q4_0",
    "mistral": "Mistral-7B-Instruct-v0.3:q4_0",
    "codegemma": "codegemma-2b:q4_0",
    "codellama": "CodeLlama-7b-Instruct:q4_0",
    "codeqwen": "Qwen2.5-Coder-1.5B-Instruct:q4_0",
    "mathqwen": "Qwen2.5-Math-1.5B-Instruct:q4_0",
    "deepseek-coder": "deepseek-coder-1.3b-instruct:q4_0",
    "dolphin-mistral": "dolphin-2.8-mistral-7b:q4_0",
    "phi2": "Phi-2:q4_0",
    "phi3": "Phi-3-mini-128k-instruct:q4_0",
    "phi3.5": "Phi-3.5-mini-instruct:q4_0",
    "llama2-uncensored": "Llama2-7b-chat-uncensored:q4_0",
    "llama3-uncensored": "Llama3-8B-Lexi-Uncensored:q4_K_M",
    "openelm": "OpenELM-3B:q4_K_M",
}

NEXA_RUN_MODEL_MAP_ONNX = {
    "gpt2": "gpt2:onnx-cpu-int8",
    "phi3": "Phi-3-mini-4k-instruct:onnx-cpu-int4",
    "llama2": "Llama-2-7b-chat:onnx-cpu-int8",
    "llama3": "Meta-Llama-3-8B-Instruct:onnx-cpu-int8",
    "llama3.1": "Meta-Llama-3.1-8B-Instruct:onnx-cpu-int8",
    "qwen2": "Qwen2-1.5B-Instruct:onnx-cpu-int8",
    "gemma": "gemma-2b-instruct:onnx-cpu-int8",
    "gemma1.1": "gemma-1.1-2b-instruct:onnx-cpu-int8",
    "mistral": "Mistral-7B-Instruct-v0.3:onnx-cpu-int8",
    "phi3v": "Phi-3-vision-128k-instruct:onnx-cpu-int4",
    "sd1-5": "stable-diffusion-v1-5:onnx-cpu-fp32",
    "lcm-dreamshaper": "lcm-dreamshaper-v7:onnx-cpu-fp32",
    "whisper-tiny.en": "whisper-tiny.en:onnx-cpu-fp32",
    "whisper-tiny": "whisper-tiny:onnx-cpu-fp32",
    "whisper-small.en": "whisper-small.en:onnx-cpu-fp32",
    "whisper-small": "whisper-small:onnx-cpu-fp32",
    "whisper-base.en": "whisper-base.en:onnx-cpu-fp32",
    "whisper-base": "whisper-base:onnx-cpu-fp32",
    "ljspeech": "ljspeech-jets:onnx-cpu-fp32",
}

NEXA_RUN_MODEL_MAP_VLM = {
    "nanollava": "nanoLLaVA:model-fp16",
    "nanoLLaVA:fp16": "nanoLLaVA:model-fp16",
    "llava-phi3": "llava-phi-3-mini:model-q4_0",
    "llava-phi-3-mini:q4_0": "llava-phi-3-mini:model-q4_0",
    "llava-phi-3-mini:fp16": "llava-phi-3-mini:model-fp16",
    "llava-llama3": "llava-llama-3-8b-v1.1:model-q4_0",
    "llava-llama-3-8b-v1.1:q4_0": "llava-llama-3-8b-v1.1:model-q4_0",
    "llava-llama-3-8b-v1.1:fp16": "llava-llama-3-8b-v1.1:model-fp16",
    "llava1.6-mistral": "llava-v1.6-mistral-7b:model-q4_0",
    "llava-v1.6-mistral-7b:q4_0": "llava-v1.6-mistral-7b:model-q4_0",
    "llava-v1.6-mistral-7b:fp16": "llava-v1.6-mistral-7b:model-fp16",
    "llava1.6-vicuna": "llava-v1.6-vicuna-7b:model-q4_0",
    "llava-v1.6-vicuna-7b:q4_0": "llava-v1.6-vicuna-7b:model-q4_0",
    "llava-v1.6-vicuna-7b:fp16": "llava-v1.6-vicuna-7b:model-fp16",
}

NEXA_RUN_MODEL_MAP_VOICE = {
    "whisper-large": "whisper-large:bin-large-v3",
    "whisper-tiny": "whisper-tiny:bin-tiny",
    "faster-whisper-tiny": "faster-whisper-tiny:bin-cpu-fp16",
    "faster-whisper-tiny.en": "faster-whisper-tiny.en:bin-cpu-fp16",
    "faster-whisper-small": "faster-whisper-small:bin-cpu-fp16",
    "faster-whisper-small.en": "faster-whisper-small.en:bin-cpu-fp16",
    "faster-whisper-medium": "faster-whisper-medium:bin-cpu-fp16",
    "faster-whisper-medium.en": "faster-whisper-medium.en:bin-cpu-fp16",
    "faster-whisper-base": "faster-whisper-base:bin-cpu-fp16",
    "faster-whisper-base.en": "faster-whisper-base.en:bin-cpu-fp16",
    "faster-whisper-large": "faster-whisper-large-v3:bin-cpu-fp16",
    "faster-whisper-large-turbo": "faster-whisper-large-v3-turbo:bin-cpu-fp16",
}

NEXA_RUN_MODEL_MAP_FUNCTION_CALLING = {
    "llama2-function-calling": "Llama2-7b-function-calling:q4_K_M",
    "Llama2-7b-function-calling:fp16": "Llama2-7b-function-calling:fp16",
    "Llama2-7b-function-calling:q2_K": "Llama2-7b-function-calling:q2_K",
    "Llama2-7b-function-calling:q3_K_L": "Llama2-7b-function-calling:q3_K_L",
    "Llama2-7b-function-calling:q3_K_M": "Llama2-7b-function-calling:q3_K_M",
    "Llama2-7b-function-calling:q3_K_S": "Llama2-7b-function-calling:q3_K_S",
    "Llama2-7b-function-calling:q4_K_M": "Llama2-7b-function-calling:q4_K_M",
    "Llama2-7b-function-calling:q4_K_S": "Llama2-7b-function-calling:q4_K_S",
    "Llama2-7b-function-calling:q5_K_M": "Llama2-7b-function-calling:q5_K_M",
    "Llama2-7b-function-calling:q5_K_S": "Llama2-7b-function-calling:q5_K_S",
    "Llama2-7b-function-calling:q6_K": "Llama2-7b-function-calling:q6_K",
    "Llama2-7b-function-calling:q8_0": "Llama2-7b-function-calling:q8_0",
}

NEXA_RUN_PROJECTOR_MAP = {
    "nanollava": "nanoLLaVA:projector-fp16",
    "nanoLLaVA:fp16": "nanoLLaVA:projector-fp16",
    "llava-phi3": "llava-phi-3-mini:projector-q4_0",
    "llava-phi-3-mini:q4_0": "llava-phi-3-mini:projector-q4_0",
    "llava-phi-3-mini:fp16": "llava-phi-3-mini:projector-fp16",
    "llava-llama3": "llava-llama-3-8b-v1.1:projector-q4_0",
    "llava-llama-3-8b-v1.1:q4_0": "llava-llama-3-8b-v1.1:projector-q4_0",
    "llava-llama-3-8b-v1.1:fp16": "llava-llama-3-8b-v1.1:projector-fp16",
    "llava1.6-mistral": "llava-v1.6-mistral-7b:projector-q4_0",
    "llava-v1.6-mistral-7b:q4_0": "llava-v1.6-mistral-7b:projector-q4_0",
    "llava-v1.6-mistral-7b:fp16": "llava-v1.6-mistral-7b:projector-fp16",
    "llava1.6-vicuna": "llava-v1.6-vicuna-7b:projector-q4_0",
    "llava-v1.6-vicuna-7b:q4_0": "llava-v1.6-vicuna-7b:projector-q4_0",
    "llava-v1.6-vicuna-7b:fp16": "llava-v1.6-vicuna-7b:projector-fp16",
}

NEXA_RUN_T5XXL_MAP = {
    "flux": "FLUX.1-schnell:t5xxl-q4_0",
    "FLUX.1-schnell:q4_0": "FLUX.1-schnell:t5xxl-q4_0",
    "FLUX.1-schnell:q5_0": "FLUX.1-schnell:t5xxl-q5_0",
    "FLUX.1-schnell:q5_1": "FLUX.1-schnell:t5xxl-q5_1",
    "FLUX.1-schnell:q8_0": "FLUX.1-schnell:t5xxl-q8_0",
    "FLUX.1-schnell:fp16": "FLUX.1-schnell:t5xxl-fp16",
}

NEXA_RUN_MODEL_MAP_IMAGE = {
    "sd1-4": "stable-diffusion-v1-4:q4_0",
    "sd1-5": "stable-diffusion-v1-5:q4_0",
    "sd2-1": "stable-diffusion-v2-1:q4_0",
    "sd3": "stable-diffusion-3-medium:q4_0",
    "sdxl-turbo": "sdxl-turbo:q8_0",
    "flux": "FLUX.1-schnell:q4_0",
    "lcm-dreamshaper": "lcm-dreamshaper-v7:fp16",
    "anything-lcm": "anything-v30-LCM:fp16",
    "hassaku-lcm": "hassaku-hentai-model-v13-LCM:fp16",
}

NEXA_RUN_MODEL_MAP_FLUX = {
    "flux": "FLUX.1-schnell:flux1-schnell-q4_0",
    "FLUX.1-schnell:q4_0": "FLUX.1-schnell:flux1-schnell-q4_0",
    "FLUX.1-schnell:q5_0": "FLUX.1-schnell:flux1-schnell-q5_0",
    "FLUX.1-schnell:q5_1": "FLUX.1-schnell:flux1-schnell-q5_1",
    "FLUX.1-schnell:q8_0": "FLUX.1-schnell:flux1-schnell-q8_0",
    "FLUX.1-schnell:fp16": "FLUX.1-schnell:flux1-schnell-fp16",
}

NEXA_RUN_MODEL_MAP_TEXT_EMBEDDING = {
    "mxbai": "mxbai-embed-large-v1:fp16",
    "mxbai-embed-large-v1": "mxbai-embed-large-v1:fp16",
    "mxbai-embed-large-v1:fp16": "mxbai-embed-large-v1:fp16",
    "nomic": "nomic-embed-text-v1.5:fp16",
    "nomic-embed-text-v1.5": "nomic-embed-text-v1.5:fp16",
    "nomic-embed-text-v1.5:fp16": "nomic-embed-text-v1.5:fp16",
    "all-MiniLM": "all-MiniLM-L6-v2:fp16",
    "all-MiniLM-L6-v2": "all-MiniLM-L6-v2:fp16",
    "all-MiniLM-L6-v2:fp16": "all-MiniLM-L6-v2:fp16",
    "all-MiniLM-L12-v2": "all-MiniLM-L12-v2:fp16",
    "all-MiniLM-L12-v2:fp16": "all-MiniLM-L12-v2:fp16",
}

NEXA_RUN_MODEL_MAP = {
    **NEXA_RUN_MODEL_MAP_TEXT,
    **NEXA_RUN_MODEL_MAP_VLM,
    **NEXA_RUN_MODEL_MAP_IMAGE,
    **NEXA_RUN_MODEL_MAP_VOICE,
    **NEXA_RUN_MODEL_MAP_FUNCTION_CALLING,
    **NEXA_RUN_MODEL_MAP_FLUX,
    **NEXA_RUN_MODEL_MAP_TEXT_EMBEDDING,
}

NEXA_RUN_CHAT_TEMPLATE_MAP = {
    "llama2": "llama-2",
    "llama-2-7b-chat": "llama-2",
    "llama3": "llama-3",
    "meta-llama-3-8b-instruct": "llama-3",
    "llama3.1": "llama-3",
    "meta-llama-3.1-8b-instruct": "llama-3",
    "llama3.2": "llama-3",
    "llama3.2-1b-instruct": "llama-3",
    "llama3.2-3b-instruct": "llama-3",
    "gemma": "gemma",
    "gemma-1.1-2b-instruct": "gemma",
    "gemma-1.1-7b-instruct": "gemma",
    "gemma-2b-instruct": "gemma",
    "gemma-7b-instruct": "gemma",
    "gemma-2-2b-instruct": "gemma",
    "gemma-2-9b-instruct": "gemma",
    "qwen1.5": "qwen",
    "qwen1.5-7b-instruct": "qwen",
    "codeqwen1.5-7b-instruct": "qwen",
    "qwen2": "qwen",
    "qwen2.5": "qwen",
    "qwen2-0.5b-instruct": "qwen",
    "qwen2-1.5b-instruct": "qwen",
    "qwen2-7b-instruct": "qwen",
    "qwen2.5-0.5b-instruct": "qwen",
    "qwen2.5-1.5b-instruct": "qwen",
    "qwen2.5-3b-instruct": "qwen",
    "qwen2.5-7b-instruct": "qwen",
    "qwen2.5-coder-1.5b-instruct": "qwen",
    "qwen2.5-coder-7b-instruct": "qwen",
    "qwen2.5-math-1.5b-instruct": "qwen",
    "qwen2.5-math-7b-instruct": "qwen",
    "mistral": "mistral-instruct",
    "mistral-7b-instruct-v0.3": "mistral-instruct",
    "mistral-7b-instruct-v0.2": "mistral-instruct",
}

NEXA_RUN_COMPLETION_TEMPLATE_MAP = {
    "octopus-v2": "Below is the query from the users, please call the correct function and generate the parameters to call the function.\n\nQuery: {input} \n\nResponse:",
    "octopus-v4": "<|system|>You are a router. Below is the query from the users, please call the correct function and generate the parameters to call the function.<|end|><|user|>{input}<|end|><|assistant|>",
}

NEXA_RUN_MODEL_PRECISION_MAP = {
    "sd1-4": "q4_0",
    "sd1-5": "q4_0",
    "sd2-1": "q4_0",
    "sd3": "q4_0",
    "flux": "q4_0",
    "lcm-dreamshaper": "f16",
    "sdxl-turbo": "q8_0",
    "anything-lcm": "f16",
    "hassaku-lcm": "f16",
}

EXIT_COMMANDS = ["/exit", "/quit", "/bye"]
EXIT_REMINDER = f"Please use Ctrl + d or one of {EXIT_COMMANDS} to exit.\n"

NEXA_STOP_WORDS_MAP = {"octopus-v2": ["<nexa_end>"], "octopus-v4": ["<nexa_end>"]}

DEFAULT_TEXT_GEN_PARAMS = {
    "temperature": 0.7,
    "max_new_tokens": 2048,
    "nctx": 2048,
    "top_k": 50,
    "top_p": 1.0,
}

DEFAULT_IMG_GEN_PARAMS = {
    "num_inference_steps": 20,
    "height": 512,
    "width": 512,
    "guidance_scale": 7.5,
    "output_path": "generated_images/image.png",
    "random_seed": 0,
}

DEFAULT_IMG_GEN_PARAMS_LCM = {
    "num_inference_steps": 4,
    "height": 512,
    "width": 512,
    "guidance_scale": 1.0,
    "output_path": "generated_images/image.png",
    "random_seed": 0,
}

DEFAULT_IMG_GEN_PARAMS_TURBO = {
    "num_inference_steps": 5,
    "height": 512,
    "width": 512,
    "guidance_scale": 5.0,
    "output_path": "generated_images/image.png",
    "random_seed": 0,
}

DEFAULT_VOICE_GEN_PARAMS = {
    "output_dir": "transcriptions",
    "beam_size": 5,
    "language": None,
    "task": "transcribe",
    "temperature": 0.0,
    "compute_type": "default",
}

NEXA_OFFICIAL_MODELS_TYPE = {
    "gemma-2b": ModelType.NLP,
    "Llama-2-7b-chat": ModelType.NLP,
    "Llama-2-7b": ModelType.NLP,
    "Meta-Llama-3-8B-Instruct": ModelType.NLP,
    "Meta-Llama-3.1-8B-Instruct": ModelType.NLP,
    "Llama3.2-3B-Instruct": ModelType.NLP,
    "Llama3.2-1B-Instruct": ModelType.NLP,
    "Mistral-7B-Instruct-v0.3": ModelType.NLP,
    "Mistral-7B-Instruct-v0.2": ModelType.NLP,
    "Phi-3-mini-128k-instruct": ModelType.NLP,
    "Phi-3-mini-4k-instruct": ModelType.NLP,
    "Phi-3.5-mini-instruct": ModelType.NLP,
    "CodeQwen1.5-7B-Instruct": ModelType.NLP,
    "Qwen2-0.5B-Instruct": ModelType.NLP,
    "Qwen2-1.5B-Instruct": ModelType.NLP,
    "Qwen2-7B-Instruct": ModelType.NLP,
    "codegemma-2b": ModelType.NLP,
    "gemma-1.1-2b-instruct": ModelType.NLP,
    "gemma-2b-instruct": ModelType.NLP,
    "gemma-2-9b-instruct": ModelType.NLP,
    "gemma-1.1-7b-instruct": ModelType.NLP,
    "gemma-7b-instruct": ModelType.NLP,
    "gemma-7b": ModelType.NLP,
    "Qwen2-1.5B": ModelType.NLP,
    "Qwen2.5-0.5B-Instruct": ModelType.NLP,
    "Qwen2.5-1.5B-Instruct": ModelType.NLP,
    "Qwen2.5-3B-Instruct": ModelType.NLP,
    "Qwen2.5-Coder-1.5B-Instruct": ModelType.NLP,
    "Qwen2.5-Coder-7B-Instruct": ModelType.NLP,
    "Qwen2.5-Math-1.5B-Instruct": ModelType.NLP,
    "Qwen2.5-Math-7B-Instruct": ModelType.NLP,
    "codegemma-7b": ModelType.NLP,
    "TinyLlama-1.1B-Chat-v1.0": ModelType.NLP,
    "CodeLlama-7b-Instruct": ModelType.NLP,
    "gpt2": ModelType.NLP,
    "CodeLlama-7b": ModelType.NLP,
    "CodeLlama-7b-Python": ModelType.NLP,
    "Qwen1.5-7B-Instruct": ModelType.NLP,
    "Qwen1.5-7B": ModelType.NLP,
    "Phi-2": ModelType.NLP,
    "deepseek-coder-1.3b-instruct": ModelType.NLP,
    "deepseek-coder-1.3b-base": ModelType.NLP,
    "deepseek-coder-6.7b-instruct": ModelType.NLP,
    "dolphin-2.8-mistral-7b": ModelType.NLP,
    "gemma-2-2b-instruct": ModelType.NLP,
    "Octopus-v2": ModelType.NLP,
    "Octopus-v4": ModelType.NLP,
    "Octo-planner": ModelType.NLP,
    "deepseek-coder-6.7b-base": ModelType.NLP,
    "Llama2-7b-chat-uncensored": ModelType.NLP,
    "Llama3-8B-Lexi-Uncensored": ModelType.NLP,
    "Llama2-7b-function-calling": ModelType.NLP,
    "OpenELM-1_1B": ModelType.NLP,
    "OpenELM-3B": ModelType.NLP,
    "AMD-Llama-135m": ModelType.NLP,
    "lcm-dreamshaper-v7": ModelType.COMPUTER_VISION,
    "stable-diffusion-v1-5": ModelType.COMPUTER_VISION,
    "stable-diffusion-v1-4": ModelType.COMPUTER_VISION,
    "stable-diffusion-v2-1": ModelType.COMPUTER_VISION,
    "stable-diffusion-3-medium": ModelType.COMPUTER_VISION,
    "sdxl-turbo": ModelType.COMPUTER_VISION,
    "hassaku-hentai-model-v13-LCM": ModelType.COMPUTER_VISION,
    "anything-v30-LCM": ModelType.COMPUTER_VISION,
    "FLUX.1-schnell": ModelType.COMPUTER_VISION,
    "Phi-3-vision-128k-instruct": ModelType.MULTIMODAL,
    "nanoLLaVA": ModelType.MULTIMODAL,
    "llava-v1.6-mistral-7b": ModelType.MULTIMODAL,
    "llava-v1.6-vicuna-7b": ModelType.MULTIMODAL,
    "llava-phi-3-mini": ModelType.MULTIMODAL,
    "llava-llama-3-8b-v1.1": ModelType.MULTIMODAL,
    "faster-whisper-tiny.en": ModelType.AUDIO,
    "faster-whisper-tiny": ModelType.AUDIO,
    "faster-whisper-small.en": ModelType.AUDIO,
    "faster-whisper-small": ModelType.AUDIO,
    "faster-whisper-medium.en": ModelType.AUDIO,
    "faster-whisper-medium": ModelType.AUDIO,
    "faster-whisper-base.en": ModelType.AUDIO,
    "faster-whisper-base": ModelType.AUDIO,
    "faster-whisper-large-v3": ModelType.AUDIO,
    "faster-whisper-large-v3-turbo": ModelType.AUDIO,
    "whisper-tiny.en": ModelType.AUDIO,
    "whisper-tiny": ModelType.AUDIO,
    "whisper-small.en": ModelType.AUDIO,
    "whisper-small": ModelType.AUDIO,
    "whisper-base.en": ModelType.AUDIO,
    "whisper-base": ModelType.AUDIO,
    "mxbai-embed-large-v1": ModelType.TEXT_EMBEDDING,
    "nomic-embed-text-v1.5": ModelType.TEXT_EMBEDDING,
    "all-MiniLM-L6-v2": ModelType.TEXT_EMBEDDING,
    "all-MiniLM-L12-v2": ModelType.TEXT_EMBEDDING,
}

from nexa.gguf.llama.llama_cpp import (
    LLAMA_FTYPE_ALL_F32,
    LLAMA_FTYPE_MOSTLY_F16,
    LLAMA_FTYPE_MOSTLY_Q4_0,
    LLAMA_FTYPE_MOSTLY_Q4_1,
    LLAMA_FTYPE_MOSTLY_Q8_0,
    LLAMA_FTYPE_MOSTLY_Q5_0,
    LLAMA_FTYPE_MOSTLY_Q5_1,
    LLAMA_FTYPE_MOSTLY_Q2_K,
    LLAMA_FTYPE_MOSTLY_Q3_K_S,
    LLAMA_FTYPE_MOSTLY_Q3_K_M,
    LLAMA_FTYPE_MOSTLY_Q3_K_L,
    LLAMA_FTYPE_MOSTLY_Q4_K_S,
    LLAMA_FTYPE_MOSTLY_Q4_K_M,
    LLAMA_FTYPE_MOSTLY_Q5_K_S,
    LLAMA_FTYPE_MOSTLY_Q5_K_M,
    LLAMA_FTYPE_MOSTLY_Q6_K,
    LLAMA_FTYPE_MOSTLY_IQ2_XXS,
    LLAMA_FTYPE_MOSTLY_IQ2_XS,
    LLAMA_FTYPE_MOSTLY_Q2_K_S,
    LLAMA_FTYPE_MOSTLY_IQ3_XS,
    LLAMA_FTYPE_MOSTLY_IQ3_XXS,
    LLAMA_FTYPE_MOSTLY_IQ1_S,
    LLAMA_FTYPE_MOSTLY_IQ4_NL,
    LLAMA_FTYPE_MOSTLY_IQ3_S,
    LLAMA_FTYPE_MOSTLY_IQ3_M,
    LLAMA_FTYPE_MOSTLY_IQ2_S,
    LLAMA_FTYPE_MOSTLY_IQ2_M,
    LLAMA_FTYPE_MOSTLY_IQ4_XS,
    LLAMA_FTYPE_MOSTLY_IQ1_M,
    LLAMA_FTYPE_MOSTLY_BF16,
    LLAMA_FTYPE_MOSTLY_Q4_0_4_4,
    LLAMA_FTYPE_MOSTLY_Q4_0_4_8,
    LLAMA_FTYPE_MOSTLY_Q4_0_8_8,
    LLAMA_FTYPE_MOSTLY_TQ1_0,
    LLAMA_FTYPE_MOSTLY_TQ2_0,
)
from nexa.gguf.llama.llama_cpp import (
    GGML_TYPE_F32,
    GGML_TYPE_F16,
    GGML_TYPE_Q4_0,
    GGML_TYPE_Q4_1,
    GGML_TYPE_Q5_0,
    GGML_TYPE_Q5_1,
    GGML_TYPE_Q8_0,
    GGML_TYPE_Q8_1,
    GGML_TYPE_Q2_K,
    GGML_TYPE_Q3_K,
    GGML_TYPE_Q4_K,
    GGML_TYPE_Q5_K,
    GGML_TYPE_Q6_K,
    GGML_TYPE_Q8_K,
    GGML_TYPE_IQ2_XXS,
    GGML_TYPE_IQ2_XS,
    GGML_TYPE_IQ3_XXS,
    GGML_TYPE_IQ1_S,
    GGML_TYPE_IQ4_NL,
    GGML_TYPE_IQ3_S,
    GGML_TYPE_IQ2_S,
    GGML_TYPE_IQ4_XS,
    GGML_TYPE_I8,
    GGML_TYPE_I16,
    GGML_TYPE_I32,
    GGML_TYPE_I64,
    GGML_TYPE_F64,
    GGML_TYPE_IQ1_M,
    GGML_TYPE_BF16,
    GGML_TYPE_Q4_0_4_4,
    GGML_TYPE_Q4_0_4_8,
    GGML_TYPE_Q4_0_8_8,
    GGML_TYPE_COUNT,
)

LLAMA_QUANTIZATION_TYPES = {
    "q4_0": LLAMA_FTYPE_MOSTLY_Q4_0,
    "q4_1": LLAMA_FTYPE_MOSTLY_Q4_1,
    "q5_0": LLAMA_FTYPE_MOSTLY_Q5_0,
    "q5_1": LLAMA_FTYPE_MOSTLY_Q5_1,
    "q8_0": LLAMA_FTYPE_MOSTLY_Q8_0,
    "q2_k": LLAMA_FTYPE_MOSTLY_Q2_K,
    "q3_k_s": LLAMA_FTYPE_MOSTLY_Q3_K_S,
    "q3_k_m": LLAMA_FTYPE_MOSTLY_Q3_K_M,
    "q3_k_l": LLAMA_FTYPE_MOSTLY_Q3_K_L,
    "q4_k_s": LLAMA_FTYPE_MOSTLY_Q4_K_S,
    "q4_k_m": LLAMA_FTYPE_MOSTLY_Q4_K_M,
    "q5_k_s": LLAMA_FTYPE_MOSTLY_Q5_K_S,
    "q5_k_m": LLAMA_FTYPE_MOSTLY_Q5_K_M,
    "q6_k": LLAMA_FTYPE_MOSTLY_Q6_K,
    "iq2_xxs": LLAMA_FTYPE_MOSTLY_IQ2_XXS,
    "iq2_xs": LLAMA_FTYPE_MOSTLY_IQ2_XS,
    "q2_k_s": LLAMA_FTYPE_MOSTLY_Q2_K_S,
    "iq3_xs": LLAMA_FTYPE_MOSTLY_IQ3_XS,
    "iq3_xxs": LLAMA_FTYPE_MOSTLY_IQ3_XXS,
    "iq1_s": LLAMA_FTYPE_MOSTLY_IQ1_S,
    "iq4_nl": LLAMA_FTYPE_MOSTLY_IQ4_NL,
    "iq3_s": LLAMA_FTYPE_MOSTLY_IQ3_S,
    "iq3_m": LLAMA_FTYPE_MOSTLY_IQ3_M,
    "iq2_s": LLAMA_FTYPE_MOSTLY_IQ2_S,
    "iq2_m": LLAMA_FTYPE_MOSTLY_IQ2_M,
    "iq4_xs": LLAMA_FTYPE_MOSTLY_IQ4_XS,
    "iq1_m": LLAMA_FTYPE_MOSTLY_IQ1_M,
    "f16": LLAMA_FTYPE_MOSTLY_F16,
    "f32": LLAMA_FTYPE_ALL_F32,
    "bf16": LLAMA_FTYPE_MOSTLY_BF16,
    "q4_0_4_4": LLAMA_FTYPE_MOSTLY_Q4_0_4_4,
    "q4_0_4_8": LLAMA_FTYPE_MOSTLY_Q4_0_4_8,
    "q4_0_8_8": LLAMA_FTYPE_MOSTLY_Q4_0_8_8,
    "tq1_0": LLAMA_FTYPE_MOSTLY_TQ1_0,
    "tq2_0": LLAMA_FTYPE_MOSTLY_TQ2_0,
}

GGML_TYPES = {
    "f32": GGML_TYPE_F32,
    "f16": GGML_TYPE_F16,
    "q4_0": GGML_TYPE_Q4_0,
    "q4_1": GGML_TYPE_Q4_1,
    "q5_0": GGML_TYPE_Q5_0,
    "q5_1": GGML_TYPE_Q5_1,
    "q8_0": GGML_TYPE_Q8_0,
    "q8_1": GGML_TYPE_Q8_1,
    "q2_k": GGML_TYPE_Q2_K,
    "q3_k": GGML_TYPE_Q3_K,
    "q4_k": GGML_TYPE_Q4_K,
    "q5_k": GGML_TYPE_Q5_K,
    "q6_k": GGML_TYPE_Q6_K,
    "q8_k": GGML_TYPE_Q8_K,
    "iq2_xxs": GGML_TYPE_IQ2_XXS,
    "iq2_xs": GGML_TYPE_IQ2_XS,
    "iq3_xxs": GGML_TYPE_IQ3_XXS,
    "iq1_s": GGML_TYPE_IQ1_S,
    "iq4_nl": GGML_TYPE_IQ4_NL,
    "iq3_s": GGML_TYPE_IQ3_S,
    "iq2_s": GGML_TYPE_IQ2_S,
    "iq4_xs": GGML_TYPE_IQ4_XS,
    "i8": GGML_TYPE_I8,
    "i16": GGML_TYPE_I16,
    "i32": GGML_TYPE_I32,
    "i64": GGML_TYPE_I64,
    "f64": GGML_TYPE_F64,
    "iq1_m": GGML_TYPE_IQ1_M,
    "bf16": GGML_TYPE_BF16,
    "q4_0_4_4": GGML_TYPE_Q4_0_4_4,
    "q4_0_4_8": GGML_TYPE_Q4_0_4_8,
    "q4_0_8_8": GGML_TYPE_Q4_0_8_8,
}