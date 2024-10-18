import json
import logging
import os
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
from PIL import Image
import base64
import re
from urllib.parse import urlparse

from nexa.constants import (
    NEXA_RUN_CHAT_TEMPLATE_MAP,
    NEXA_RUN_MODEL_MAP_VLM,
    NEXA_RUN_PROJECTOR_MAP,
    NEXA_RUN_COMPLETION_TEMPLATE_MAP,
    NEXA_RUN_MODEL_PRECISION_MAP,
    NEXA_RUN_MODEL_MAP_FUNCTION_CALLING,
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama.llama_chat_format import (
    Llava15ChatHandler,
    Llava16ChatHandler,
    NanoLlavaChatHandler,
)
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model
from nexa.gguf.llama.llama import Llama
from nexa.gguf.sd.stable_diffusion import StableDiffusion
from faster_whisper import WhisperModel
import argparse

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

app = FastAPI()
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Allows all origins
    allow_credentials=True,
    allow_methods=["*"],  # Allows all methods
    allow_headers=["*"],  # Allows all headers
)

model = None
chat_format = None
completion_template = None
hostname = socket.gethostname()
chat_completion_system_prompt = [{"role": "system", "content": "You are a helpful assistant"}]
function_call_system_prompt = [{"role": "system", "content": "A chat between a curious user and an artificial intelligence assistant. The assistant gives helpful, detailed, and polite answers to the user's questions. The assistant calls functions with appropriate input when necessary"}]
model_path = None
n_ctx = None
is_local_path = False
model_type = None
is_huggingface = False
projector_path = None
# Request Classes
class GenerationRequest(BaseModel):
    prompt: str = "Tell me a story"
    temperature: float = 1.0
    max_new_tokens: int = 128
    top_k: int = 50
    top_p: float = 1.0
    stop_words: Optional[List[str]] = []
    logprobs: Optional[int] = None
    stream: Optional[bool] = False

class TextContent(BaseModel):
    type: Literal["text"] = "text"
    text: str

class ImageUrlContent(BaseModel):
    type: Literal["image_url"] = "image_url"
    image_url: Dict[str, Union[HttpUrl, str]]

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
    temperature: Optional[float] = 0.1
    stream: Optional[bool] = False
    stop_words: Optional[List[str]] = []
    logprobs: Optional[bool] = False
    top_logprobs: Optional[int] = 4
    top_k: Optional[int] = 40
    top_p: Optional[float] = 0.95

class FunctionDefinitionRequestClass(BaseModel):
    type: str = "function"
    function: Dict[str, Any]

    class Config:
        extra = "allow"

class FunctionCallRequest(BaseModel):
    messages: List[Message] = [
        Message(role="user", content="Extract Jason is 25 years old")]
    tools: List[FunctionDefinitionRequestClass] = [
        FunctionDefinitionRequestClass(
            type="function",
            function={
                "name": "UserDetail",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "name": {
                            "type": "string",
                            "description": "The user's name"
                        },
                        "age": {
                            "type": "integer",
                            "description": "The user's age"
                        }
                    },
                    "required": ["name", "age"]
                }
            }
        )
    ]
    tool_choice: Optional[str] = "auto"

class ImageGenerationRequest(BaseModel):
    prompt: str = "A girl, standing in a field of flowers, vivid"
    image_path: Optional[str] = ""
    cfg_scale: float = 7.0
    width: int = 256
    height: int = 256
    sample_steps: int = 20
    seed: int = 0
    negative_prompt: Optional[str] = ""

# New request class for embeddings
class EmbeddingRequest(BaseModel):
    input: Union[str, List[str]] = Field(..., description="The input text to get embeddings for. Can be a string or an array of strings.")
    normalize: Optional[bool] = False
    truncate: Optional[bool] = True

