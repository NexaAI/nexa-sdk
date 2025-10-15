BRIDGE_VERSION ?= v1.0.19-rc19

VERSION ?= $(shell git rev-parse --short HEAD)
ifeq ($(OS), Windows_NT)
	OS      := windows
	ARCH    ?= $(shell powershell -NoProfile -NonInteractive "switch ((Get-CimInstance Win32_Processor).Architecture) { 0 {'x86'} 5 {'arm'} 9 {'x86_64'} 12 {'arm64'} Default {'unknown'} }")
	EXE     := .exe
	RM      := powershell -NoProfile -NonInteractive "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -Path"
	MKDIR   := powershell -NoProfile -NonInteractive "New-Item -ItemType Directory -Force -Path"
	MKLINK  := powershell -NoProfile -NonInteractive "New-Item -ItemType Junction -Path 'build' -Target '..\..\nexasdk-bridge\build\out'"
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
endif

.PHONY: build link download clean package

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
	cd build && tar -xf nexasdk-bridge.zip && $(RM_FILE) nexasdk-bridge.zip

package: build
	-$(RM) artifacts
	@echo "====> Creating installer package for $(OS)_$(ARCH)"
	@if [ "$(OS)" = "macos" ]; then \
		release/package_darwin.sh $(VERSION) $(ARCH); \
	else \
		echo "Error: Package support only available for macOS currently"; \
		exit 1; \
	fi

clean:
	-$(RM) build
	-$(RM) artifacts
