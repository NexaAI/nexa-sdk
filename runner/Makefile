BRIDGE_VERSION ?= v1.0.6-rc1

VERSION ?= $(shell git rev-parse --short HEAD)
ifeq ($(OS), Windows_NT)
    OS      := windows
    ARCH    ?= x86_64
    EXE     := .exe

    RM_FILE := cmd /c del /f /q
    RM_DIR  := cmd /c rmdir /s /q
    MKDIR   := cmd /c mkdir
    UNZIP   := powershell -Command "Expand-Archive -Force -DestinationPath . -Path"
else
    UNAME := $(shell uname -s)
    ARCH  := $(shell uname -m)
    ifeq ($(UNAME), Linux)
        OS  := linux
    else ifeq ($(UNAME), Darwin)
        OS  := macos
    endif
    EXE     :=

    RM_FILE := rm -f
    RM_DIR  := rm -rf
    MKDIR   := mkdir -p
    UNZIP   := unzip -q
endif

.PHONY: build link xcopy download clean package

build:
	go build -o ./build/nexa$(EXE) ./cmd/nexa-launcher
	go build \
		-tags="sonic avx" \
		-ldflags "-X 'main.Version=$(VERSION)'" \
		-o ./build/nexa-cli$(EXE) \
		./cmd/nexa-cli

link: clean
	$(MKDIR) build
	@echo "====> Linking bridge lib to build"
	cd build && ln -s ../../../nexasdk-bridge/build/out/* ./
	cd nexa-sdk && ln -sf  ../../../nexasdk-bridge/modelfiles .

xcopy: clean
	$(MKDIR) build
	xcopy ..\\..\\nexasdk-bridge\\build\\out\\* build /E /R /Y

download: clean
	$(MKDIR) build
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "ARCH: $(ARCH)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/nexasdk-bridge.zip \
		https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(OS)_$(ARCH)/nexasdk-bridge.zip
	cd build && $(UNZIP) nexasdk-bridge.zip && $(RM_FILE) nexasdk-bridge.zip

package: build
	-$(RM_DIR) artifacts
	@echo "====> Creating installer package for $(OS)_$(ARCH)"
	@if [ "$(OS)" = "windows" ]; then \
		bash release/package_windows.sh $(VERSION); \
	elif [ "$(OS)" = "macos" ]; then \
		release/package_darwin.sh $(VERSION) $(ARCH); \
	elif [ "$(OS)" = "linux" ]; then \
		bash release/package_linux.sh $(VERSION); \
	fi

clean:
	-$(RM_DIR) build
	-$(RM_DIR) artifacts