# helper functions
async def load_model():
    global model, chat_format, completion_template, model_path, n_ctx, is_local_path, model_type, is_huggingface, projector_path
    if is_local_path:
        if model_type == "Multimodal":
            if not projector_path:
                raise ValueError("Projector path must be provided when using local path for Multimodal models")
            downloaded_path = model_path
            projector_downloaded_path = projector_path
        else:
            downloaded_path = model_path
    elif is_huggingface:
        # TODO: currently Multimodal models and Audio models are not supported for Hugging Face
        if model_type == "Multimodal" or model_type == "Audio":
            raise ValueError("Multimodal and Audio models are not supported for Hugging Face")
        downloaded_path, _ = pull_model(model_path, hf=True)
    else:
        if model_path in NEXA_RUN_MODEL_MAP_VLM: # for Multimodal models
            downloaded_path, _ = pull_model(NEXA_RUN_MODEL_MAP_VLM[model_path])
            projector_downloaded_path, _ = pull_model(NEXA_RUN_PROJECTOR_MAP[model_path])
            model_type = "Multimodal"
        else:
            downloaded_path, model_type = pull_model(model_path)
            
    print(f"model_type: {model_type}")
    
    if model_type == "NLP" or model_type == "Text Embedding":
        if model_path in NEXA_RUN_MODEL_MAP_FUNCTION_CALLING:
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
            completion_template = NEXA_RUN_COMPLETION_TEMPLATE_MAP.get(model_name, None)
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
                chat_format = model.metadata.get("tokenizer.chat_template", None)
            
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
            model = StableDiffusion(
                model_path=downloaded_path,
                wtype=NEXA_RUN_MODEL_PRECISION_MAP.get(
                    model_path, "f32"
                ),  # Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
                n_threads=multiprocessing.cpu_count(),
            )
        logging.info(f"model loaded as {model}")
    elif model_type == "Multimodal":
        with suppress_stdout_stderr():
            projector_handler = NEXA_PROJECTOR_HANDLER_MAP.get(model_path, Llava15ChatHandler)
            projector = (projector_handler(
                clip_model_path=projector_downloaded_path, verbose=False
            ) if projector_downloaded_path else None)
            
            chat_format = NEXA_RUN_CHAT_TEMPLATE_MAP.get(model_path, None)
            try:
                model = Llama(
                    model_path=downloaded_path,
                    chat_handler=projector,
                    verbose=False,
                    chat_format=chat_format,
                    n_ctx=n_ctx,
                    n_gpu_layers=-1 if is_gpu_available() else 0,
                )
            except Exception as e:
                logging.error(
                    f"Failed to load model: {e}. Falling back to CPU.",
                    exc_info=True,
                )
                model = Llama(
                    model_path=downloaded_path,
                    chat_handler=projector,
                    verbose=False,
                    chat_format=chat_format,
                    n_ctx=n_ctx,
                    n_gpu_layers=0,  # hardcode to use CPU
                )

        logging.info(f"Model loaded as {model}")
    elif model_type == "Audio":
        with suppress_stdout_stderr():
            model = WhisperModel(
                downloaded_path,
                device="cpu", # only support cpu for now because cuDNN needs to be installed on user's machine
                compute_type="default"
            )
        logging.info(f"model loaded as {model}")
    else:
        raise ValueError(f"Model {model_path} not found in Model Hub. If you are using local path, be sure to add --local_path and --model_type flags.")
    
def nexa_run_text_generation(
    prompt, temperature, stop_words, max_new_tokens, top_k, top_p, logprobs=None, stream=False, is_chat_completion=True
) -> Dict[str, Any]:
    global model, chat_format, completion_template
    if model is None:
        raise ValueError("Model is not loaded. Please check the model path and try again.")
    
    generated_text = ""
    logprobs_or_none = None

    if is_chat_completion:
        if is_local_path or is_huggingface: # do not add system prompt if local path or huggingface
            messages = [{"role": "user", "content": prompt}]
        else:
            messages = chat_completion_system_prompt + [{"role": "user", "content": prompt}]

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
                            logprobs_or_none[key].extend(chunk["choices"][0]["logprobs"][key])  # accumulate data from each chunk

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
    negative_prompt = "",
):
    global model
    if model is None:
        raise ValueError("Model is not loaded. Please check the model path and try again.")

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

def process_image_input(image_input: Union[str, AnyUrl]) -> str:
    """Process image input, returning a data URI for both URL and base64 inputs."""
    if isinstance(image_input, AnyUrl) or is_url(image_input):
        return image_url_to_base64(str(image_input))
    elif is_base64(image_input):
        if image_input.startswith('data:image'):
            return image_input
        else:
            return f"data:image/png;base64,{image_input}"
    else:
        raise ValueError("Invalid image input. Must be a URL or base64 encoded image.")

def image_url_to_base64(image_url: str) -> str:
    response = requests.get(image_url)
    img = Image.open(BytesIO(response.content))
    buffered = BytesIO()
    img.save(buffered, format="PNG")
    return f"data:image/png;base64,{base64.b64encode(buffered.getvalue()).decode()}"


