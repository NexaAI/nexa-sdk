BRIDGE_VERSION ?= v1.0.17

ifeq ($(OS), Windows_NT)
	OS      := windows
	ARCH    ?= $(shell powershell -NoProfile -NonInteractive "switch ((Get-CimInstance Win32_Processor).Architecture) { 0 {'x86'} 5 {'arm'} 9 {'x86_64'} 12 {'arm64'} Default {'unknown'} }")
	EXE     := .exe

	RM      := powershell -NoProfile -NonInteractive Remove-Item -Recurse -Force -ErrorAction SilentlyContinue -Path
	MKDIR   := powershell -NoProfile -NonInteractive New-Item -ItemType Directory -Force -Path
	UNZIP   := powershell -NoProfile -NonInteractive Expand-Archive -Force -DestinationPath . -Path
else
	UNAME   := $(shell uname -s)
	ARCH    := $(shell uname -m)
	ifeq ($(UNAME), Linux)
		OS  := linux
    else ifeq ($(UNAME), Darwin)
		OS  := macos
	endif
	EXE     :=

	RM      := rm -rf
	MKDIR   := mkdir -p
	UNZIP   := unzip -q
endif

.PHONY: build link xcopy download clean serve

build:
	go build \
		-ldflags "-s -w" \
		-o ./build/nexa$(EXE) \
		./cmd/nexa-launcher
	go build \
		-tags="sonic avx" \
		-ldflags "-X 'main.Version=$(VERSION)' -s -w" \
		-o ./build/nexa-cli$(EXE) \
		./cmd/nexa-cli

link: clean
	$(MKDIR) build
	@echo "====> Linking bridge lib to build"
	cd build && ln -s ../../../nexasdk-bridge/build/out/* ./
	cd nexa-sdk && ln -sf  ../../../nexasdk-bridge/modelfiles .

xcopy: clean
	$(MKDIR) build
	powershell -NoProfile -NonInteractive -Command "Copy-Item -Path '..\..\nexasdk-bridge\build\out\*' -Destination 'build' -Recurse -Force"

download: clean
	$(MKDIR) build
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "ARCH: $(ARCH)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/nexasdk-bridge.zip \
		https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(OS)_$(ARCH)/nexasdk-bridge.zip
	cd build && $(UNZIP) nexasdk-bridge.zip && $(RM) nexasdk-bridge.zip && $(RM) nexa_bridge.lib

clean:
	-$(RM) build

serve:
	./build/nexa serve
