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

from typing import List, Union

try:
    from typing import Unpack
except ImportError:
    from typing_extensions import Unpack

from transformers.feature_extraction_utils import BatchFeature
from transformers.image_utils import ImageInput, VideoInput
from transformers.processing_utils import (
    ProcessingKwargs,
    ProcessorMixin,
)
from transformers.tokenization_utils_base import PreTokenizedInput, TextInput
from transformers.utils import logging


logger = logging.get_logger(__name__)
NUM_IMAGE_TOKENS = 81

class NanoVLMProcessorKwargs(ProcessingKwargs, total=False):
    _defaults = {
        "text_kwargs": {
            "padding": False,
        },
    }


class NanoVLMProcessor(ProcessorMixin):
    attributes = ["image_processor", "tokenizer"]
    valid_kwargs = ["chat_template"]
    image_processor_class = "SiglipImageProcessor"
    tokenizer_class = ("Qwen2Tokenizer", "Qwen2TokenizerFast")

    def __init__(self, image_processor=None, tokenizer=None, chat_template=None, **kwargs):
        if chat_template is None:
            chat_template = self.default_chat_template
        super().__init__(image_processor, tokenizer, chat_template=chat_template)

    def __call__(
        self,
        images: ImageInput = None,
        text: Union[TextInput, PreTokenizedInput, List[TextInput], List[PreTokenizedInput]] = None,
        **kwargs: Unpack[NanoVLMProcessorKwargs],
    ) -> BatchFeature:
        """
        Main method to prepare for the model one or several sequences(s) and image(s). This method forwards the `text`
        and `kwargs` arguments to Gemma2TokenizerFast's [`~Gemma2TokenizerFast.__call__`] if `text` is not `None` to encode
        the text. To prepare the vision inputs, this method forwards the `vision_infos` and `kwrags` arguments to
        Gemma2VLImageProcessor's [`~Gemma2VLImageProcessor.__call__`] if `vision_infos` is not `None`.

        Args:
            images (`PIL.Image.Image`, `np.ndarray`, `torch.Tensor`, `List[PIL.Image.Image]`, `List[np.ndarray]`, `List[torch.Tensor]`):
                The image or batch of images to be prepared. Each image can be a PIL image, NumPy array or PyTorch
                tensor. Both channels-first and channels-last formats are supported.
            text (`str`, `List[str]`, `List[List[str]]`):
                The sequence or batch of sequences to be encoded. Each sequence can be a string or a list of strings
                (pretokenized string). If the sequences are provided as list of strings (pretokenized), you must set
                `is_split_into_words=True` (to lift the ambiguity with a batch of sequences).
            return_tensors (`str` or [`~utils.TensorType`], *optional*):
                If set, will return tensors of a particular framework. Acceptable values are:
                - `'tf'`: Return TensorFlow `tf.constant` objects.
                - `'pt'`: Return PyTorch `torch.Tensor` objects.
                - `'np'`: Return NumPy `np.ndarray` objects.
                - `'jax'`: Return JAX `jnp.ndarray` objects.

        Returns:
            [`BatchFeature`]: A [`BatchFeature`] with the following fields:

            - **input_ids** -- List of token ids to be fed to a model. Returned when `text` is not `None`.
            - **attention_mask** -- List of indices specifying which tokens should be attended to by the model (when
              `return_attention_mask=True` or if *"attention_mask"* is in `self.model_input_names` and if `text` is not
              `None`).
            - **pixel_values** -- Pixel values to be fed to a model. Returned when `images` is not `None`.
        """
        output_kwargs = self._merge_kwargs(
            NanoVLMProcessorKwargs,
            tokenizer_init_kwargs=self.tokenizer.init_kwargs,
            **kwargs,
        )
        
        # check the number of images is equal to the number of all image_pad tokens
        assert len(images) == sum([t.count("<|image_pad|>") for t in text]), "The number of images must be equal to the number of all image_pad tokens in the text."
        
        if images is not None:
            image_inputs = self.image_processor(images=images, **output_kwargs["images_kwargs"])
        else:
            image_inputs = {}

        if not isinstance(text, list):
            text = [text]

        if image_inputs is not None:
            index = 0
            for i in range(len(text)):
                while "<|image_pad|>" in text[i]:
                    text[i] = text[i].replace(
                        "<|image_pad|>", "<|placeholder|>" * NUM_IMAGE_TOKENS, 1
                    )
                    index += 1
                text[i] = text[i].replace("<|placeholder|>", "<|image_pad|>")

        _ = output_kwargs["text_kwargs"].pop("padding_side", None)
        text_inputs = self.tokenizer(text, **output_kwargs["text_kwargs"])
        
        return BatchFeature(data={**text_inputs, **image_inputs})

    def batch_decode(self, *args, **kwargs):
        """
        This method forwards all its arguments to Gemma2TokenizerFast's [`~PreTrainedTokenizer.batch_decode`]. Please
        refer to the docstring of this method for more information.
        """
        return self.tokenizer.batch_decode(*args, **kwargs)

    def decode(self, *args, **kwargs):
        """
        This method forwards all its arguments to Gemma2TokenizerFast's [`~PreTrainedTokenizer.decode`]. Please refer to
        the docstring of this method for more information.
        """
        return self.tokenizer.decode(*args, **kwargs)

    @property
    def model_input_names(self):
        tokenizer_input_names = self.tokenizer.model_input_names
        image_processor_input_names = self.image_processor.model_input_names
        return list(dict.fromkeys(tokenizer_input_names + image_processor_input_names))
    
    
    @property
    def default_chat_template(self):
        return (
            "{%- if tools %}"
                "{{- '<|im_start|>system\n' }}"
                "{%- if messages[0]['role'] == 'system' %}"
                    "{{- messages[0]['content'] }}"
                "{%- else %}"
                    "{{- 'You are Nano-Omni-VLM, created by Nexa AI. You are a helpful assistant.' }}"
                "{%- endif %}"
                "{{- \"\n\n# Tools\n\nYou may call one or more functions to assist with the user query.\n\nYou are provided with function signatures within <tools></tools> XML tags:\n<tools>\" }}"
                "{%- for tool in tools %}"
                    "{{- \"\n\" }}"
                    "{{- tool | tojson }}"
                "{%- endfor %}"
                "{{- \"\n</tools>\n\nFor each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:\n<tool_call>\n{\\\"name\\\": <function-name>, \\\"arguments\\\": <args-json-object>}\n</tool_call><|im_end|>\n\" }}"
            "{%- else %}"
                "{%- if messages[0]['role'] == 'system' %}"
                    "{{- '<|im_start|>system\n' + messages[0]['content'] + '<|im_end|>\n' }}"
                "{%- else %}"
                    "{{- '<|im_start|>system\nYou are Nano-Omni-VLM, created by Nexa AI. You are a helpful assistant.<|im_end|>\n' }}"
                "{%- endif %}"
            "{%- endif %}"
            "{%- for message in messages %}"
                "{%- if (message.role == \"user\") or (message.role == \"system\" and not loop.first) or (message.role == \"assistant\" and not message.tool_calls) %}"
                    "{{- '<|im_start|>' + message.role + '\n' + message.content + '<|im_end|>' + '\n' }}"
                "{%- elif message.role == \"assistant\" %}"
                    "{{- '<|im_start|>' + message.role }}"
                    "{%- if message.content %}"
                        "{{- '\n' + message.content }}"
                    "{%- endif %}"
                    "{%- for tool_call in message.tool_calls %}"
                        "{%- if tool_call.function is defined %}"
                            "{%- set tool_call = tool_call.function %}"
                        "{%- endif %}"
                        "{{- '\n<tool_call>\n{\"name\": \"' }}"
                        "{{- tool_call.name }}"
                        "{{- '\", \"arguments\": ' }}"
                        "{{- tool_call.arguments | tojson }}"
                        "{{- '}\n</tool_call>' }}"
                    "{%- endfor %}"
                    "{{- '<|im_end|>\n' }}"
                "{%- elif message.role == \"tool\" %}"
                    "{%- if (loop.index0 == 0) or (messages[loop.index0 - 1].role != \"tool\") %}"
                        "{{- '<|im_start|>user' }}"
                    "{%- endif %}"
                    "{{- '\n<tool_response>\n' }}"
                    "{{- message.content }}"
                    "{{- '\n</tool_response>' }}"
                    "{%- if loop.last or (messages[loop.index0 + 1].role != \"tool\") %}"
                        "{{- '<|im_end|>\n' }}"
                    "{%- endif %}"
                "{%- endif %}"
            "{%- endfor %}"
            "{%- if add_generation_prompt %}"
                "{{- '<|im_start|>assistant\n' }}"
            "{%- endif %}"
        )