def run_nexa_ai_service(model_path_arg=None, is_local_path_arg=False, model_type_arg=None, huggingface=False, projector_local_path_arg=None, **kwargs):
    global model_path, n_ctx, is_local_path, model_type, is_huggingface, projector_path
    is_local_path = is_local_path_arg
    is_huggingface = huggingface
    projector_path = projector_local_path_arg
    if is_local_path_arg or huggingface:
        if not model_path_arg:
            raise ValueError("model_path must be provided when using --local_path or --huggingface")
        if is_local_path_arg and not model_type_arg:
            raise ValueError("--model_type must be provided when using --local_path")
        model_path = os.path.abspath(model_path_arg) if is_local_path_arg else model_path_arg
        model_type = model_type_arg
    else:
        model_path = model_path_arg or "gemma"
        model_type = None
    os.environ["MODEL_PATH"] = model_path
    os.environ["IS_LOCAL_PATH"] = str(is_local_path_arg)
    os.environ["MODEL_TYPE"] = model_type if model_type else ""
    os.environ["HUGGINGFACE"] = str(huggingface)
    os.environ["PROJECTOR_PATH"] = projector_path if projector_path else ""
    n_ctx = kwargs.get("nctx", 2048)
    host = kwargs.get("host", "localhost")
    port = kwargs.get("port", 8000)
    reload = kwargs.get("reload", False)
    uvicorn.run(app, host=host, port=port, reload=reload)

# Endpoints
@app.on_event("startup")
async def startup_event():
    global model_path, is_local_path, model_type, is_huggingface, projector_path
    model_path = os.getenv("MODEL_PATH", "gemma")
    is_local_path = os.getenv("IS_LOCAL_PATH", "False").lower() == "true"
    model_type = os.getenv("MODEL_TYPE", None)
    is_huggingface = os.getenv("HUGGINGFACE", "False").lower() == "true"
    projector_path = os.getenv("PROJECTOR_PATH", None)
    await load_model()


@app.get("/", response_class=HTMLResponse, tags=["Root"])
async def read_root(request: Request):
    return HTMLResponse(
        content=f"<h1>Welcome to Nexa AI</h1><p>Hostname: {hostname}</p>"
    )


def _resp_async_generator(streamer):
    _id = str(uuid.uuid4())
    for token in streamer:
        chunk = {
            "id": _id,
            "object": "chat.completion.chunk",
            "created": time.time(),
            "choices": [{"delta": {"content": token}}],
        }
        yield f"data: {json.dumps(chunk)}\n\n"
    yield "data: [DONE]\n\n"

@app.post("/v1/completions", tags=["NLP"])
async def generate_text(request: GenerationRequest):
    try:
        generation_kwargs = request.dict()

        if request.stream:
            # Run the generation and stream the response
            streamer = nexa_run_text_generation(is_chat_completion=False, **generation_kwargs)
            return StreamingResponse(_resp_async_generator(streamer), media_type="application/x-ndjson")
        else:
            # Generate text synchronously and return the response
            result = nexa_run_text_generation(is_chat_completion=False, **generation_kwargs)
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
async def chat_completions(request: ChatCompletionRequest):
    try:
        is_vlm = any(isinstance(msg.content, list) for msg in request.messages)
        
        if is_vlm:
            if model_type != "Multimodal":
                raise HTTPException(status_code=400, detail="The model that is loaded is not a Multimodal model. Please use a Multimodal model (e.g. llava1.6-vicuna) for VLM.")
            # Process VLM request
            processed_messages = []
            for msg in request.messages:
                if isinstance(msg.content, list):
                    processed_content = []
                    for item in msg.content:
                        if isinstance(item, TextContent):
                            processed_content.append({"type": "text", "text": item.text})
                        elif isinstance(item, ImageUrlContent):
                            try:
                                image_input = item.image_url["url"]
                                image_data_uri = process_image_input(image_input)
                                processed_content.append({"type": "image_url", "image_url": {"url": image_data_uri}})
                            except ValueError as e:
                                raise HTTPException(status_code=400, detail=str(e))
                    processed_messages.append({"role": msg.role, "content": processed_content})
                else:
                    processed_messages.append({"role": msg.role, "content": msg.content})
                    
            response = model.create_chat_completion(
                messages=processed_messages,
                temperature=request.temperature,
                max_tokens=request.max_tokens,
                top_k=request.top_k,
                top_p=request.top_p,
                stream=request.stream,
            )
        else:
            # Process regular chat completion request
            generation_kwargs = GenerationRequest(
                prompt="" if len(request.messages) == 0 else request.messages[-1].content,
                temperature=request.temperature,
                max_new_tokens=request.max_tokens,
                stop_words=request.stop_words,
                logprobs=request.logprobs,
                top_logprobs=request.top_logprobs,
                stream=request.stream,
                top_k=request.top_k,
                top_p=request.top_p
            ).dict()

            if request.stream:
                streamer = nexa_run_text_generation(is_chat_completion=True, **generation_kwargs)
                return StreamingResponse(_resp_async_generator(streamer), media_type="application/x-ndjson")
            else:
                result = nexa_run_text_generation(is_chat_completion=True, **generation_kwargs)
                return {
                    "id": str(uuid.uuid4()),
                    "object": "chat.completion",
                    "created": time.time(),
                    "choices": [{
                        "message": Message(role="assistant", content=result["result"]),
                        "logprobs": result["logprobs"] if "logprobs" in result else None,
                    }],
                }

        if request.stream:
            return StreamingResponse(_resp_async_generator(response), media_type="application/x-ndjson")
        else:
            return response
    
    except HTTPException as e:
        raise e

    except Exception as e:
        logging.error(f"Error in chat completions: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/function-calling", tags=["NLP"])
