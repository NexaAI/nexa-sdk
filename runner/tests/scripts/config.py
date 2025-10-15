import platform

PLUGIN_MAP = {
    'Linux': {
        'x86_64': ['cpu_gpu']
    },
    'Windows': {
        'x86_64': ['cpu_gpu'],
        'arm64': ['cpu_gpu', 'npu']
    },
    'Darwin': {
        'x86_64': ['cpu_gpu'],
        'arm64': ['cpu_gpu', 'metal']
    }
}

TESTCASE_MAP: list[tuple[str, str, list[str]]] = [
    ('cpu_gpu', 'Qwen/Qwen3-1.7B-GGUF', ['multi_round']),  # TODO: add quant support like Qwen3-1.7B-GGUF:Q4_0
    ('cpu_gpu', 'ggml-org/gemma-3-4b-it-GGUF', ['multi_round', 'image_multi_round']),
    ('cpu_gpu', 'ggml-org/Qwen2.5-Omni-3B-GGUF', ['multi_round', 'audio_multi_round']),
    ('cpu_gpu', 'djuna/jina-embeddings-v2-small-en-Q5_K_M-GGUF', ['multi_round']),
]


def get_plugins() -> list[str]:
    system = platform.system()
    machine = platform.machine()
    return PLUGIN_MAP.get(system, {}).get(machine.lower(), [])


def get_testcases(plugins: list[str]) -> list[tuple[str, str, list[str]]]:
    return [tc for tc in TESTCASE_MAP if tc[0] in plugins]
