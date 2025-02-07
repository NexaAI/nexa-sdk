import json
import logging
import os
from pathlib import Path
import socket
import time
import uuid
from typing import List, Optional, Dict, Any, Union, Literal
import base64
import multiprocessing
from PIL import Image
import tempfile
import uvicorn
from fastapi import FastAPI, HTTPException, Request, File, UploadFile, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse, JSONResponse, StreamingResponse
from pydantic import BaseModel, HttpUrl, AnyUrl, Field
import requests
from io import BytesIO
from urllib.parse import urlparse
import asyncio


from nexa.constants import (
    NEXA_MODELS_HUB_OFFICIAL_DIR,
    NEXA_OFFICIAL_MODELS_TYPE,
    NEXA_RUN_CHAT_TEMPLATE_MAP,
    NEXA_RUN_MODEL_MAP,
    NEXA_RUN_MODEL_MAP_TEXT,
    NEXA_RUN_MODEL_MAP_VLM,
    NEXA_RUN_MODEL_MAP_VOICE,
    NEXA_RUN_PROJECTOR_MAP,
    NEXA_RUN_OMNI_VLM_MAP,
    NEXA_RUN_OMNI_VLM_PROJECTOR_MAP,
    NEXA_RUN_MODEL_MAP_AUDIO_LM,
    NEXA_RUN_AUDIO_LM_PROJECTOR_MAP,
    NEXA_RUN_COMPLETION_TEMPLATE_MAP,
    NEXA_RUN_MODEL_PRECISION_MAP,
    NEXA_RUN_MODEL_MAP_FUNCTION_CALLING,
    NEXA_MODEL_LIST_PATH,
    NEXA_OFFICIAL_BUCKET,
)

from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama.llama_chat_format import (
    Llava15ChatHandler,
    Llava16ChatHandler,
    NanoLlavaChatHandler,
)
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import add_model_to_list, default_use_processes, download_file_with_progress, get_model_info, is_model_exists, pull_model
from nexa.gguf.llama.llama import Llama
from faster_whisper import WhisperModel
import numpy as np
import argparse
import soundfile as sf
import librosa
import io

logging.basicConfig(level=logging.INFO)

# HACK: This is moved from nexa.constants to avoid circular imports
NEXA_PROJECTOR_HANDLER_MAP: dict[str, Llava15ChatHandler] = {
    "nanollava": NanoLlavaChatHandler,
    "nanoLLaVA:fp16": NanoLlavaChatHandler,
    "llava-phi3": Llava15ChatHandler,
    "llava-phi-3-mini:q4_0": Llava15ChatHandler,
    "llava-phi-3-mini:fp16": Llava15ChatHandler,
    "llava-llama3": Llava15ChatHandler,
    "llava-llama-3-8b-v1.1:q4_0": Llava15ChatHandler,
    "llava-llama-3-8b-v1.1:fp16": Llava15ChatHandler,
    "llava1.6-mistral": Llava16ChatHandler,
    "llava-v1.6-mistral-7b:q4_0": Llava16ChatHandler,
    "llava-v1.6-mistral-7b:fp16": Llava16ChatHandler,
    "llava1.6-vicuna": Llava16ChatHandler,
    "llava-v1.6-vicuna-7b:q4_0": Llava16ChatHandler,
    "llava-v1.6-vicuna-7b:fp16": Llava16ChatHandler,
}

app = FastAPI(title="Nexa SDK Server")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Allows all origins
    allow_credentials=True,
    allow_methods=["*"],  # Allows all methods
    allow_headers=["*"],  # Allows all headers
)

model = None
whisper_model = None
chat_format = None
completion_template = None
hostname = socket.gethostname()
default_chat_completion_system_prompt = [
    {"role": "system", "content": "You are a helpful assistant"}]
default_function_call_system_prompt = [
    {"role": "system", "content": "A chat between a curious user and an artificial intelligence assistant. The assistant gives helpful, detailed, and polite answers to the user's questions. The assistant calls functions with appropriate input when necessary"}]
model_path = None
whisper_model_path = "faster-whisper-tiny"  # by default, use tiny whisper model
n_ctx = None
is_local_path = False
model_type = None
is_huggingface = False
is_modelscope = False
projector_path = None
SAMPLING_RATE = 16000


# Request Classes
class GenerationRequest(BaseModel):
    prompt: str = "Tell me a story"
    temperature: float = 0.8
    max_new_tokens: int = 128
    top_k: int = 40
    top_p: float = 0.95
    stop_words: Optional[List[str]] = []
    logprobs: Optional[int] = None
    stream: Optional[bool] = False


class TextContent(BaseModel):
    type: Literal["text"] = "text"
    text: str


class ImageUrlContent(BaseModel):
    type: Literal["image_url"] = "image_url"
    image_url: Dict[str, Union[HttpUrl, str, None]] = Field(
        default={"url": None, "path": None},
        description="Either url or path must be provided"
    )


ContentItem = Union[str, TextContent, ImageUrlContent]


class Message(BaseModel):
    role: str
    content: Union[str, List[ContentItem]]


class ImageResponse(BaseModel):
    base64: str
    url: str


class ChatCompletionRequest(BaseModel):
    messages: List[Message] = [
        {"role": "user", "content": "Tell me a story"}]
    max_tokens: Optional[int] = 128
    temperature: Optional[float] = 0.2
    stream: Optional[bool] = False
    stop_words: Optional[List[str]] = []
    logprobs: Optional[bool] = False
    top_logprobs: Optional[int] = 4
    top_k: Optional[int] = 40
    top_p: Optional[float] = 0.95


class VLMChatCompletionRequest(BaseModel):
    messages: List[Message] = [
        {"role": "user", "content": [
            {"type": "text", "text": "What’s in this image?"},
            {"type": "image_url", "image_url": {
                "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
            }}
        ]
        }
    ]
    max_tokens: Optional[int] = 128
    temperature: Optional[float] = 0.2
    stream: Optional[bool] = False
    stop_words: Optional[List[str]] = []
    top_k: Optional[int] = 40
    top_p: Optional[float] = 0.95


class FunctionDefinitionRequestClass(BaseModel):
    type: str = "function"
    function: Dict[str, Any]

    class Config:
        extra = "allow"

# class FunctionCallRequest(BaseModel):
#     messages: List[Message] = [
#         Message(role="user", content="Extract Jason is 25 years old")]
#     tools: List[FunctionDefinitionRequestClass] = [
#         FunctionDefinitionRequestClass(
#             type="function",
#             function={
#                 "name": "UserDetail",
#                 "parameters": {
#                     "type": "object",
#                     "properties": {
#                         "name": {
#                             "type": "string",
#                             "description": "The user's name"
#                         },
#                         "age": {
#                             "type": "integer",
#                             "description": "The user's age"
#                         }
#                     },
#                     "required": ["name", "age"]
#                 }
#             }
#         )
#     ]
#     tool_choice: Optional[str] = "auto"


class TextToImageRequest(BaseModel):
    prompt: str = "a lovely cat holding a sign says 'Nexa Server'"
    negative_prompt: Optional[str] = ""
    cfg_scale: float = 7.0
    width: int = 256
    height: int = 256
    sample_steps: int = 20
    seed: int = 0


class ImageToImageRequest(BaseModel):
    prompt: str = "a lovely cat holding a sign says 'Nexa Server'"
    negative_prompt: Optional[str] = ""
    image_path: Optional[str] = ""
    cfg_scale: float = 7.0
    width: int = 256
    height: int = 256
    sample_steps: int = 20
    seed: int = 0
    

class TextToSpeechRequest(BaseModel):
    text: str = "Hello, this is a text-to-speech interface."
    seed: int = 42
    sampling_rate: int = 24000
    language: Optional[str] = "en"  # Only for 'outetts'


