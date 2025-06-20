# nexasdk-go

## setup

do with `make build`, or follow manually steps.

> for darwin users, upgrade go to 1.24 `brew install go@1.24`
> for windows users, use [WSL](https://learn.microsoft.com/en-us/windows/wsl/), `ubuntu 22.04` is official support

1. download binding library (choose one)
   - `make download`
   - download dist from [nexasdk-bridge release page](), then put file in `build` like `build/include/ml.h` and `build/lib/libnexa_bridge.so/dylib`
2. build app
   1. `cd runner`
   1. build nexa-cli `go build -o ../build/nexa ./cmd/nexa-cli`
   1. TODO

## Run project

on Linux/WSL:

```shell
# helper manual
LD_LIBRARY_PATH=./build/lib ./build/nexa -h

# download model
LD_LIBRARY_PATH=./build/lib ./build/nexa pull Qwen/Qwen3-0.6B-GGUF

# inference
LD_LIBRARY_PATH=./build/lib ./build/nexa infer Qwen/Qwen3-0.6B-GGUF
```

On Mac:

```shell
# helper manual
DYLD_LIBRARY_PATH=./build/lib ./build/nexa -h

# download model
DYLD_LIBRARY_PATH=./build/lib ./build/nexa pull Qwen/Qwen3-0.6B-GGUF

# inference
DYLD_LIBRARY_PATH=./build/lib ./build/nexa infer Qwen/Qwen3-0.6B-GGUF
```

## Roadmap

- [ ] `nexa infer`, `nexa pull`, `nexa list`, `nexa clean` works E2E for LLM model
  - `nexa infer` will eject resource after inference finishes
  - multiple round conversation with kv cache (default behavior)
  - download model from huggingface
  - list all local downloadable models, saved in user cache dir, print in table format
- [ ] Remove submodule link, automatically pull dynamic C-lib based on device OS (windows, linux, macos) and architecture (x86, arm64), and GO binding works E2E for LLM model
  - Prepare `llama.cpp` shared c-lib in `nexa-sdk-internal` repo with Github Action
  - Setup shellscript to pull shared lib to local path
- [ ] `nexa serve` works E2E for LLM model with OpenAI compatible API
  - Follow OpenAI spec for LLM, VLM, ASR, TTS, image generation, etc.
  - `nexa run`, keep model loading alive for 5 min.
- [ ] Release nexa sdk as an App to download on windows & macos.
  - Add `launcher` for auto-update

## Code Design

```
.
├── cmd                         #
│   ├── nexa-cli                # cli app
│   │   ├── main.go             # cli entrypoint
│   │   ├── repl.go             # repl
│   │   └── runner.go           # simple runner implement
│   └── nexa-launcher           # runner updater
│       ├── downloader.go       #
│       ├── main.go             #
│       └── updater.go          #
├── nexa-sdk                    #
│   ├── nexa.go                 # nexa sdk go api
│   ├── wrap.c                  # wrap code
│   └── wrap.go                 # wrap code
├── internal                    #
│   ├── auth                    # http auth function
│   │   └── auth.go             #
│   ├── config                  # load config from env or file
│   │   └── config.go           #
│   └── store                   # local data management
│       ├── downloader.go       # model file downloader
│       ├── manager.go          # local data manager
│       ├── manifest.go         # manifest management
│       └── model.go            # model file management
├── server                      #
│   ├── frontend                #
│   │   ├── frontend.go         # go embed entrypoint
│   │   └── index.html          # simple debug page
│   ├── handler                 # openai compatible router
│   │   ├── asr.go              #
│   │   ├── cv.go               #
│   │   ├── embedding.go        #
│   │   ├── llm.go              #
│   │   ├── multimodal.go       #
│   │   ├── rerank.go           #
│   │   ├── tts.go              #
│   │   └── system.go           # server status api
│   ├── middleware              # http middlewares
│   │   └── auth.go             #
│   ├── route.go                # handlers route register
│   └── server.go               # http server starter
├── Makefile                    # build, test, package commands
└── go.mod                      # go dependencies file
```

## GO optimization

- c-go binding optmization
  - [unsafe](https://pkg.go.dev/unsafe) provide type convert without copy, but be careful
- go server optimization
  - use [goroutine](https://go.dev/doc/effective_go#goroutines) and [channel](https://go.dev/doc/effective_go#channels) to speed up Concurrency performance
  - [sync.Pool](https://pkg.go.dev/sync#Pool) object pool, reduce malloc and gc pressure
  - [atomic](https://pkg.go.dev/sync/atomic) lightweight mutex replacement
  - [sonic](https://github.com/bytedance/sonic) fast json library, but amd64/arm64 only
  - use [c.Request.Body](https://pkg.go.dev/net/http#Request) directly when body is large
- cross-compile optimization and environment setup
  - set `GOPATH` env, go will download dependencies here
  - cross-compile pure go
    - `CGO_ENABLED=0` disable cgo, then c compiler is no need
    - export `GOOS` `GOARCH`, then you can cross-compile to almost platform
  - cross-compile cgo
    - `CGO_ENABLED=1` enable c-go
    - export `CC` for cross compile toolchain
    - export `CGO_CFLAGS` `CGO_LDFLAGS`, but it's better to declare in wrapper code, like [llm.go](./runner/nexa-sdk/llm.go)