async def function_call(request: FunctionCallRequest):
    try:
        messages = function_call_system_prompt + [
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
async def txt2img(request: ImageGenerationRequest):
    try:
        generation_kwargs = request.dict()

        generated_images = await nexa_run_image_generation(**generation_kwargs)

        resp = {"created": time.time(), "data": []}

        for image in generated_images:
            id = int(time.time())
            if not os.path.exists("nexa_server_output"):
                os.makedirs("nexa_server_output")
            image_path = os.path.join("nexa_server_output", f"txt2img_{id}.png")
            image.save(image_path)
            img = ImageResponse(base64=base64_encode_image(image_path), url=os.path.abspath(image_path))
            resp["data"].append(img)

        return resp

    except Exception as e:
        logging.error(f"Error in txt2img generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/v1/img2img", tags=["Computer Vision"])
async def img2img(request: ImageGenerationRequest):
    try:
        generation_kwargs = request.dict()

        generated_images = await nexa_run_image_generation(**generation_kwargs)
        resp = {"created": time.time(), "data": []}

        for image in generated_images:
            id = int(time.time())
            if not os.path.exists("nexa_server_output"):
                os.makedirs("nexa_server_output")
            image_path = os.path.join("nexa_server_output", f"img2img_{id}.png")
            image.save(image_path)
            img = ImageResponse(base64=base64_encode_image(image_path), url=os.path.abspath(image_path))
            resp["data"].append(img)

        return resp


    except Exception as e:
        logging.error(f"Error in img2img generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/v1/audio/transcriptions", tags=["Audio"])
async def transcribe_audio(
    file: UploadFile = File(...),
    beam_size: Optional[int] = Query(5, description="Beam size for transcription"),
    language: Optional[str] = Query(None, description="Language code (e.g., 'en', 'fr')"),
    temperature: Optional[float] = Query(0.0, description="Temperature for sampling"),
):

    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(file.filename)[1]) as temp_audio:
            temp_audio.write(await file.read())
            temp_audio_path = temp_audio.name

        transcribe_params = {
            "beam_size": beam_size,
            "language": language,
            "task": "transcribe",
            "temperature": temperature,
            "vad_filter": True
        }
        segments, _ = model.transcribe(temp_audio_path, **transcribe_params)
        transcription = "".join(segment.text for segment in segments)
        return JSONResponse(content={"text": transcription})
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error during transcription: {str(e)}")
    finally:
        os.unlink(temp_audio_path)

@app.post("/v1/audio/translations", tags=["Audio"])
async def translate_audio(
    file: UploadFile = File(...),
    beam_size: Optional[int] = Query(5, description="Beam size for translation"),
    temperature: Optional[float] = Query(0.0, description="Temperature for sampling"),
):
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(file.filename)[1]) as temp_audio:
            temp_audio.write(await file.read())
            temp_audio_path = temp_audio.name

        translate_params = {
            "beam_size": beam_size,
            "task": "translate",
            "temperature": temperature,
            "vad_filter": True
        }
        segments, _ = model.transcribe(temp_audio_path, **translate_params)
        translation = "".join(segment.text for segment in segments)
        return JSONResponse(content={"text": translation})
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error during translation: {str(e)}")
    finally:
        os.unlink(temp_audio_path)

@app.post("/v1/embeddings", tags=["Embedding"])
async def create_embedding(request: EmbeddingRequest):
    try:
        if isinstance(request.input, list):
            embeddings_results = [model.embed(text, normalize=request.normalize, truncate=request.truncate) for text in request.input]
        else:
            embeddings_results = model.embed(request.input, normalize=request.normalize, truncate=request.truncate)

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
        input_texts = request.input if isinstance(request.input, list) else [request.input]
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

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run the Nexa AI Text Generation Service"
    )
    parser.add_argument("model_path", type=str, nargs='?', default="gemma", help="Folder Path on Amazon S3")
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
    args = parser.parse_args()
    run_nexa_ai_service(
        args.model_path,
        is_local_path_arg=args.local_path,
        model_type_arg=args.model_type,
        huggingface=args.huggingface,
        nctx=args.nctx,
        host=args.host,
        port=args.port,
        reload=args.reload
    )