import platform
from typing import TypeAlias

from cases import *

PLUGIN_MAP = {
    "Linux": {
        "x86_64": ["cpu_gpu", "nexaml"],
        "arm64": ["cpu_gpu"]
    },
    "Windows": {
        "x86_64": ["cpu_gpu"],
        "arm64": ["cpu_gpu", "npu", "nexaml"]
    },
    "Darwin": {
        "x86_64": ["cpu_gpu"],
        "arm64": ["cpu_gpu", "metal"]
    },
}

# (plugin, model_id, cases)
TESTCASE_MAP: dict[str, dict[str, dict[str, list[type[BaseCase]]]]] = {
    'cpu_gpu': {
        'llm': {
            'Qwen/Qwen3-1.7B-GGUF:Q8_0': [MultiRound],
            # 'ggml-org/gemma-3-4b-it-GGUF:F16': [MultiRound, ImageMultiRound],
            'ggml-org/Qwen2.5-Omni-3B-GGUF:Q4_K_M': [MultiRound, AudioMultiRound],
        },
        "vlm": {},
        'embedder': {
            'djuna/jina-embeddings-v2-small-en-Q5_K_M-GGUF:Q5_K_M': [SingleRound],
        },
        "reranker": {},
        "tts": {},
        "asr": {},
        "diarize": {},
        "cv": {},
        "image_gen": {},
    },
    'npu': {
        'llm': {
            'NexaAI/Granite-4-Micro-NPU': [MultiRound],
            'NexaAI/phi4-mini-npu-turbo': [MultiRound],
            'NexaAI/Qwen3-4B-Instruct-2507-npu': [MultiRound],
            'NexaAI/Qwen3-4B-Thinking-2507-npu': [MultiRound],
            'NexaAI/Llama3.2-3B-NPU-Turbo': [MultiRound],
            'NexaAI/jan-v1-4B-npu': [MultiRound],
            'NexaAI/qwen3-4B-npu': [MultiRound],
            'NexaAI/phi3.5-mini-npu': [MultiRound],
        },
        'vlm': {
            'NexaAI/Qwen3-VL-4B-Instruct-NPU': [MultiRound, ImageMultiRound],
            'NexaAI/OmniNeural-4B': [MultiRound, ImageMultiRound, AudioMultiRound],
            'NexaAI/LFM2-1.2B-npu': [MultiRound],
        },
        'embedder': {
            'NexaAI/embeddinggemma-300m-npu': [SingleRound],
        },
        'asr': {
            'NexaAI/parakeet-tdt-0.6b-v3-npu': [ASR],
        },
        'cv': {
            'NexaAI/convnext-tiny-npu': [ImageRecognition],
            'NexaAI/paddleocr-npu': [OCR],
            'NexaAI/yolov12-npu': [ImageRecognition],
        },
        'reranker': {
            'NexaAI/jina-v2-rerank-npu': [QueryDocument],
        }
    },
    'nexaml': {
        'vlm': {
            'NexaAI/Qwen3-VL-4B-Instruct-GGUF:Q4_0': [MultiRound, ImageMultiRound],
            'NexaAI/Qwen3-VL-4B-Thinking-GGUF:Q4_0': [MultiRound, ImageMultiRound],
        }
    }
}


def get_plugins() -> list[str]:
    system = platform.system()
    machine = platform.machine()
    return PLUGIN_MAP.get(system, {}).get(machine.lower(), [])


case: TypeAlias = tuple[str, str, str, list[type[BaseCase]]]


def get_testcases(plugins: list[str]) -> list[case]:
    res: list[case] = []
    for plugin in TESTCASE_MAP:
        if plugin not in plugins:
            continue
        for modal, model_cases in TESTCASE_MAP[plugin].items():
            for model, cases in model_cases.items():
                res.append((plugin, model, modal, cases))

    return res
