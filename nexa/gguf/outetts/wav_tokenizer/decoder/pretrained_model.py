from typing import Tuple, Any, Union, Dict

import torch
import yaml
from huggingface_hub import hf_hub_download
from torch import nn
from decoder.feature_extractors import FeatureExtractor, EncodecFeatures
from decoder.heads import FourierHead
from decoder.models import Backbone
from decoder.discriminators import MultiPeriodDiscriminator, MultiResolutionDiscriminator


def instantiate_class(args: Union[Any, Tuple[Any, ...]], init: Dict[str, Any]) -> Any:
    """Instantiates a class with the given args and init.

    Args:
        args: Positional arguments required for instantiation.
        init: Dict of the form {"class_path":...,"init_args":...}.

    Returns:
        The instantiated class object.
    """
    kwargs = init.get("init_args", {})
    if not isinstance(args, tuple):
        args = (args,)
    class_module, class_name = init["class_path"].rsplit(".", 1)
    module = __import__(class_module, fromlist=[class_name])
    args_class = getattr(module, class_name)
    return args_class(*args, **kwargs)


class WavTokenizer(nn.Module):
    """
    The Vocos class represents a Fourier-based neural vocoder for audio synthesis.
    This class is primarily designed for inference, with support for loading from pretrained
    model checkpoints. It consists of three main components: a feature extractor,
    a backbone, and a head.
    """

    def __init__(
        self, feature_extractor: FeatureExtractor, backbone: Backbone, head: FourierHead,
            multiperioddisc: MultiPeriodDiscriminator, multiresddisc: MultiResolutionDiscriminator,
    ):
        super().__init__()
        self.feature_extractor = feature_extractor
        self.backbone = backbone
        self.head = head

        self.multiperioddisc = multiperioddisc
        self.multiresddisc = multiresddisc

    @classmethod
    def from_hparams0828(cls, config_path: str) -> "Vocos":
        """
        Class method to create a new Vocos model instance from hyperparameters stored in a yaml configuration file.
        """
        with open(config_path, "r") as f:
            config = yaml.safe_load(f)
        feature_extractor = instantiate_class(args=(), init=config['model']['init_args']["feature_extractor"])
        backbone = instantiate_class(args=(), init=config['model']['init_args']["backbone"])
        head = instantiate_class(args=(), init=config['model']['init_args']["head"])
        model = cls(feature_extractor=feature_extractor, backbone=backbone, head=head,
                    multiperioddisc=MultiPeriodDiscriminator(num_embeddings=4),
                    multiresddisc=MultiResolutionDiscriminator(num_embeddings=4))
        return model

    @classmethod
    def from_pretrained0828(self, config_path, model_path):
        """
        Class method to create a new Vocos model instance from a pre-trained model stored in the Hugging Face model hub.
        """
        model = self.from_hparams0828(config_path)
        state_dict_raw = torch.load(model_path, map_location="cpu")['state_dict']
        state_dict = dict()
        for k, v in state_dict_raw.items():
            if k.startswith('backbone.') or k.startswith('head.') or k.startswith('feature_extractor.') \
                    or k.startswith('multiperioddisc.') or k.startswith('multiresddisc.'):
                state_dict[k] = v
        # if isinstance(model.feature_extractor, EncodecFeatures):
        #     encodec_parameters = {
        #         "feature_extractor.encodec." + key: value
        #         for key, value in model.feature_extractor.encodec.state_dict().items()
        #     }
        #     state_dict.update(encodec_parameters)
        model.load_state_dict(state_dict)
        return model

    @classmethod
    def from_hparams0802(cls, config_path: str) -> "Vocos":
        """
        Class method to create a new Vocos model instance from hyperparameters stored in a yaml configuration file.
        """
        with open(config_path, "r") as f:
            config = yaml.safe_load(f)
        feature_extractor = instantiate_class(args=(), init=config['model']['init_args']["feature_extractor"])
        backbone = instantiate_class(args=(), init=config['model']['init_args']["backbone"])
        head = instantiate_class(args=(), init=config['model']['init_args']["head"])
        model = cls(feature_extractor=feature_extractor, backbone=backbone, head=head)
        return model

    @classmethod
    def from_pretrained0802(self, config_path, model_path):
        """
        Class method to create a new Vocos model instance from a pre-trained model stored in the Hugging Face model hub.
        """
        model = self.from_hparams0802(config_path)
        state_dict_raw = torch.load(model_path, map_location="cpu")['state_dict']
        state_dict = dict()
        for k, v in state_dict_raw.items():
            if k.startswith('backbone.') or k.startswith('head.') or k.startswith('feature_extractor.'):
                state_dict[k] = v
        # if isinstance(model.feature_extractor, EncodecFeatures):
        #     encodec_parameters = {
        #         "feature_extractor.encodec." + key: value
        #         for key, value in model.feature_extractor.encodec.state_dict().items()
        #     }
        #     state_dict.update(encodec_parameters)
        model.load_state_dict(state_dict)
        model.eval()
        return model

    @torch.inference_mode()
    def forward(self, audio_input: torch.Tensor, **kwargs: Any) -> torch.Tensor:
        """
        Method to run a copy-synthesis from audio waveform. The feature extractor first processes the audio input,
        which is then passed through the backbone and the head to reconstruct the audio output.

        Args:
            audio_input (Tensor): The input tensor representing the audio waveform of shape (B, T),
                                        where B is the batch size and L is the waveform length.


        Returns:
            Tensor: The output tensor representing the reconstructed audio waveform of shape (B, T).
        """
        features, _, _ = self.feature_extractor(audio_input, **kwargs)  # 0818
        audio_output = self.decode(features, **kwargs)
        return audio_output


    # 0818
    @torch.inference_mode()
    def encode(self, audio_input: torch.Tensor, **kwargs: Any) -> torch.Tensor:
        features, _, _ = self.feature_extractor(audio_input, **kwargs)
        return features


    @torch.inference_mode()
    def decode(self, features_input: torch.Tensor, **kwargs: Any) -> torch.Tensor:
        """
        Method to decode audio waveform from already calculated features. The features input is passed through
        the backbone and the head to reconstruct the audio output.

        Args:
            features_input (Tensor): The input tensor of features of shape (B, C, L), where B is the batch size,
                                     C denotes the feature dimension, and L is the sequence length.

        Returns:
            Tensor: The output tensor representing the reconstructed audio waveform of shape (B, T).
        """
        x = self.backbone(features_input, **kwargs)
        audio_output = self.head(x)
        return audio_output

    @torch.inference_mode()
    def codes_to_features(self, codes: torch.Tensor) -> torch.Tensor:
        """
        Transforms an input sequence of discrete tokens (codes) into feature embeddings using the feature extractor's
        codebook weights.

        Args:
            codes (Tensor): The input tensor. Expected shape is (K, L) or (K, B, L),
                            where K is the number of codebooks, B is the batch size and L is the sequence length.

        Returns:
            Tensor: Features of shape (B, C, L), where B is the batch size, C denotes the feature dimension,
                    and L is the sequence length.
        """
        assert isinstance(
            self.feature_extractor, EncodecFeatures
        ), "Feature extractor should be an instance of EncodecFeatures"

        if codes.dim() == 2:
            codes = codes.unsqueeze(1)

        n_bins = self.feature_extractor.encodec.quantizer.bins
        offsets = torch.arange(0, n_bins * len(codes), n_bins, device=codes.device)
        embeddings_idxs = codes + offsets.view(-1, 1, 1)
        features = torch.nn.functional.embedding(embeddings_idxs, self.feature_extractor.codebook_weights).sum(dim=0)
        features = features.transpose(1, 2)

        return features