class FunctionCallRequest(BaseModel):
    """
    Represents the request schema for an OpenAI-style function calling API.

    Attributes:
        tools (List[Dict[str, Any]]):
            Defines the available function calls that can be executed.
        messages (List[Dict[str, Any]]):
            A list of messages representing the conversation history.
        model_path (str):
            The path to the model used for function calling.
    """
    tools: List[Dict[str, Any]] = [
        {
            "type": "function",
            "function": {
                "name": "add_integer",
                "description": "Returns the addition of input integers.",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "num1": {"type": "integer", "description": "An integer to add."},
                        "num2": {"type": "integer", "description": "An integer to add."}
                    },
                    "required": ["num1", "num2"],
                    "additionalProperties": False
                },
                "strict": True
            }
        }
    ]
    messages: List[Dict[str, Any]] = [
        {
            "role": "system",
            "content": "You are an AI assistant that generates structured function calling responses. "
            "Identify the correct function from the available tools and return a JSON object "
            "containing the function name and all required parameters. Ensure the parameters "
            "are accurately derived from the user's input and formatted correctly."
        },
        {"role": "user", "content": "Please calculate the sum of 42 and 100."}
    ]


# New request class for embeddings
class EmbeddingRequest(BaseModel):
    input: Union[str, List[str]] = Field(
        ..., description="The input text to get embeddings for. Can be a string or an array of strings.")
    normalize: Optional[bool] = False
    truncate: Optional[bool] = True


class LoadModelRequest(BaseModel):
    model_path: str = "llama3.2"
    model_type: Optional[str] = None
    is_local_path: Optional[bool] = False
    is_huggingface: Optional[bool] = False
    is_modelscope: Optional[bool] = False
    local_projector_path: Optional[str] = None

    model_config = {
        "protected_namespaces": ()
    }


class LoadWhisperModelRequest(BaseModel):
    whisper_model_path: str = "faster-whisper-tiny"


class DownloadModelRequest(BaseModel):
    model_path: str = "llama3.2"

    model_config = {
        "protected_namespaces": ()
    }


class ActionRequest(BaseModel):
    prompt: str = ""


class StreamASRProcessor:
    def __init__(self, asr, task, language):
        self.asr = asr
        self.task = task
        self.language = None if language == "auto" else language
        self.audio_buffer = np.array([], dtype=np.float32)
        self.commited = []
        self.buffer_time_offset = 0.0

    def insert_audio_chunk(self, audio):
        self.audio_buffer = np.append(self.audio_buffer, audio)

    def process_iter(self):
        if len(self.audio_buffer) == 0:
            return (None, None, "")
        res = self.transcribe(self.audio_buffer)
        tsw = self.ts_words(res)
        if len(tsw) == 0:
            return (None, None, "")

        self.commited = tsw
        text = " ".join([w[2] for w in self.commited])
        beg = self.commited[0][0] + self.buffer_time_offset
        end = self.commited[-1][1] + self.buffer_time_offset
        return (beg, end, text)

    def finish(self):
        if len(self.commited) == 0:
            return (None, None, "")
        text = " ".join([w[2] for w in self.commited])
        beg = self.commited[0][0] + self.buffer_time_offset
        end = self.commited[-1][1] + self.buffer_time_offset
        return (beg, end, text)

    def transcribe(self, audio, prompt=""):
        segments, info = self.asr.transcribe(
            audio,
            language=self.language,
            task=self.task,
            beam_size=5,
            word_timestamps=True,
            condition_on_previous_text=True,
            initial_prompt=prompt
        )
        return list(segments)

    def ts_words(self, segments):
        words = []
        for seg in segments:
            if seg.no_speech_prob > 0.9:
                continue
            for w in seg.words:
                words.append((w.start, w.end, w.word))
        return words


class MetricsResult:
    def __init__(self, ttft: float, decoding_speed: float):
        self.ttft = ttft
        self.decoding_speed = decoding_speed

    def to_dict(self):
        return {
            'ttft': round(self.ttft, 2),
            'decoding_speed': round(self.decoding_speed, 2)
        }

    def to_json(self):
        return json.dumps(self.to_dict())


# helper functions
async def load_model():
    global model, chat_format, completion_template, model_path, n_ctx, is_local_path, model_type, is_huggingface, is_modelscope, projector_path
    global use_function_calling

    if is_local_path:
        if model_type == "Multimodal":
            if not projector_path:
                raise ValueError(
                    "Projector path must be provided when using local path for Multimodal models")
            downloaded_path = model_path
            projector_downloaded_path = projector_path
        else:
            downloaded_path = model_path
    elif is_huggingface or is_modelscope:
        # TODO: currently Multimodal models and Audio models are not supported for Hugging Face
        if model_type == "Multimodal" or model_type == "Audio":
            raise ValueError(
                "Multimodal and Audio models are not supported for Hugging Face")
        downloaded_path, _ = pull_model(
            model_path, hf=is_huggingface, ms=is_modelscope)
    else:
        if model_path in NEXA_RUN_MODEL_MAP_VLM or model_path in NEXA_RUN_OMNI_VLM_MAP or model_path in NEXA_RUN_MODEL_MAP_AUDIO_LM:
            if model_path in NEXA_RUN_OMNI_VLM_MAP:
                downloaded_path, model_type = pull_model(
                    NEXA_RUN_OMNI_VLM_MAP[model_path])
                projector_downloaded_path, _ = pull_model(
                    NEXA_RUN_OMNI_VLM_PROJECTOR_MAP[model_path])
            elif model_path in NEXA_RUN_MODEL_MAP_VLM:
                downloaded_path, model_type = pull_model(
                    NEXA_RUN_MODEL_MAP_VLM[model_path])
                projector_downloaded_path, _ = pull_model(
                    NEXA_RUN_PROJECTOR_MAP[model_path])
            elif model_path in NEXA_RUN_MODEL_MAP_AUDIO_LM:
                downloaded_path, model_type = pull_model(
                    NEXA_RUN_MODEL_MAP_AUDIO_LM[model_path])
                projector_downloaded_path, _ = pull_model(
                    NEXA_RUN_AUDIO_LM_PROJECTOR_MAP[model_path])
        else:
            downloaded_path, model_type = pull_model(model_path)

    print(f"model_type: {model_type}")
    if use_function_calling:
        print('Function calling option is enabled')

    if use_function_calling and model_type != "NLP":
        raise ValueError(
            "Function calling is only supported for NLP models. "
            "Please ensure that you are using a compatible NLP model before enabling this feature."
        )

    if model_type == "NLP" or model_type == "Text Embedding":
        if model_type == "NLP" and use_function_calling:
            from nexa.gguf.nexa_inference_text import NexaTextInference
            model = NexaTextInference(
                model_path=model_path, function_calling=True)
        elif model_path in NEXA_RUN_MODEL_MAP_FUNCTION_CALLING:
            chat_format = "chatml-function-calling"
            with suppress_stdout_stderr():
                try:
                    model = Llama(
                        model_path=downloaded_path,
                        verbose=False,
                        chat_format=chat_format,
                        n_gpu_layers=-1 if is_gpu_available() else 0,
                        logits_all=True,
                        n_ctx=n_ctx,
                        embedding=False
                    )
                except Exception as e:
                    logging.error(
                        f"Failed to load model: {e}. Falling back to CPU.", exc_info=True
                    )
                    model = Llama(
                        model_path=downloaded_path,
                        verbose=False,
                        chat_format=chat_format,
                        n_gpu_layers=0,  # hardcode to use CPU,
                        logits_all=True,
                        n_ctx=n_ctx,
                        embedding=False
                    )

                logging.info(f"model loaded as {model}")
        else:
            model_name = model_path.split(":")[0].lower()
            chat_format = NEXA_RUN_CHAT_TEMPLATE_MAP.get(model_name, None)
            completion_template = NEXA_RUN_COMPLETION_TEMPLATE_MAP.get(
                model_name, None)
            with suppress_stdout_stderr():
                try:
                    model = Llama(
                        model_path=downloaded_path,
                        verbose=False,
                        chat_format=chat_format,
                        n_gpu_layers=-1 if is_gpu_available() else 0,
                        logits_all=True,
                        n_ctx=n_ctx,
                        embedding=model_type == "Text Embedding"
                    )
                except Exception as e:
                    logging.error(
                        f"Failed to load model: {e}. Falling back to CPU.", exc_info=True
                    )
                    model = Llama(
                        model_path=downloaded_path,
                        verbose=False,
                        chat_format=chat_format,
                        n_gpu_layers=0,  # hardcode to use CPU
                        logits_all=True,
                        n_ctx=n_ctx,
                        embedding=model_type == "Text Embedding"
                    )
                logging.info(f"model loaded as {model}")
                chat_format = model.metadata.get(
                    "tokenizer.chat_template", None)

            if (
                completion_template is None
                and (
                    chat_format := model.metadata.get("tokenizer.chat_template", None)
                )
                is not None
            ):
                chat_format = chat_format
                logging.debug("Chat format detected")
    elif model_type == "Computer Vision":
        with suppress_stdout_stderr():
            # TODO: add flux support
            from nexa.gguf.sd.stable_diffusion import StableDiffusion
            model = StableDiffusion(
                model_path=downloaded_path,
                wtype=NEXA_RUN_MODEL_PRECISION_MAP.get(
                    model_path, "f32"
                    # Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
                ),
                n_threads=multiprocessing.cpu_count(),
            )
        logging.info(f"model loaded as {model}")
    elif model_type == "TTS":
        # The TTS model requires parameters that are only available upon receiving a user request.
        # Therefore, model initialization is deferred until the text-to-speech API is called.
        model = None
    elif model_type == "Multimodal":
        from nexa.gguf.nexa_inference_vlm_omni import NexaOmniVlmInference
        with suppress_stdout_stderr():
            if 'omni' in model_path.lower():
                try:
                    model = NexaOmniVlmInference(
                        model_path=model_path,
                        device="gpu" if is_gpu_available() else "cpu"
                    )
                except Exception as e:
                    logging.error(
                        f"Failed to load OmniVLM model: {e}. Falling back to CPU.",
                        exc_info=True,
                    )
                    model = NexaOmniVlmInference(
                        model_path=model_path,
                        device="cpu"
                    )
            else:
                from nexa.gguf.nexa_inference_vlm import NexaVLMInference
                try:
                    model = NexaVLMInference(
                        model_path=model_path,
                        device="gpu" if is_gpu_available() else "cpu"
                    )
                except Exception as e:
                    logging.error(
                        f"Failed to load Vlm model: {e}. Falling back to CPU.",
                        exc_info=True,
                    )
                    model = NexaVLMInference(
                        model_path=model_path,
                        device="cpu"
                    )
        logging.info(f"Model loaded as {model}")
    elif model_type == "AudioLM":
        from nexa.gguf.nexa_inference_audio_lm import NexaAudioLMInference
        with suppress_stdout_stderr():
            try:
                model = NexaAudioLMInference(
                    model_path=model_path,
                    device="gpu" if is_gpu_available() else "cpu"
                )
            except Exception as e:
                logging.error(
                    f"Failed to load model: {e}. Falling back to CPU.",
                    exc_info=True,
                )
                model = NexaAudioLMInference(
                    model_path=model_path,
                    device="cpu"
                )
        logging.info(f"model loaded as {model}")
    else:
        raise ValueError(
            f"Model {model_path} not found in Model Hub. If you are using local path, be sure to add --local_path and --model_type flags.")


