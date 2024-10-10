## Start Local Server

You can start a local server using models on your local computer with the `nexa server` command. Here's the usage syntax:

```
usage: nexa server [-h] [--host HOST] [--port PORT] [--reload] model_path
```

### Options:

- `-lp, --local_path`: Indicate that the model path provided is the local path, must be used with -mt
- `-mt, --model_type`: Indicate the model running type, must be used with -lp or -hf, choose from [NLP, COMPUTER_VISION, MULTIMODAL, AUDIO]
- `-hf, --huggingface`: Load model from Hugging Face Hub, must be used with -mt
- `--host`: Host to bind the server to
- `--port`: Port to bind the server to
- `--reload`: Enable automatic reloading on code changes
- `--nctx`: Maximum context length of the model you're using

### Example Commands:

```
nexa server gemma
nexa server llama2-function-calling
nexa server sd1-5
nexa server faster-whipser-large
nexa server ../models/llava-v1.6-vicuna-7b/ -lp -mt MULTIMODAL
```

By default, `nexa server` will run gguf models. To run onnx models, simply add `onnx` after `nexa server`.

## API Endpoints

### 1. Text Generation: <code>/v1/completions</code>

Generates text based on a single prompt.

#### Request body:

```json
{
  "prompt": "Tell me a story",
  "temperature": 1,
  "max_new_tokens": 128,
  "top_k": 50,
  "top_p": 1,
  "stop_words": ["string"],
  "stream": false
}
```

#### Example Response:

```json
{
  "result": "Once upon a time, in a small village nestled among rolling hills..."
}
```

### 2. Chat Completions: <code>/v1/chat/completions</code>

Update: Now supports multimodal inputs when using Multimodal models.

Handles chat completions with support for conversation history.

#### Request body:

Multimodal models (VLM):

```json
{
  "model": "anything",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "What’s in this image?"
        },
        {
          "type": "image_url",
          "image_url": {
            "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
          }
        }
      ]
    }
  ],
  "max_tokens": 300,
  "temperature": 0.7,
  "top_p": 0.95,
  "top_k": 40,
  "stream": false
}
```

Traditional NLP models:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Tell me a story"
    }
  ],
  "max_tokens": 128,
  "temperature": 0.1,
  "stream": false,
  "stop_words": []
}
```

#### Example Response:

```json
{
  "id": "f83502df-7f5a-4825-a922-f5cece4081de",
  "object": "chat.completion",
  "created": 1723441724.914671,
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "In the heart of a mystical forest..."
      }
    }
  ]
}
```

### 3. Function Calling: <code>/v1/function-calling</code>

Call the most appropriate function based on user's prompt.

#### Request body:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Extract Jason is 25 years old"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "UserDetail",
        "parameters": {
          "properties": {
            "name": {
              "description": "The user's name",
              "type": "string"
            },
            "age": {
              "description": "The user's age",
              "type": "integer"
            }
          },
          "required": ["name", "age"],
          "type": "object"
        }
      }
    }
  ],
  "tool_choice": "auto"
}
```

#### Function format:

```json
{
  "type": "function",
  "function": {
    "name": "function_name",
    "description": "function_description",
    "parameters": {
      "type": "object",
      "properties": {
        "property_name": {
          "type": "string | number | boolean | object | array",
          "description": "string"
        }
      },
      "required": ["array_of_required_property_names"]
    }
  }
}
```

#### Example Response:

```json
{
  "id": "chatcmpl-7a9b0dfb-878f-4f75-8dc7-24177081c1d0",
  "object": "chat.completion",
  "created": 1724186442,
  "model": "/home/ubuntu/.cache/nexa/hub/official/Llama2-7b-function-calling/q3_K_M.gguf",
  "choices": [
    {
      "finish_reason": "tool_calls",
      "index": 0,
      "logprobs": null,
      "message": {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "id": "call__0_UserDetail_cmpl-8d5cf645-7f35-4af2-a554-2ccea1a67bdd",
            "type": "function",
            "function": {
              "name": "UserDetail",
              "arguments": "{ \"name\": \"Jason\", \"age\": 25 }"
            }
          }
        ],
        "function_call": {
          "name": "",
          "arguments": "{ \"name\": \"Jason\", \"age\": 25 }"
        }
      }
    }
  ],
  "usage": {
    "completion_tokens": 15,
    "prompt_tokens": 316,
    "total_tokens": 331
  }
}
```

### 4. Text-to-Image: <code>/v1/txt2img</code>

Generates images based on a single prompt.

#### Request body:

```json
{
  "prompt": "A girl, standing in a field of flowers, vivid",
  "image_path": "",
  "cfg_scale": 7,
  "width": 256,
  "height": 256,
  "sample_steps": 20,
  "seed": 0,
  "negative_prompt": ""
}
```

#### Example Response:

```json
{
  "created": 1724186615.5426757,
  "data": [
    {
      "base64": "base64_of_generated_image",
      "url": "path/to/generated_image"
    }
  ]
}
```

### 5. Image-to-Image: <code>/v1/img2img</code>

Modifies existing images based on a single prompt.

#### Request body:

```json
{
  "prompt": "A girl, standing in a field of flowers, vivid",
  "image_path": "path/to/image",
  "cfg_scale": 7,
  "width": 256,
  "height": 256,
  "sample_steps": 20,
  "seed": 0,
  "negative_prompt": ""
}
```

#### Example Response:

```json
{
  "created": 1724186615.5426757,
  "data": [
    {
      "base64": "base64_of_generated_image",
      "url": "path/to/generated_image"
    }
  ]
}
```

### 6. Audio Transcriptions: <code>/v1/audio/transcriptions</code>

Transcribes audio files to text.

#### Parameters:

- `beam_size` (integer): Beam size for transcription (default: 5)
- `language` (string): Language code (e.g., 'en', 'fr')
- `temperature` (number): Temperature for sampling (default: 0)

#### Request body:

```
{
  "file" (form-data): The audio file to transcribe (required)
}
```

#### Example Response:

```json
{
  "text": " And so my fellow Americans, ask not what your country can do for you, ask what you can do for your country."
}
```

### 7. Audio Translations: <code>/v1/audio/translations</code>

Translates audio files to text in English.

#### Parameters:

- `beam_size` (integer): Beam size for transcription (default: 5)
- `temperature` (number): Temperature for sampling (default: 0)

#### Request body:

```
{
  "file" (form-data): The audio file to transcribe (required)
}
```

#### Example Response:

```json
{
  "text": " Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday"
}
```

### 8. Generate Embeddings: <code>/v1/embeddings</code>

Generate embeddings for a given text.

#### Request body:

```json
{
  "input": "I love Nexa AI.",
  "normalize": false,
  "truncate": true
}
```

#### Example Response:

```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [
        -0.006929283495992422,
        -0.005336422007530928,
        ... (omitted for spacing)
        -4.547132266452536e-05,
        -0.024047505110502243
      ],
    }
  ],
  "model": "/home/ubuntu/models/embedding_models/mxbai-embed-large-q4_0.gguf",
  "usage": {
    "prompt_tokens": 5,
    "total_tokens": 5
  }
}
```
