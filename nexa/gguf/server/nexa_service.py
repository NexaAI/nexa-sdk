import json
import logging
import os
import socket
import time
import uuid
from typing import List, Optional, Dict, Any
import base64
import multiprocessing
from PIL import Image
import tempfile
import uvicorn
from fastapi import FastAPI, HTTPException, Request, File, UploadFile, Query
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse, JSONResponse, StreamingResponse
from pydantic import BaseModel

from nexa.constants import (
    NEXA_RUN_CHAT_TEMPLATE_MAP,
    NEXA_RUN_COMPLETION_TEMPLATE_MAP,
    NEXA_RUN_MODEL_PRECISION_MAP,
    NEXA_RUN_MODEL_MAP_FUNCTION_CALLING,
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model
from nexa.gguf.llama.llama import Llama
from nexa.gguf.sd.stable_diffusion import StableDiffusion
from faster_whisper import WhisperModel
import argparse

logging.basicConfig(level=logging.INFO)

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

# Request Classes
class GenerationRequest(BaseModel):
    prompt: str = "Tell me a story"
    temperature: float = 1.0
    max_new_tokens: int = 128
    top_k: int = 50
    top_p: float = 1.0
    stop_words: Optional[List[str]] = []
    logprobs: Optional[bool] = False
    top_logprobs: Optional[int] = 4

class Message(BaseModel):
    role: str
    content: str

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

# helper functions
async def load_model():
    global model, chat_format, completion_template, model_path, n_ctx
    downloaded_path, run_type = pull_model(model_path)
    if run_type == "NLP":
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
                        n_ctx=n_ctx
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
                        n_ctx=n_ctx
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
                        n_ctx=n_ctx
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
                        n_ctx=n_ctx
                    )
                logging.info(f"model loaded as {model}")
                chat_format = model.metadata.get("tokenizer.chat_template", None)               
    elif run_type == "Computer Vision":
        with suppress_stdout_stderr():
            model = StableDiffusion(
                model_path=downloaded_path,
                wtype=NEXA_RUN_MODEL_PRECISION_MAP.get(
                    model_path, "f32"
                ),  # Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
                n_threads=multiprocessing.cpu_count(),
            )
        logging.info(f"model loaded as {model}")
    elif run_type == "Audio":
        with suppress_stdout_stderr():
            model = WhisperModel(
                downloaded_path,
                device="auto",
                compute_type="default"
            )
        logging.info(f"model loaded as {model}")
    else:
        raise ValueError(f"Model {model_path} not found in Model Hub")

async def nexa_run_text_generation(
    prompt, temperature, stop_words, max_new_tokens, top_k, top_p, logprobs=None, top_logprobs=None
) -> Dict[str, Any]:
    global model, chat_format, completion_template
    if model is None:
        raise ValueError("Model is not loaded. Please check the model path and try again.")

    generated_text = ""
    logprobs_or_none = None  # init to store the logprobs if requested

    if chat_format:
        messages = chat_completion_system_prompt + [{"role": "user", "content": prompt}]

        params = {
            'messages': messages,
            'temperature': temperature,
            'max_tokens': max_new_tokens,
            'top_k': top_k,
            'top_p': top_p,
            'stream': True,
            'stop': stop_words,
            'logprobs': logprobs,
            'top_logprobs': top_logprobs,
        }

        streamer = model.create_chat_completion(**params)

        for chunk in streamer:
            delta = chunk["choices"][0]["delta"]
            if "content" in delta:
                generated_text += delta["content"]

            if logprobs and "logprobs" in chunk["choices"][0]:
                if logprobs_or_none is None:
                    logprobs_or_none = chunk["choices"][0]["logprobs"]
                else:
                    for key in logprobs_or_none:  # tokens, token_logprobs, top_logprobs, text_offset
                        if key in chunk["choices"][0]["logprobs"]:
                            logprobs_or_none[key].extend(chunk["choices"][0]["logprobs"][key])  # accumulate data from each chunk                            
    else:
        if completion_template:
            formatted_prompt = completion_template.format(input=prompt)
        else:
            formatted_prompt = prompt

        streamer = model.create_completion(
            prompt=formatted_prompt,
            temperature=temperature,
            max_tokens=max_new_tokens,
            top_k=top_k,
            top_p=top_p,
            stream=True,
            stop=stop_words,
            logprobs=logprobs,
            top_logprobs=top_logprobs,
        )

        for chunk in streamer:
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


def run_nexa_ai_service(model_path_arg, **kwargs):
    global model_path, n_ctx
    model_path = model_path_arg or "gemma"
    os.environ["MODEL_PATH"] = model_path
    n_ctx = kwargs.get("nctx", 2048)
    host = kwargs.get("host", "0.0.0.0")
    port = kwargs.get("port", 8000)
    reload = kwargs.get("reload", False)
    uvicorn.run(app, host=host, port=port, reload=reload)

# Endpoints
@app.on_event("startup")
async def startup_event():
    global model_path
    model_path = os.getenv("MODEL_PATH", "gemma")
    logging.info(f"Model Path: {model_path}")
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
        result = await nexa_run_text_generation(**request.dict())

        return JSONResponse(content={
            "choices": [{
                "text": result["result"],
                "logprobs": result.get("logprobs")
            }]
        })
    except Exception as e:
        logging.error(f"Error in text generation: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/v1/chat/completions", tags=["NLP"])
async def chat_completions(request: ChatCompletionRequest):
    try:
        generation_kwargs = GenerationRequest(
            prompt="" if len(request.messages) == 0 else request.messages[-1].content,
            temperature=request.temperature,
            max_new_tokens=request.max_tokens,
            stop_words=request.stop_words,
            logprobs=request.logprobs,
            top_logprobs=request.top_logprobs,
        ).dict()

        if request.stream:
            # run the generation and stream the response:
            async def stream_generator():
                streamer = await nexa_run_text_generation(**generation_kwargs)
                async for chunk in _resp_async_generator(streamer):
                    yield chunk

            return StreamingResponse(stream_generator(), media_type="application/x-ndjson")

        else:
            # generate text synchronously and return the response:
            result = await nexa_run_text_generation(**generation_kwargs)
            return {
                "id": str(uuid.uuid4()),
                "object": "chat.completion",
                "created": time.time(),
                "choices": [{
                    "message": Message(role="assistant", content=result["result"]),
                    "logprobs": result["logprobs"] if "logprobs" in result else None,
                }],
            }
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

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run the Nexa AI Text Generation Service"
    )
    parser.add_argument("model_path", type=str, nargs='?', default="gemma", help="Folder Path on Amazon S3")
    parser.add_argument(
        "--nctx", type=int, default=2048, help="Length of context window"
    )
    parser.add_argument(
        "--host", type=str, default="0.0.0.0", help="Host to bind the server to"
    )
    parser.add_argument(
        "--port", type=int, default=8000, help="Port to bind the server to"
    )
    parser.add_argument(
        "--reload",
        action="store_true",
        help="Enable automatic reloading on code changes",
    )
    args = parser.parse_args()
    run_nexa_ai_service(args.model_path, nctx=args.nctx, host=args.host, port=args.port, reload=args.reload)