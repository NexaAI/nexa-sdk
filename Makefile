AGS ?= infer Qwen/Qwen3-0.6B-GGUF
#ARGS ?= serve

BRIDGE_VERSION ?= latest
BRIDGE_BACKEND ?= llama

UNAME := $(shell uname -s)
ifeq ($(UNAME), Linux)
	OS := linux
	EXT := so
else ifeq ($(UNAME), Darwin)
	MACOS_VERSION := $(shell sw_vers -productVersion))
	MACOS_MAJOR_VERSION := $(firstword $(subst ., ,$(MACOS_VERSION)))
	OS := macos-$(MACOS_MAJOR_VERSION)
	EXT := dylib
endif

.PHONY: run build doc test download clean

run:
	./build/nexa $(ARGS)

build:
	cd runner/nexa-sdk/stub && g++ -O3 -fPIC -shared -o libnexa_bridge.$(EXT) *.cpp
	cd runner && CGO_ENABLED=0 go build -o ../build/nexa ./cmd/nexa-launcher
	cd runner && go build -tags="sonic avx" -o ../build/nexa-cli ./cmd/nexa-cli

doc:
	swag init -d ./runner/server -g ./server.go -o ./runner/server/docs

test:
	cd runner && LD_LIBRARY_PATH=$(PWD)/build/lib go test -v ./nexa-sdk --run VLM

download:
	rm -rf build/lib
	mkdir -p build/lib
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "BACKEND: $(BRIDGE_BACKEND)"
	@echo "VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/nexasdk-bridge.zip https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(BRIDGE_BACKEND)/$(OS)/nexasdk-bridge.zip
	cd build/lib && unzip ../nexasdk-bridge.zip
	rm build/nexasdk-bridge.zip

clean:
	rm -rf build/nexa
	rm -rf build/nexa-cli
	rm -rf build/lib
	rm -rf runner/nexa-sdk/stub/libnexa_bridge.$(EXT)

