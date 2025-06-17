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

## code design

## Roadmap
- [ ] `nexa run`, `nexa pull`, `nexa list`, `nexa clean` works E2E for LLM model form huggingface
- [ ]  Remove submodule link, automatically pull dynamic C-lib based on device OS (windows, linux, macos) and architecture (x86, arm64), and GO binding works E2E for LLM model
- [ ] `nexa serve` works E2E for LLM model with OpenAI compatible API