async def load_whisper_model(custom_whisper_model_path=None):
    global whisper_model, whisper_model_path
    try:
        if custom_whisper_model_path:
            whisper_model_path = custom_whisper_model_path
        downloaded_path, _ = pull_model(whisper_model_path)
        with suppress_stdout_stderr():
            whisper_model = WhisperModel(
                downloaded_path,
                device="cpu",  # only support cpu for now because cuDNN needs to be installed on user's machine
                compute_type="default"
            )
        logging.info(f"whisper model loaded as {whisper_model}")
    except Exception as e:
        logging.error(f"Error loading Whisper model: {e}")
        raise ValueError(f"Failed to load Whisper model: {str(e)}")


def nexa_run_text_generation(
    prompt, temperature, stop_words, max_new_tokens, top_k, top_p, messages=[], logprobs=None, stream=False, is_chat_completion=True, **kwargs
) -> Dict[str, Any]:
    global model, chat_format, completion_template
    if model is None:
        raise ValueError(
            "Model is not loaded. Please check the model path and try again.")

    generated_text = ""
    logprobs_or_none = None

    if is_chat_completion:
        # do not add system prompt if local path or huggingface or modelscope
        if is_local_path or is_huggingface or is_modelscope:
            pass
        else:
            if messages[0]['role'] != 'system':
                messages = default_chat_completion_system_prompt + messages

        params = {
            'messages': messages,
            'temperature': temperature,
            'max_tokens': max_new_tokens,
            'top_k': top_k,
            'top_p': top_p,
            'stream': True,
            'stop': stop_words,
            'logprobs': logprobs
        }

        streamer = model.create_chat_completion(**params)
    else:
        if completion_template:
            formatted_prompt = completion_template.format(input=prompt)
        else:
            formatted_prompt = prompt

        params = {
            'prompt': formatted_prompt,
            'temperature': temperature,
            'max_tokens': max_new_tokens,
            'top_k': top_k,
            'top_p': top_p,
            'stream': True,
            'stop': stop_words,
            'logprobs': logprobs,
        }

        streamer = model.create_completion(**params)

    if stream:
        def stream_with_logprobs():
            for chunk in streamer:
                if is_chat_completion:
                    delta = chunk["choices"][0]["delta"]
                    content = delta.get("content", "")
                else:
                    delta = chunk["choices"][0]["text"]
                    content = delta

                chunk_logprobs = None
                if logprobs and "logprobs" in chunk["choices"][0]:
                    chunk_logprobs = chunk["choices"][0]["logprobs"]

                yield {
                    "content": content,
                    "logprobs": chunk_logprobs
                }

        return stream_with_logprobs()
    else:
        for chunk in streamer:
            if is_chat_completion:
                delta = chunk["choices"][0]["delta"]
                if "content" in delta:
                    generated_text += delta["content"]
            else:
                delta = chunk["choices"][0]["text"]
                generated_text += delta

            if logprobs and "logprobs" in chunk["choices"][0]:
                if logprobs_or_none is None:
                    logprobs_or_none = chunk["choices"][0]["logprobs"]
                else:
                    for key in logprobs_or_none:  # tokens, token_logprobs, top_logprobs, text_offset
                        if key in chunk["choices"][0]["logprobs"]:
                            # accumulate data from each chunk
                            logprobs_or_none[key].extend(
                                chunk["choices"][0]["logprobs"][key])

    result = {
        "result": generated_text,
        "logprobs": logprobs_or_none
    }
    return result


async def nexa_run_image_generation(
    prompt,
    image_path,
    cfg_scale,
    width,
    height,
    sample_steps,
    seed,
    negative_prompt="",
):
    global model
    if model is None:
        raise ValueError(
            "Model is not loaded. Please check the model path and try again.")

    if image_path and image_path.strip():
        image_path = image_path.strip()
        if not os.path.exists(image_path):
            raise ValueError(f"Image file not found: {image_path}")
        image = Image.open(image_path)
        generated_image = model.img_to_img(
            image=image,
            prompt=prompt,
            negative_prompt=negative_prompt,
            cfg_scale=cfg_scale,
            width=width,
            height=height,
            sample_steps=sample_steps,
            seed=seed,
        )
    else:
        generated_image = model.txt_to_img(
            prompt=prompt,
            negative_prompt=negative_prompt,
            cfg_scale=cfg_scale,
            width=width,
            height=height,
            sample_steps=sample_steps,
            seed=seed,
        )
    return generated_image


def base64_encode_image(image_path):
    with open(image_path, "rb") as image_file:
        return base64.b64encode(image_file.read()).decode("utf-8")


