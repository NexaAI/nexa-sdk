import math
import warnings
from typing import Optional

import numpy as np
import coremltools as ct
from transformers import AutoTokenizer, AutoImageProcessor
from PIL import Image


# ============================ Helper Functions ============================
def get_alibi_slopes(nheads):
    def get_slopes_power_of_2(nheads):
        start = 2 ** (-(2 ** -(math.log2(nheads) - 3)))
        ratio = start
        return [start * ratio**i for i in range(nheads)]

    if math.log2(nheads).is_integer():
        return get_slopes_power_of_2(nheads)
    else:
        closest_power_of_2 = 2 ** math.floor(math.log2(nheads))
        return (
            get_slopes_power_of_2(closest_power_of_2)
            + get_alibi_slopes(2 * closest_power_of_2)[0::2][: nheads - closest_power_of_2]
        )

def create_alibi_linear_biases(alibi_slopes, seq_len):
    context_position = np.arange(seq_len)[:, None]
    memory_position = np.arange(seq_len)[None, :]
    distance = np.abs(memory_position - context_position)
    linear_biases = (distance[None, ...] * alibi_slopes[:, None, None])[None, ...]
    return linear_biases

def create_padding_mask(attention_mask):
    bsz, seq_len = attention_mask.shape
    padding_mask = np.full((bsz, seq_len), -10000.0, dtype=np.float32)
    padding_mask[attention_mask.astype(bool)] = 0.0
    padding_mask = padding_mask[:, None, None, :]
    return padding_mask

def create_rope_embeddings(dim, pt_seq_len, ft_seq_len, theta=10000):
    freqs = 1.0 / (theta ** (np.arange(0, dim, 2)[: (dim // 2)].astype(np.float32) / dim))
    t = np.arange(ft_seq_len, dtype=np.float32) / ft_seq_len * pt_seq_len
    freqs = np.einsum('..., f -> ... f', t, freqs)
    freqs = np.repeat(freqs, 2, axis=-1)
    freqs_x = np.broadcast_to(freqs[:, None, :], (ft_seq_len, ft_seq_len, freqs.shape[-1]))
    freqs_y = np.broadcast_to(freqs[None, :, :], (ft_seq_len, ft_seq_len, freqs.shape[-1]))
    freqs_2d = np.concatenate([freqs_x, freqs_y], axis=-1)
    rope_cos = np.cos(freqs_2d).reshape(-1, freqs_2d.shape[-1])
    rope_sin = np.sin(freqs_2d).reshape(-1, freqs_2d.shape[-1])
    return rope_cos, rope_sin


# ============================ EmbedNeural ============================
class EmbedNeural:
    MODEL_NAME = "jinaai/jina-clip-v1"
    DEFAULT_NUM_ATTENTION_HEADS = 12
    
    def __init__(
        self, 
        text_model_path: Optional[str] = None, 
        image_model_path: Optional[str] = None,
        compute_units: ct.ComputeUnit = ct.ComputeUnit.CPU_AND_NE
    ):
        self.compute_units = compute_units
        
        self.text_encoder = None
        self.image_encoder = None
        self.tokenizer = None
        self.image_processor = None
        
        # storing the input dict to predict() as an instance variable to allow reuse and prevent python GC to conflict with
        # CoreML's async reset from accessing freed memory
        self._last_text_input_dict = None
        self._last_image_input_dict = None
        
        if text_model_path is not None:
            self.load_text_encoder(text_model_path)
        
        if image_model_path is not None:
            self.load_image_encoder(image_model_path)
        
        if text_model_path is None and image_model_path is None:
            warnings.warn(
                "Neither text_model_path nor image_model_path are provided, "
                "inference won't be available until load_text_encoder() or load_image_encoder() is called."
            )

    def __del__(self):
        self.unload_text_encoder()
        self.unload_image_encoder()
    
    def load_text_encoder(self, model_path: str):
        self.text_encoder = ct.models.CompiledMLModel(model_path, compute_units=self.compute_units)
        if self.tokenizer is None:
            self.tokenizer = AutoTokenizer.from_pretrained(self.MODEL_NAME, trust_remote_code=True)
    
    def load_image_encoder(self, model_path: str):
        self.image_encoder = ct.models.CompiledMLModel(model_path, compute_units=self.compute_units)
        if self.image_processor is None:
            self.image_processor = AutoImageProcessor.from_pretrained(self.MODEL_NAME, trust_remote_code=True, use_fast=False)      # the model does not have a fast processor
    
    def unload_text_encoder(self):
        self.text_encoder = None
    
    def unload_image_encoder(self):
        self.image_encoder = None
    
    def _create_alibi_linear_biases(self, seq_len):
        alibi_slopes = np.array(get_alibi_slopes(self.DEFAULT_NUM_ATTENTION_HEADS), dtype=np.float32)
        linear_biases = create_alibi_linear_biases(alibi_slopes, seq_len)
        return linear_biases
    
    def encode_text(self, text: str) -> np.ndarray:
        if self.text_encoder is None:
            raise RuntimeError("Text encoder is not loaded. Call load_text_encoder() first.")
        
        if self.tokenizer is None:
            self.tokenizer = AutoTokenizer.from_pretrained(self.MODEL_NAME, trust_remote_code=True)
        
        inputs = self.tokenizer(text, return_tensors='np', padding='max_length', max_length=256, truncation=True)
        input_ids = np.array(inputs['input_ids'])
        attention_mask = np.array(inputs['attention_mask'])
        linear_biases = self._create_alibi_linear_biases(input_ids.shape[1])
        padding_mask = create_padding_mask(attention_mask)

        # Store as instance variable to keep arrays alive across inferences
        # This prevents CoreML's async reset from accessing freed memory (see __init__ for details)
        self._last_text_input_dict = {
            "input_ids": input_ids.astype(np.int32),
            "attention_mask": attention_mask.astype(np.float16),
            "linear_biases": linear_biases.astype(np.float16),
            "padding_mask": padding_mask.astype(np.float16),
        }
        
        embeddings = self.text_encoder.predict(self._last_text_input_dict)["text_embeddings"]
        return embeddings.flatten()
    
    def encode_image(self, image: Image.Image) -> np.ndarray:
        if self.image_encoder is None:
            raise RuntimeError("Image encoder is not loaded. Call load_image_encoder() first.")
        
        if self.image_processor is None:
            self.image_processor = AutoImageProcessor.from_pretrained(self.MODEL_NAME, trust_remote_code=True)
        
        processed = self.image_processor(images=[image], return_tensors='np')
        # processor returns pytorch tensor, despite passing in 'np'
        pixel_values = np.array(processed['pixel_values'])
        rope_cos, rope_sin = create_rope_embeddings(dim=32, pt_seq_len=14, ft_seq_len=14, theta=10000)
        rope_cos_input = rope_cos[None, None, :, :]
        rope_sin_input = rope_sin[None, None, :, :]

        # Store as instance variable to keep arrays alive across inferences
        # This prevents CoreML's async reset from accessing freed memory (see __init__ for details)
        self._last_image_input_dict = {
            "pixel_values": pixel_values.astype(np.float16),
            "rope_cos": rope_cos_input.astype(np.float16),
            "rope_sin": rope_sin_input.astype(np.float16),
        }
        
        embeddings = self.image_encoder.predict(self._last_image_input_dict)["image_embeddings"]
        return embeddings.flatten()

