AGS ?= infer Qwen/Qwen3-0.6B-GGUF
#ARGS ?= serve

# ubuntu 22.04: llama-cpp-cpu llama-cpp-cuda
# macos 13: llama-cpp-metal
# macos 14: llama-cpp-metal mlx
# macos 15: llama-cpp-metal mlx
# windows: llama-cpp-cpu llama-cpp-vulkan llama-cpp-cuda
BRIDGE_VERSION ?= latest

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

.PHONY: run build doc test download clean swagger-ui

run:
	./build/nexa $(ARGS)

build:
	cd ./runner/nexa-sdk/stub && g++ -O3 -fPIC -shared -o $(LIB)nexa_bridge.$(EXT) *.cpp
	cd ./runner && go build -o ../build/nexa$(EXE) ./cmd/nexa-launcher
	cd ./runner && go build -tags="sonic avx" -o ../build/nexa-cli$(EXE) ./cmd/nexa-cli

doc:
	swag init --parseDependency --parseDepth 1 -d ./runner/server -g ./server.go -o ./runner/server/docs

test:
	cd runner && LD_LIBRARY_PATH=$(PWD)/build/lib go test -v ./nexa-sdk --run VLM

download:
	-$(RM) ./build/lib
	$(MKDIR) ./build/lib
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "BRIDGE_BACKEND: $(BRIDGE_BACKEND)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/nexasdk-bridge.zip https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(BRIDGE_BACKEND)/$(OS)/nexasdk-bridge.zip
	cd ./build/lib && unzip ../nexasdk-bridge.zip
	-$(RM) ./build/nexasdk-bridge.zip

swagger-ui:
	-$(RM) ./runner/server/swagger-ui/dist
	$(MKDIR) ./runner/server/swagger-ui/dist
	@echo "====> Download Swagger UI dist"
	$(eval SWAGGER_VERSION := $(shell curl -s https://api.github.com/repos/swagger-api/swagger-ui/releases/latest | grep '"tag_name"' | sed 's/.*"tag_name": "v\([^"]*\)".*/\1/'))
	curl -L https://github.com/swagger-api/swagger-ui/archive/refs/tags/v$(SWAGGER_VERSION).tar.gz | tar -xz --strip-components=2 -C ./runner/server/swagger-ui/dist --wildcards swagger-ui-$(SWAGGER_VERSION)/dist/*
	@echo "====> Configure Swagger UI"
	@sed -i 's|https://petstore.swagger.io/v2/swagger.json|/docs/swagger.json|g' ./runner/server/swagger-ui/dist/swagger-initializer.js
	@echo "====> Swagger UI dist downloaded and configured successfully"

clean:
	-$(RM) ./build/nexa{,.exe}
	-$(RM) ./build/nexa-cli{,.exe}
	-$(RM) ./build/lib
	-$(RM) ./runner/nexa-sdk/stub/$(LIB)nexa_bridge.$(EXT)