def is_base64(s: str) -> bool:
    """Check if a string is base64 encoded."""
    try:
        return base64.b64encode(base64.b64decode(s)).decode() == s
    except Exception:
        return False


def is_url(s: Union[str, AnyUrl]) -> bool:
    """Check if a string or AnyUrl object is a valid URL."""
    if isinstance(s, AnyUrl):
        return True
    try:
        result = urlparse(s)
        return all([result.scheme, result.netloc])
    except ValueError:
        return False


def process_image_input(image_data: Dict[str, Union[HttpUrl, str, None]]) -> str:
    """Process image input from either URL or file path, returning a data URI."""
    url = image_data.get("url")
    path = image_data.get("path")
    if url:
        if isinstance(url, str) and (url.startswith('data:image') or is_base64(url)):
            return url if url.startswith('data:image') else f"data:image/png;base64,{url}"
        return image_url_to_base64(str(url))
    elif path:
        if not os.path.exists(path):
            raise ValueError(f"Image file not found: {path}")
        return image_path_to_base64(path)
    else:
        raise ValueError(
            "Either 'url' or 'path' must be provided in image_url")


def image_url_to_base64(image_url: str) -> str:
    response = requests.get(image_url)
    img = Image.open(BytesIO(response.content))
    buffered = BytesIO()
    img.save(buffered, format="PNG")
    return f"data:image/png;base64,{base64.b64encode(buffered.getvalue()).decode()}"


def voice_url_to_base64(file_path: str) -> str:
    """
    Converts a WAV file to a base64 string.

    :param file_path: Path to the WAV file.
    :return: Base64 encoded string.
    """
    try:
        with open(file_path, "rb") as wav_file:
            base64_string = base64.b64encode(wav_file.read()).decode('utf-8')
        return base64_string
    except Exception as e:
        raise ValueError(f"Error converting file to base64: {e}")


def base64_to_wav(base64_string: str, save_path: str):
    """
    Converts a base64 string back to a WAV file.

    :param base64_string: Base64 encoded string.
    :param save_path: Path where the WAV file will be saved.
    """
    try:
        with open(save_path, "wb") as wav_file:
            wav_file.write(base64.b64decode(base64_string))
    except Exception as e:
        raise ValueError(f"Error converting base64 string to WAV file: {e}")


def image_path_to_base64(file_path):
    if file_path and os.path.exists(file_path):
        with open(file_path, "rb") as img_file:
            base64_data = base64.b64encode(img_file.read()).decode("utf-8")
            return f"data:image/png;base64,{base64_data}"
    return None


def load_audio_from_bytes(audio_bytes: bytes):
    buffer = io.BytesIO(audio_bytes)
    a, sr = sf.read(buffer, dtype='float32')
    if sr != SAMPLING_RATE:
        a = librosa.resample(a, orig_sr=sr, target_sr=SAMPLING_RATE)
    return a


def run_nexa_ai_service(model_path_arg=None, is_local_path_arg=False, model_type_arg=None, huggingface=False, modelscope=False, function_calling=False, projector_local_path_arg=None, **kwargs):
    global model_path, n_ctx, is_local_path, model_type, is_huggingface, is_modelscope, projector_path, use_function_calling
    is_local_path = is_local_path_arg
    is_huggingface = huggingface
    is_modelscope = modelscope
    projector_path = projector_local_path_arg
    use_function_calling = function_calling
    if is_local_path_arg or huggingface or modelscope:
        if not model_path_arg:
            raise ValueError(
                "model_path must be provided when using --local_path or --huggingface or --modelscope")
        if is_local_path_arg and not model_type_arg:
            raise ValueError(
                "--model_type must be provided when using --local_path")
        model_path = os.path.abspath(
            model_path_arg) if is_local_path_arg else model_path_arg
        model_type = model_type_arg
    else:
        model_path = model_path_arg
        model_type = None
    n_ctx = kwargs.get("nctx", 2048)
    host = kwargs.get("host", "localhost")
    port = kwargs.get("port", 8000)
    reload = kwargs.get("reload", False)

    uvicorn.run(app, host=host, port=port, reload=reload)


# Endpoints
@app.on_event("startup")
async def startup_event():
    global model_path
    if model_path:
        await load_model()
    else:
        logging.info(
            "No model path provided. Server started without loading a model.")


@app.get("/", response_class=HTMLResponse, tags=["Root"])
async def read_root(request: Request):
    return HTMLResponse(
        content=f"<h1>Welcome to Nexa AI</h1><p>Hostname: {hostname}</p>"
    )


def _resp_async_generator(streamer, start_time):
    _id = str(uuid.uuid4())
    ttft = 0
    decoding_times = 0
    first_token_time = 0
    for token in streamer:
        ttft = time.perf_counter() - start_time if ttft == 0 else ttft
        first_token_time = time.perf_counter() if first_token_time == 0 else first_token_time
        decoding_times += 1
        chunk = {
            "id": _id,
            "object": "chat.completion.chunk",
            "created": time.time(),
            "choices": [{"delta": {"content": token}}],
        }
        yield f"data: {json.dumps(chunk)}\n\n"

    yield f"metrics: {MetricsResult(ttft=ttft, decoding_speed=decoding_times / (time.perf_counter() - first_token_time)).to_json()}\n\n"
    yield "data: [DONE]\n\n"


# Global variable for download progress tracking
download_progress = {}


def pull_model_with_progress(model_path, progress_key, **kwargs):
    """
    Wrapper for pull_model to track download progress using download_file_with_progress.
    """
    model_path = NEXA_RUN_MODEL_MAP.get(model_path, model_path)

    try:
        # Initialize progress tracking
        download_progress[progress_key] = 0
        current_file_completed = False

        # Extract local download path
        local_download_path = kwargs.get('local_download_path')
        base_download_dir = Path(
            local_download_path) if local_download_path else NEXA_MODELS_HUB_OFFICIAL_DIR
        model_name, model_version = model_path.split(":")
        file_extension = ".zip" if kwargs.get("model_type") in [
            "onnx", "bin"] else ".gguf"
        filename = f"{model_version}{file_extension}"
        file_path = base_download_dir / model_name / filename

        # Record expected file details
        expected_files = [
            {
                "path": file_path,
                "size": int(requests.head(f"{NEXA_OFFICIAL_BUCKET}/{model_name}/{filename}").headers.get("Content-Length", 0))
            }
        ]

        # Progress tracker
        def monitor_progress(file_path, total_size):
            """
            Monitor file size growth to estimate progress.
            """
            nonlocal current_file_completed
            while not current_file_completed:
                # Update downloading progress
                time.sleep(0.5)

        def progress_callback(downloaded_chunks, total_chunks, stage="downloading"):
            """
            Callback to update progress based on downloaded chunks.
            """
            nonlocal current_file_completed
            if stage == "downloading":
                if total_chunks > 0:
                    progress = int((downloaded_chunks / total_chunks) * 100)
                    download_progress[progress_key] = min(progress, 99)
                if downloaded_chunks == total_chunks:
                    current_file_completed = True  # Mark file as completed
            elif stage == "verifying":
                download_progress[progress_key] = 100

        url = f"{NEXA_OFFICIAL_BUCKET}/{model_name}/{filename}"
        response = requests.head(url)
        total_size = int(response.headers.get("Content-Length", 0))

        # Start monitoring progress in a background thread
        from threading import Thread
        progress_thread = Thread(
            target=monitor_progress, args=(file_path, total_size))
        progress_thread.start()

        # Call pull_model to start the download
        result = pull_model(
            model_path=model_path,
            hf=kwargs.get("hf", False),
            ms=kwargs.get("ms", False),
            progress_callback=lambda downloaded, total, stage: progress_callback(
                downloaded, total, stage),
            **kwargs,
        )

        if not result or len(result) != 2:
            raise ValueError("Invalid response from pull_model.")

        final_file_path, run_type = result

        if not final_file_path or not run_type:
            raise ValueError(
                "Failed to download model or invalid response from pull_model.")

        # Extract model type from the returned file path or extension
        model_type = Path(final_file_path).suffix.strip(".") or "undefined"

        download_progress[progress_key] = 100

        return {
            "local_path": str(final_file_path),
            "model_type": model_type,
            "run_type": run_type,
        }
    except Exception as e:
        download_progress[progress_key] = -1  # Mark download as failed
        raise ValueError(f"Error in pull_model_with_progress: {e}")


