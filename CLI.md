## CLI Reference

### Overview
```
usage: nexa [-h] [-V] {run,onnx,server,pull,remove,clean,list,login,whoami,logout} ...

Nexa CLI tool for handling various model operations.

positional arguments:
  {run,onnx,server,pull,remove,clean,list,login,whoami,logout}
                        sub-command help
    run                 Run inference for various tasks using GGUF models.
    onnx                Run inference for various tasks using ONNX models.
    server              Run the Nexa AI Text Generation Service
    pull                Pull a model from official or hub.
    remove              Remove a model from local machine.
    clean               Clean up all model files.
    list                List all models in the local machine.
    login               Login to Nexa API.
    whoami              Show current user information.
    logout              Logout from Nexa API.

options:
  -h, --help            show this help message and exit
  -V, --version         Show the version of the Nexa SDK.
```

### List Local Models

List all models on your local computer.

```
nexa list
```

### Download a Model

Download a model file to your local computer from Nexa Model Hub.

```
nexa pull MODEL_PATH
usage: nexa pull [-h] model_path

positional arguments:
  model_path  Path or identifier for the model in Nexa Model Hub

options:
  -h, --help  show this help message and exit
```

#### Example

```
nexa pull llama2
```

### Remove a Model

Remove a model from your local computer.

```
nexa remove MODEL_PATH
usage: nexa remove [-h] model_path

positional arguments:
  model_path  Path or identifier for the model in Nexa Model Hub

options:
  -h, --help  show this help message and exit
```

#### Example

```
nexa remove llama2
```

### Remove All Downloaded Models

Remove all downloaded models on your local computer.

```
nexa clean
```

### Run a Model

Run a model on your local computer. If the model file is not yet downloaded, it will be automatically fetched first.

By default, `nexa` will run gguf models. To run onnx models, use `nexa onnx MODEL_PATH`

#### Run Text-Generation Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-t TEMPERATURE] [-m MAX_NEW_TOKENS] [-k TOP_K] [-p TOP_P] [-sw [STOP_WORDS ...]] [-pf] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -pf, --profiling      Enable profiling logs for the inference process
  -st, --streamlit      Run the inference in Streamlit UI

Text generation options:
  -t, --temperature TEMPERATURE
                        Temperature for sampling
  -m, --max_new_tokens MAX_NEW_TOKENS
                        Maximum number of new tokens to generate
  -k, --top_k TOP_K     Top-k sampling parameter
  -p, --top_p TOP_P     Top-p sampling parameter
  -sw, --stop_words [STOP_WORDS ...]
                        List of stop words for early stopping
```

##### Example

```
nexa run llama2
```

#### Run Image-Generation Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-i2i] [-ns NUM_INFERENCE_STEPS] [-np NUM_IMAGES_PER_PROMPT] [-H HEIGHT] [-W WIDTH] [-g GUIDANCE_SCALE] [-o OUTPUT] [-s RANDOM_SEED] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -st, --streamlit      Run the inference in Streamlit UI

Image generation options:
  -i2i, --img2img       Whether to run image-to-image generation
  -ns, --num_inference_steps NUM_INFERENCE_STEPS
                        Number of inference steps
  -np, --num_images_per_prompt NUM_IMAGES_PER_PROMPT
                        Number of images to generate per prompt
  -H, --height HEIGHT   Height of the output image
  -W, --width WIDTH     Width of the output image
  -g, --guidance_scale GUIDANCE_SCALE
                        Guidance scale for diffusion
  -o, --output OUTPUT   Output path for the generated image
  -s, --random_seed RANDOM_SEED
                        Random seed for image generation
  --lora_dir LORA_DIR   Path to directory containing LoRA files
  --wtype WTYPE         Weight type (f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
  --control_net_path CONTROL_NET_PATH
                        Path to control net model
  --control_image_path CONTROL_IMAGE_PATH
                        Path to image condition for Control Net
  --control_strength CONTROL_STRENGTH
                        Strength to apply Control Net
```

##### Example

```
nexa run sd1-4
```

#### Run Vision-Language Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-t TEMPERATURE] [-m MAX_NEW_TOKENS] [-k TOP_K] [-p TOP_P] [-sw [STOP_WORDS ...]] [-pf] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -pf, --profiling      Enable profiling logs for the inference process
  -st, --streamlit      Run the inference in Streamlit UI

VLM generation options:
  -t, --temperature TEMPERATURE
                        Temperature for sampling
  -m, --max_new_tokens MAX_NEW_TOKENS
                        Maximum number of new tokens to generate
  -k, --top_k TOP_K     Top-k sampling parameter
  -p, --top_p TOP_P     Top-p sampling parameter
  -sw, --stop_words [STOP_WORDS ...]
                        List of stop words for early stopping
```

##### Example

```
nexa run nanollava
```

#### Run Audio Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-o OUTPUT_DIR] [-b BEAM_SIZE] [-l LANGUAGE] [--task TASK] [-t TEMPERATURE] [-c COMPUTE_TYPE] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -st, --streamlit      Run the inference in Streamlit UI

Automatic Speech Recognition options:
  -b, --beam_size BEAM_SIZE
                        Beam size to use for transcription
  -l, --language LANGUAGE
                        The language spoken in the audio. It should be a language code such as 'en' or 'fr'.
  --task TASK           Task to execute (transcribe or translate)
  -c, --compute_type COMPUTE_TYPE
                        Type to use for computation (e.g., float16, int8, int8_float16)
```

##### Example

```
nexa run faster-whisper-tiny
```

### Start Local Server

Start a local server using models on your local computer.

```
nexa server MODEL_PATH
usage: nexa server [-h] [--host HOST] [--port PORT] [--reload] model_path

positional arguments:
  model_path   Path or identifier for the model in S3

options:
  -h, --help   show this help message and exit
  --host HOST  Host to bind the server to
  --port PORT  Port to bind the server to
  --reload     Enable automatic reloading on code changes
```

#### Example

```
nexa server llama2
```

### Model Path Format

For `model_path` in nexa commands, it's better to follow the standard format to ensure correct model loading and execution. The standard format for `model_path` is:

- `[user_name]/[repo_name]:[tag_name]` (user's model)
- `[repo_name]:[tag_name]` (official model)

#### Examples:

- `gemma-2b:q4_0`
- `Meta-Llama-3-8B-Instruct:onnx-cpu-int8`
- `alanzhuly/Qwen2-1B-Instruct:q4_0`