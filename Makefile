AGS ?= infer Qwen/Qwen3-0.6B-GGUF
#ARGS ?= serve

# ubuntu 22.04: llama-cpp-cpu llama-cpp-cuda
# macos 13: llama-cpp-metal
# macos 14: llama-cpp-metal mlx
# macos 15: llama-cpp-metal mlx
# windows: llama-cpp-cpu llama-cpp-vulkan llama-cpp-cuda
BRIDGE_VERSION ?= v0.1.1-rc13

ifeq ($(OS), Windows_NT)
	OS := windows
	LIB :=
	EXT := dll
	EXE := .exe
	RM := powershell Remove-Item -Recurse -Force -Path
	MKDIR := powershell New-Item -ItemType Directory -Force -Path

	BRIDGE_BACKEND ?= llama-cpp-cpu
else
	UNAME := $(shell uname -s)
	ifeq ($(UNAME), Linux)
		OS := linux
		LIB := lib
		EXT := so
		BRIDGE_BACKEND ?= llama-cpp-cpu
	else ifeq ($(UNAME), Darwin)
		MACOS_VERSION := $(shell sw_vers -productVersion))
		MACOS_MAJOR_VERSION := $(firstword $(subst ., ,$(MACOS_VERSION)))
		OS := macos-$(MACOS_MAJOR_VERSION)
		LIB := lib
		EXT := dylib
		BRIDGE_BACKEND ?= llama-cpp-metal
	endif

	EXE :=
	RM := rm -rf
	MKDIR := mkdir -p
endif

.PHONY: run build doc test download clean

# Allow passing arguments to run target
# Usage: make run ARGS="pull Qwen/Qwen2.5-7B-Instruct"
# Or use: make run pull Qwen/Qwen2.5-7B-Instruct
run:
	@if [ "$(filter-out run,$(MAKECMDGOALS))" != "" ]; then \
		./build/nexa$(EXE) $(filter-out run,$(MAKECMDGOALS)); \
	else \
		./build/nexa$(EXE) $(ARGS); \
	fi

build:
	cd ./runner/nexa-sdk/stub && g++ -std=c++11 -O3 -fPIC -shared -o $(LIB)nexa_bridge.$(EXT) *.cpp
	cd ./runner && go build -o ../build/nexa$(EXE) ./cmd/nexa-launcher
	cd ./runner && go build \
		-tags="sonic avx" \
		-ldflags "-X 'main.Version=$(VERSION)'" \
		-o ../build/nexa-cli$(EXE) \
		./cmd/nexa-cli

test:
	cd runner && LD_LIBRARY_PATH=$(PWD)/build/lib go test -v ./nexa-sdk --run VLM

download:
	-$(RM) ./build/lib/${BRIDGE_BACKEND}
	$(MKDIR) ./build/lib/${BRIDGE_BACKEND}
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "BRIDGE_BACKEND: $(BRIDGE_BACKEND)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/lib/${BRIDGE_BACKEND}/nexasdk-bridge.zip https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(BRIDGE_BACKEND)/$(OS)/nexasdk-bridge.zip
	cd ./build/lib/${BRIDGE_BACKEND} && unzip ./nexasdk-bridge.zip
	-$(RM) ./build/lib/${BRIDGE_BACKEND}/nexasdk-bridge.zip

clean:
	-$(RM) ./build/nexa$(EXE)
	-$(RM) ./build/nexa-cli$(EXE)
	-$(RM) ./build/lib
	-$(RM) ./runner/nexa-sdk/stub/$(LIB)nexa_bridge.$(EXT)