@app.get("/v1/check_model_type", tags=["Model"])
async def check_model_type(model_path: str):
    """
    Check if the model exists and return its type.
    """
    model_name = NEXA_RUN_MODEL_MAP.get(model_path, model_path)

    if ":" in model_name:
        model_name = model_name.split(":")[0]
    else:
        model_name = model_name

    if model_name in NEXA_RUN_MODEL_MAP or NEXA_RUN_CHAT_TEMPLATE_MAP:

        model_type = NEXA_OFFICIAL_MODELS_TYPE[model_name].value
        return {
            "model_name": model_name,
            "model_type": model_type
        }
    else:
        raise HTTPException(
            status_code=404,
            detail=f"Model '{model_name}' not found in the official model list."
        )


@app.post("/v1/download_model", tags=["Model"])
async def download_model(request: DownloadModelRequest):
    """
    Download a model from the model hub with progress tracking.
    """
    try:
        # Initialize progress tracking
        progress_key = request.model_path
        download_progress[progress_key] = 0

        def perform_download():
            """
            Perform the download process with progress tracking.
            """
            try:
                if request.model_path in NEXA_RUN_MODEL_MAP_VLM:
                    downloaded_path = pull_model_with_progress(
                        NEXA_RUN_MODEL_MAP_VLM[request.model_path], progress_key=progress_key
                    )
                    projector_downloaded_path = pull_model_with_progress(
                        NEXA_RUN_PROJECTOR_MAP[request.model_path], progress_key=progress_key
                    )
                    return {
                        "status": "success",
                        "message": "Successfully downloaded model and projector",
                        "model_path": request.model_path,
                        "model_local_path": downloaded_path["local_path"],
                        "projector_local_path": projector_downloaded_path["local_path"],
                        "model_type": downloaded_path["run_type"]
                    }
                elif request.model_path in NEXA_RUN_OMNI_VLM_MAP:
                    downloaded_path = pull_model_with_progress(
                        NEXA_RUN_OMNI_VLM_MAP[request.model_path], progress_key=progress_key
                    )
                    projector_downloaded_path = pull_model_with_progress(
                        NEXA_RUN_OMNI_VLM_PROJECTOR_MAP[request.model_path], progress_key=progress_key
                    )
                    return {
                        "status": "success",
                        "message": "Successfully downloaded model and projector",
                        "model_path": request.model_path,
                        "model_local_path": downloaded_path["local_path"],
                        "projector_local_path": projector_downloaded_path["local_path"],
                        "model_type": downloaded_path["run_type"]
                    }
                elif request.model_path in NEXA_RUN_MODEL_MAP_AUDIO_LM:
                    downloaded_path = pull_model_with_progress(
                        NEXA_RUN_MODEL_MAP_AUDIO_LM[request.model_path], progress_key=progress_key
                    )
                    projector_downloaded_path = pull_model_with_progress(
                        NEXA_RUN_AUDIO_LM_PROJECTOR_MAP[request.model_path], progress_key=progress_key
                    )
                    return {
                        "status": "success",
                        "message": "Successfully downloaded model and projector",
                        "model_path": request.model_path,
                        "model_local_path": downloaded_path["local_path"],
                        "projector_local_path": projector_downloaded_path["local_path"],
                        "model_type": downloaded_path["run_type"]
                    }
                else:
                    downloaded_path = pull_model_with_progress(
                        request.model_path, progress_key=progress_key
                    )
                    return {
                        "status": "success",
                        "message": "Successfully downloaded model",
                        "model_path": request.model_path,
                        "model_local_path": downloaded_path["local_path"],
                        "model_type": downloaded_path["run_type"]
                    }
            except Exception as e:
                logging.error(f"Error during download: {e}")
                download_progress[progress_key] = -1  # Mark download as failed
                raise

        # Execute the download in a background thread
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(None, perform_download)

        # Return the result of the download
        return result

    except Exception as e:
        # Log error and raise HTTP exception
        logging.error(f"Error downloading model: {e}")
        raise HTTPException(
            status_code=500,
            detail=f"Failed to download model: {str(e)}"
        )


async def progress_generator(model_path: str):
    """
    A generator to stream download progress updates.
    """
    try:
        while True:
            progress = download_progress.get(model_path, -1)
            # Check if the model_path exists in download_progress
            if progress == -1:
                yield f"data: {{\"error\": \"Download failed or invalid model path.\"}}\n\n"
                break

            yield f"data: {{\"model_name\": \"{model_path}\", \"progress\": {progress}}}\n\n"

            if progress == 100:
                # if model_path in download_progress and download_progress[model_path] == 100:
                break

            await asyncio.sleep(1)
    except Exception as e:
        yield f"data: {{\"error\": \"Error streaming progress: {str(e)}\"}}\n\n"


@app.get("/v1/download_progress", tags=["Model"])
async def get_download_progress(model_path: str):
    """
    Stream the download progress for a specific model.
    """
    if model_path not in download_progress:
        raise HTTPException(
            status_code=404, detail="No download progress found for the specified model.")

    # Return a StreamingResponse
    return StreamingResponse(progress_generator(model_path), media_type="text/event-stream")


@app.post("/v1/load_model", tags=["Model"])
async def load_different_model(request: LoadModelRequest):
    """Load a different model while maintaining the global model state"""
    try:
        global model_path, is_local_path, model_type, is_huggingface, is_modelscope, projector_path

        # Update global variables with new configuration
        model_path = request.model_path
        is_local_path = request.is_local_path
        model_type = request.model_type
        is_huggingface = request.is_huggingface
        is_modelscope = request.is_modelscope
        projector_path = request.local_projector_path

        # Load the new model
        await load_model()

        return {
            "status": "succeed",
            "message": f"Successfully loaded model: {model_path}",
            "model_type": model_type
        }

    except Exception as e:
        logging.error(f"Error loading model: {e}")
        raise HTTPException(
            status_code=500,
            detail=f"Failed to load model: {str(e)}"
        )


@app.post("/v1/unload_model", tags=["Model"])
async def unload_different_model(request: LoadModelRequest):
    """Load a different model while maintaining the global model state"""
    try:
        global model
        if model:
            model.close()

        return {
            "status": "succeed",
            "message": f"Successfully unloaded model: {model_path}",
            "model_type": model_type
        }

    except Exception as e:
        logging.error(f"Error unloading model: {e}")

        return JSONResponse(
            status_code=500,
            content={
                "error": "Internal server error",
                "detail": f"Failed to unload model: {str(e)}"
            }
        )


@app.post("/v1/load_whisper_model", tags=["Model"])
async def load_different_whisper_model(request: LoadWhisperModelRequest):
    """Load a different Whisper model while maintaining the global model state"""
    try:
        global whisper_model_path
        whisper_model_path = request.whisper_model_path
        await load_whisper_model(custom_whisper_model_path=whisper_model_path)

        return {
            "status": "success",
            "message": f"Successfully loaded Whisper model: {whisper_model_path}",
            "model_type": "Audio",
        }
    except Exception as e:
        logging.error(f"Error loading Whisper model: {e}")
        raise HTTPException(
            status_code=500,
            detail=f"Failed to load Whisper model: {str(e)}"
        )


@app.get("/v1/list_models", tags=["Model"])
async def list_models():
    """List all models available in the model hub"""
    try:
        if NEXA_MODEL_LIST_PATH.exists():
            with open(NEXA_MODEL_LIST_PATH, "r") as f:
                model_list = json.load(f)
        else:
            model_list = {}
        return JSONResponse(content=model_list)
    except Exception as e:
        logging.error(f"Error listing models: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/completions", tags=["NLP"])
