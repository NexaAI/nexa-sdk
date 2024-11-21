# Copyright (c) 2024 Nexa AI Inc., Alibaba Group (Qwen team), and HuggingFace Inc.
# All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

""" Qwen2 model configuration"""

from transformers.configuration_utils import PretrainedConfig
from transformers.utils import logging
from typing import Union
from transformers import PretrainedConfig
import os
from transformers.models.auto import CONFIG_MAPPING

logger = logging.get_logger(__name__)


class SigLipVisionConfig(PretrainedConfig):
    model_type = "siglip_vision_model"
    def __init__(
            self,
            hidden_size=1152,
            image_mean=(0.5, 0.5, 0.5),
            intermediate_size=4304,
            num_hidden_layers=27,
            num_attention_heads=16,
            num_channels=3,
            image_size=384,
            patch_size=14,
            hidden_act="gelu_pytorch_tanh",
            layer_norm_eps=1e-6,
            attention_dropout=0.0,
            **kwargs,
    ):
        super().__init__(**kwargs)
        self.hidden_size = hidden_size
        self.intermediate_size = intermediate_size
        self.num_hidden_layers = num_hidden_layers
        self.num_attention_heads = num_attention_heads
        self.num_channels = num_channels
        self.patch_size = patch_size
        self.image_size = image_size
        self.attention_dropout = attention_dropout
        self.layer_norm_eps = layer_norm_eps
        self.hidden_act = hidden_act
        self.image_mean = image_mean

    @classmethod
    def from_pretrained(cls, pretrained_model_name_or_path: Union[str, os.PathLike], **kwargs) -> "PretrainedConfig":
        cls._set_token_in_kwargs(kwargs)

        config_dict, kwargs = cls.get_config_dict(pretrained_model_name_or_path, **kwargs)

        # get the vision config dict if we are loading from SigLipConfig
        if config_dict.get("model_type") == "siglip":
            config_dict = config_dict["vision_config"]

        if "model_type" in config_dict and hasattr(cls, "model_type") and config_dict["model_type"] != cls.model_type:
            logger.warning(
                f"You are using a model of type {config_dict['model_type']} to instantiate a model of type "
                f"{cls.model_type}. This is not supported for all configurations of models and can yield errors."
            )
        return cls.from_dict(config_dict, **kwargs)
        
        
""" Nexa AI model configuration"""
class OminiVLMConfig(PretrainedConfig):
    model_type = "nano-omini-vlm"
    
    model_type = "omini_vlm"
    keys_to_ignore_at_inference = ["past_key_values"]
    
    def __init__(
        self,
        vision_config=None,
        text_config=None,
        hidden_size=4096,
        mm_hidden_size=1152,
        mm_projector_lr=None,
        mm_projector_type="mlp2x_gelu",
        image_token_index=151655,
        initializer_range=0.02,
        **kwargs,
    ):
        self.hidden_size = hidden_size
        self.mm_hidden_size = mm_hidden_size
        self.mm_projector_lr = mm_projector_lr
        self.mm_projector_type = mm_projector_type
        self.image_token_index = image_token_index
        self.initializer_range = initializer_range
        if isinstance(vision_config, dict):
            vision_config = SigLipVisionConfig(**vision_config)
        elif vision_config is None:
            vision_config = SigLipVisionConfig(
                hidden_size=1152,
                image_mean=(0.5, 0.5, 0.5),
                intermediate_size=4304,
                num_hidden_layers=27,
                num_attention_heads=16,
                num_channels=3,
                image_size=384,
                patch_size=14,
                hidden_act="gelu_pytorch_tanh",
                layer_norm_eps=1e-6,
                attention_dropout=0.0,
            )
        self.vision_config = vision_config
        
        if isinstance(text_config, dict):
            text_config["model_type"] = (
                text_config["model_type"] if "model_type" in text_config else "qwen2"
            )
            text_config = CONFIG_MAPPING[text_config["model_type"]](**text_config)
        elif text_config is None:
            text_config = CONFIG_MAPPING["qwen2"]()

        self.text_config = text_config

        super().__init__(**kwargs)
            