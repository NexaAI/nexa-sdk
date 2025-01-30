from typing import List

import torch
import torchaudio
from torch import nn
import math
from decoder.modules import safe_log
from encoder.modules import SEANetEncoder, SEANetDecoder
from encoder import EncodecModel
from encoder.quantization import ResidualVectorQuantizer


class FeatureExtractor(nn.Module):
    """Base class for feature extractors."""

    def forward(self, audio: torch.Tensor, **kwargs) -> torch.Tensor:
        """
        Extract features from the given audio.

        Args:
            audio (Tensor): Input audio waveform.

        Returns:
            Tensor: Extracted features of shape (B, C, L), where B is the batch size,
                    C denotes output features, and L is the sequence length.
        """
        raise NotImplementedError("Subclasses must implement the forward method.")


class MelSpectrogramFeatures(FeatureExtractor):
    def __init__(self, sample_rate=24000, n_fft=1024, hop_length=256, n_mels=100, padding="center"):
        super().__init__()
        if padding not in ["center", "same"]:
            raise ValueError("Padding must be 'center' or 'same'.")
        self.padding = padding
        self.mel_spec = torchaudio.transforms.MelSpectrogram(
            sample_rate=sample_rate,
            n_fft=n_fft,
            hop_length=hop_length,
            n_mels=n_mels,
            center=padding == "center",
            power=1,
        )

    def forward(self, audio, **kwargs):
        if self.padding == "same":
            pad = self.mel_spec.win_length - self.mel_spec.hop_length
            audio = torch.nn.functional.pad(audio, (pad // 2, pad // 2), mode="reflect")
        mel = self.mel_spec(audio)
        features = safe_log(mel)
        return features


class EncodecFeatures(FeatureExtractor):
    def __init__(
        self,
        encodec_model: str = "encodec_24khz",
        bandwidths: List[float] = [1.5, 3.0, 6.0, 12.0],
        train_codebooks: bool = False,
        num_quantizers: int = 1, 
        dowmsamples: List[int] = [6, 5, 5, 4],
        vq_bins: int = 16384,
        vq_kmeans: int = 800,
    ):
        super().__init__()

        # breakpoint()
        self.frame_rate = 25  # not use
        # n_q = int(bandwidths[-1]*1000/(math.log2(2048) * self.frame_rate))
        n_q = num_quantizers   # important
        encoder = SEANetEncoder(causal=False, n_residual_layers=1, norm='weight_norm', pad_mode='reflect', lstm=2,
                                dimension=512, channels=1, n_filters=32, ratios=dowmsamples, activation='ELU',
                                kernel_size=7, residual_kernel_size=3, last_kernel_size=7, dilation_base=2,
                                true_skip=False, compress=2)
        decoder = SEANetDecoder(causal=False, n_residual_layers=1, norm='weight_norm', pad_mode='reflect', lstm=2,
                                dimension=512, channels=1, n_filters=32, ratios=[8, 5, 4, 2], activation='ELU',
                                kernel_size=7, residual_kernel_size=3, last_kernel_size=7, dilation_base=2,
                                true_skip=False, compress=2)
        quantizer = ResidualVectorQuantizer(dimension=512, n_q=n_q, bins=vq_bins, kmeans_iters=vq_kmeans,
                                            decay=0.99, kmeans_init=True)

        # breakpoint()
        if encodec_model == "encodec_24khz":
            self.encodec = EncodecModel(encoder=encoder, decoder=decoder, quantizer=quantizer,
                                        target_bandwidths=bandwidths, sample_rate=24000, channels=1)
        else:
            raise ValueError(
                f"Unsupported encodec_model: {encodec_model}. Supported options are 'encodec_24khz'."
            )
        for param in self.encodec.parameters():
            param.requires_grad = True
        # self.num_q = n_q
        # codebook_weights = torch.cat([vq.codebook for vq in self.encodec.quantizer.vq.layers[: self.num_q]], dim=0)
        # self.codebook_weights = torch.nn.Parameter(codebook_weights, requires_grad=train_codebooks)
        self.bandwidths = bandwidths

    # @torch.no_grad()
    # def get_encodec_codes(self, audio):
    #     audio = audio.unsqueeze(1)
    #     emb = self.encodec.encoder(audio)
    #     codes = self.encodec.quantizer.encode(emb, self.encodec.frame_rate, self.encodec.bandwidth)
    #     return codes

    def forward(self, audio: torch.Tensor, bandwidth_id: torch.Tensor):
        if self.training:
            self.encodec.train()

        audio = audio.unsqueeze(1)                  # audio(16,24000)

        # breakpoint()

        emb = self.encodec.encoder(audio)
        q_res = self.encodec.quantizer(emb, self.frame_rate, bandwidth=self.bandwidths[bandwidth_id])
        quantized = q_res.quantized
        codes = q_res.codes
        commit_loss = q_res.penalty                 # codes(8,16,75),features(16,128,75)

        return quantized, codes, commit_loss

        # codes = self.get_encodec_codes(audio)
        # # Instead of summing in the loop, it stores subsequent VQ dictionaries in a single `self.codebook_weights`
        # # with offsets given by the number of bins, and finally summed in a vectorized operation.
        # offsets = torch.arange(
        #     0, self.encodec.quantizer.bins * len(codes), self.encodec.quantizer.bins, device=audio.device
        # )
        # embeddings_idxs = codes + offsets.view(-1, 1, 1)
        # features = torch.nn.functional.embedding(embeddings_idxs, self.codebook_weights).sum(dim=0)
        # return features.transpose(1, 2)

    def infer(self, audio: torch.Tensor, bandwidth_id: torch.Tensor):
        if self.training:
            self.encodec.train()

        audio = audio.unsqueeze(1)                  # audio(16,24000)
        emb = self.encodec.encoder(audio)
        q_res = self.encodec.quantizer.infer(emb, self.frame_rate, bandwidth=self.bandwidths[bandwidth_id])
        quantized = q_res.quantized
        codes = q_res.codes
        commit_loss = q_res.penalty                 # codes(8,16,75),features(16,128,75)

        return quantized, codes, commit_loss