async def generate_text(request: GenerationRequest):
    try:
        if model_type != "NLP":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not an NLP model. Please use an NLP model for text generation."
            )
        generation_kwargs = request.dict()
        if request.stream:
            # Run the generation and stream the response
            start_time = time.perf_counter()
            streamer = nexa_run_text_generation(
                is_chat_completion=False, **generation_kwargs)
            return StreamingResponse(_resp_async_generator(streamer, start_time), media_type="application/x-ndjson")
        else:
            # Generate text synchronously and return the response
            result = nexa_run_text_generation(
                is_chat_completion=False, **generation_kwargs)
            return JSONResponse(content={
                "id": str(uuid.uuid4()),
                "object": "text_completion",
                "created": int(time.time()),
                "model": model_path,
                "choices": [{
                    "text": result["result"],
                    "index": 0,
                    "logprobs": result.get("logprobs"),
                    "finish_reason": "stop"
                }]
            })
    except Exception as e:
        logging.error(f"Error in text generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/chat/completions", tags=["NLP"])
async def text_chat_completions(request: ChatCompletionRequest):
    """Endpoint for text-only chat completions using NLP models"""
    try:
        if model_type != "NLP":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not an NLP model. Please use an NLP model for text chat completion."
            )

        if not request.messages:
            raise HTTPException(
                status_code=400,
                detail="No messages provided in the request."
            )

        if request.stream:
            start_time = time.perf_counter()
            streamer = nexa_run_text_generation(
                None, max_new_tokens=request.max_tokens, is_chat_completion=True, **request.dict())
            return StreamingResponse(_resp_async_generator(streamer, start_time), media_type="application/x-ndjson")

        result = nexa_run_text_generation(
            None, max_new_tokens=request.max_tokens, is_chat_completion=True, **request.dict())
        return {
            "id": str(uuid.uuid4()),
            "object": "chat.completion",
            "created": time.time(),
            "choices": [{
                "message": Message(role="assistant", content=result["result"]),
                "logprobs": result["logprobs"] if "logprobs" in result else None,
            }],
        }

    except HTTPException as e:
        raise e
    except Exception as e:
        logging.error(f"Error in text chat completions: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/vlm/chat/completions", tags=["Multimodal"])
async def multimodal_chat_completions(request: VLMChatCompletionRequest):
    """Endpoint for multimodal chat completions using VLM models"""
    try:
        if model_type != "Multimodal" or 'omni' in model_path.lower():
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not a Multimodal model. Please use a Multimodal model (e.g. nanollava) for VLM."
            )

        processed_messages = []
        for msg in request.messages:
            if isinstance(msg.content, list):
                processed_content = []
                for item in msg.content:
                    if isinstance(item, TextContent):
                        processed_content.append(
                            {"type": "text", "text": item.text})
                    elif isinstance(item, ImageUrlContent):
                        try:
                            image_data_uri = process_image_input(
                                item.image_url)
                            processed_content.append({
                                "type": "image_url",
                                "image_url": {"url": image_data_uri}
                            })
                        except ValueError as e:
                            raise HTTPException(status_code=400, detail=str(e))
                processed_messages.append(
                    {"role": msg.role, "content": processed_content})
            else:
                processed_messages.append(
                    {"role": msg.role, "content": msg.content})

        start_time = time.perf_counter()
        response = model.create_chat_completion(
            messages=processed_messages,
            max_tokens=request.max_tokens,
            temperature=request.temperature,
            top_k=request.top_k,
            top_p=request.top_p,
            stream=request.stream,
            stop=request.stop_words,
        )

        if request.stream:

            return StreamingResponse(_resp_async_generator(response, start_time), media_type="application/x-ndjson")
        return response

    except HTTPException as e:
        raise e
    except Exception as e:
        logging.error(f"Error in multimodal chat completions: {e}")
        raise HTTPException(status_code=500, detail=str(e))


async def _resp_omnivlm_async_generator(model, prompt: str, image_path: str):
    _id = str(uuid.uuid4())
    ttft = 0
    start_time = time.perf_counter()
    first_token_time = 0
    decoding_times = 0
    try:
        if not os.path.exists(image_path):
            raise FileNotFoundError(f"Image file not found: {image_path}")

        for token in model.inference_streaming(prompt, image_path):
            ttft = time.perf_counter() - start_time if ttft == 0 else ttft
            first_token_time = time.perf_counter() if first_token_time == 0 else first_token_time
            decoding_times += 1
            chunk = {
                "id": _id,
                "object": "chat.completion.chunk",
                "created": time.time(),
                "choices": [{
                    "delta": {"content": token},
                    "index": 0,
                    "finish_reason": None
                }]
            }
            yield f"data: {json.dumps(chunk)}\n\n"
        yield f"metrics: {MetricsResult(ttft=ttft, decoding_speed=decoding_times / (time.perf_counter() - first_token_time)).to_json()}\n\n"
        yield "data: [DONE]\n\n"
    except Exception as e:
        logging.error(f"Error in OmniVLM streaming: {e}")
        raise


@app.post("/v1/omnivlm/chat/completions", tags=["Multimodal"])
async def omnivlm_chat_completions(request: VLMChatCompletionRequest):
    """Endpoint for Multimodal chat completions using OmniVLM models"""
    temp_file = None
    image_path = None

    try:
        if model_type != "Multimodal" or 'omni' not in model_path.lower():
            raise HTTPException(
                status_code=400,
                detail="Please use an OmniVLM model for this endpoint."
            )

        prompt = ""
        last_message = request.messages[-1]

        if isinstance(last_message.content, list):
            for item in last_message.content:
                if isinstance(item, TextContent):
                    prompt = item.text
                elif isinstance(item, ImageUrlContent):
                    try:
                        base64_image = process_image_input(item.image_url)
                        base64_data = base64_image.split(',')[1]
                        image_data = base64.b64decode(base64_data)

                        temp_file = tempfile.NamedTemporaryFile(
                            suffix='.jpg', delete=False)
                        temp_file.write(image_data)
                        temp_file.flush()
                        os.fsync(temp_file.fileno())
                        temp_file.close()

                        image_path = temp_file.name

                        if not os.path.exists(image_path):
                            raise ValueError(
                                f"Failed to create temporary file at {image_path}")

                    except Exception as e:
                        if temp_file and os.path.exists(temp_file.name):
                            os.unlink(temp_file.name)
                        raise ValueError(f"Failed to process image: {str(e)}")
                else:
                    raise ValueError(
                        "Either url or path must be provided for image")
        else:
            prompt = last_message.content

        if not image_path:
            raise HTTPException(
                status_code=400,
                detail="Image is required for OmniVLM inference"
            )

        if request.stream:
            async def stream_with_cleanup():
                try:
                    async for chunk in _resp_omnivlm_async_generator(model, prompt, image_path):
                        yield chunk
                finally:
                    if image_path and os.path.exists(image_path):
                        try:
                            os.unlink(image_path)
                        except Exception as e:
                            logging.error(
                                f"Error cleaning up file {image_path}: {e}")

            return StreamingResponse(
                stream_with_cleanup(),
                media_type="text/event-stream"
            )
        else:
            try:
                response = model.inference(prompt, image_path)
                return {
                    "id": str(uuid.uuid4()),
                    "object": "chat.completion",
                    "created": time.time(),
                    "choices": [{
                        "message": {"role": "assistant", "content": response},
                        "index": 0,
                        "finish_reason": "stop"
                    }],
                }
            finally:
                if image_path and os.path.exists(image_path):
                    os.unlink(image_path)

    except Exception as e:
        if image_path and os.path.exists(image_path):
            try:
                os.unlink(image_path)
            except Exception as cleanup_error:
                logging.error(
                    f"Error cleaning up file {image_path}: {cleanup_error}")

        if isinstance(e, HTTPException):
            raise e
        logging.error(f"Error in OmniVLM chat completions: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/function-calling", tags=["NLP"])
