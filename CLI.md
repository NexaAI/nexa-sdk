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
    embed               Generate embeddings for text.
    convert             Convert and quantize a Hugging Face model to GGUF format.
    server              Run the Nexa AI Text Generation Service.
    eval                Run the Nexa AI Evaluation Tasks.
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

List all models on your local computer. You can use `nexa run <model_name>` to run any model shown in the list.

```
nexa list
```

### Download a Model

Download a model file to your local computer from Nexa Model Hub.

```
nexa pull MODEL_PATH
usage: nexa pull [-h] model_path

positional arguments:
  model_path  Path or identifier for the model in Nexa Model Hub, Hugging Face repo ID when using -hf flag, or ModelScope model ID when using -ms flag

options:
  -h, --help            show this help message and exit
  -hf, --huggingface    Pull model from Hugging Face Hub
  -ms, --modelscope     Pull model from ModelScope Hub
  -o, --output_path OUTPUT_PATH
                        Custom output path for the pulled model
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

You can run any model shown in `nexa list` command.

#### Run Text-Generation Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-t TEMPERATURE] [-m MAX_NEW_TOKENS] [-k TOP_K] [-p TOP_P] [-sw [STOP_WORDS ...]] [-pf] [-st] [-lp] [-mt {NLP, COMPUTER_VISION, MULTIMODAL, AUDIO}] [-hf] [-ms] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -pf, --profiling      Enable profiling logs for the inference process
  -st, --streamlit      Run the inference in Streamlit UI, can be used with -lp or -hf
  -lp, --local_path     Indicate that the model path provided is the local path
  -mt, --model_type     Indicate the model running type, must be used with -lp or -hf or -ms, choose from [NLP, COMPUTER_VISION, MULTIMODAL, AUDIO]
  -hf, --huggingface    Load model from Hugging Face Hub
  -ms, --modelscope     Load model from ModelScope Hub

Text generation options:
  -t, --temperature TEMPERATURE
                        Temperature for sampling
  -m, --max_new_tokens MAX_NEW_TOKENS
                        Maximum number of new tokens to generate
  -k, --top_k TOP_K     Top-k sampling parameter
  -p, --top_p TOP_P     Top-p sampling parameter
  -sw, --stop_words [STOP_WORDS ...]
                        List of stop words for early stopping
  --nctx TEXT_CONTEXT   Length of context window
```

##### Example

```
nexa run llama2
```

#### Run Image-Generation Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-i2i] [-ns NUM_INFERENCE_STEPS] [-np NUM_IMAGES_PER_PROMPT] [-H HEIGHT] [-W WIDTH] [-g GUIDANCE_SCALE] [-o OUTPUT] [-s RANDOM_SEED] [-st] [-lp] [-mt {NLP, COMPUTER_VISION, MULTIMODAL, AUDIO}] [-hf] [-ms] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -st, --streamlit      Run the inference in Streamlit UI, can be used with -lp or -hf
  -lp, --local_path     Indicate that the model path provided is the local path
  -mt, --model_type     Indicate the model running type, must be used with -lp or -hf or -ms, choose from [NLP, COMPUTER_VISION, MULTIMODAL, AUDIO]
  -hf, --huggingface    Load model from Hugging Face Hub
  -ms, --modelscope     Load model from ModelScope Hub

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
                        Path to the control net model
  --control_image_path CONTROL_IMAGE_PATH
                        Path to the image condition for Control Net
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
usage: nexa run [-h] [-t TEMPERATURE] [-m MAX_NEW_TOKENS] [-k TOP_K] [-p TOP_P] [-sw [STOP_WORDS ...]] [-pf] [-st] [-lp] [-mt {NLP, COMPUTER_VISION, MULTIMODAL, AUDIO}] [-hf] [-ms] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -pf, --profiling      Enable profiling logs for the inference process
  -st, --streamlit      Run the inference in Streamlit UI, can be used with -lp or -hf
  -lp, --local_path     Indicate that the model path provided is the local path
  -mt, --model_type     Indicate the model running type, must be used with -lp or -hf or -ms, choose from [NLP, COMPUTER_VISION, MULTIMODAL, AUDIO]
  -hf, --huggingface    Load model from Hugging Face Hub
  -ms, --modelscope     Load model from ModelScope Hub

VLM generation options:
  -t, --temperature TEMPERATURE
                        Temperature for sampling
  -m, --max_new_tokens MAX_NEW_TOKENS
                        Maximum number of new tokens to generate
  -k, --top_k TOP_K     Top-k sampling parameter
  -p, --top_p TOP_P     Top-p sampling parameter
  -sw, --stop_words [STOP_WORDS ...]
                        List of stop words for early stopping
  --nctx TEXT_CONTEXT   Length of context window
```

##### Example

```
nexa run nanollava
```

#### Run Audio Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-o OUTPUT_DIR] [-b BEAM_SIZE] [-l LANGUAGE] [--task TASK] [-t TEMPERATURE] [-c COMPUTE_TYPE] [-st] [-lp] [-mt {NLP, COMPUTER_VISION, MULTIMODAL, AUDIO}] [-hf] [-ms] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -st, --streamlit      Run the inference in Streamlit UI, can be used with -lp or -hf
  -lp, --local_path     Indicate that the model path provided is the local path
  -mt, --model_type     Indicate the model running type, must be used with -lp or -hf or -ms, choose from [NLP, COMPUTER_VISION, MULTIMODAL, AUDIO]
  -hf, --huggingface    Load model from Hugging Face Hub
  -ms, --modelscope     Load model from ModelScope Hub

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

### Generate Embeddings

#### Generate Text Embeddings

```
nexa embed MODEL_PATH
usage: nexa embed [-h] [-lp] [-hf] [-ms] [-n] [-nt] model_path prompt

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub
  prompt                Prompt to generate embeddings

