# nexasdk-go

## setup

1. Install dependencies
   - darwin
     `brew install go@1.24`
   - windows
     `choco install mingw make go -y`
   - linux (ubuntu 22.04)
     `apt install g++ make go -y`
1. download bridge library
   `make download` for default version, or specify version and backend, for example: `make download BRIDGE_VERSION=latest BRIDGE_BACKEND=mlx`.
1. build app
   `make build`

## Run project

do with `make run ARGS="serve"`, or follow steps bellow.

### Test inference

On Linux/WSL:

```shell
# helper manual
./build/nexa -h
# download model
./build/nexa pull Qwen/Qwen3-0.6B-GGUF
# inference
./build/nexa infer Qwen/Qwen3-0.6B-GGUF
```

Note that the model will be downloaded to [UserCacheDir](https://pkg.go.dev/os#UserCacheDir) with base64 encoded name.

````

### Test server

On Linux/WSL:

```shell
# server
./build/nexa serve

# Test server
curl -X POST http://127.0.0.1:8080/v1/completions \
-H "Content-Type: application/json" \
-d '{
  "model": "Qwen/Qwen3-0.6B-GGUF",
  "prompt": "Write a hello world program in Python",
  "max_tokens": 150
}'
````

We can also set environment variables

```shell
# Setup environment variables (need prefix `NEXA_`)
# Set custom host and port
export NEXA_HOST="0.0.0.0:8080"
# Set custom keep-alive timeout (in seconds)
export NEXA_KEEPALIVE=600

# Run server
./build/nexa serve
```

## Roadmap

- [x] `nexa infer`, `nexa pull`, `nexa list`, `nexa clean` works E2E for LLM model
  - [x] `nexa infer` will eject resource after inference finishes
  - [x] multiple round conversation with kv cache (default behavior)
  - [x] download model from huggingface
  - [x] list all local downloadable models, saved in user cache dir, print in table format
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
│   │   └── infer.go            # local infer command
│   │   └── run.go              # keepalive client
│   │   └── serve.go            # keepalive server
│   └── nexa-launcher           # runner updater
│       ├── main.go             #
├── nexa-sdk                    #
│   ├── nexa.go                 # nexa sdk go wrapper
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

## dependencies

- [cobra](https://pkg.go.dev/github.com/spf13/cobra) is a commander providing a simple interface to create powerful modern CLI interfaces
- [viper](https://github.com/spf13/viper) is a library for reading configuration files
- [sonic](https://github.com/bytedance/sonic) A blazingly fast JSON serializing & deserializing library, accelerated by JIT (just-in-time compiling) and SIMD (single-instruction-multiple-data).
- [gin](https://github.com/gin-gonic/gin) is a HTTP web framework

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