async def function_call(request: FunctionCallRequest):
    try:
        if model_type != "NLP":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not an NLP model. Please use an NLP model for function calling."
            )
        messages = default_function_call_system_prompt + [
            {"role": msg.role, "content": msg.content} for msg in request.messages
        ]
        tools = [tool.dict() for tool in request.tools]

        response = model.create_chat_completion(
            messages=messages,
            tools=tools,
            tool_choice=request.tool_choice,
        )

        return response

    except Exception as e:
        logging.error(f"Error in function calling: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/txt2img", tags=["Computer Vision"])
async def txt2img(request: TextToImageRequest):
    try:
        if model_type != "Computer Vision":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not a Computer Vision model. Please use a Computer Vision model for image generation."
            )
        generation_kwargs = request.dict()
        generated_images = await nexa_run_image_generation(**generation_kwargs)

        resp = {"created": time.time(), "data": []}

        for image in generated_images:
            id = int(time.time())
            if not os.path.exists("nexa_server_output"):
                os.makedirs("nexa_server_output")
            image_path = os.path.join(
                "nexa_server_output", f"txt2img_{id}.png")
            image.save(image_path)
            img = ImageResponse(base64=base64_encode_image(
                image_path), url=os.path.abspath(image_path))
            resp["data"].append(img)

        return resp

    except Exception as e:
        logging.error(f"Error in txt2img generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/img2img", tags=["Computer Vision"])
async def img2img(request: ImageToImageRequest):
    try:
        if model_type != "Computer Vision":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not a Computer Vision model. Please use a Computer Vision model for image generation."
            )
        generation_kwargs = request.dict()

        generated_images = await nexa_run_image_generation(**generation_kwargs)
        resp = {"created": time.time(), "data": []}

        for image in generated_images:
            id = int(time.time())
            if not os.path.exists("nexa_server_output"):
                os.makedirs("nexa_server_output")
            image_path = os.path.join(
                "nexa_server_output", f"img2img_{id}.png")
            image.save(image_path)
            img = ImageResponse(base64=base64_encode_image(
                image_path), url=os.path.abspath(image_path))
            resp["data"].append(img)

        return resp

    except Exception as e:
        logging.error(f"Error in img2img generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/txt2speech", tags=["Text-to-Speech"])
async def txt2speech(request: TextToSpeechRequest):
    global model
    try:
        # Initialize a new model if:
        # 1. No model has been initialized.
        # 2. The user provides new model initialization parameters.
        if (
            model is None
            or model.seed != request.seed
            or model.sampling_rate != request.sampling_rate
            or model.language != request.language
        ):
            from nexa.gguf.nexa_inference_tts import NexaTTSInference
            model = NexaTTSInference(
                model_path=model_path,
                tts_engine='bark' if 'bark' in model_path.lower() else 'outetts',
                seed=request.seed,
                sampling_rate=request.sampling_rate,
                language=request.language
            )

        if model_type != "TTS":
            raise HTTPException(
                status_code=400,
                detail="The model loaded is not a Text-to-Speech model. Please use a Text-to-Speech model for this api."
            )

        audio_data = model.audio_generation(request.text)
        output_dir = "nexa_server_output"
        os.makedirs(output_dir, exist_ok=True)
        file_path = model._save_audio(
            audio_data, request.sampling_rate, output_dir)

        resp = {
            "created": time.time(),
            "data": {
                "url": os.path.abspath(file_path),
                "base64": voice_url_to_base64(file_path)
            }
        }

        return resp

    except Exception as e:
        logging.error(f"Error in text-to-speech generation: {e}")
        if isinstance(e, ImportError):
            logging.error(
                "To resolve this issue, please run "
                "`pip install nexaai[tts]` to install necessary dependencies for TTS models."
            )
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/func_calling", tags=["Function Calling"])
async def function_calling(request: FunctionCallRequest):
    try:
        json_response = model.function_calling(
            messages=request.messages, tools=request.tools)

        return {
            "created": time.time(),
            "response": json_response
        }
    except Exception as e:
        logging.error(f"Error in function calling: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/audio/processing", tags=["Audio"])
async def process_audio(
    file: UploadFile = File(...),
    task: str = Query("transcribe",
                      description="Task to perform on the audio. Options are: 'transcribe' or 'translate'.",
                      regex="^(transcribe|translate)$"
                      ),
    beam_size: Optional[int] = Query(5, description="Beam size for decoding."),
    language: Optional[str] = Query(
        None, description="Language code (e.g. 'en', 'fr') for transcription."),
    temperature: Optional[float] = Query(
        0.0, description="Temperature for sampling.")
):
    try:
        if not whisper_model:
            raise HTTPException(
                status_code=400,
                detail="Whisper model is not loaded. Please load a Whisper model first."
            )

        with tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(file.filename)[1]) as temp_audio:
            temp_audio.write(await file.read())
            temp_audio_path = temp_audio.name

        # Set up parameters for Whisper or similar model
        task_params = {
            "beam_size": beam_size,
            "temperature": temperature,
            "vad_filter": True,
            "task": task
        }

        # Only include language parameter if task is "transcribe"
        # For "translate", the language is always defined as "en"
        if task == "transcribe" and language:
            task_params["language"] = language

        segments, _ = whisper_model.transcribe(temp_audio_path, **task_params)
        result_text = "".join(segment.text for segment in segments)
        return JSONResponse(content={"text": result_text})

    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error during {task}: {str(e)}")
    finally:
        if 'temp_audio_path' in locals() and os.path.exists(temp_audio_path):
            os.unlink(temp_audio_path)


@app.post("/v1/audio/processing_stream", tags=["Audio"])
async def processing_stream_audio(
    file: UploadFile = File(...),
    task: str = Query("transcribe",
                      description="Task to perform on the audio. Options are: 'transcribe' or 'translate'.",
                      regex="^(transcribe|translate)$"
                      ),
    language: Optional[str] = Query(
        "auto", description="Language code (e.g., 'en', 'fr')"),
    min_chunk: Optional[float] = Query(
        1.0, description="Minimum chunk duration for streaming"),
):
    try:
        if not whisper_model:
            raise HTTPException(
                status_code=400,
                detail="Whisper model is not loaded. Please load a Whisper model first."
            )

        # Read the entire file into memory
        audio_bytes = await file.read()
        a_full = load_audio_from_bytes(audio_bytes)
        duration = len(a_full) / SAMPLING_RATE

        # Only include language parameter if task is "transcribe"
        # For "translate", the language is always defined as "en"
        if task == "transcribe" and language != "auto":
            used_language = language
        else:
            used_language = None

        warmup_audio = a_full[:SAMPLING_RATE]  # first second
        whisper_model.transcribe(warmup_audio)

        streamer = StreamASRProcessor(whisper_model, task, used_language)

        start = time.time()
        beg = 0.0

        def stream_generator():
            nonlocal beg
            while beg < duration:
                now = time.time() - start
                if now < beg + min_chunk:
                    time.sleep((beg + min_chunk) - now)
                end = time.time() - start
                if end > duration:
                    end = duration

                chunk_samples = int((end - beg)*SAMPLING_RATE)
                chunk_audio = a_full[int(
                    beg*SAMPLING_RATE):int(beg*SAMPLING_RATE)+chunk_samples]
                beg = end

                streamer.insert_audio_chunk(chunk_audio)
                o = streamer.process_iter()
                if o[0] is not None:
                    data = {
                        "emission_time_ms": (time.time()-start)*1000,
                        "segment_start_ms": o[0]*1000,
                        "segment_end_ms": o[1]*1000,
                        "text": o[2]
                    }
                    yield f"data: {json.dumps(data)}\n\n".encode("utf-8")

            # Final flush
            o = streamer.finish()
            if o[0] is not None:
                data = {
                    "emission_time_ms": (time.time()-start)*1000,
                    "segment_start_ms": o[0]*1000,
                    "segment_end_ms": o[1]*1000,
                    "text": o[2],
                    "final": True
                }
                yield f"data: {json.dumps(data)}\n\n".encode("utf-8")

        return StreamingResponse(stream_generator(), media_type="application/x-ndjson")

    except Exception as e:
        logging.error(f"Error in audio processing stream: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/audiolm/chat/completions", tags=["AudioLM"])
