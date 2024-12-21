from .version.v1.interface import InterfaceGGUF as _InterfaceGGUF_v1
from .version.v1.interface import GGUFModelConfig as GGUFModelConfig_v1

MODEL_CONFIGS = {
    "0.1": {
        "tokenizer": "OuteAI/OuteTTS-0.1-350M",
        "sizes": ["350M"],
        "links": ["https://huggingface.co/OuteAI/OuteTTS-0.1-350M", "https://huggingface.co/OuteAI/OuteTTS-0.1-350M-GGUF"],
        "languages": ["en"],
        "max_seq_length": 4096
    },
    "0.2": {
        "tokenizer": "OuteAI/OuteTTS-0.2-500M",
        "sizes": ["500M"],
        "links": ["https://huggingface.co/OuteAI/OuteTTS-0.2-500M", "https://huggingface.co/OuteAI/OuteTTS-0.2-500M-GGUF"],
        "languages": ["en", "ja", "ko", "zh"],
        "max_seq_length": 4096
    },
}

def get_model_config(version: str):
    """
    Retrieve the configuration for a given model version.
    """
    if version not in MODEL_CONFIGS:
        raise ValueError(f"Unsupported model version '{version}'. Supported versions are: {list(MODEL_CONFIGS.keys())}")
    return MODEL_CONFIGS[version]

def check_max_length(max_seq_length: int, model_max_seq_length: int):
    if max_seq_length is None:
        raise ValueError("max_seq_length must be specified.")
    if max_seq_length > model_max_seq_length:
        raise ValueError(f"Requested max_seq_length ({max_seq_length}) exceeds the maximum supported length ({model_max_seq_length}).")

def InterfaceGGUF(
        model_version: str,
        cfg: GGUFModelConfig_v1
    ) -> _InterfaceGGUF_v1:
    """
    Creates and returns a GGUF model interface for OuteTTS.

    Parameters
    ----------
    model_version : str
        Version identifier for the model to be loaded
    cfg : GGUFModelConfig_v1
        Configuration object containing parameters

    Returns
    -------
    An instance of interface based on the specified version.
    """

    if not cfg.model_path.lower().endswith('.gguf'):
        raise ValueError(f"Model path must point to a .gguf file, got: '{cfg.model_path}'")

    config = get_model_config(model_version)
    cfg.tokenizer_path = cfg.tokenizer_path or config["tokenizer"]
    languages = config["languages"]
    if cfg.language not in languages:
        raise ValueError(f"Language '{cfg.language}' is not supported by model version '{model_version}'. Supported languages are: {languages}")
    cfg.languages = languages

    check_max_length(cfg.max_seq_length, config["max_seq_length"])
    return _InterfaceGGUF_v1(cfg)
