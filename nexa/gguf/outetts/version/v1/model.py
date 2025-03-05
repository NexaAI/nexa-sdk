from dataclasses import dataclass, field

try:
    from nexa.gguf.llama.llama import Llama
    from nexa.gguf.llama.llama_cpp import llama_vocab_is_eog, llama_model_get_vocab
    _GGUF_AVAILABLE = True
except ImportError:
    _GGUF_AVAILABLE = False

@dataclass
class GenerationConfig:
    temperature: float = 0.1
    repetition_penalty: float = 1.1
    max_length: int = 4096
    additional_gen_config: dict = field(default_factory=lambda: {})

class GGUFModel:
    def __init__(
            self,
            model_path: str,
            n_gpu_layers: int = 0,
            max_seq_length: int = 4096,
            additional_model_config: dict = {}
    ) -> None:

        if not _GGUF_AVAILABLE:
            raise ImportError(
                "llama_cpp python module not found."
                "To use the GGUF model you must install llama cpp python manually."
            )

        additional_model_config["n_ctx"] = max_seq_length
        self.model = Llama(
            model_path=model_path,
            n_gpu_layers=n_gpu_layers,
            verbose=False,
            **additional_model_config
        )

    def generate(self, input_ids: list[int], config: GenerationConfig):
        return self._generate(input_ids, config)

    def _generate(self, input_ids: list[int], config: GenerationConfig) -> list:
        tokens = []
        for token in self.model.generate(
            input_ids,
            temp=config.temperature,
            repeat_penalty=config.repetition_penalty,
            **config.additional_gen_config,
        ):
            tokens.append(token)
            token_value = token[0] if isinstance(token, tuple) else token
            if (llama_vocab_is_eog(llama_model_get_vocab(self.model._model.model), token_value) or 
                len(tokens) >= config.max_length):
                break
        return tokens