async def audio_chat_completions(
    file: UploadFile = File(...),
    prompt: Optional[str] = Query(
        None, description="Prompt for audio chat completions"),
    stream: Optional[bool] = Query(
        False, description="Whether to stream the response"),
):
    temp_file = None
    ttft = 0
    start_time = time.perf_counter()
    decoding_times = 0

    try:
        if model_type != "AudioLM":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not an AudioLM model. Please use an AudioLM model for audio chat completions."
            )

        temp_file = tempfile.NamedTemporaryFile(
            suffix=os.path.splitext(file.filename)[1], delete=False)
        temp_file.write(await file.read())
        temp_file.flush()
        os.fsync(temp_file.fileno())
        audio_path = temp_file.name

        if stream:
            async def stream_with_cleanup():
                nonlocal ttft, decoding_times, start_time
                first_token_time = 0
                try:
                    for token in model.inference_streaming(audio_path, prompt or ""):
                        ttft = time.perf_counter() - start_time if ttft == 0 else ttft
                        first_token_time = time.perf_counter() if first_token_time == 0 else first_token_time
                        decoding_times += 1
                        chunk = {
                            "id": str(uuid.uuid4()),
                            "object": "chat.completion.chunk",
                            "created": time.time(),
                            "choices": [{
                                "delta": {"content": token},
                                "index": 0,
                                "finish_reason": None
                            }]
                        }
                        yield f"data: {json.dumps(chunk)}\n\n"
                    yield f"metrics: {MetricsResult(ttft=ttft, decoding_speed=decoding_times / (time.perf_counter() - first_token_time)).to_json()}\n\n"
                    yield "data: [DONE]\n\n"
                finally:
                    temp_file.close()
                    if os.path.exists(audio_path):
                        os.unlink(audio_path)

            return StreamingResponse(
                stream_with_cleanup(),
                media_type="text/event-stream"
            )
        else:
            try:
                print("audio_path: ", audio_path)
                response = model.inference(audio_path, prompt or "")
                return {
                    "id": str(uuid.uuid4()),
                    "object": "chat.completion",
                    "created": time.time(),
                    "choices": [{
                        "message": {"role": "assistant", "content": response},
                        "index": 0,
                        "finish_reason": "stop"
                    }],
                }
            finally:
                temp_file.close()
                if os.path.exists(audio_path):
                    os.unlink(audio_path)

    except Exception as e:
        if temp_file:
            temp_file.close()
            if os.path.exists(temp_file.name):
                try:
                    os.unlink(temp_file.name)
                except Exception as cleanup_error:
                    logging.error(
                        f"Error cleaning up file {temp_file.name}: {cleanup_error}")

        if isinstance(e, HTTPException):
            raise e
        logging.error(f"Error in audio chat completions: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/embeddings", tags=["Embedding"])
async def create_embedding(request: EmbeddingRequest):
    try:
        if model_type != "Text Embedding":
            raise HTTPException(
                status_code=400,
                detail="The model that is loaded is not a Text Embedding model. Please use a Text Embedding model for embedding generation."
            )
        if isinstance(request.input, list):
            embeddings_results = [model.embed(
                text, normalize=request.normalize, truncate=request.truncate) for text in request.input]
        else:
            embeddings_results = model.embed(
                request.input, normalize=request.normalize, truncate=request.truncate)

        # Prepare the response data
        if isinstance(request.input, list):
            data = [
                {
                    "object": "embedding",
                    "embedding": embedding,
                    "index": i
                } for i, embedding in enumerate(embeddings_results)
            ]
        else:
            data = [
                {
                    "object": "embedding",
                    "embedding": embeddings_results,
                    "index": 0
                }
            ]

        # Calculate token usage
        input_texts = request.input if isinstance(
            request.input, list) else [request.input]
        total_tokens = sum(len(text.split()) for text in input_texts)

        return {
            "object": "list",
            "data": data,
            "model": model_path,
            "usage": {
                "prompt_tokens": total_tokens,
                "total_tokens": total_tokens
            }
        }
    except Exception as e:
        logging.error(f"Error in embedding generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/action", tags=["Actions"])
async def action(request: ActionRequest):
    try:
        # Extract content between <nexa_X> and <nexa_end>
        prompt = request.prompt
        import re

        # Use regex to match <nexa_X> pattern
        match = re.match(r"<nexa_\d+>(.*?)<nexa_end>", prompt)
        if not match:
            raise ValueError(
                "Invalid prompt format. Must be wrapped in <nexa_X> and <nexa_end>")

        # Extract the function call content
        function_content = match.group(1)

        # Parse function name and parameters
        function_name = function_content[:function_content.index("(")]
        params_str = function_content[function_content.index(
            "(")+1:function_content.rindex(")")]

        # Parse parameters into dictionary
        params = {}
        for param in params_str.split(","):
            if "=" in param:
                key, value = param.split("=")
                params[key.strip()] = value.strip().strip("'").strip('"')

        # Handle different function types
        if function_name == "query_plane_ticket":
            # Validate required parameters
            required_params = ["year", "date",
                               "time", "departure", "destination"]
            for param in required_params:
                if param not in params:
                    raise ValueError(f"Missing required parameter: {param}")
            # Corner case handling for date
            if params['date'] == '00-00':
                from datetime import datetime, timedelta
                tomorrow = datetime.now() + timedelta(days=1)
                params['date'] = tomorrow.strftime('%m-%d')

            if params['year'] == '2024':
                params['year'] = datetime.now().year

            if params['departure'] == 'current':
                params['departure'] = 'LAS'
            # Construct the date string in required format
            date_str = f"{params['date']}/{params['year']}"
            from urllib.parse import quote
            params['departure'] = quote(params['departure'])
            params['destination'] = quote(params['destination'])
            date_str = quote(f"{params['date']}/{params['year']}")
            # Build the URL
            url = (f"https://www.expedia.com/Flights-Search?"
                   f"leg1=from:{params['departure']},to:{params['destination']},"
                   f"departure:{date_str}T&"
                   f"passengers=adults:1&trip=oneway&mode=search")

            return {
                "status": "success",
                "function": function_name,
                "parameters": params,
                "url": url
            }
        else:
            # Handle other function types in the future
            return {
                "status": "error",
                "message": f"Unsupported function: {function_name}"
            }

    except Exception as e:
        return {
            "status": "error",
            "message": str(e)
        }

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run the Nexa AI Text Generation Service"
    )
    parser.add_argument(
        "--model_path", type=str, help="Path or identifier for the model in Nexa Model Hub"
    )
    parser.add_argument(
        "--nctx", type=int, default=2048, help="Length of context window"
    )
    parser.add_argument(
        "--host", type=str, default="localhost", help="Host to bind the server to"
    )
    parser.add_argument(
        "--port", type=int, default=8000, help="Port to bind the server to"
    )
    parser.add_argument(
        "--reload",
        action="store_true",
        help="Enable automatic reloading on code changes",
    )
    parser.add_argument(
        "--local_path",
        action="store_true",
        help="Use a local model path instead of pulling from S3",
    )
    parser.add_argument(
        "--model_type",
        type=str,
        choices=["NLP", "Computer Vision", "Audio", "Multimodal"],
        help="Type of the model (required when using --local_path)",
    )
    parser.add_argument(
        "--huggingface",
        action="store_true",
        help="Use a Hugging Face model",
    )
    parser.add_argument(
        "--modelscope",
        action="store_true",
        help="Use a ModelScope model",
    )
    args = parser.parse_args()
    run_nexa_ai_service(
        model_path_arg=args.model_path,
        is_local_path_arg=args.local_path,
        model_type_arg=args.model_type,
        huggingface=args.huggingface,
        modelscope=args.modelscope,
        nctx=args.nctx,
        host=args.host,
        port=args.port,
        reload=args.reload
    )