options:
  -h, --help            show this help message and exit
  -lp, --local_path     Indicate that the model path provided is the local path
  -hf, --huggingface    Load model from Hugging Face Hub
  -ms, --modelscope     Load model from ModelScope Hub
  -n, --normalize       Normalize the embeddings
  -nt, --no_truncate    Not truncate the embeddings
```

#### Example

```
nexa embed mxbai "I love Nexa AI."
nexa embed nomic "I love Nexa AI." >> generated_embeddings.txt
nexa embed nomic-embed-text-v1.5:fp16 "I love Nexa AI."
nexa embed sentence-transformers/all-MiniLM-L6-v2:gguf-fp16 "I love Nexa AI." >> generated_embeddings.txt
```

### Convert and quantize a Hugging Face Model to GGUF

Additional package `nexa-gguf` is required to run this command.

You can install it by `pip install "nexaai[convert]"` or `pip install nexa-gguf`.

```
nexa convert HF_MODEL_PATH [ftype] [output_file]
usage: nexa convert [-h] [-t NTHREAD] [--convert_type CONVERT_TYPE] [--bigendian] [--use_temp_file] [--no_lazy]
                    [--metadata METADATA] [--split_max_tensors SPLIT_MAX_TENSORS] [--split_max_size SPLIT_MAX_SIZE]
                    [--no_tensor_first_split] [--vocab_only] [--dry_run] [--output_tensor_type OUTPUT_TENSOR_TYPE]
                    [--token_embedding_type TOKEN_EMBEDDING_TYPE] [--allow_requantize] [--quantize_output_tensor]
                    [--only_copy] [--pure] [--keep_split] input_path [ftype] [output_file]

positional arguments:
  input_path            Path to the input Hugging Face model directory or GGUF file
  ftype                 Quantization type (default: q4_0)
  output_file           Path to the output quantized GGUF file

options:
  -h, --help            show this help message and exit
  -t, --nthread NTHREAD Number of threads to use (default: 4)
  --convert_type CONVERT_TYPE
                        Conversion type for safetensors to GGUF (default: f16)
  --bigendian           Use big endian format
  --use_temp_file       Use a temporary file during conversion
  --no_lazy             Disable lazy loading
  --metadata METADATA   Additional metadata as JSON string
  --split_max_tensors SPLIT_MAX_TENSORS
                        Maximum number of tensors per split
  --split_max_size SPLIT_MAX_SIZE
  --no_tensor_first_split
                        Disable tensor-first splitting
  --vocab_only          Only process vocabulary
  --dry_run             Perform a dry run without actual conversion
  --output_tensor_type  Output tensor type
  --token_embedding_type
                        Token embedding type
  --allow_requantize    Allow quantizing non-f32/f16 tensors
  --quantize_output_tensor
                        Quantize output.weight
  --only_copy           Only copy tensors (ignores ftype, allow_requantize, and quantize_output_tensor)
  --pure                Quantize all tensors to the default type
  --keep_split          Quantize to the same number of shards
  -ms --modelscope      Load model from ModelScope Hub
```

#### Example

```
# Default quantization type (q4_0) and output file in current directory
nexa convert meta-llama/Llama-3.2-1B-Instruct

# Equivalent to:
# nexa convert meta-llama/Llama-3.2-1B-Instruct q4_0 ./Llama-3.2-1B-Instruct-q4_0.gguf

# Specifying quantization type and output file
nexa convert meta-llama/Llama-3.2-1B-Instruct q6_k llama3.2-1b-instruct-q6_k.gguf
```

Note: When not specified, the default quantization type is set to q4_0, and the output file will be created in the current directory with the name format: `<model_name>-q4_0.gguf`.

### Start Local Server

Start a local server using models on your local computer.

```
nexa server MODEL_PATH
usage: nexa server [-h] [--host HOST] [--port PORT] [--reload] [-lp] [-mt {NLP, COMPUTER_VISION, MULTIMODAL, AUDIO}] [-hf] [-ms] model_path

positional arguments:
  model_path   Path or identifier for the model in S3

options:
  -h, --help   show this help message and exit
  -lp, --local_path     Indicate that the model path provided is the local path
  -mt, --model_type     Indicate the model running type, must be used with -lp or -hf or -ms, choose from [NLP, COMPUTER_VISION, MULTIMODAL, AUDIO]
  -hf, --huggingface    Load model from Hugging Face Hub
  -ms, --modelscope     Load model from ModelScope Hub
  --host HOST  Host to bind the server to
  --port PORT  Port to bind the server to
  --reload     Enable automatic reloading on code changes
```

#### Example

```
nexa server llama2
```

### Run Model Evaluation

Run evaluation using models on your local computer.

```
usage: nexa eval model_path [-h] [--tasks TASKS] [--limit LIMIT]

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  --tasks TASKS         Tasks to evaluate, comma-separated
  --limit LIMIT         Limit the number of examples per task. If <1, limit is a percentage of the total number of examples.
```

#### Examples

```
nexa eval phi3 --tasks ifeval --limit 0.5
```

### Model Path Format

For `model_path` in nexa commands, it's better to follow the standard format to ensure correct model loading and execution. The standard format for `model_path` is:

- `[user_name]/[repo_name]:[tag_name]` (user's model)
- `[repo_name]:[tag_name]` (official model)

#### Examples:

- `gemma-2b:q4_0`
- `Meta-Llama-3-8B-Instruct:onnx-cpu-int8`
- `liuhaotian/llava-v1.6-vicuna-7b:gguf-q4_0`

```
</rewritten_chunk>
```
