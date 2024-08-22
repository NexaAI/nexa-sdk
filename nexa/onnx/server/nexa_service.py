import asyncio
import json
import logging
import os
import socket
import time
import uuid
from threading import Thread
from typing import List, Optional
import argparse
import uvicorn
from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse, JSONResponse, StreamingResponse
from optimum.onnxruntime import ORTModelForCausalLM
from pydantic import BaseModel
from transformers import AutoTokenizer, TextIteratorStreamer

from nexa.constants import NEXA_RUN_MODEL_MAP_ONNX
from nexa.general import pull_model

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
is_chat_mode = False
chat_format = None
completion_template = None
hostname = socket.gethostname()
conversation_history = [{"role": "system", "content": "You are a helpful assistant"}]
model_path = None

class GenerationRequest(BaseModel):
    prompt: str
    temperature: float = 1.0
    max_new_tokens: int = 256
    min_new_tokens: int = 1
    top_k: int = 50
    top_p: float = 1.0


async def load_model_and_tokenizer(model_path):
    global model, tokenizer, streamer
    model = ORTModelForCausalLM.from_pretrained(model_path)
    tokenizer = AutoTokenizer.from_pretrained(model_path)
    streamer = TextIteratorStreamer(
        tokenizer, skip_prompt=True, skip_special_tokens=True
    )
    return model, tokenizer, streamer


async def nexa_run_text_generation_preperation(model_path):
    global model, tokenizer, streamer, chat_template, is_chat_mode
    # Step 1: Check if the model_path is a key in NEXA_RUN_MODEL_MAP_ONNX, if so, get the full path
    full_model_path = NEXA_RUN_MODEL_MAP_ONNX.get(model_path, model_path)
    downloaded_onnx_folder, run_type = pull_model(full_model_path)
    logging.info(f"Downloaded ONNX folder: {downloaded_onnx_folder}")

    # Step 2: Load the model and tokenizer
    model, tokenizer, streamer = await load_model_and_tokenizer(downloaded_onnx_folder)
    # Step 3: Determine whether to use chat or completion mode
    if hasattr(tokenizer, "chat_template") and tokenizer.chat_template is not None:
        chat_template = tokenizer.chat_template
        is_chat_mode = True


@app.on_event("startup")
async def startup_event():
    model_path = os.getenv("MODEL_PATH", "gemma")
    logging.info(f"Model Path: {model_path}")
    await nexa_run_text_generation_preperation(model_path)


async def nexa_run_text_generation(
    prompt, temperature, max_new_tokens, min_new_tokens, top_k, top_p
) -> str:
    # Brian TODO : move the genai API logic here
    global model, tokenizer, streamer, chat_template, is_chat_mode, conversation_history
    if is_chat_mode:
        conversation_history.append({"role": "user", "content": prompt})
        full_prompt = tokenizer.apply_chat_template(
            conversation_history, chat_template=chat_template, tokenize=False
        )
        inputs = tokenizer(full_prompt, return_tensors="pt")
    else:
        inputs = tokenizer(prompt, return_tensors="pt")
    output = model.generate(
        **inputs,
        min_new_tokens=min_new_tokens,
        max_new_tokens=max_new_tokens,
        do_sample=True,
        temperature=temperature,
        streamer=streamer,
        top_k=top_k,
        top_p=top_p,
        pad_token_id=tokenizer.eos_token_id,
    )
    response = tokenizer.decode(
        output[0][len(inputs["input_ids"][0]) :], skip_special_tokens=True
    )
    if is_chat_mode:
        conversation_history.append({"role": "assistant", "content": response})
    return response


@app.get("/", response_class=HTMLResponse)
async def read_root(request: Request):
    return HTMLResponse(
        content=f"<h1>Welcome to Nexa AI</h1><p>Hostname: {hostname}</p>"
    )


# follow https://platform.openai.com/docs/api-reference/completions/create?lang=python
@app.post("/v1/completions")
async def generate_text(request: GenerationRequest):
    try:
        result = await nexa_run_text_generation(**request.model_dump())
        return JSONResponse(content={"result": result})
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# OpenAI compatible
class Message(BaseModel):
    role: str
    content: str


class ChatCompletionRequest(BaseModel):
    model_path: Optional[str] = "gemma"  # Default model path or make it optional
    messages: List[Message] = [
        Message(role="system", content="You are a helpful assistant"),
        Message(role="user", content="tell me a story")
    ]
    max_tokens: Optional[int] = 128
    temperature: Optional[float] = 0.5
    stream: Optional[bool] = False


def _resp_async_generator(streamer: TextIteratorStreamer):
    _id = str(uuid.uuid4())

    for token in streamer:
        chunk = {
            "id": _id,
            "object": "chat.completion.chunk",
            "created": time.time(),
            "model": model,
            "choices": [{"delta": {"content": token}}],
        }
        yield f"data: {json.dumps(chunk)}\n\n"

    yield "data: [DONE]\n\n"


def run_async_function(async_func, **kwargs):
    asyncio.run(async_func(**kwargs))


@app.post("/v1/chat/completions")
async def chat_completions(request: ChatCompletionRequest):
    await nexa_run_text_generation_preperation(model_path=args.model_path)

    generation_kwargs = GenerationRequest.model_construct(
        prompt="" if len(request.messages) == 0 else request.messages[-1].content,
        temperature=request.temperature,
        max_new_tokens=request.max_tokens,
    ).model_dump()

    if request.stream:
        global streamer

        thread = Thread(
            target=run_async_function,
            args=(nexa_run_text_generation,),
            kwargs=generation_kwargs,
        )
        thread.start()

        return StreamingResponse(
            _resp_async_generator(streamer),
            media_type="application/x-ndjson",
        )

    else:
        resp_content = await nexa_run_text_generation(**generation_kwargs)

        return {
            "id": str(uuid.uuid4()),
            "object": "chat.completion",
            "created": time.time(),
            "model": request.model_path,
            "choices": [{"message": Message(role="assistant", content=resp_content)}],
        }

def run_nexa_ai_service(model_path_arg, **kwargs):
    global model_path
    model_path = model_path_arg or "gemma"
    os.environ["MODEL_PATH"] = model_path
    host = kwargs.get("host", "0.0.0.0")
    port = kwargs.get("port", 8000)
    reload = kwargs.get("reload", False)
    uvicorn.run(app, host=host, port=port, reload=reload)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run the Nexa AI Text Generation Service"
    )
    parser.add_argument("model_path", type=str, nargs='?', default="gemma", help="Folder Path on Amazon S3")
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
    run_nexa_ai_service(args.model_path, host=args.host, port=args.port, reload=args.reload)