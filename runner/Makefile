BRIDGE_VERSION ?= v1.0.39

ifeq ($(OS), Windows_NT)
	OS      := windows
	ARCH    ?= $(shell powershell -NoProfile -NonInteractive "switch ((Get-CimInstance Win32_Processor).Architecture) { 0 {'x86'} 5 {'arm'} 9 {'x86_64'} 12 {'arm64'} Default {'unknown'} }")
	EXE     := .exe
	RM      := powershell -NoProfile -NonInteractive "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -Path"
	MKDIR   := powershell -NoProfile -NonInteractive "New-Item -ItemType Directory -Force -Path"
	MKLINK  := powershell -NoProfile -NonInteractive "New-Item -ItemType Junction -Path 'build' -Target '..\..\nexasdk-bridge\build\out'"
	UNZIP   := powershell -NoProfile -NonInteractive -Command "Expand-Archive -Path build/nexasdk-bridge.zip -DestinationPath build -Force"
else
	OS      := $(shell echo $(shell uname -s) | tr '[:upper:]' '[:lower:]')
	ifeq ($(OS), darwin)
		OS := macos
	endif
	ARCH    := $(shell uname -m)
	ifeq ($(ARCH), aarch64)
		ARCH := arm64
	endif
	EXE     :=
	RM      := rm -rf
	MKDIR   := mkdir -p
	MKLINK  := ln -s ../../nexasdk-bridge/build/out build
	UNZIP   := unzip -o build/nexasdk-bridge.zip -d build
endif

DOCKER_TAG ?= latest
BASE_IMAGE ?= ubuntu:22.04
DOCKER_ARCH := $(ARCH)
ifeq ($(DOCKER_ARCH), x86_64)
	DOCKER_ARCH := amd64
endif

.PHONY: build link download clean docker docker-cuda

build:
	go build -ldflags "-s -w" -o build/nexa$(EXE) ./cmd/nexa-launcher
	go build -tags="sonic avx" -ldflags "-X 'main.Version=$(VERSION)' -s -w" -o build/nexa-cli$(EXE) ./cmd/nexa-cli

link: clean
	$(MKLINK)

download: clean
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "ARCH: $(ARCH)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	$(MKDIR) build
	curl -L -o build/nexasdk-bridge.zip \
		https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(OS)_$(ARCH)/nexasdk-bridge.zip
	$(UNZIP)
	$(RM) build/nexasdk-bridge.zip
	$(RM) build/nexa_bridge.lib

docker: build
	$(MKDIR) release/linux/nexa_sdk.$(DOCKER_ARCH)
	cp -r build/* release/linux/nexa_sdk.$(DOCKER_ARCH)/
	chmod +x release/linux/nexa_sdk.$(DOCKER_ARCH)/nexa
	chmod +x release/linux/nexa_sdk.$(DOCKER_ARCH)/nexa-cli
	docker build \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg TARGETARCH=$(DOCKER_ARCH) \
		-t nexa4ai/nexasdk:$(DOCKER_TAG) \
		-f release/linux/Dockerfile \
		release/linux

docker-cuda:
	$(MAKE) docker BASE_IMAGE=nvidia/cuda:12.9.1-runtime-ubuntu22.04 DOCKER_TAG=$(DOCKER_TAG)-cuda

clean:
	-$(RM) build
	-$(RM) release/linux/nexa_sdk.*
