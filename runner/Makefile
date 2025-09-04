BRIDGE_VERSION ?= v1.0.4

ifeq ($(OS), Windows_NT)
    OS   := windows
    ARCH ?= x86_64
	EXE  := .exe
else
    UNAME := $(shell uname -s)
    ARCH  := $(shell uname -m)
    ifeq ($(UNAME), Linux)
        OS  := linux
        LIB := lib
        EXT := so
    else ifeq ($(UNAME), Darwin)
        OS  := macos
        LIB := lib
        EXT := dylib
    endif

    EXE :=
endif

.PHONY: build link xcopy download clean

build:
	go build -o ./build/nexa$(EXE) ./cmd/nexa-launcher
	go build \
		-tags="sonic avx" \
		-ldflags "-X 'main.Version=$(VERSION)'" \
		-o ./build/nexa-cli$(EXE) \
		./cmd/nexa-cli

link: clean
	mkdir -p build
	@echo "====> Linking bridge lib to build"
	cd build && ln -s ../../../nexasdk-bridge/build/out/* ./
	cd nexa-sdk && ln -sf  ../../../nexasdk-bridge/modelfiles .

xcopy: clean
	mkdir -p build
	xcopy ..\\..\\nexasdk-bridge\\build\\out\\* build /E /R /Y

download: clean
	mkdir -p build
	@echo "====> Download runtime"
	@echo "OS: $(OS)"
	@echo "ARCH: $(ARCH)"
	@echo "BRIDGE_VERSION: $(BRIDGE_VERSION)"
	curl -L -o build/nexasdk-bridge.zip \
		https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(BRIDGE_VERSION)/$(OS)_$(ARCH)/nexasdk-bridge.zip
	cd build && unzip -q nexasdk-bridge.zip && rm -f nexasdk-bridge.zip

clean:
	rm -rf build
