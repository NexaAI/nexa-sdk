# Nexa SDK

The Nexa SDK is a comprehensive toolkit for supporting **ONNX** and **GGML** models. It supports text generation, image generation, vision-language models (VLM), and text-to-speech (TTS) capabilities. Additionally, it offers an OpenAI-compatible API server with JSON schema mode for function calling and streaming support, and a user-friendly Streamlit UI.

## Features

- **Model Support:**
  - **ONNX & GGML models**
  - **Conversion Engine**
  - **Inference Engine**:
    - **Text Generation**
    - **Image Generation**
    - **Vision-Language Models (VLM)**
    - **Text-to-Speech (TTS)**
- **Server:**
  - OpenAI-compatible API
  - JSON schema mode for function calling
  - Streaming support
- **Streamlit UI** for interactive model deployment and testing

## Installation
### install from PyPI
```bash
pip install nexaai
pip install nexaai[onnx] # if you need ONNX support
```

### build from source
```bash
git clone --recursive https://github.com/NexaAI/nexa-sdk.git
cd nexa-sdk
pip install -e .
pip install -e .[onnx] # if you need ONNX support
```

## Publishing to PYPI
Firstly build the wheel
```bash
python -m build
```
Then upload the wheel to PyPI
```bash
pip install twine
twine upload dist/*
```

### add a tag
```
git tag
git tag -d <version>
git tag <version>
git push origin <version>
```

## Testing

### Test Inference with GGUF Files

```bash
python -m nexa.gguf.nexa_inference_text gemma
python -m nexa.gguf.nexa_inference_text octopusv2 --stop_words "<nexa_end>"
wget https://assets-c4akfrf5b4d3f4b7.z01.azurefd.net/assets/2024/04/BMDataViz_661fb89f3845e.png -O test.png
python -m nexa.gguf.nexa_inference_vlm nanollava
python -m nexa.gguf.nexa_inference_image sd1-4
python -m nexa.gguf.nexa_inference_image sd1-4 --img2img
wget -O control_normal-fp16.safetensors https://huggingface.co/webui/ControlNet-modules-safetensors/resolve/main/control_normal-fp16.safetensors
wget -O controlnet_test.png https://huggingface.co/takuma104/controlnet_dev/resolve/main/gen_compare/control_images/converted/control_human_normal.png
python -m nexa.gguf.nexa_inference_image sd1-5 --control_net_path control_normal-fp16.safetensors --control_image_path controlnet_test.png
python -m nexa.gguf.nexa_inference_voice whisper-tiny
```

### Test with Streamlit UI

```bash
python -m nexa.gguf.nexa_inference_text gemma --streamlit
python -m nexa.gguf.nexa_inference_image sd1-4 --streamlit
python -m nexa.gguf.nexa_inference_vlm nanollava --streamlit
```

### Test CLI with GGUF Files

```bash
python -m nexa.cli.entry pull gpt2
python -m nexa.cli.entry gen-text gpt2
python -m nexa.cli.entry gen-text gemma
python -m nexa.cli.entry gen-image sd1-4
python -m nexa.cli.entry gen-image sd1-4 -i2i
python -m nexa.cli.entry vlm nanollava
python -m nexa.cli.entry server gemma
```

### Test CLI with ONNX Files

```bash
python -m nexa.cli.entry onnx pull gpt2
python -m nexa.cli.entry onnx gen-text gpt2
python -m nexa.cli.entry onnx gen-text gemma
python -m nexa.cli.entry onnx gen-image sd1-4
python -m nexa.cli.entry onnx gen-voice whisper
python -m nexa.cli.entry onnx tts ljspeech
python -m nexa.cli.entry onnx server gemma
```

### Test Inference with ONNX Files

```bash
python -m nexa.cli.entry onnx pull gpt2
python -m nexa.cli.entry onnx gen-text gpt2
python -m nexa.cli.entry onnx gen-text phi3
python -m nexa.cli.entry onnx gen-image sd1-4
wget https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav -O test.wav
python -m nexa.cli.entry onnx gen-voice whisper
python -m nexa.cli.entry onnx tts ljspeech
```

### Test Server

```bash
python -m nexa.gguf.server.nexa_service gemma
python -m nexa.onnx.server.nexa_service gemma
python -m nexa.gguf.server.nexa_service llama2-function-calling
```

### Test CLI with GGML Files

For CLI, no UI:

```shell
nexa-cli pull gpt2
nexa-cli gen-text gpt2
nexa-cli gen-text gemma
nexa-cli gen-text llama3.1
nexa-cli gen-text octopusv2 --stop_words "<nexa_end>"
nexa-cli vlm nanollava
nexa-cli gen-image sd1-4
nexa-cli gen-image sdxl-turbo
nexa-cli gen-image lcm-dreamshaper
```

For CLI with Streamlit UI:

```shell
nexa-cli gen-text gpt2 --streamlit
nexa-cli gen-image stable-diffusion --streamlit
nexa-cli vlm nanollava --streamlit
```

For CLI with Server Engine:

```shell
nexa-cli server gemma
nexa-cli server onnx gemma
```

### Test CLI with ONNX Files

```shell
nexa-cli onnx pull gpt2
nexa-cli onnx gen-text gpt2
nexa-cli onnx gen-text gemma
nexa-cli onnx gen-image sd1-4
nexa-cli onnx gen-voice whisper
nexa-cli onnx tts ljspeech
```

### Test Modules (Not Stable)

Test individual modules with downloaded GGUF files:

```shell
python  tests/verify_text_generation.py
python tests/verify_vlm.py
python tests/verify_image_generation.py
```

## Notes

### Display File Architecture

To ignore specific folders:

```shell
tree -I 'vendor|tests'
```

### Find and Manage Files

To find all `.so` files in Linux:

```shell
find . -name "*.so"
find . -name "*.so" -exec rm -f {} \; # delete
```

To find all `.dll` files in Windows:

```shell
dir /s /b *.dll
del /s /q *.dll # delete
Get-ChildItem -Recurse -Filter *.dll  # in PowerShell
dumpbin /dependents your_executable_or_dll.dll  # in Developer PowerShell for Visual Studio
```
