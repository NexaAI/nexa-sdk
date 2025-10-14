import platform

PLUGIN_MAP = {
    'Linux': {
        'x86_64': ['cpu_gpu']
    },
    'Windows': {
        'x86_64': ['cpu_gpu']
    },
    'macOS': {
        'x86_64': ['cpu_gpu'],
        'arm64': ['cpu_gpu', 'metal']
    }
}

TESTCASE_MAP: list[tuple[str, str, list[str]]] = [
    ('cpu_gpu', 'Qwen/Qwen3-1.7B-GGUF', ['multi_round']),  # TODO: add quant support like Qwen3-1.7B-GGUF:Q4_0
    ('cpu_gpu', 'NexaAI/Qwen3-0.6B-GGUF', ['multi_round']),
]


def get_plugins() -> list[str]:
    system = platform.system()
    machine = platform.machine()
    return PLUGIN_MAP.get(system, {}).get(machine, [])


def get_testcases(plugins: list[str]) -> list[tuple[str, str, list[str]]]:
    return [tc for tc in TESTCASE_MAP if tc[0] in plugins]
