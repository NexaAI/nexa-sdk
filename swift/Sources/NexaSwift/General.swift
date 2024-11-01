import Foundation

let NEXA_RUN_MODEL_MAP_TEXT: [String: String] = [
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
    "openelm": "OpenELM-3B:q4_K_M"
]

let NEXA_RUN_MODEL_MAP_VLM: [String: String] = [
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
    "llava-v1.6-vicuna-7b:fp16": "llava-v1.6-vicuna-7b:model-fp16"
]

let NEXA_RUN_MODEL_MAP_IMAGE : [String: String] = [
    "sd1-4": "stable-diffusion-v1-4:q4_0",
    "sd1-5": "stable-diffusion-v1-5:q4_0",
    "sd2-1": "stable-diffusion-v2-1:q4_0",
    "sd3": "stable-diffusion-3-medium:q4_0",
    "sdxl-turbo": "sdxl-turbo:q8_0",
    "flux": "FLUX.1-schnell:q4_0",
    "lcm-dreamshaper": "lcm-dreamshaper-v7:fp16",
    "anything-lcm": "anything-v30-LCM:fp16",
    "hassaku-lcm": "hassaku-hentai-model-v13-LCM:fp16",
]

let NEXA_RUN_MODEL_MAP_VOICE:[String: String] = [
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
]

let NEXA_RUN_MODEL_MAP_FUNCTION_CALLING:[String: String] = [
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
]
let NEXA_RUN_MODEL_MAP_FLUX:[String: String] = [
    "flux": "FLUX.1-schnell:flux1-schnell-q4_0",
    "FLUX.1-schnell:q4_0": "FLUX.1-schnell:flux1-schnell-q4_0",
    "FLUX.1-schnell:q5_0": "FLUX.1-schnell:flux1-schnell-q5_0",
    "FLUX.1-schnell:q5_1": "FLUX.1-schnell:flux1-schnell-q5_1",
    "FLUX.1-schnell:q8_0": "FLUX.1-schnell:flux1-schnell-q8_0",
    "FLUX.1-schnell:fp16": "FLUX.1-schnell:flux1-schnell-fp16",
]

let NEXA_RUN_MODEL_MAP_TEXT_EMBEDDING:[String: String] = [
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
]

let NEXA_RUN_MODEL_MAP: [String: String] = {
    var combinedMap = NEXA_RUN_MODEL_MAP_TEXT
    combinedMap.merge(NEXA_RUN_MODEL_MAP_IMAGE) { (_, new) in new }
    combinedMap.merge(NEXA_RUN_MODEL_MAP_VLM) { (_, new) in new }
    combinedMap.merge(NEXA_RUN_MODEL_MAP_VOICE) { (_, new) in new }
    combinedMap.merge(NEXA_RUN_MODEL_MAP_FUNCTION_CALLING) { (_, new) in new }
    combinedMap.merge(NEXA_RUN_MODEL_MAP_FLUX) { (_, new) in new }
    combinedMap.merge(NEXA_RUN_MODEL_MAP_TEXT_EMBEDDING) { (_, new) in new }
    // Merge other maps as needed
    return combinedMap
}()
