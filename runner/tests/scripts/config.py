import platform

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
TESTCASE_MAP: dict[str, dict[str, list[str]]] = {
    'cpu_gpu': {
        'Qwen/Qwen3-1.7B-GGUF:Q8_0': ['multi_round'],
        # 'ggml-org/gemma-3-4b-it-GGUF': ['multi_round', 'image_multi_round'],
        'ggml-org/Qwen2.5-Omni-3B-GGUF:Q4_K_M': ['multi_round', 'audio_multi_round'],
        'djuna/jina-embeddings-v2-small-en-Q5_K_M-GGUF:Q5_K_M': ['multi_round'],
    },
    'npu': {
        'NexaAI/Qwen3-VL-4B-Instruct-NPU': ['multi_round', 'image_multi_round'],
    },
    'nexaml': {
        'NexaAI/Qwen3-VL-4B-Instruct-GGUF:Q4_0': ['multi_round', 'image_multi_round'],
        # 'NexaAI/Qwen3-VL-4B-Thinking-GGUF:Q4_0': ['multi_round', 'image_multi_round'],
    }
}


def get_plugins() -> list[str]:
    system = platform.system()
    machine = platform.machine()
    return PLUGIN_MAP.get(system, {}).get(machine.lower(), [])


def get_testcases(plugins: list[str]) -> list[tuple[str, str, list[str]]]:
    res: list[tuple[str, str, list[str]]] = []
    for tc in TESTCASE_MAP:
        if tc not in plugins:
            continue

        for model_id, cases in TESTCASE_MAP[tc].items():
            res.append((tc, model_id, cases))

    return res
