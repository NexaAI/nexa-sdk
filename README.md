# nexasdk-go

## setup

```
cd runner
# build dependence
make build

# run cli
make run
# run cli with other model file
make run MODEL=./path/to/model
```

## Roadmap
- [ ] `nexa infer`, `nexa pull`, `nexa list`, `nexa clean` works E2E for LLM model
    - `nexa infer` will eject resource after inference finishes
    - multiple round conversation with kv cache (default behavior)
    - download model from huggingface
    - list all local downloadable models, saved in `~/.cache/nexasdk`, print in table format
- [ ]  Remove submodule link, automatically pull dynamic C-lib based on device OS (windows, linux, macos) and architecture (x86, arm64), and GO binding works E2E for LLM model
    - Prepare `llama.cpp` shared c-lib in `nexa-sdk-internal` repo with Github Action
    - Setup shellscript to pull shared lib to local path
- [ ] `nexa serve` works E2E for LLM model with OpenAI compatible API
    - Follow OpenAI spec for LLM, VLM, ASR, TTS, image generation, etc.
    - `nexa run`, keep model loading alive for 5 min.
- [ ] Release nexa sdk as an App to download on windows & macos.
    - Add `launcher` for auto-update
- [ ] Integrate more backend
    - mlx-c
    - onnx-c
    - coreml-c

## Code Design