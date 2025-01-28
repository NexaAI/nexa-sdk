import torchaudio
import torch
import os
import platform
from .model import WavEncoder, WavDecoder
from huggingface_hub import snapshot_download

class AudioCodec:
    def __init__(self, device: str = None, model_path: str = None):
        self.device = torch.device(device if device is not None else "cuda" if torch.cuda.is_available() else "cpu")
        self.cache_dir = self.get_cache_dir()
        if model_path is None:
            self.ensure_model_exists()
        else:
            if not os.path.isdir(model_path):
                raise ValueError(f"Model path {model_path} is not a directory. Please provide a valid directory path.")
            self.cache_dir = model_path

        self.encoder_path = os.path.join(self.cache_dir , 'encoder')
        if not os.path.isdir(self.encoder_path):
            raise ValueError(f"Encoder directory not found at {self.encoder_path}. The model path must contain an 'encoder' subdirectory.")
        
        self.decoder_path = os.path.join(self.cache_dir, 'decoder')
        if not os.path.isdir(self.decoder_path):
            raise ValueError(f"Decoder directory not found at {self.decoder_path}. The model path must contain a 'decoder' subdirectory.")
        
        self.sr = 24000
        self.bandwidth_id = torch.tensor([0])

        self.load_decoder()
        self.load_encoder()

    def get_cache_dir(self):
        return os.path.join(
            os.getenv('APPDATA') if platform.system() == "Windows" else os.path.join(os.path.expanduser("~"), ".cache"),
            "outeai", "tts", "wavtokenizer_75_token_interface")
    
    def ensure_model_exists(self):
        snapshot_download("OuteAI/wavtokenizer-large-75token-interface", local_dir=self.cache_dir)
    
    def load_encoder(self):
        self.encoder = WavEncoder.from_pretrained(self.encoder_path).to(self.device)
        self.encoder.eval()

    def load_decoder(self):
        self.decoder = WavDecoder.from_pretrained(self.decoder_path).to(self.device)
        self.decoder.eval()

    def convert_audio(self, wav: torch.Tensor, sr: int, target_sr: int, target_channels: int):
        # Implementation from: https://github.com/jishengpeng/WavTokenizer/blob/afdec2512c0778746250f6fc40d4bac7ff82b742/encoder/utils.py#L79
        assert wav.dim() >= 2, "Audio tensor must have at least 2 dimensions"
        assert wav.shape[-2] in [1, 2], "Audio must be mono or stereo."
        *shape, channels, length = wav.shape
        if target_channels == 1:
            wav = wav.mean(-2, keepdim=True)
        elif target_channels == 2:
            wav = wav.expand(*shape, target_channels, length)
        elif channels == 1:
            wav = wav.expand(target_channels, -1)
        else:
            raise RuntimeError(f"Impossible to convert from {channels} to {target_channels}")
        wav = torchaudio.transforms.Resample(sr, target_sr)(wav)
        return wav
    
    def convert_audio_tensor(self, audio: torch.Tensor, sr):
        return self.convert_audio(audio, sr, self.sr, 1)
    
    def load_audio(self, path):
        wav, sr = torchaudio.load(path)
        return self.convert_audio_tensor(wav, sr).to(self.device)

    def encode(self, audio: torch.Tensor):
        _, discrete_code = self.encoder(audio, bandwidth_id=self.bandwidth_id.to(self.device))
        return discrete_code

    def decode(self, codes):
        features = self.decoder.codes_to_features(codes)
        audio_out = self.decoder(features, bandwidth_id=self.bandwidth_id.to(self.device))
        return audio_out

    def save_audio(self, audio: torch.Tensor, path: str):
        torchaudio.save(path, audio.cpu(), sample_rate=self.sr, encoding='PCM_S', bits_per_sample=16)




    

       