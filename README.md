# nexasdk-go

## setup

do with `make build`, or follow manually steps.

for darwin users, upgrade go to 1.24 then run:
```
brew install go@1.24
make -f Makefile.darwin build
```

### nexa-sdk-binding

`nexa-sdk-binding` is hosted in this repository. It will be moved out in the future, currently requires manually compiling.

#### Build llama.cpp

1. download llama.cpp
   1. `mkdir -p build/include build/lib`
   1. `git clone https://github.com/ggml-org/llama.cpp.git build/llama.cpp`
1. compile llama.cpp
   1. `cmake -B build/llama.cpp-build build/llama.cpp -DBUILD_SHARED_LIBS=ON -DLLAMA_BUILD_TESTS=OFF -DLLAMA_BUILD_TOOLS=OFF -DLLAMA_BUILD_EXAMPLES=OFF -DLLAMA_BUILD_SERVER=OFF -DLLAMA_CURL=OFF`
   1. `cmake --build build/llama.cpp-build -j --config Release`
1. copy shared lib and header files
   1. `cp ./build/llama.cpp/include/llama.h ./build/include`
   1. `cp ./build/llama.cpp/ggml/include/*.h ./build/include`
   1. `cp ./build/llama.cpp-build/bin/*.so ./build/lib`

#### Build nexa-sdk-binding

1. compile shared lib
   1. `mkdir -p build/binding-build`
   1. `g++ -O2 -fPIC -shared -I./build/include -L./build/lib -lllama -o ./build/binding-build/libbinding.so ./binding/binding.cpp`
2. copy shared lib and header file
   1. `cp ./build/binding-build/*.so ./build/lib/`
   1. `cp ./binding/binding.h ./build/include`

### runner

1. download binding library (skip currently) (choose one)
   - `make download`
   - download dist from [nexa-sdk-binding release page](), then put file in `build` like `build/include/binding.h` and `build/lib/libbinding.so`
2. `nexa-cli`
   1. build `cd runner && go build -o ../build/nexa ./cmd/nexa-cli`
   1. run with `LD_LIBRARY_PATH=./build/lib ./build/nexa`, will show usage
      1. pull model `LD_LIBRARY_PATH=./build/lib ./build/nexa pull Qwen/Qwen3-0.6B-GGUF`
      1. list models `LD_LIBRARY_PATH=./build/lib ./build/nexa list`
      1. run models `LD_LIBRARY_PATH=./build/lib ./build/nexa infer Qwen/Qwen3-0.6B-GGUF`
3. `nexa-launcher`
   1. TODO

## Run project
on Linux:
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
