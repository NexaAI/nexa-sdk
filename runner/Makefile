AGS ?= infer Qwen/Qwen3-0.6B-GGUF
#ARGS ?= serve

BRIDGE_VERSION ?= v1.0.3-rc2

ifeq ($(OS), Windows_NT)
	OS := windows
	ARCH ?= x86_64
	LIB :=
	EXT := dll
	EXE := .exe
	RM := powershell Remove-Item -Recurse -Force -Path
	MKDIR := powershell New-Item -ItemType Directory -Force -Path
else
	UNAME := $(shell uname -s)
	ARCH := $(shell uname -m)
	ifeq ($(UNAME), Linux)
		OS := linux
		LIB := lib
		EXT := so
	else ifeq ($(UNAME), Darwin)
		OS := macos
		LIB := lib
		EXT := dylib
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
	go build -o ./build/nexa$(EXE) ./cmd/nexa-launcher
	go build \
		-tags="sonic avx" \
		-ldflags "-X 'main.Version=$(VERSION)'" \
		-o ./build/nexa-cli$(EXE) \
		./cmd/nexa-cli

test:
	go test -v ./nexa-sdk --run VLM

link:
	@echo "====> Linking bridge lib to build"
	-$(RM) ./build
	$(MKDIR) ./build
	cd build && ln -s ../../../nexasdk-bridge/build/out/* ./
	cd nexa-sdk && ln -sf  ../../../nexasdk-bridge/modelfiles .

download:
	-$(RM) ./build
	$(MKDIR) ./build
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "ARCH: $(ARCH)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/nexasdk-bridge.zip https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(OS)_$(ARCH)/nexasdk-bridge.zip
	cd ./build && unzip ./nexasdk-bridge.zip
	-$(RM) ./build/nexasdk-bridge.zip

release:
	pass

clean:
	-$(RM) ./build
