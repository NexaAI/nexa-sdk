from ...wav_tokenizer.audio_codec import AudioCodec
from .prompt_processor import PromptProcessor
from .model import GGUFModel, GenerationConfig
import torch
import torchaudio
from dataclasses import dataclass, field
import logging
import os
import json

# Configure basic logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
# Create logger instance
logger = logging.getLogger(__name__)

try:
    import sounddevice as sd
    ENABLE_PLAYBACK = True
except Exception as e:
    ENABLE_PLAYBACK = False
    logger.error(e)
    logger.warning("Failed to import sounddevice. Audio playback is disabled.")

BASE_DIR = os.path.dirname(__file__)

DEFAULT_SPEAKERS_DIR = os.path.join(BASE_DIR, "default_speakers")

def get_speaker_path(speaker_name):
    return os.path.join(DEFAULT_SPEAKERS_DIR, f"{speaker_name}.json")

_DEFAULT_SPEAKERS = {
    "en": {
        "male_1": get_speaker_path("en_male_1"),
        "female_1": get_speaker_path("en_female_1"),
    },
    "ja": {
        "male_1": get_speaker_path("ja_male_1"),
        "female_1": get_speaker_path("ja_female_1"),
    },
    "ko": {
        "male_1": get_speaker_path("ko_male_1"),
        "female_1": get_speaker_path("ko_female_1"),
    },
    "zh": {
        "male_1": get_speaker_path("zh_male_1"),
        "female_1": get_speaker_path("zh_female_1"),
    }
}

@dataclass
class GGUFModelConfig:
    model_path: str = "OuteAI/OuteTTS-0.2-500M"
    language: str = "en"
    tokenizer_path: str = None
    languages: list = field(default_factory=list)
    verbose: bool = False
    device: str = None
    dtype: torch.dtype = None
    additional_model_config: dict = field(default_factory=dict)
    wavtokenizer_model_path: str = None
    max_seq_length: int = 4096
    n_gpu_layers: int = 0

@dataclass
class ModelOutput:
    audio: torch.Tensor
    sr: int
    enable_playback: bool = ENABLE_PLAYBACK

    def save(self, path: str):
        if self.audio is None:
            logger.warning("Audio is empty, skipping save.")
            return

        torchaudio.save(path, self.audio.cpu(), sample_rate=self.sr, encoding='PCM_S', bits_per_sample=16)

class InterfaceGGUF:
    def __init__(self, config: GGUFModelConfig) -> None:
        self.device = torch.device(
            config.device if config.device is not None
            else "cuda" if torch.cuda.is_available()
            else "cpu"
        )
        self.config = config
        self._device = config.device
        self.languages = config.languages
        self.language = config.language
        self.verbose = config.verbose

        self.audio_codec = AudioCodec(self.device, config.wavtokenizer_model_path)
        self.prompt_processor = PromptProcessor(config.tokenizer_path, self.languages)
        self.model = GGUFModel(
            model_path=config.model_path,
            n_gpu_layers=config.n_gpu_layers,
            max_seq_length=config.max_seq_length,
            additional_model_config=config.additional_model_config
        )

    def prepare_prompt(self, text: str, speaker: dict = None):
        prompt = self.prompt_processor.get_completion_prompt(text, self.language, speaker)
        # Return a list of token IDs for GGUFModel
        return self.prompt_processor.tokenizer.encode(prompt, add_special_tokens=False)

    def get_audio(self, tokens):
        output = self.prompt_processor.extract_audio_from_tokens(tokens)
        # print(f"InterfaceGGUF get_audio: {output}")
        if not output:
            logger.warning("No audio tokens found in the output")
            return None

        return self.audio_codec.decode(
            torch.tensor([[output]], dtype=torch.int64).to(self.audio_codec.device)
        )

    def load_speaker(self, path: str):
        with open(path, "r") as f:
            return json.load(f)

    def load_default_speaker(self, name: str):
        name = name.lower().strip()
        language = self.language.lower().strip()
        if language not in _DEFAULT_SPEAKERS:
            raise ValueError(f"Speaker for language {language} not found")

        speakers = _DEFAULT_SPEAKERS[language]
        if name not in speakers:
            raise ValueError(f"Speaker {name} not found for language {language}")

        return self.load_speaker(speakers[name])

    def check_generation_max_length(self, max_length):
        if max_length is None:
            raise ValueError("max_length must be specified.")
        if max_length > self.config.max_seq_length:
            raise ValueError(
                f"Requested max_length ({max_length}) exceeds the current max_seq_length ({self.config.max_seq_length})."
            )

    def generate(
        self,
        text: str,
        speaker: dict = None,
        temperature: float = 0.1,
        repetition_penalty: float = 1.1,
        max_length = 4096,
        additional_gen_config = {}
    ) -> ModelOutput:
        input_ids = self.prepare_prompt(text, speaker)
        if self.verbose:
            logger.info(f"Input tokens: {len(input_ids)}")
            logger.info("Generating audio...")

        self.check_generation_max_length(max_length)

        output = self.model.generate(
            input_ids=input_ids,
            config=GenerationConfig(
                temperature=temperature,
                max_length=max_length,
                repetition_penalty=repetition_penalty,
                additional_gen_config=additional_gen_config,
            )
        )

        audio = self.get_audio(output)
        if self.verbose:
            logger.info("Audio generation completed")

        return ModelOutput(audio, self.audio_codec.sr)

    def _create_audio_chunk(self, tokens: list[int], idx: int):
        audio = self.get_audio(tokens)
        size = audio.size()
        audio = audio[:, idx:]
        return ModelOutput(audio, self.audio_codec.sr), size[-1]